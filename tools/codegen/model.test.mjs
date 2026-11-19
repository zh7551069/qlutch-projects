import assert from 'node:assert/strict'
import test from 'node:test'
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

test('rejects a request body instead of dropping it', () => {
  const withBody = structuredClone(spec)
  withBody.paths['/healthz'].get.requestBody = { required: true }
  assert.throws(() => buildModel(withBody), /request body/)
})
