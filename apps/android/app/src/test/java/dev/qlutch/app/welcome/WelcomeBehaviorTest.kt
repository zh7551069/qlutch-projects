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
            "welcome_cheer_001.png" to "0ceecb088cb5f389573cb5371747e373c402aeab314829f50f15ad80a44b5508",
            "welcome_cheer_002.png" to "28070bf2e501b4931a40bddc758d5c63b7ec1448541763c36b45f48d1dcf72bb",
            "welcome_cheer_003.png" to "57699e122e744bb71206097d5c1f6bcc5f6c3def6630f8328ee33e7a836995af",
            "welcome_cheer_004.png" to "a7cc203f3f648eb9793fa27d5045f935ad5a9b18eb9e285622a21c5d9aeb314a",
            "welcome_cheer_005.png" to "60b6f57e1199d440b2c985c7a4074383d0371eab111d59d772ef5938f8375ad5",
            "welcome_cheer_006.png" to "95b0e0d467b64c1f596dae7eaabc6a4e69b5f8df03229cd71806f5fbe7129c73",
            "welcome_cheer_007.png" to "6085ff52e923c7b57e5e6eeaf76a10d38a46b13fd31f3c8e7b1bdc64b9ceaf26",
            "welcome_cheer_008.png" to "eba773cc156d65e7575d7448a16d0881e4b00d8e8cc7b325c34b41c37eda4f6c",
            "welcome_cheer_009.png" to "6dd068edd5109422ef353fc13f6dfffca6d29089b91b5188b253ab55c68d4aa0",
            "welcome_cheer_010.png" to "8b101828dc5454fa82606f8295a7b740539db559955ee0b97020e71ddedce161",
            "welcome_cheer_011.png" to "046a02d5bca28c5de301794cf3514625ec27a494937453229fc541fa0b282996",
            "welcome_cheer_012.png" to "6ca245b6db84dd3b424a42976faa794baecfcc2533df7e1d4864a0c38b34dd6b",
            "welcome_cheer_013.png" to "b0c0151ad4e895cfad37817a79d87bedb9c2635585b8a79e79c42f00f3d8a6b5",
            "welcome_cheer_014.png" to "b0c0151ad4e895cfad37817a79d87bedb9c2635585b8a79e79c42f00f3d8a6b5",
            "welcome_cheer_015.png" to "d21884a9624b9d3829b3615f6bf8f38030dd0aa19dce83ecc783cabbaedbe736",
            "welcome_cheer_016.png" to "538be4501a6bafb02605e65b595c247d412301d709c9c5cd6400a956945fe737",
            "welcome_cheer_017.png" to "e834ca744d826df6a56c4e2f12a96737014b922e751e6943c5207667c0b67736",
            "welcome_cheer_018.png" to "2a940217e462eb922ce2c21fc87302cc4fe27b84fe6831b34a87011634537497",
            "welcome_cheer_019.png" to "ce46f8fb24fe9459a008dd440f5c7dc9dce06e3ef196c4f1aa6f47881b380a8f",
            "welcome_cheer_020.png" to "41dccfcfc237db6ad19f1f565214cb0e4be2065efb902fe982a9f208056de497",
            "welcome_cheer_021.png" to "a3d393f699ea3961c87db18b3d6b7ae9b6474d5882824f22f8afcb14f6f6567d",
            "welcome_cheer_022.png" to "f7757e6cb67aecb0882f9346c7c132096209bb75763ea51368052743c187bb6a",
            "welcome_cheer_023.png" to "1144e8c6bfc2a4eddc37fad722467eb2447ed273080bcddcff8f2c42c92b7d78",
            "welcome_cheer_024.png" to "9715bf4650e41f52d781740b89bbeee6908a34cc6a136163534cf023f40c01ef",
            "welcome_cheer_025.png" to "ce7f4976815db4a3f2d717380b003b8d2ecffd4ec7ad4d57c71a40576a194db3",
            "welcome_cheer_026.png" to "a3dc3df09062ff930a2de2cf735d9bae3a3eb582b794e9f9b8d2a8bdfe371867",
            "welcome_cheer_027.png" to "1a361a9a87030ccbf0e7c66730e624fb67bec325ce633287b8fb4514cad2b093",
            "welcome_cheer_028.png" to "23d4fea0577a8bb1a2ed4703442eecc14981ba7500dccde67cb3320636ae511f",
            "welcome_cheer_029.png" to "01e9fa698dcc50c58677378c5382136b2efbcd3765628e9d575919da73b5e6ee",
            "welcome_cheer_030.png" to "7f3282edd6ed25c232f51e1839bb5d8092fc218b87c5447cbb9ffd08efa8a813",
            "welcome_cheer_031.png" to "9cb4cd80b53ef8790b92cb903148585214fcc90c6fc790c26896514e7bdf1c36",
            "welcome_cheer_032.png" to "698b86cb909bd8d2943db3251539c2f9ca13b551f659d2bb9dd8fda6129aeec0",
            "welcome_cheer_033.png" to "7cc5707c981cddbccf6030d15ac1ac8ba0ca696823bdfdc837f882668423b021",
            "welcome_cheer_034.png" to "8e6cba0d28fa148771683c2ac9861721df57950755d02fbae893c229ec466e85",
            "welcome_cheer_035.png" to "6fa04f8cdf1b3b8b2e0c98272b76467beda2e5f54848a186e84261742de8c298",
            "welcome_cheer_036.png" to "c87da137f5fda6e82921083ae2217002867e675ad03e29735668e0911be28e87",
            "welcome_cheer_037.png" to "ff006158f98b57ef523bf36575543fc16ddaf8ad2ae4ba6b404d5b47e0df8253",
            "welcome_cheer_038.png" to "9d90830937baba40f2c3df86cb514d074a657d6f66d6c661c742a6c36c8b1e6f",
            "welcome_cheer_039.png" to "9d90830937baba40f2c3df86cb514d074a657d6f66d6c661c742a6c36c8b1e6f",
            "welcome_cheer_040.png" to "e18b7835592cf7831e074dbcd16607fcb79912159e204c2014cb9f2272d45799",
            "welcome_cheer_041.png" to "f5426c6a76347424786ef40f6709e8350932abd25bae49d76b29ee0f76598d58",
            "welcome_cheer_042.png" to "8074b158132e168e10588ce00043e8cf5a31fbf63feef9bfc710c95c83eadc4f",
            "welcome_cheer_043.png" to "717ae39e4b7e9a576a84a993aa2c39a311bb1c578dc5c2503f81b31b5e15f745",
            "welcome_cheer_044.png" to "6d5f6e4464c605309f2dc098f185ce480cb499ca79bcd3ff6358292d8b1bba92",
            "welcome_cheer_045.png" to "585562996a74dcd84cbbbb34327df40757f01ac53dbc68c70d405ff7b58d66ad",
            "welcome_cheer_046.png" to "e0b92426ace6356931ba15584c36b7ef5a861abbd04f1768386500d3fdbe6c27",
            "welcome_cheer_047.png" to "961712cd3f709e71f7ec84a550384deda64960e219d2fc1485a5f7b68953f038",
            "welcome_cheer_048.png" to "278aa31f1d5c52829487d14de459b5f645849d1696c19101e205982db3a088e9",
            "welcome_cheer_049.png" to "edf9bd3134524c141a6feb17b588260b11f3dba06b764c9d70b7f545cb0fec8b",
            "welcome_cheer_050.png" to "de819af9f688a72f1b6581ae46a8d4147360b429da944fb376720ad34f885ae1",
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
        assertTrue(gradle.contains("versionCode = 7"))
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
