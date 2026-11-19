import { spawnSync } from 'node:child_process'
import { mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { parse } from 'yaml'
import { emitArkTS, emitGo, emitKotlin, emitSwift, emitTS } from './emit.mjs'
import { buildModel } from './model.mjs'

const root = resolve(fileURLToPath(new URL('.', import.meta.url)), '../..')
const spec = parse(readFileSync(resolve(root, 'packages/contracts/openapi.yaml'), 'utf8'))
const model = buildModel(spec)

const files = new Map([
  ['packages/contracts/gen/go/client.go', emitGo(model)],
  ['packages/contracts/gen/typescript/client.ts', emitTS(model)],
  ['packages/contracts/gen/swift/QlutchClient.swift', emitSwift(model)],
  ['packages/contracts/gen/kotlin/dev/qlutch/contract/QlutchClient.kt', emitKotlin(model)],
  ['packages/contracts/gen/arkts/QlutchClient.ets', emitArkTS(model)],
])

const copies = [
  ['packages/contracts/gen/swift/QlutchClient.swift', 'apps/ios/Qlutch/Generated/QlutchClient.swift'],
  [
    'packages/contracts/gen/kotlin/dev/qlutch/contract/QlutchClient.kt',
    'apps/android/app/src/main/java/dev/qlutch/contract/QlutchClient.kt',
  ],
  ['packages/contracts/gen/arkts/QlutchClient.ets', 'apps/harmony/entry/src/main/ets/generated/QlutchClient.ets'],
]

for (const [rel, contents] of files) {
  writeText(resolve(root, rel), contents)
}
for (const [from, to] of copies) {
  writeText(resolve(root, to), files.get(from))
}

const formatted = spawnSync('gofmt', ['-w', resolve(root, 'packages/contracts/gen/go/client.go')], {
  stdio: 'inherit',
})
if (formatted.status !== 0) {
  process.exit(formatted.status ?? 1)
}

function writeText(path, contents) {
  mkdirSync(dirname(path), { recursive: true })
  const text = contents.endsWith('\n') ? contents : `${contents}\n`
  writeFileSync(path, text)
}
