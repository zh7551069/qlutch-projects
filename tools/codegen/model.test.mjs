import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'
import { parse } from 'yaml'
import { buildModel } from './model.mjs'

const spec = {
  openapi: '3.0.3',
  paths: {
    '/healthz': {
      get: {
        operationId: 'getHealth',
        responses: {
          200: {
            content: {
              'application/json': { schema: { $ref: '#/components/schemas/HealthStatus' } },
            },
          },
        },
      },
    },
  },
  components: {
    schemas: {
      HealthStatus: {
        type: 'object',
        required: ['status'],
        properties: { status: { type: 'string' } },
      },
    },
  },
}

test('builds a probe operation from component schemas', () => {
  const model = buildModel(spec)
  assert.equal(model.operations[0].goName, 'GetHealth')
  assert.equal(model.operations[0].path, '/healthz')
  assert.equal(model.schemas[0].fields[0].goName, 'Status')
  assert.equal(model.schemas[0].fields[0].required, true)
})

test('keeps a component request body and required header', () => {
  const withBody = structuredClone(spec)
  withBody.components.schemas.Echo = {
    type: 'object',
    required: ['status'],
    properties: { status: { type: 'string' } },
  }
  withBody.paths['/healthz'].post = {
    operationId: 'postHealth',
    parameters: [{ name: 'Idempotency-Key', in: 'header', required: true, schema: { type: 'string' } }],
    requestBody: {
      required: true,
      content: { 'application/json': { schema: { $ref: '#/components/schemas/Echo' } } },
    },
    responses: {
      200: {
        content: { 'application/json': { schema: { $ref: '#/components/schemas/HealthStatus' } } },
      },
    },
  }
  const model = buildModel(withBody)
  const post = model.operations.find((operation) => operation.operationId === 'postHealth')
  assert.equal(post.bodyName, 'Echo')
  assert.equal(post.bodyRequired, true)
  assert.equal(post.headers[0].name, 'Idempotency-Key')
  assert.equal(post.headers[0].required, true)
})

test('builds the merged identity contract without dropping operations', () => {
  const file = new URL('../../packages/contracts/openapi.yaml', import.meta.url)
  const model = buildModel(parse(readFileSync(file, 'utf8')))
  const ids = new Set(model.operations.map((operation) => operation.operationId))
  for (const id of [
    'getHealth',
    'getReady',
    'sendIdentityOtp',
    'verifyIdentityOtp',
    'refreshIdentitySession',
    'revokeIdentitySession',
    'getIdentityOnboarding',
    'putIdentityOnboarding',
    'getIdentityProfile',
    'putIdentityProfile',
  ]) {
    assert.equal(ids.has(id), true, id)
  }
  const send = model.operations.find((operation) => operation.operationId === 'sendIdentityOtp')
  assert.equal(send.bodyName, 'OtpSendRequest')
  assert.equal(send.headers.some((item) => item.name === 'Idempotency-Key' && item.required), true)
  const revoke = model.operations.find((operation) => operation.operationId === 'revokeIdentitySession')
  assert.equal(revoke.responseName, null)
  assert.equal(revoke.auth, 'optional')
})

test('rejects a request body without a component schema', () => {
  const withBody = structuredClone(spec)
  withBody.paths['/healthz'].get.requestBody = { required: true, content: { 'application/json': {} } }
  assert.throws(() => buildModel(withBody), /request body/)
})
