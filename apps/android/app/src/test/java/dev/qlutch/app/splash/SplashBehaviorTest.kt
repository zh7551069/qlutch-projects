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
    fun welcomeFrameIsTwoThirdsWideAndNineBySixteen() {
        val (width, height) = SplashLayout.frameSize(1080f)
        assertEquals(720f, width, 0.001f)
        assertEquals(1280f, height, 0.001f)
        assertEquals(width * SplashLayout.HEIGHT_OVER_WIDTH, height, 0f)
        assertTrue(width < 1080f)
        assertTrue(height < 1920f)

        val (narrow, narrowHeight) = SplashLayout.frameSize(393f)
        assertEquals(262f, narrow, 0.001f)
        assertEquals(262f * 16f / 9f, narrowHeight, 0.001f)
        assertEquals(SplashLayout.WIDTH_FRACTION, 2f / 3f, 0f)
    }

    @Test
    fun containScaleFitsTheClipWithoutCroppingOrStretching() {
        val (matchedX, matchedY) = containScale(720, 1280, 1080, 1920)
        assertEquals(1f, matchedX, 0.001f)
        assertEquals(1f, matchedY, 0.001f)

        val (wideX, wideY) = containScale(640, 640, 1080, 1920)
        assertEquals((1080f / 1920f) / (640f / 640f), wideX, 0.001f)
        assertEquals(1f, wideY, 0.001f)
        assertTrue(wideX < 1f)

        val (tallX, tallY) = containScale(360, 800, 1080, 1920)
        assertEquals(1f, tallX, 0.001f)
        assertTrue(tallY < 1f)

        assertEquals(1f to 1f, containScale(0, 100, 1080, 1920))
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
    fun packedWelcomeClipBytesAreTheV4File() {
        val file = moduleFile("src/main/res/raw/splash.mp4")
        val bytes = file.readBytes()
        assertEquals(849742, bytes.size)
        assertEquals("41f48c6521c33b0d04e64f3925751225796597fff713e2208ba39d59136c468a", sha256(bytes))
        assertTrue(bytes.containsSequence("vide".encodeToByteArray()))
        assertFalse(bytes.containsSequence("soun".encodeToByteArray()))
    }

    @Test
    fun wordmarkStaysBelowTheLandingShadowAndDoesNotFollowTheJump() {
        assertEquals(1462f, SplashLayout.LANDING_SHADOW_BOTTOM_Y, 0f)
        assertTrue(SplashLayout.LANDING_SHADOW_BOTTOM_Y > 1454f)
        assertTrue(SplashLayout.WORDMARK_GAP_AT_375 in 20f..32f)

        val place = SplashLayout.placement(1080f, 2340f)
        assertEquals(720f, place.frameWidth, 0.001f)
        assertEquals(1280f, place.frameHeight, 0.001f)
        assertEquals(180f, place.frameLeft, 0.001f)
        assertEquals(530f, place.frameTop, 0.001f)
        assertEquals(396f, place.wordmarkWidth, 0.001f)
        assertEquals(396f * 238f / 941f, place.wordmarkHeight, 0.001f)
        assertEquals(342f, place.wordmarkLeft, 0.001f)
        assertEquals(1580f, place.wordmarkTop, 0.5f)
        assertEquals(26f * 1080f / 375f, place.wordmarkGap, 0.01f)
        assertEquals(place.shadowBottom + place.wordmarkGap, place.wordmarkTop, 0.001f)
        assertTrue(place.wordmarkTop > place.shadowBottom)
        assertTrue(place.wordmarkWidth < place.frameWidth)
        assertTrue(place.wordmarkWidth > place.frameWidth * 0.45f)
        assertTrue(place.frameWidth < 1080f)

        val local = SplashLayout.wordmarkInFrame(place.frameWidth, place.frameHeight)
        assertEquals(place.wordmarkTop, place.frameTop + local.top, 0.001f)
        assertEquals(place.wordmarkLeft, place.frameLeft + local.left, 0.001f)
        assertEquals(place, SplashLayout.placement(1080f, 2340f))

        val design = SplashLayout.placement(375f, 812f)
        assertEquals(26f, design.wordmarkGap, 0.001f)
        assertTrue(design.wordmarkTop > design.shadowBottom)

        listOf(360f to 780f, 393f to 852f, 411f to 891f).forEach { (width, height) ->
            val sized = SplashLayout.placement(width, height)
            assertEquals(width * 2f / 3f, sized.frameWidth, 0.05f)
            assertEquals(sized.frameWidth * 16f / 9f, sized.frameHeight, 0.05f)
            assertEquals(
                SplashLayout.WORDMARK_GAP_AT_375 * (width / SplashLayout.DESIGN_WIDTH),
                sized.wordmarkGap,
                0.01f,
            )
            assertTrue(sized.wordmarkTop > sized.shadowBottom)
            assertTrue(sized.wordmarkWidth < width)
        }
    }

    @Test
    fun packedWordmarkIsPurpleG4WithoutAWhitePlate() {
        val file = moduleFile("src/main/res/drawable-nodpi/splash_g4_wordmark.png")
        val bytes = file.readBytes()
        assertEquals(
            "ac47a0ff712f9af63632ac44c7a2014ab7fe71ca18afd48587450b2dd88cc92b",
            sha256(bytes),
        )
        val image = decodePng(bytes)
        assertEquals(941, image.width)
        assertEquals(238, image.height)
        listOf(0 to 0, image.width - 1 to 0, 0 to image.height - 1, image.width - 1 to image.height - 1)
            .forEach { (x, y) ->
                assertEquals(0, image.alpha(x, y))
            }
        var colored = 0
        for (y in 0 until image.height) {
            for (x in 0 until image.width) {
                val alpha = image.alpha(x, y)
                if (alpha == 0) continue
                colored += 1
                assertEquals(0x6C, image.red(x, y))
                assertEquals(0x5C, image.green(x, y))
                assertEquals(0xE7, image.blue(x, y))
            }
        }
        assertTrue(colored > 0)
    }

    @Test
    fun previewSplashKeepsTheV4BackgroundAndOverlaysTheFixedWordmark() {
        val colors = moduleFile("src/main/res/values/colors.xml").readText()
        assertTrue(colors.contains("<color name=\"splash_background\">#F2F5F7</color>"))
        val themes = moduleFile("src/main/res/values/themes.xml").readText()
        assertTrue(themes.contains("android:windowBackground\">@color/splash_background"))
        assertTrue(themes.contains("android:windowSplashScreenBackground\" tools:targetApi=\"31\">@color/splash_background"))
        assertFalse(themes.contains("@color/canvas"))

        val splashDir = moduleFile("src/main/java/dev/qlutch/app/splash")
        assertFalse(splashDir.list().orEmpty().any { it.contains("StaticBrand") })
        val shown = listOf(
            "ColdStartVideo.kt",
            "StaticWelcomeStill.kt",
            "SplashActivity.kt",
            "SplashFrame.kt",
            "SplashVideoView.kt",
        ).joinToString("\n") { name ->
            File(splashDir, name).readText()
        }
        assertFalse(shown.contains("splash_slogan"))
        assertFalse(shown.contains("让记知识"))
        assertFalse(shown.contains("coverScale"))
        assertFalse(shown.contains("setColorFilter"))
        assertFalse(shown.contains("colorFilter"))
        assertFalse(shown.contains("R.color.ink"))
        assertFalse(shown.contains("R.color.spark"))
        assertFalse(shown.contains("#20244A"))
        assertFalse(shown.contains("#FF6B4A"))
        assertFalse(shown.contains("currentPosition"))
        assertTrue(shown.contains("splash_background"))
        assertTrue(shown.contains("StaticWelcomeStill"))
        assertTrue(shown.contains("containScale"))
        assertTrue(shown.contains("setVolume(0f, 0f)"))
        assertTrue(shown.contains("OnBackPressedCallback"))
        assertTrue(shown.contains("ContentScale.Fit"))
        assertTrue(shown.contains("ScaleType.FIT_CENTER"))
        assertTrue(shown.contains("getFrameAtIndex(0)"))
        assertTrue(shown.contains("splash_g4_wordmark"))
        assertTrue(shown.contains("LANDING_SHADOW_BOTTOM_Y"))
        assertTrue(shown.contains("wordmarkInFrame"))
        assertTrue(shown.contains("WIDTH_FRACTION = 2f / 3f"))
        assertTrue(shown.contains("Color.TRANSPARENT"))
        assertTrue(shown.contains("imageTintList = null"))
        assertTrue(moduleFile("build.gradle.kts").readText().contains("noCompress += \"mp4\""))

        val still = File(splashDir, "StaticWelcomeStill.kt").readText()
        val playback = File(splashDir, "ColdStartVideo.kt").readText()
        val video = File(splashDir, "SplashVideoView.kt").readText()
        val frame = File(splashDir, "SplashFrame.kt").readText()
        assertTrue(still.contains("G4Wordmark()"))
        assertFalse(playback.contains("G4Wordmark"))
        assertTrue(playback.contains("SplashPlaybackView"))
        assertTrue(frame.contains("fun G4Wordmark"))
        assertTrue(frame.contains("R.drawable.splash_g4_wordmark"))
        assertTrue(video.contains("R.drawable.splash_g4_wordmark"))
        assertTrue(video.contains("wordmarkInFrame"))
        assertTrue(video.indexOf("addView(\n            video") < video.indexOf("addView(\n            wordmark"))
        assertFalse(video.contains("splash_slogan"))
        assertFalse(frame.contains("splash_slogan"))
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

    private fun sha256(bytes: ByteArray): String {
        val digest = MessageDigest.getInstance("SHA-256").digest(bytes)
        return digest.joinToString("") { "%02x".format(it) }
    }

    private class PngImage(val width: Int, val height: Int, private val pixels: IntArray) {
        fun alpha(x: Int, y: Int) = (pixels[y * width + x] ushr 24) and 0xff
        fun red(x: Int, y: Int) = (pixels[y * width + x] ushr 16) and 0xff
        fun green(x: Int, y: Int) = (pixels[y * width + x] ushr 8) and 0xff
        fun blue(x: Int, y: Int) = pixels[y * width + x] and 0xff
    }

    private fun decodePng(bytes: ByteArray): PngImage {
        val signature = byteArrayOf(
            0x89.toByte(), 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
        )
        require(bytes.copyOfRange(0, 8).contentEquals(signature))
        var pos = 8
        var width = 0
        var height = 0
        var colorType = -1
        val idat = java.io.ByteArrayOutputStream()
        while (pos + 12 <= bytes.size) {
            val length = readInt(bytes, pos)
            val type = bytes.copyOfRange(pos + 4, pos + 8).toString(Charsets.US_ASCII)
            val start = pos + 8
            when (type) {
                "IHDR" -> {
                    width = readInt(bytes, start)
                    height = readInt(bytes, start + 4)
                    require(bytes[start + 8].toInt() and 0xff == 8)
                    colorType = bytes[start + 9].toInt() and 0xff
                    require(bytes[start + 12].toInt() and 0xff == 0)
                }
                "IDAT" -> idat.write(bytes, start, length)
                "IEND" -> break
            }
            pos = start + length + 4
        }
        val channels = when (colorType) {
            6 -> 4
            2 -> 3
            else -> error("unsupported color type $colorType")
        }
        val stride = width * channels
        val expected = height * (1 + stride)
        val raw = ByteArray(expected)
        val inflater = java.util.zip.Inflater()
        inflater.setInput(idat.toByteArray())
        var produced = 0
        while (produced < expected) {
            val count = inflater.inflate(raw, produced, expected - produced)
            check(count > 0)
            produced += count
        }
        inflater.end()
        val pixels = IntArray(width * height)
        val prev = IntArray(stride)
        val row = IntArray(stride)
        var cursor = 0
        for (y in 0 until height) {
            val filter = raw[cursor].toInt() and 0xff
            cursor += 1
            for (x in 0 until stride) {
                row[x] = raw[cursor].toInt() and 0xff
                cursor += 1
            }
            when (filter) {
                0 -> Unit
                1 -> for (x in 0 until stride) {
                    val left = if (x >= channels) row[x - channels] else 0
                    row[x] = (row[x] + left) and 0xff
                }
                2 -> for (x in 0 until stride) {
                    row[x] = (row[x] + prev[x]) and 0xff
                }
                3 -> for (x in 0 until stride) {
                    val left = if (x >= channels) row[x - channels] else 0
                    row[x] = (row[x] + ((left + prev[x]) / 2)) and 0xff
                }
                4 -> for (x in 0 until stride) {
                    val left = if (x >= channels) row[x - channels] else 0
                    val up = prev[x]
                    val upLeft = if (x >= channels) prev[x - channels] else 0
                    val estimate = left + up - upLeft
                    val leftDelta = kotlin.math.abs(estimate - left)
                    val upDelta = kotlin.math.abs(estimate - up)
                    val upLeftDelta = kotlin.math.abs(estimate - upLeft)
                    val predictor = when {
                        leftDelta <= upDelta && leftDelta <= upLeftDelta -> left
                        upDelta <= upLeftDelta -> up
                        else -> upLeft
                    }
                    row[x] = (row[x] + predictor) and 0xff
                }
                else -> error("png filter $filter")
            }
            for (x in 0 until width) {
                val offset = x * channels
                val red = row[offset]
                val green = row[offset + 1]
                val blue = row[offset + 2]
                val alpha = if (channels == 4) row[offset + 3] else 255
                pixels[y * width + x] = (alpha shl 24) or (red shl 16) or (green shl 8) or blue
            }
            row.copyInto(prev)
        }
        return PngImage(width, height, pixels)
    }

    private fun readInt(bytes: ByteArray, offset: Int): Int {
        return ((bytes[offset].toInt() and 0xff) shl 24) or
            ((bytes[offset + 1].toInt() and 0xff) shl 16) or
            ((bytes[offset + 2].toInt() and 0xff) shl 8) or
            (bytes[offset + 3].toInt() and 0xff)
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
