import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = resolve(fileURLToPath(new URL('.', import.meta.url)), '../..')
const tabs = ['首页', '做题', 'PK', '好友', '我的']

function text(rel) {
  return readFileSync(resolve(root, rel), 'utf8')
}

function same(left, right) {
  assert.equal(text(left), text(right), `${left} drifted from ${right}`)
}

const shells = [
  ['apps/ios/Qlutch/ContentView.swift', '.tabItem'],
  ['apps/android/app/src/main/java/dev/qlutch/app/MainActivity.kt', 'NavigationBarItem('],
  ['apps/harmony/entry/src/main/ets/pages/Index.ets', 'TabContent()'],
]

for (const [rel, marker] of shells) {
  const source = text(rel)
  const count = source.split(marker).length - 1
  assert.equal(count, 5, `${rel} should reserve exactly five root tabs`)
  for (const label of tabs) {
    assert.equal(source.includes(label), true, `${rel} missing ${label}`)
  }
  assert.equal(source.includes('知识地图'), false, `${rel} must not add a knowledge-map root tab`)
}

for (const rel of [
  'packages/contracts/gen/go/client.go',
  'packages/contracts/gen/typescript/client.ts',
  'packages/contracts/gen/swift/QlutchClient.swift',
  'packages/contracts/gen/kotlin/dev/qlutch/contract/QlutchClient.kt',
  'packages/contracts/gen/arkts/QlutchClient.ets',
]) {
  const source = text(rel)
  assert.equal(source.includes('DO NOT EDIT'), true, `${rel} is not marked generated`)
  assert.equal(/getHealth|GetHealth/.test(source), true, `${rel} missing health operation`)
}

same(
  'packages/contracts/gen/swift/QlutchClient.swift',
  'apps/ios/Qlutch/Generated/QlutchClient.swift',
)
same(
  'packages/contracts/gen/kotlin/dev/qlutch/contract/QlutchClient.kt',
  'apps/android/app/src/main/java/dev/qlutch/contract/QlutchClient.kt',
)
same(
  'packages/contracts/gen/arkts/QlutchClient.ets',
  'apps/harmony/entry/src/main/ets/generated/QlutchClient.ets',
)

for (const rel of [
  'apps/harmony/build-profile.json5',
  'apps/harmony/oh-package.json5',
  'apps/harmony/entry/src/main/module.json5',
  'apps/harmony/AppScope/app.json5',
]) {
  JSON.parse(text(rel))
}

const ios = text('apps/ios/Qlutch.xcodeproj/project.pbxproj')
assert.equal(ios.includes('PBXFileSystemSynchronizedRootGroup'), true)
assert.equal(ios.includes('PRODUCT_BUNDLE_IDENTIFIER = dev.qlutch.app'), true)
const gradle = text('apps/android/gradle/libs.versions.toml')
assert.equal(gradle.includes('agp = "9.5.1"'), true)

console.log('client shell checks passed')
