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
assert.equal(ios.includes('QLUTCH_E1_API'), true)
assert.equal(ios.includes('INFOPLIST_FILE = Info-E1.plist'), true)
assert.equal(ios.includes('name = Staging'), true)
const gradle = text('apps/android/gradle/libs.versions.toml')
assert.equal(gradle.includes('agp = "9.5.1"'), true)

const androidGradle = text('apps/android/app/build.gradle.kts')
assert.equal(androidGradle.includes('qlutch.e1.api'), true)
assert.equal(androidGradle.includes('http://10.0.2.2:4710'), true)
assert.equal(androidGradle.includes('http://47.93.180.37'), true)
assert.equal(androidGradle.includes('http://47.93.180.37:4710'), false)
const androidNet = text('apps/android/app/src/main/res/xml/network_security_config.xml')
for (const host of ['10.0.2.2', '127.0.0.1', 'localhost', '47.93.180.37']) {
  assert.equal(androidNet.includes(host), true, `android cleartext missing ${host}`)
}
assert.equal(text('apps/android/app/src/main/java/dev/qlutch/app/ApiConfig.kt').includes('BuildConfig.API_BASE_URL'), true)

const iosConfig = text('apps/ios/Qlutch/ApiConfig.swift')
assert.equal(iosConfig.includes('QLUTCH_E1_API'), true)
assert.equal(iosConfig.includes('http://127.0.0.1:4710'), true)
assert.equal(iosConfig.includes('http://47.93.180.37'), true)
assert.equal(iosConfig.includes(':4710")!'), true)
const iosLocal = text('apps/ios/Info.plist')
assert.equal(iosLocal.includes('NSAllowsLocalNetworking'), true)
assert.equal(iosLocal.includes('47.93.180.37'), false)
assert.equal(iosLocal.includes('NSAllowsArbitraryLoads'), false)
const iosE1 = text('apps/ios/Info-E1.plist')
assert.equal(iosE1.includes('47.93.180.37'), true)
assert.equal(iosE1.includes('NSExceptionAllowsInsecureHTTPLoads'), true)
assert.equal(iosE1.includes('NSAllowsArbitraryLoads'), false)
const iosScheme = text('apps/ios/Qlutch.xcodeproj/xcshareddata/xcschemes/Qlutch.xcscheme')
assert.equal(iosScheme.includes('buildConfiguration = "Debug"'), true)
assert.equal(iosScheme.includes('Staging'), false)
const e1Scheme = text('apps/ios/Qlutch.xcodeproj/xcshareddata/xcschemes/Qlutch-E1.xcscheme')
assert.equal(e1Scheme.includes('buildConfiguration = "Staging"'), true)
assert.equal(e1Scheme.includes('buildConfiguration = "Debug"'), false)

const harmony = text('apps/harmony/build-profile.json5')
const harmonyProfile = JSON.parse(harmony)
const products = harmonyProfile.app.products
assert.equal(products.find((item) => item.name === 'default').buildOption.arkOptions.buildProfileFields.API_BASE_URL, 'http://127.0.0.1:4710')
assert.equal(products.find((item) => item.name === 'e1').buildOption.arkOptions.buildProfileFields.API_BASE_URL, 'http://47.93.180.37')
assert.equal(harmony.includes('http://47.93.180.37:4710'), false)
assert.equal(text('apps/harmony/entry/src/main/ets/config/ApiConfig.ets').includes('BuildProfile.API_BASE_URL'), true)
const harmonyNet = JSON.parse(text('apps/harmony/entry/src/main/resources/base/profile/network_config.json'))
assert.equal(harmonyNet['network-security-config']['base-config'].cleartextTrafficPermitted, true)

for (const rel of [
  'apps/android/app/src/main/java/dev/qlutch/app/flow/IdentityApi.kt',
  'apps/ios/Qlutch/Flow/IdentityClient.swift',
  'apps/harmony/entry/src/main/ets/flow/IdentityApi.ets',
]) {
  const source = text(rel)
  assert.equal(source.includes('/internal/staging/otp/reveal'), false, `${rel} calls the staging reveal route`)
  const lines = source.split('\n')
  const onboarding = lines.filter((line) => line.includes('/v1/identity/onboarding'))
  const methods = onboarding
    .map((line) => line.match(/["'](GET|PUT|POST|PATCH|DELETE)["']/)?.[1])
    .filter(Boolean)
  assert.deepEqual(methods.slice().sort(), ['GET', 'PUT'], `${rel} onboarding verbs`)
  assert.equal(onboarding.some((line) => /\bpost\s*\(/i.test(line)), false, `${rel} posts onboarding`)
  const writeLine = lines.find((line, index) => {
    if (!line.includes('/v1/identity/onboarding')) return false
    return lines.slice(Math.max(0, index - 15), index).some((prev) => prev.includes('writeOnboarding'))
  })
  assert.equal(typeof writeLine, 'string', `${rel} writeOnboarding does not call /v1/identity/onboarding`)
  assert.match(writeLine, /["']PUT["']/, `${rel} writeOnboarding must PUT`)
  assert.doesNotMatch(writeLine, /["']POST["']|\bpost\s*\(/i, `${rel} writeOnboarding must not POST`)
}

console.log('client shell checks passed')
