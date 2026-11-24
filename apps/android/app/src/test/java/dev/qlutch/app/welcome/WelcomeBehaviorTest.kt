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
        // Bridge export stays at gap 20. Slot-to-wordmark is 0, wordmark-to-tagline stays 12, group still shifts up 24.
        assertEquals(0, WelcomeLayout.SLOT_TO_WORDMARK_GAP_DP)
        assertEquals(12, WelcomeLayout.WORDMARK_TO_TAGLINE_GAP_DP)
        assertEquals(24, WelcomeLayout.CONTENT_GROUP_SHIFT_UP_DP)
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
        // The rest pose sits 14 artboard px left of centre; 716px of artwork maps
        // to 358dp, so 7dp puts the pose the eye rests on back on the midline.
        assertEquals(-7, WelcomeLayout.SLOT_CONTENT_SHIFT_RIGHT_DP)
        // Frames reserve jump headroom; nudging down brings the squirrel closer
        // to the wordmark without resizing it. Value set by 辉哥 on 2026-11-24.
        assertEquals(20, WelcomeLayout.SLOT_CONTENT_SHIFT_DOWN_DP)
        assertEquals(941, WelcomeLayout.WORDMARK_ART_WIDTH_PX)
        assertEquals(238, WelcomeLayout.WORDMARK_ART_HEIGHT_PX)
        assertEquals(174, WelcomeLayout.WORDMARK_IMAGE_WIDTH_DP)
        assertEquals(72, WelcomeLayout.WORDMARK_BOX_HEIGHT_DP)
        assertEquals(5, WelcomeLayout.WORDMARK_INK_SHIFT_DOWN_DP)
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
        assertTrue(screen.contains("WelcomeLayout.SLOT_TO_WORDMARK_GAP_DP"))
        assertTrue(screen.contains("WelcomeLayout.WORDMARK_TO_TAGLINE_GAP_DP"))
        assertFalse(screen.contains("CONTENT_GAP_DP"))
        val welcomeScreen = welcomeSource("WelcomeScreen.kt")
        val slotGap = welcomeScreen.indexOf("WelcomeLayout.SLOT_TO_WORDMARK_GAP_DP")
        val centered = welcomeScreen.indexOf("Alignment.CenterVertically")
        val taglineGap = welcomeScreen.indexOf("WelcomeLayout.WORDMARK_TO_TAGLINE_GAP_DP")
        assertTrue(slotGap >= 0)
        assertTrue(centered > slotGap)
        assertTrue(taglineGap > centered)
        val slotArrangement = welcomeScreen.substring(slotGap, taglineGap)
        assertTrue(slotArrangement.contains("Alignment.CenterVertically"))
        assertFalse(slotArrangement.contains("WORDMARK_TO_TAGLINE_GAP_DP"))
        assertTrue(screen.contains("WelcomeLayout.CONTENT_GROUP_SHIFT_UP_DP"))
        assertTrue(screen.contains("offset(y = (-WelcomeLayout.CONTENT_GROUP_SHIFT_UP_DP).dp)"))
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
        // P02 wordmark is the splash's baked PNG, not live text: one typeface on both screens.
        assertTrue(screen.contains("R.drawable.splash_g4_wordmark"))
        assertTrue(screen.contains("WelcomeLayout.WORDMARK_IMAGE_WIDTH_DP"))
        assertTrue(screen.contains("WelcomeLayout.WORDMARK_BOX_HEIGHT_DP"))
        assertFalse(screen.contains("WORDMARK_SIZE_SP"))
        assertFalse(screen.contains("0xFFFFFFFF"))
        assertFalse(screen.contains("#FFFFFF"))
        assertFalse(screen.contains("#ffffff"))

        val colors = moduleFile("src/main/res/values/colors.xml").readText()
        assertTrue(colors.contains("<color name=\"page_background\">#F2F5F7</color>"))
        assertTrue(colors.contains("<color name=\"splash_background\">#F2F5F8</color>"))
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
        assertTrue(screen.contains("ContentScale.Fit"))
        assertTrue(screen.contains("WelcomeLayout.WORDMARK_ART_WIDTH_PX"))
        assertTrue(screen.contains("WelcomeLayout.WORDMARK_ART_HEIGHT_PX"))
        assertTrue(screen.contains("WelcomeType.notoSemiBold"))
        assertTrue(screen.contains("WelcomeType.notoRegular"))
        assertTrue(screen.contains("R.string.welcome_tagline"))
        assertTrue(screen.contains("R.string.welcome_start"))
        assertTrue(screen.contains("R.string.welcome_existing_account"))
        assertTrue(screen.contains("R.string.welcome_training_note"))
        assertTrue(screen.contains("R.string.welcome_legal"))
        assertTrue(screen.contains("R.drawable.splash_g4_wordmark"))
    }

    @Test
    fun cheerLoopsFiftyFramesAtFortyMillisWithNoEndHold() {
        assertEquals(50, WelcomeLayout.CHEER_FRAME_COUNT)
        assertEquals(40L, WelcomeLayout.CHEER_FRAME_MILLIS)
        assertEquals(50 * 40L, WelcomeLayout.CHEER_PLAY_MILLIS)
        assertEquals(2_000L, WelcomeLayout.CHEER_CYCLE_MILLIS)
        assertEquals(WelcomeLayout.CHEER_PLAY_MILLIS, WelcomeLayout.CHEER_CYCLE_MILLIS)
        assertEquals(716, WelcomeLayout.CHEER_FRAME_WIDTH_PX)
        assertEquals(480, WelcomeLayout.CHEER_FRAME_HEIGHT_PX)

        assertEquals(0, welcomeCheerFrameIndex(0))
        assertEquals(0, welcomeCheerFrameIndex(39))
        assertEquals(1, welcomeCheerFrameIndex(40))
        assertEquals(49, welcomeCheerFrameIndex(49 * 40L))
        assertEquals(49, welcomeCheerFrameIndex(50 * 40L - 1))
        assertEquals(0, welcomeCheerFrameIndex(50 * 40L))
        assertEquals(1, welcomeCheerFrameIndex(50 * 40L + 40L))
        assertEquals(49, welcomeCheerFrameIndex(2 * 50 * 40L - 1))
        assertEquals(0, welcomeCheerFrameIndex(2 * 50 * 40L))

        val delays = mutableListOf<Long>()
        var cursor = 0L
        while (cursor < WelcomeLayout.CHEER_CYCLE_MILLIS) {
            val delay = welcomeCheerDelayUntilNextBoundary(cursor)
            val index = welcomeCheerFrameIndex(cursor)
            assertEquals(index, welcomeCheerFrameIndex(cursor + delay - 1))
            delays += delay
            cursor += delay
        }
        assertEquals(List(50) { 40L }, delays)
        assertEquals(WelcomeLayout.CHEER_CYCLE_MILLIS, cursor)
        assertEquals(0, welcomeCheerFrameIndex(cursor))

        val layout = welcomeSource("WelcomeLayout.kt")
        assertFalse(layout.contains("CHEER_HOLD"))
        assertFalse(layout.contains("1_000L"))
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
        assertFalse(cheer.contains("1_000L"))
        assertFalse(cheer.contains("CHEER_HOLD"))
        assertEquals(50, WelcomeCheerFrames.ids.size)
    }

    @Test
    fun slotFramesAreTheUnmodified716By480Sequence() {
        val expected = listOf(
            "welcome_cheer_001.png" to "b0ca707383cb5b9fc46d283993ffec2887f2d94a81326dbf8532a5551fcf2920",
            "welcome_cheer_002.png" to "a6a8abaa052ff28b58847fbf647265d3eb0aff62a837a6d831b30563fa88af1b",
            "welcome_cheer_003.png" to "b020c031f4ebb4348ddd1ccaadb21e172da1adbecacf191fdad5ff94e2e6bbfd",
            "welcome_cheer_004.png" to "b73add99f6c472df66062d7cb47c1309b0e5f9b1ff6d326179b25e666df564d4",
            "welcome_cheer_005.png" to "56ec2f970b2654f4820d759aa882aed10257f3d803720d11299c9a5361f66b93",
            "welcome_cheer_006.png" to "2c9159005c6fa6df79ad11bf8fe48373b664c3e278c8b463a1154134135bc2ca",
            "welcome_cheer_007.png" to "80b7d999d0b7fc2c45927ed59281bc091e1c06dd9ad9c8575b4960466841570c",
            "welcome_cheer_008.png" to "c4558a06dbc138909903f47d60abced56e71c5046e90437cf8af91fe1d8f035c",
            "welcome_cheer_009.png" to "a5106a083c9d37682b2dff60c69afd67cfc883ae16afbc62f3d8bb15f43d5553",
            "welcome_cheer_010.png" to "410c65816f16c83e708a3489c5453efbfebfd7ca8b810ceaf132082d778bac0c",
            "welcome_cheer_011.png" to "ae9facaac26deb8eb318800856d51ce35aca527dfdc08568dd4f8f8314fb9134",
            "welcome_cheer_012.png" to "16b97d776c01f8b836371215d970f030bd21bc9b9741022e6846b31cb0f57a04",
            "welcome_cheer_013.png" to "d652ffc490d1d523074932772f1edf2e9268a7d7a3abda63752f92c28965f7ad",
            "welcome_cheer_014.png" to "d652ffc490d1d523074932772f1edf2e9268a7d7a3abda63752f92c28965f7ad",
            "welcome_cheer_015.png" to "11520a8e0771588e3fcbbad281e2467e5dd0d5ec1f39d0e11e23e902ce000784",
            "welcome_cheer_016.png" to "b5703bbba44f3f108e82cc5b6dc5ebeb08081e4fa8b74c041654152abdc1eec6",
            "welcome_cheer_017.png" to "c532c577f8d3a45f2eed52a0dde3e48a15230caf65b4b9a17ac3dd9edb3a55f0",
            "welcome_cheer_018.png" to "7bc629f035f254f6a5349172b02e51b384b378b32a11e0290ff574f6b5238e40",
            "welcome_cheer_019.png" to "dfcc73eade9556ce756e4fd74c55bd77e69cae88a145c16e7b969bd9542b09c9",
            "welcome_cheer_020.png" to "4f9b68adf70b4e2d9324105a00618829c98ca5209666c6124417bdf4a773cd2f",
            "welcome_cheer_021.png" to "1b1ce21489578915d05cb3bfd82334fea00711278a3cbaa0b0f065e4f2198878",
            "welcome_cheer_022.png" to "7dd3b5c1dce39f971c4dba96d8437bbe7b241d20012e761aa616229f6847aea2",
            "welcome_cheer_023.png" to "272ea223f3ee65874c646aba47327df513d7e58650d50087f1530eb4dd5c4e74",
            "welcome_cheer_024.png" to "3c934d2ed9f6e66fea7eb671ec51d3f8d088e1689944798c3cfee76264380624",
            "welcome_cheer_025.png" to "292032d40812ec6680c591a405c0a317693ebd4c059de7c44e5e502cf543ad9e",
            "welcome_cheer_026.png" to "2f3cc48add8d21b60efd88bdea2cc194def2a5aafbb19047f9f8b77ad00bba31",
            "welcome_cheer_027.png" to "1ee18b1ef333fed1db7cc7e9cf23bd8e6d7a9b5ef7bf29fda38ee1521f0083cd",
            "welcome_cheer_028.png" to "fdd6007b70be0450a192574b593c04b5fe2a8ade7b00712ccc3df32ec04e2395",
            "welcome_cheer_029.png" to "3906b6ec5421a88d280030968e8b35a94960634361334804eca8fba269eaca64",
            "welcome_cheer_030.png" to "ca1524cb3afdf768b1eef361d13a9dd1dd64f113e0dbdfe3b3184c93eddf2e23",
            "welcome_cheer_031.png" to "66cebfb8a5a76c10c1820065f390853026449f897fe3b9c59b2305ec1a701cf4",
            "welcome_cheer_032.png" to "19262224141ef07684b30b349e38594d170e79572e2a8f6d81a2745505be1b83",
            "welcome_cheer_033.png" to "60e3a2f61100e175c70dcda00b6ae07c95e6b13637c6a0dff51bcf1cdfbffccf",
            "welcome_cheer_034.png" to "2797575b9e295c2d03725091b67d45af806b80d7988745825ee9af4eecbac0f8",
            "welcome_cheer_035.png" to "609d415e3cfbf3ae731b285963ff4f1924db836362c4707342b187ddba58568e",
            "welcome_cheer_036.png" to "d7d44f43ec0855ca0d05bc381048fc2095e6050e8c50420fd8a5eda9a5711cf7",
            "welcome_cheer_037.png" to "abed19cbd465c11e65260866a04c4da8d3ca95e2e74f909dbca802fe51b7ee1d",
            "welcome_cheer_038.png" to "98ad4ac0d7d906a56097cf6bc676b81689dcf547eff4d21cf1a12a363048e0cb",
            "welcome_cheer_039.png" to "98ad4ac0d7d906a56097cf6bc676b81689dcf547eff4d21cf1a12a363048e0cb",
            "welcome_cheer_040.png" to "ebae018a13caaee27365b8a499cbde2ed1438847de3a1cdd0857d9541df468a9",
            "welcome_cheer_041.png" to "431757869177e08b60826f6bb4f5d2c5d4c42d7100b25d782d7ae49e7e304218",
            "welcome_cheer_042.png" to "2372c9abb5a9a383eb61d8e512203b8e4b2116c3fe39f2d3d9c2a2e73560df49",
            "welcome_cheer_043.png" to "ad68a5ff80b306ab5230a78ce1a2dd8c8b6fd658a7e1d834914e39119cf7a04e",
            "welcome_cheer_044.png" to "078fe0cfb0740145292511a775a9f22dfadc56baafaf48783740bf2e2d95fb68",
            "welcome_cheer_045.png" to "5162ebe4e6d0b4dd489fc7099c55b07b4166d76a1de165af3abfea7236dcf3e2",
            "welcome_cheer_046.png" to "569ca8d11d3f52ab59cb9e08ade619324504b3d95d4642c7a93a675508736ebc",
            "welcome_cheer_047.png" to "16e6aca50678ff1d250d3aeb8efa4098b7fb180a0fe8068f5d036de57247403d",
            "welcome_cheer_048.png" to "cb268e3b08892136324018b85b1253723b9054edfe080979b4a838745c9001f1",
            "welcome_cheer_049.png" to "230d65a56de08f08277bb462a972b9cad8fee4f71964641b3f32cdeeebfab3f4",
            "welcome_cheer_050.png" to "ec9ab92345e2a2012b0133123431d97ebc8b22c388a35d2b28ec2aebfedcfa8c",
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
        assertTrue(welcome.contains("splash_g4_wordmark"))
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
        assertTrue(gradle.contains("versionCode = 6"))
    }

    @Test
    fun welcomeWordmarkIsTheSameBakedArtworkTheSplashDraws() {
        val artwork = moduleFile("src/main/res/drawable-nodpi/splash_g4_wordmark.png")
        val bytes = artwork.readBytes()
        assertEquals(
            "ac47a0ff712f9af63632ac44c7a2014ab7fe71ca18afd48587450b2dd88cc92b",
            sha256(bytes),
        )
        assertEquals(WelcomeLayout.WORDMARK_ART_WIDTH_PX, readInt(bytes, 16))
        assertEquals(WelcomeLayout.WORDMARK_ART_HEIGHT_PX, readInt(bytes, 20))

        val welcome = welcomeSource("WelcomeScreen.kt")
        val splashSources = listOf("SplashFrame.kt", "SplashVideoView.kt", "SplashLaunchArtwork.kt")
            .joinToString("\n") { moduleFile("src/main/java/dev/qlutch/app/splash/$it").readText() }
        assertTrue(welcome.contains("R.drawable.splash_g4_wordmark"))
        assertTrue(splashSources.contains("R.drawable.splash_g4_wordmark"))

        // Exactly one copy of the artwork, and no font-rendered wordmark left behind.
        assertTrue(
            moduleFile("src/main/res").walk().none { it.isFile && it.name.startsWith("wordmark") },
        )
        assertFalse(welcome.contains("WelcomeType.interExtraBold"))
        assertFalse(welcome.contains("WelcomeType.notoExtraBold"))
        assertFalse(welcome.contains("WORDMARK_SIZE_SP"))
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
