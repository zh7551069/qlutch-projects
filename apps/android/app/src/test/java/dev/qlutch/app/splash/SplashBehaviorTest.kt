package dev.qlutch.app.splash

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import java.io.File
import java.security.MessageDigest

class SplashBehaviorTest {
    @Test
    fun animatorOrTransitionScaleOffSkipsMotion() {
        assertTrue(isReduceMotion(animatorScaleOff = true, transitionScaleOff = false, platformReduceMotion = false))
        assertTrue(isReduceMotion(animatorScaleOff = false, transitionScaleOff = true, platformReduceMotion = false))
        assertTrue(isReduceMotion(animatorScaleOff = false, transitionScaleOff = false, platformReduceMotion = true))
        assertFalse(isReduceMotion(animatorScaleOff = false, transitionScaleOff = false, platformReduceMotion = false))
    }

    @Test
    fun scaleOffOnlyWhenTheStoredValueIsZero() {
        assertFalse(scaleIsOff(null))
        assertFalse(scaleIsOff(""))
        assertFalse(scaleIsOff("1"))
        assertFalse(scaleIsOff("1.0"))
        assertFalse(scaleIsOff("0.5"))
        assertTrue(scaleIsOff("0"))
        assertTrue(scaleIsOff("0.0"))
        assertTrue(scaleIsOff("0.0f"))
        assertTrue(scaleIsOff("0,0"))
    }

    @Test
    fun coverScalePreservesAspectAndFills() {
        val (phoneX, phoneY) = coverScale(1080, 2400, 1080, 1920)
        assertEquals(1.25f, phoneX, 0.001f)
        assertEquals(1f, phoneY, 0.001f)

        val (sameX, sameY) = coverScale(1080, 1920, 1080, 1920)
        assertEquals(1f, sameX, 0f)
        assertEquals(1f, sameY, 0f)

        val (wideX, wideY) = coverScale(1920, 1080, 1080, 1920)
        assertEquals(1f, wideX, 0.001f)
        assertEquals((1920f / 1080f) / (1080f / 1920f), wideY, 0.001f)

        assertEquals(1f to 1f, coverScale(0, 100, 1080, 1920))
    }

    @Test
    fun staticHoldStaysInsideTheEntranceBudget() {
        assertTrue(SplashTiming.STATIC_HOLD_MILLIS in 1_200L..2_000L)
    }

    @Test
    fun brandCopyIsWordmarkPeriodAndSloganOnly() {
        val strings = moduleFile("src/main/res/values/strings.xml").readText()
        assertTrue(strings.contains("<string name=\"splash_wordmark\">Qlutch</string>"))
        assertTrue(strings.contains("<string name=\"splash_wordmark_period\">\u3002</string>"))
        assertTrue(strings.contains("<string name=\"splash_slogan\">让记知识，变得轻松有趣</string>"))
        assertFalse(strings.contains("把课件或教材"))
        assertFalse(strings.contains("好玩"))
        assertFalse(strings.contains("Gizmo"))

        val colors = moduleFile("src/main/res/values/colors.xml").readText()
        assertTrue(colors.contains("<color name=\"canvas\">#F6F8FF</color>"))
        assertTrue(colors.contains("<color name=\"ink\">#20244A</color>"))
        assertTrue(colors.contains("<color name=\"spark\">#FF6B4A</color>"))
    }

    @Test
    fun splashSourcesDoNotInventPositioningOrIp() {
        val source = moduleFile("src/main/java/dev/qlutch/app/splash")
            .walk()
            .filter { it.extension == "kt" }
            .joinToString("\n") { it.readText() }
        assertFalse(source.contains("把课件或教材上传给AI"))
        assertFalse(source.contains("Gizmo"))
        assertFalse(source.contains("美西螈"))
        assertFalse(source.contains("好玩"))
    }

    @Test
    fun officialVideoBytesAreUnchanged() {
        val file = moduleFile("src/main/res/raw/splash.mp4")
        val bytes = file.readBytes()
        assertEquals(266684, bytes.size)
        val digest = MessageDigest.getInstance("SHA-256").digest(bytes)
        val hex = digest.joinToString("") { "%02x".format(it) }
        assertEquals("0307ab2061c71133b336761f4c45de1ecb27376ce140237d5f96e36df0347922", hex)
        assertTrue(bytes.containsSequence("vide".encodeToByteArray()))
        assertFalse(bytes.containsSequence("soun".encodeToByteArray()))
    }

    private fun ByteArray.containsSequence(needle: ByteArray): Boolean {
        if (needle.isEmpty() || needle.size > size) return false
        for (start in 0..size - needle.size) {
            var matched = true
            for (offset in needle.indices) {
                if (this[start + offset] != needle[offset]) {
                    matched = false
                    break
                }
            }
            if (matched) return true
        }
        return false
    }

    @Test
    fun launcherIsThePortraitSplashAndTheShellStaysNext() {
        val manifest = moduleFile("src/main/AndroidManifest.xml").readText()
        assertTrue(manifest.contains("android:name=\".splash.SplashActivity\""))
        assertTrue(manifest.contains("android:screenOrientation=\"portrait\""))
        assertTrue(manifest.contains("android.intent.action.MAIN"))
        assertTrue(manifest.contains("android:name=\".MainActivity\""))
        val shell = moduleFile("src/main/java/dev/qlutch/app/MainActivity.kt").readText()
        assertTrue(shell.contains("首页"))
        assertTrue(shell.contains("做题"))
        assertTrue(shell.contains("PK"))
        assertTrue(shell.contains("好友"))
        assertTrue(shell.contains("我的"))
    }

    private fun moduleFile(relative: String): File {
        val cwd = File(System.getProperty("user.dir") ?: ".")
        val candidates = listOf(
            File(cwd, relative),
            File(cwd, "app/$relative"),
            File(cwd, "apps/android/app/$relative"),
        )
        return candidates.firstOrNull { it.exists() }
            ?: error("Missing $relative from ${cwd.absolutePath}")
    }
}
