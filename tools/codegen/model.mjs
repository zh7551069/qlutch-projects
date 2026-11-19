// Turns a small OpenAPI 3 document into a model the emitters can render.
// Unsupported constructs fail the process so a contract change cannot be dropped silently.

export function buildModel(spec) {
  if (!String(spec?.openapi ?? '').startsWith('3.')) {
    throw new Error('OpenAPI 3 document required')
  }
  const schemas = Object.entries(spec.components?.schemas ?? {}).map(([name, schema]) =>
    compileSchema(name, schema),
  )
  const operations = []
  for (const [path, item] of Object.entries(spec.paths ?? {})) {
    for (const method of ['get', 'post', 'put', 'patch', 'delete']) {
      const operation = item?.[method]
      if (!operation) continue
      operations.push(compileOperation(method, path, operation))
    }
  }
  if (operations.length === 0) {
    throw new Error('OpenAPI document has no operations')
  }
  return { schemas, operations }
}

function compileOperation(method, path, operation) {
  if (!operation.operationId) {
    throw new Error(`${method.toUpperCase()} ${path} is missing operationId`)
  }
  if (operation.parameters?.length) {
    throw new Error(`${operation.operationId} has parameters, which this generator does not emit yet`)
  }
  if (operation.requestBody) {
    throw new Error(`${operation.operationId} has a request body, which this generator does not emit yet`)
  }
  if (path.includes('{')) {
    throw new Error(`${operation.operationId} has a path parameter, which this generator does not emit yet`)
  }
  const responseName = responseSchemaName(operation)
  return {
    operationId: operation.operationId,
    goName: exportName(operation.operationId),
    method: method.toUpperCase(),
    path,
    responseName,
  }
}

function responseSchemaName(operation) {
  const responses = operation.responses ?? {}
  const names = new Set()
  for (const [code, response] of Object.entries(responses)) {
    if (!code.startsWith('2') && code !== '503') continue
    const schema = response?.content?.['application/json']?.schema
    if (!schema?.$ref) {
      throw new Error(`${operation.operationId} response ${code} must use a component schema`)
    }
    names.add(schema.$ref.split('/').pop())
  }
  if (names.size !== 1) {
    throw new Error(`${operation.operationId} must use one JSON schema for success and 503`)
  }
  return [...names][0]
}

function compileSchema(name, schema) {
  if (schema?.type !== 'object') {
    throw new Error(`${name} must be an object schema`)
  }
  if (schema.allOf || schema.oneOf || schema.anyOf) {
    throw new Error(`${name} uses a combinator, which this generator does not emit yet`)
  }
  const required = new Set(schema.required ?? [])
  const fields = Object.entries(schema.properties ?? {}).map(([jsonName, property]) => ({
    jsonName,
    goName: exportName(jsonName),
    type: compileType(name, jsonName, property),
    required: required.has(jsonName),
  }))
  return { name, fields }
}

function compileType(parent, field, property) {
  if (property?.$ref) {
    return { kind: 'ref', name: property.$ref.split('/').pop() }
  }
  if (property?.type === 'array') {
    return { kind: 'array', item: compileType(parent, field, property.items) }
  }
  if (property?.type === 'object' && property.additionalProperties && property.additionalProperties !== true) {
    return { kind: 'map', value: compileType(parent, field, property.additionalProperties) }
  }
  if (['string', 'integer', 'number', 'boolean'].includes(property?.type)) {
    return { kind: property.type }
  }
  throw new Error(`${parent}.${field} uses an unsupported schema`)
}

function exportName(name) {
  return name.slice(0, 1).toUpperCase() + name.slice(1)
}
