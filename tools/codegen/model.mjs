// Turns an OpenAPI 3 document into a model the emitters can render.
// Unsupported constructs fail the process so a contract change cannot be dropped silently.

export function buildModel(spec) {
  if (!String(spec?.openapi ?? '').startsWith('3.')) {
    throw new Error('OpenAPI 3 document required')
  }
  const lifted = liftSchemas(spec)
  const schemas = Object.entries(lifted.schemas).map(([name, schema]) => compileSchema(name, schema))
  const operations = []
  for (const [path, item] of Object.entries(spec.paths ?? {})) {
    for (const method of ['get', 'post', 'put', 'patch', 'delete']) {
      const operation = item?.[method]
      if (!operation) continue
      operations.push(compileOperation(spec, method, path, operation))
    }
  }
  if (operations.length === 0) {
    throw new Error('OpenAPI document has no operations')
  }
  return { schemas, operations }
}

function liftSchemas(spec) {
  const schemas = structuredClone(spec.components?.schemas ?? {})
  let added = true
  while (added) {
    added = false
    for (const [name, schema] of Object.entries(schemas)) {
      if (schema?.type !== 'object' || !schema.properties) continue
      for (const [field, property] of Object.entries(schema.properties)) {
        if (property?.$ref) continue
        if (property?.type !== 'object' || !property.properties) continue
        if (property.additionalProperties === true) {
          throw new Error(`${name}.${field} mixes properties with additionalProperties`)
        }
        const child = `${name}${exportName(field)}`
        if (schemas[child] && JSON.stringify(schemas[child]) !== JSON.stringify(property)) {
          throw new Error(`${child} collides while lifting ${name}.${field}`)
        }
        if (!schemas[child]) {
          schemas[child] = structuredClone(property)
          added = true
        }
        schema.properties[field] = { $ref: `#/components/schemas/${child}` }
      }
    }
  }
  return { schemas }
}

function compileOperation(spec, method, path, operation) {
  if (!operation.operationId) {
    throw new Error(`${method.toUpperCase()} ${path} is missing operationId`)
  }
  const headers = []
  const declaredPath = new Map()
  for (const parameter of operation.parameters ?? []) {
    const resolved = resolveRef(spec, parameter)
    if (!resolved?.name || !resolved.in) {
      throw new Error(`${operation.operationId} has an unreadable parameter`)
    }
    if (resolved.in === 'header') {
      headers.push({
        name: resolved.name,
        required: resolved.required === true,
      })
      continue
    }
    if (resolved.in === 'path') {
      if (resolved.schema?.type && resolved.schema.type !== 'string') {
        throw new Error(`${operation.operationId} path parameter ${resolved.name} must be a string`)
      }
      if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(resolved.name)) {
        throw new Error(`${operation.operationId} path parameter ${resolved.name} is not an identifier`)
      }
      declaredPath.set(resolved.name, { name: resolved.name, required: resolved.required !== false })
      continue
    }
    throw new Error(`${operation.operationId} has a ${resolved.in} parameter, which this generator does not emit yet`)
  }
  const namesInPath = [...path.matchAll(/\{([^}]+)\}/g)].map((match) => match[1])
  if (namesInPath.length > 1) {
    throw new Error(`${operation.operationId} has more than one path parameter`)
  }
  const pathParams = namesInPath.map((name) => {
    const declared = declaredPath.get(name)
    if (!declared) {
      throw new Error(`${operation.operationId} path {${name}} is not declared`)
    }
    return declared
  })
  for (const name of declaredPath.keys()) {
    if (!namesInPath.includes(name)) {
      throw new Error(`${operation.operationId} declares path parameter ${name} that is not in the path`)
    }
  }
  let bodyName = null
  let bodyRequired = false
  if (operation.requestBody) {
    const body = resolveRef(spec, operation.requestBody)
    const content = body?.content ?? {}
    const types = Object.keys(content)
    if (types.length !== 1 || types[0] !== 'application/json') {
      throw new Error(`${operation.operationId} request body must be application/json only`)
    }
    const schema = content['application/json']?.schema
    if (!schema?.$ref) {
      throw new Error(`${operation.operationId} request body must use a component schema`)
    }
    bodyName = schema.$ref.split('/').pop()
    bodyRequired = body.required !== false
  }
  const responseName = responseSchemaName(spec, operation)
  return {
    operationId: operation.operationId,
    goName: exportName(operation.operationId),
    method: method.toUpperCase(),
    path,
    pathParams,
    responseName,
    bodyName,
    bodyRequired,
    headers,
    auth: operationAuth(spec, operation),
  }
}

function operationAuth(spec, operation) {
  const security = operation.security ?? spec.security ?? []
  if (security.length === 0) return 'none'
  let bearer = false
  let optional = false
  for (const entry of security) {
    const names = Object.keys(entry ?? {})
    if (names.length === 0) {
      optional = true
      continue
    }
    if (names.length === 1 && names[0] === 'BearerAuth') {
      bearer = true
      continue
    }
    throw new Error(`${operation.operationId} uses an unsupported security requirement`)
  }
  if (!bearer) {
    throw new Error(`${operation.operationId} does not declare BearerAuth`)
  }
  return optional ? 'optional' : 'required'
}

function responseSchemaName(spec, operation) {
  const responses = operation.responses ?? {}
  const names = new Set()
  let sawEmpty = false
  for (const [code, raw] of Object.entries(responses)) {
    if (!String(code).startsWith('2') && String(code) !== '503') continue
    const response = resolveRef(spec, raw)
    const schema = response?.content?.['application/json']?.schema
    if (!schema) {
      if (String(code) === '204') {
        sawEmpty = true
        continue
      }
      throw new Error(`${operation.operationId} response ${code} must use a component schema`)
    }
    if (!schema.$ref) {
      throw new Error(`${operation.operationId} response ${code} must use a component schema`)
    }
    names.add(schema.$ref.split('/').pop())
  }
  if (sawEmpty && names.size === 0) return null
  if (names.size !== 1 || sawEmpty) {
    throw new Error(`${operation.operationId} must use one JSON schema for success and 503`)
  }
  return [...names][0]
}

function compileSchema(name, schema) {
  if (schema?.type === 'string' && !schema.properties) {
    return {
      name,
      kind: 'alias',
      enumValues: (schema.enum ?? []).filter((value) => value !== null && value !== undefined).map(String),
    }
  }
  if (schema?.type !== 'object') {
    throw new Error(`${name} must be an object schema or a string alias`)
  }
  if (schema.allOf || schema.oneOf || schema.anyOf) {
    throw new Error(`${name} uses a combinator, which this generator does not emit yet`)
  }
  const required = new Set(schema.required ?? [])
  const fields = Object.entries(schema.properties ?? {}).map(([jsonName, property]) => {
    const type = compileType(name, jsonName, property)
    const nullable = property?.nullable === true || (Array.isArray(property?.enum) && property.enum.includes(null))
    return {
      jsonName,
      goName: exportName(jsonName),
      type,
      required: required.has(jsonName),
      nullable,
      enumValues: (property?.enum ?? []).filter((value) => value !== null && value !== undefined).map(String),
    }
  })
  return { name, kind: 'object', fields }
}

function compileType(parent, field, property) {
  if (property?.$ref) {
    return { kind: 'ref', name: property.$ref.split('/').pop() }
  }
  if (property?.type === 'array') {
    if (!property.items) {
      throw new Error(`${parent}.${field} array is missing items`)
    }
    return { kind: 'array', item: compileType(parent, field, property.items) }
  }
  if (property?.type === 'object' && property.additionalProperties === true) {
    return { kind: 'json' }
  }
  if (property?.type === 'object' && property.additionalProperties && property.additionalProperties !== true) {
    return { kind: 'map', value: compileType(parent, field, property.additionalProperties) }
  }
  if (['string', 'integer', 'number', 'boolean'].includes(property?.type)) {
    return { kind: property.type }
  }
  throw new Error(`${parent}.${field} uses an unsupported schema`)
}

function resolveRef(spec, node) {
  let current = node
  for (let hop = 0; hop < 8; hop += 1) {
    if (!current?.$ref) return current
    const parts = String(current.$ref).replace(/^#\//, '').split('/')
    let cursor = spec
    for (const part of parts) {
      cursor = cursor?.[part]
    }
    if (!cursor) {
      throw new Error(`unresolved $ref ${current.$ref}`)
    }
    current = cursor
  }
  throw new Error('ref cycle')
}

export function exportName(name) {
  return name.slice(0, 1).toUpperCase() + name.slice(1)
}

export function enumConstName(typeName, value) {
  const parts = String(value)
    .split(/[^A-Za-z0-9]+/)
    .filter(Boolean)
    .map((part) => part.slice(0, 1).toUpperCase() + part.slice(1))
  const body = parts.join('')
  if (!body || /^[0-9]/.test(body)) {
    return `${typeName}Value${body}`
  }
  return `${typeName}${body}`
}
