package dev.qlutch.app.welcome

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import java.awt.Font
import java.io.File
import java.security.MessageDigest

class WelcomeBehaviorTest {
    @Test
    fun boardMatchesTheBridgeArtboard() {
        assertEquals(390, WelcomeLayout.ARTBOARD_WIDTH_DP)
        assertEquals(844, WelcomeLayout.ARTBOARD_HEIGHT_DP)
        assertEquals(0xFFF2F5F7, WelcomeLayout.CANVAS)
        assertEquals(20, WelcomeLayout.CONTENT_GAP_DP)
        assertEquals(16, WelcomeLayout.CONTENT_PADDING_HORIZONTAL_DP)
        assertEquals(24, WelcomeLayout.CONTENT_PADDING_TOP_DP)
        assertEquals(
            WelcomeLayout.SLOT_WIDTH_DP,
            WelcomeLayout.ARTBOARD_WIDTH_DP - 2 * WelcomeLayout.CONTENT_PADDING_HORIZONTAL_DP,
        )
        assertEquals(358, WelcomeLayout.SLOT_WIDTH_DP)
        assertEquals(240, WelcomeLayout.SLOT_HEIGHT_DP)
        assertEquals(22, WelcomeLayout.SLOT_RADIUS_DP)
        assertEquals(0xFFF1F4F9, WelcomeLayout.SLOT_COLOR)
        assertEquals(50, WelcomeLayout.WORDMARK_SIZE_SP)
        assertEquals(800, WelcomeLayout.WORDMARK_WEIGHT)
        assertEquals(0xFF6C5CE7, WelcomeLayout.WORDMARK_COLOR)
        assertEquals(16, WelcomeLayout.TAGLINE_SIZE_SP)
        assertEquals(600, WelcomeLayout.TAGLINE_WEIGHT)
        assertEquals(0xFF0F172A, WelcomeLayout.TAGLINE_COLOR)
        assertEquals(358, WelcomeLayout.TAGLINE_WIDTH_DP)
        assertEquals(16, WelcomeLayout.FOOTER_GAP_DP)
        assertEquals(16, WelcomeLayout.FOOTER_PADDING_HORIZONTAL_DP)
        assertEquals(28, WelcomeLayout.FOOTER_PADDING_BOTTOM_DP)
        assertEquals(358, WelcomeLayout.BUTTON_WIDTH_DP)
        assertEquals(50, WelcomeLayout.BUTTON_HEIGHT_DP)
        assertEquals(25, WelcomeLayout.BUTTON_RADIUS_DP)
        assertEquals(16, WelcomeLayout.BUTTON_TEXT_SIZE_SP)
        assertEquals(600, WelcomeLayout.BUTTON_TEXT_WEIGHT)
        assertEquals(0xFF0F172A, WelcomeLayout.PRIMARY_COLOR)
        assertEquals(0xFFFFFFFF, WelcomeLayout.PRIMARY_TEXT_COLOR)
        assertEquals(0xFFFFFFFF, WelcomeLayout.SECONDARY_COLOR)
        assertEquals(0xFF0F172A, WelcomeLayout.SECONDARY_TEXT_COLOR)
        assertEquals(0xFFE1E8F0, WelcomeLayout.SECONDARY_BORDER_COLOR)
        assertEquals(1, WelcomeLayout.SECONDARY_BORDER_DP)
        assertEquals(13, WelcomeLayout.FOOTER_TEXT_SIZE_SP)
        assertEquals(400, WelcomeLayout.FOOTER_TEXT_WEIGHT)
        assertEquals(0xFF7C848F, WelcomeLayout.FOOTER_TEXT_COLOR)
        assertEquals(358, WelcomeLayout.LEGAL_WIDTH_DP)
        assertEquals(WelcomeLayout.SLOT_WIDTH_DP, WelcomeLayout.BUTTON_WIDTH_DP)
        assertEquals(WelcomeLayout.SLOT_WIDTH_DP, WelcomeLayout.TAGLINE_WIDTH_DP)
        assertEquals(WelcomeLayout.SLOT_WIDTH_DP, WelcomeLayout.LEGAL_WIDTH_DP)

        val html = repoFile("design/bridge-export/p02-welcome/index.html").readText()
        val css = repoFile("design/bridge-export/p02-welcome/styles.css").readText()
        assertEquals(5671, repoFile("design/bridge-export/p02-welcome/index.html").length())
        assertEquals(
            "d4bb4aa63265df3b4ce6da168069e6941ca87694339e3bdb214b473ffe56d06d",
            sha256(repoFile("design/bridge-export/p02-welcome/index.html").readBytes()),
        )
        assertEquals(
            "4c0441bf5b3e1a67f74d85a8935a422e11a095790ce092ed7821c4eafa8d5223",
            sha256(repoFile("design/bridge-export/p02-welcome/styles.css").readBytes()),
        )
        assertTrue(html.contains("w-[390px]"))
        assertTrue(html.contains("h-[844px]"))
        val artboard = html.substringAfter("class=\"frame p02").substringBefore(">")
        assertTrue(artboard.contains("style=\"background:rgb(242,245,247);\""))
        assertFalse(artboard.contains("255,255,255"))
        assertEquals(1, html.split("background:rgb(242,245,247)").size - 1)
        assertEquals(2, html.split("background:rgb(255,255,255)").size - 1)
        assertTrue(html.contains("gap-[20px]"))
        assertTrue(html.contains("pt-[24px]"))
        assertTrue(html.contains("pl-[16px] pr-[16px]"))
        assertTrue(html.contains("h-[240px]"))
        assertTrue(html.contains("w-[358px]"))
        assertTrue(html.contains("rounded-[22px]"))
        assertTrue(html.contains("background:rgb(241,244,249);"))
        assertTrue(html.contains("text-[50px]"))
        assertTrue(html.contains("text-[rgb(108,92,231)]"))
        assertTrue(html.contains("font-extrabold"))
        assertTrue(html.contains("text-[16px]"))
        assertTrue(html.contains("font-semibold"))
        assertTrue(html.contains("text-[rgb(15,23,42)]"))
        assertTrue(html.contains("pb-[28px]"))
        assertTrue(html.contains("gap-[16px]"))
        assertTrue(html.contains("h-[50px]"))
        assertTrue(html.contains("rounded-[25px]"))
        assertTrue(html.contains("background:rgb(15,23,42);"))
        assertTrue(html.contains("box-shadow:inset 0 0 0 1px rgb(225,232,240);"))
        assertTrue(html.contains("text-[13px]"))
        assertTrue(html.contains("font-normal"))
        assertTrue(html.contains("text-[rgb(124,132,143)]"))
        assertTrue(css.contains(".gap-\\[20px\\] {\n  gap: 20px;\n}"))
        assertTrue(css.contains(".text-\\[50px\\] {\n  font-size: 50px;\n}"))
        assertTrue(css.contains(".font-extrabold {\n  font-weight: 800;\n}"))
        assertTrue(css.contains(".font-semibold {\n  font-weight: 600;\n}"))
        assertTrue(css.contains(".font-normal {\n  font-weight: 400;\n}"))

        val screen = welcomeSource("WelcomeScreen.kt") + "\n" + welcomeSource("WelcomeCheer.kt")
        assertTrue(screen.contains("Color(WelcomeLayout.CANVAS)"))
        assertTrue(screen.contains("WelcomeLayout.ARTBOARD_WIDTH_DP"))
        assertTrue(screen.contains("WelcomeLayout.CONTENT_GAP_DP"))
        assertTrue(screen.contains("WelcomeLayout.CONTENT_PADDING_HORIZONTAL_DP"))
        assertTrue(screen.contains("WelcomeLayout.CONTENT_PADDING_TOP_DP"))
        assertTrue(screen.contains("WelcomeLayout.SLOT_WIDTH_DP"))
        assertTrue(screen.contains("WelcomeLayout.SLOT_HEIGHT_DP"))
        assertTrue(screen.contains("WelcomeLayout.SLOT_RADIUS_DP"))
        assertTrue(screen.contains("WelcomeLayout.SLOT_COLOR"))
        assertTrue(screen.contains("WelcomeLayout.BUTTON_WIDTH_DP"))
        assertTrue(screen.contains("WelcomeLayout.BUTTON_HEIGHT_DP"))
        assertTrue(screen.contains("WelcomeLayout.BUTTON_RADIUS_DP"))
        assertTrue(screen.contains("WelcomeLayout.FOOTER_GAP_DP"))
        assertTrue(screen.contains("WelcomeLayout.FOOTER_PADDING_BOTTOM_DP"))
        assertTrue(screen.contains("Arrangement.spacedBy"))
        assertTrue(screen.contains("Alignment.CenterVertically"))
        assertTrue(screen.contains("weight(1f)"))
        assertFalse(screen.contains("引导插画占位"))
        assertFalse(screen.contains("101.dp"))
        assertFalse(screen.contains("splash_g4_wordmark"))
        assertFalse(screen.contains("0xFFFFFFFF"))
        assertFalse(screen.contains("#FFFFFF"))
        assertFalse(screen.contains("#ffffff"))

        val colors = moduleFile("src/main/res/values/colors.xml").readText()
        assertTrue(colors.contains("<color name=\"page_background\">#F2F5F7</color>"))
        assertTrue(colors.contains("<color name=\"splash_background\">#F2F5F7</color>"))
        val themes = moduleFile("src/main/res/values/themes.xml").readText()
        assertTrue(themes.contains("android:windowBackground\">@color/page_background"))
        assertTrue(themes.contains("android:statusBarColor\">@color/page_background"))
        assertTrue(themes.contains("android:navigationBarColor\">@color/page_background"))
        assertFalse(themes.contains("@android:color/white"))
        assertFalse(themes.contains("#FFFFFF"))
        assertFalse(themes.contains("#ffffff"))
        val shell = moduleFile("src/main/java/dev/qlutch/app/MainActivity.kt").readText()
        assertTrue(shell.contains("colorResource(R.color.page_background)"))
    }

    @Test
    fun copyIsTheCurrentWelcomeLine() {
        assertEquals("Qlutch。", WelcomeLayout.WORDMARK)
        assertEquals("Qlutch", WelcomeLayout.WORDMARK_LATIN)
        assertEquals("。", WelcomeLayout.WORDMARK_PERIOD)
        assertEquals(WelcomeLayout.WORDMARK, WelcomeLayout.WORDMARK_LATIN + WelcomeLayout.WORDMARK_PERIOD)
        assertEquals("沉迷于学习，对学习上瘾", WelcomeLayout.TAGLINE)
        assertEquals("开始", WelcomeLayout.PRIMARY_LABEL)
        assertEquals("已有账号", WelcomeLayout.SECONDARY_LABEL)
        assertEquals("教材不用于训练模型", WelcomeLayout.TRAINING_NOTE)
        assertEquals("继续即表示同意《用户协议》和《隐私政策》", WelcomeLayout.LEGAL)
        assertFalse(WelcomeLayout.TAGLINE.contains("让记知识"))

        val strings = moduleFile("src/main/res/values/strings.xml").readText()
        assertTrue(strings.contains("<string name=\"welcome_wordmark_latin\">Qlutch</string>"))
        assertTrue(strings.contains("<string name=\"welcome_wordmark_period\">。</string>"))
        assertTrue(strings.contains("<string name=\"welcome_tagline\">沉迷于学习，对学习上瘾</string>"))
        assertTrue(strings.contains("<string name=\"welcome_start\">开始</string>"))
        assertTrue(strings.contains("<string name=\"welcome_existing_account\">已有账号</string>"))
        assertTrue(strings.contains("<string name=\"welcome_training_note\">教材不用于训练模型</string>"))
        assertTrue(strings.contains("<string name=\"welcome_legal\">继续即表示同意《用户协议》和《隐私政策》</string>"))
        assertTrue(strings.contains("<string name=\"splash_slogan\">让记知识，更轻松有趣</string>"))

        val screen = welcomeSource("WelcomeScreen.kt")
        assertTrue(screen.contains("R.string.welcome_wordmark_latin"))
        assertTrue(screen.contains("R.string.welcome_wordmark_period"))
        assertTrue(screen.contains("WelcomeType.interExtraBold"))
        assertTrue(screen.contains("WelcomeType.notoExtraBold"))
        assertTrue(screen.contains("FontWeight.ExtraBold"))
        assertTrue(screen.contains("WelcomeType.notoSemiBold"))
        assertTrue(screen.contains("WelcomeType.notoRegular"))
        assertTrue(screen.contains("R.string.welcome_tagline"))
        assertTrue(screen.contains("R.string.welcome_start"))
        assertTrue(screen.contains("R.string.welcome_existing_account"))
        assertTrue(screen.contains("R.string.welcome_training_note"))
        assertTrue(screen.contains("R.string.welcome_legal"))
        assertFalse(screen.contains("splash_g4_wordmark"))
        assertFalse(screen.contains("R.drawable.splash_"))
    }

    @Test
    fun cheerPlaysTwentyFiveFramesThenHoldsTheLastForOneSecond() {
        assertEquals(25, WelcomeLayout.CHEER_FRAME_COUNT)
        assertEquals(40L, WelcomeLayout.CHEER_FRAME_MILLIS)
        assertEquals(1_000L, WelcomeLayout.CHEER_HOLD_MILLIS)
        assertEquals(25 * 40L, WelcomeLayout.CHEER_PLAY_MILLIS)
        assertEquals(25 * 40L + 1_000L, WelcomeLayout.CHEER_CYCLE_MILLIS)
        assertEquals(716, WelcomeLayout.CHEER_FRAME_WIDTH_PX)
        assertEquals(480, WelcomeLayout.CHEER_FRAME_HEIGHT_PX)

        assertEquals(0, welcomeCheerFrameIndex(0))
        assertEquals(0, welcomeCheerFrameIndex(39))
        assertEquals(1, welcomeCheerFrameIndex(40))
        assertEquals(24, welcomeCheerFrameIndex(24 * 40L))
        assertEquals(24, welcomeCheerFrameIndex(25 * 40L - 1))
        assertEquals(24, welcomeCheerFrameIndex(25 * 40L))
        assertEquals(24, welcomeCheerFrameIndex(25 * 40L + 999))
        assertEquals(0, welcomeCheerFrameIndex(25 * 40L + 1_000L))
        assertEquals(1, welcomeCheerFrameIndex(25 * 40L + 1_000L + 40L))
        assertEquals(24, welcomeCheerFrameIndex(2 * (25 * 40L + 1_000L) - 1))
        assertEquals(0, welcomeCheerFrameIndex(2 * (25 * 40L + 1_000L)))

        val delays = mutableListOf<Long>()
        var cursor = 0L
        while (cursor < WelcomeLayout.CHEER_CYCLE_MILLIS) {
            val delay = welcomeCheerDelayUntilNextBoundary(cursor)
            val index = welcomeCheerFrameIndex(cursor)
            assertEquals(index, welcomeCheerFrameIndex(cursor + delay - 1))
            delays += delay
            cursor += delay
        }
        assertEquals(List(25) { 40L } + listOf(1_000L), delays)
        assertEquals(WelcomeLayout.CHEER_CYCLE_MILLIS, cursor)
        assertEquals(0, welcomeCheerFrameIndex(cursor))

        val cheer = welcomeSource("WelcomeCheer.kt")
        assertTrue(cheer.contains("welcomeCheerFrameIndex"))
        assertTrue(cheer.contains("welcomeCheerDelayUntilNextBoundary"))
        assertTrue(cheer.contains("SystemClock.elapsedRealtime()"))
        assertTrue(cheer.contains("ContentScale.Fit"))
        assertTrue(cheer.contains("WelcomeLayout.SLOT_COLOR"))
        assertTrue(cheer.contains("WelcomeLayout.SLOT_RADIUS_DP"))
        assertFalse(cheer.contains("ContentScale.Crop"))
        assertFalse(cheer.contains("ContentScale.FillBounds"))
        assertFalse(cheer.contains(".gif"))
        assertFalse(cheer.contains(".webp"))
        assertFalse(cheer.contains("1080"))
        assertFalse(cheer.contains("1920"))
        assertEquals(25, WelcomeCheerFrames.ids.size)
    }

    @Test
    fun slotFramesAreTheUnmodified716By480Sequence() {
        val expected = listOf(
            "welcome_cheer_001.png" to "d40111611a80575d5374b486a3a05de28d7177e32d5a5ef9f551c4168dada27c",
            "welcome_cheer_002.png" to "bc4fd2c4e328b0a69b182e7f2d289acedf25dba128593ecec311e010023c7170",
            "welcome_cheer_003.png" to "86cac22f37501268a0b1db5b2331c1e368d093516c05559efa5428a0346f5621",
            "welcome_cheer_004.png" to "a9b2a4ac4350a8d0b13bac4f69b3fb1c686c044716a0b723e580939a1676f6b1",
            "welcome_cheer_005.png" to "78630bb7b004f29237076bf08b6f8844d9c715bd5708d7188de17455c6643524",
            "welcome_cheer_006.png" to "ae3c06672519538194a83674fe4bf6a31b28c59f12f0d642ea525146a752ee3b",
            "welcome_cheer_007.png" to "62a948927ec4f86fd6755d13076602e54a0772940f17bd380033221678f31625",
            "welcome_cheer_008.png" to "51f5c8b0410b4a68bf172bf5b15a5712e4d500bc4df12ba9ffb740c1cc6bf102",
            "welcome_cheer_009.png" to "f6a010e8b02dd5a608275cf8f1c3c6409d725ee2e7a2291895ab2a8fbd6138c0",
            "welcome_cheer_010.png" to "c58322a60e880ea5f30315ab200857b09b55ec2bc195023817e978d2fa5ed8cc",
            "welcome_cheer_011.png" to "1ca2e8724e5161d660c4da657e2ee3bcbfc3ec1385e04347be9f07e4a05c2d0a",
            "welcome_cheer_012.png" to "10da78a20412fea261826b1be95565f8452f3b3387a523093fd3b830318338bc",
            "welcome_cheer_013.png" to "b54ba5811aa553058fea5897b6bbf545bbd562fe0205bcffbac4bd96a8b229c6",
            "welcome_cheer_014.png" to "1ea6bcabd5680450a1d8334ac5a31731572b92d709afe2a797450b3693758265",
            "welcome_cheer_015.png" to "409ee2a595ecab5cc27f6e0219e443ef84bae7ba55127957f6204cc6d5f2fff9",
            "welcome_cheer_016.png" to "dd953af143588585ff4375ffe631daed2f5a30782450a67845d7aff09850fe09",
            "welcome_cheer_017.png" to "68a7cd14f5e1a16d3a0de4306cd844bf9610677fb02f4b9d69e20fc5fa7b3d57",
            "welcome_cheer_018.png" to "524479d0c7071ff288f36a23cff1688c6c7007dc65d4a775716e337b01635493",
            "welcome_cheer_019.png" to "adf607c1076edd70f01b5f8a30cf30a6512bf8d41bd29259a5c784a7806e241b",
            "welcome_cheer_020.png" to "d4b9edfeb927a2901c60a11976d23874ac9f9a7cc23f7d88e49f66e9633a0b5a",
            "welcome_cheer_021.png" to "b80c8f65f8f4c0422a75eacaac937b56404bc1ca656ed46dfade22ce600c2981",
            "welcome_cheer_022.png" to "a9d59eb0341accb753c46a9d0fee76604c673fba2b45951b1ae6dfea3b2b2634",
            "welcome_cheer_023.png" to "55d467ff31fd9d1a17404194ff45ef049c62eb4540de47aef119eeac990a6ef4",
            "welcome_cheer_024.png" to "01dcd8970aecb78649565ad79008e825bd42597ea5ee0fc85f4c3d3c00023666",
            "welcome_cheer_025.png" to "646cb2f5c9815d50485cd74c3a0170ab17fc9140b9d1da069f2e6be7083e3b31",
        )
        val dir = moduleFile("src/main/res/drawable-nodpi")
        val frames = dir.listFiles { file -> file.name.startsWith("welcome_cheer_") }.orEmpty().sortedBy { it.name }
        assertEquals(expected.map { it.first }, frames.map { it.name })
        expected.forEach { (name, hash) ->
            val bytes = File(dir, name).readBytes()
            assertEquals(hash, sha256(bytes))
            assertSlotPng(bytes)
        }
        val listed = welcomeSource("WelcomeCheerFrames.kt")
        expected.forEach { (name, _) ->
            val id = name.removePrefix("welcome_cheer_").removeSuffix(".png")
            assertTrue(listed.contains("R.drawable.welcome_cheer_$id"))
        }
        assertTrue(moduleFile("src/main/res").walk().none { it.extension.equals("gif", ignoreCase = true) })
        assertTrue(moduleFile("src/main/res").walk().none { it.extension.equals("webp", ignoreCase = true) })
    }

    @Test
    fun chineseCopyUsesNotoAtTheMatchingWeight() {
        val fontDir = moduleFile("src/main/res/font")
        assertDisplays(File(fontDir, "inter_extrabold.ttf"), "Qlutch")
        assertDisplays(File(fontDir, "noto_sans_sc_extrabold.ttf"), "。")
        assertDisplays(File(fontDir, "noto_sans_sc_semibold.ttf"), "沉迷于学习，对学习上瘾开始已有账号")
        assertDisplays(File(fontDir, "noto_sans_sc_regular.ttf"), WelcomeLayout.TRAINING_NOTE + WelcomeLayout.LEGAL)
        val inter = Font.createFont(Font.TRUETYPE_FONT, File(fontDir, "inter_extrabold.ttf"))
        assertFalse(inter.canDisplay('沉'))
        assertFalse(inter.canDisplay('。'))
        assertEquals(
            "d63fb4d20d431d9904ba0b589be9436c372db16eed1f9eb5d66439e23ccef906",
            sha256(File(fontDir, "inter_extrabold.ttf").readBytes()),
        )
        assertEquals(
            "92a21de8c58402ec6b7134b7f630d5cdc8219784e6223571c7b652bbe926e076",
            sha256(File(fontDir, "noto_sans_sc_extrabold.ttf").readBytes()),
        )
        assertEquals(
            "df388bfa1b089668daf62d74cdbdd9713059c7cc48b5a4ad2ededc281a3bb035",
            sha256(File(fontDir, "noto_sans_sc_semibold.ttf").readBytes()),
        )
        assertEquals(
            "a0c71ea499d1cd03b5bd0df2c07602c2f50fbc893cd7070e571f4214f6478db0",
            sha256(File(fontDir, "noto_sans_sc_regular.ttf").readBytes()),
        )
        val type = welcomeSource("WelcomeType.kt")
        assertTrue(type.contains("R.font.inter_extrabold"))
        assertTrue(type.contains("FontWeight.ExtraBold"))
        assertTrue(type.contains("R.font.noto_sans_sc_extrabold"))
        assertTrue(type.contains("R.font.noto_sans_sc_semibold"))
        assertTrue(type.contains("FontWeight.SemiBold"))
        assertTrue(type.contains("R.font.noto_sans_sc_regular"))
        assertTrue(type.contains("FontWeight.Normal"))
    }

    @Test
    fun splashLandsOnWelcomeWithoutASessionCheck() {
        val activity = moduleFile("src/main/java/dev/qlutch/app/splash/SplashActivity.kt").readText()
        assertTrue(activity.contains("WelcomeActivity::class.java"))
        assertFalse(activity.contains("MainActivity::class.java"))
        assertTrue(activity.contains("ColdStartVideo(onFinished = ::openNext)"))
        assertTrue(activity.contains("StaticWelcomeStill(onContinue = ::openNext)"))
        assertTrue(
            activity.contains(
                "override fun handleOnBackPressed() {\n                    openNext()\n                }",
            ),
        )
        val timing = moduleFile("src/main/java/dev/qlutch/app/splash/SplashTiming.kt").readText()
        assertTrue(timing.contains("const val POST_COMPLETION_HOLD_MILLIS = 1_000L"))
        assertTrue(timing.contains("const val STATIC_HOLD_MILLIS = 1_600L"))

        val welcome = moduleFile("src/main/java/dev/qlutch/app/welcome")
            .walk()
            .filter { it.extension == "kt" }
            .joinToString("\n") { it.readText() }
        assertFalse(welcome.contains("startActivity"))
        assertFalse(welcome.contains("MainActivity"))
        assertFalse(welcome.contains("登录"))
        assertFalse(welcome.contains("SharedPreferences"))
        assertFalse(welcome.contains("splash_g4_wordmark"))
        assertTrue(welcome.contains("onClick = {}"))
        val manifest = moduleFile("src/main/AndroidManifest.xml").readText()
        assertTrue(manifest.contains("android:name=\".welcome.WelcomeActivity\""))
        assertTrue(manifest.contains("android:screenOrientation=\"portrait\""))
        assertTrue(manifest.contains("android:name=\".splash.SplashActivity\""))
        assertTrue(manifest.contains("android:name=\".MainActivity\""))
        val shell = moduleFile("src/main/java/dev/qlutch/app/MainActivity.kt").readText()
        assertTrue(shell.contains("首页"))
        assertFalse(shell.contains("WelcomeActivity"))

        val gradle = moduleFile("build.gradle.kts").readText()
        assertTrue(gradle.contains("versionName = \"0.0.0\""))
        assertTrue(gradle.contains("versionCode = 3"))
    }

    private fun assertDisplays(file: File, text: String) {
        val font = Font.createFont(Font.TRUETYPE_FONT, file)
        text.forEach { ch ->
            assertTrue("${file.name} missing U+${ch.code.toString(16)}", font.canDisplay(ch))
        }
    }

    private fun assertSlotPng(bytes: ByteArray) {
        val signature = byteArrayOf(
            0x89.toByte(), 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
        )
        assertTrue(bytes.copyOfRange(0, 8).contentEquals(signature))
        assertEquals("IHDR", bytes.copyOfRange(12, 16).toString(Charsets.US_ASCII))
        assertEquals(WelcomeLayout.CHEER_FRAME_WIDTH_PX, readInt(bytes, 16))
        assertEquals(WelcomeLayout.CHEER_FRAME_HEIGHT_PX, readInt(bytes, 20))
        assertEquals(8, bytes[24].toInt() and 0xff)
        assertEquals(6, bytes[25].toInt() and 0xff)
        assertTrue(bytes.size < 200_000)
    }

    private fun readInt(bytes: ByteArray, offset: Int): Int {
        return ((bytes[offset].toInt() and 0xff) shl 24) or
            ((bytes[offset + 1].toInt() and 0xff) shl 16) or
            ((bytes[offset + 2].toInt() and 0xff) shl 8) or
            (bytes[offset + 3].toInt() and 0xff)
    }

    private fun sha256(bytes: ByteArray): String {
        val digest = MessageDigest.getInstance("SHA-256").digest(bytes)
        return digest.joinToString("") { "%02x".format(it) }
    }

    private fun welcomeSource(name: String): String {
        return moduleFile("src/main/java/dev/qlutch/app/welcome/$name").readText()
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

    private fun repoFile(relative: String): File {
        var dir = File(System.getProperty("user.dir") ?: ".").absoluteFile
        repeat(6) {
            val candidate = File(dir, relative)
            if (candidate.exists()) return candidate
            dir = dir.parentFile ?: return candidate
        }
        error("Missing $relative")
    }
}
