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
            "welcome_cheer_001.png" to "b0514fbf21878ceace35b67686d31900397b7e9a4619bf69213ef47b8794a31f",
            "welcome_cheer_002.png" to "ccc55774ab251b2d564fa2bef1c2508482acd90dd6236e0b2540d5203feaabf2",
            "welcome_cheer_003.png" to "ee086aba17f0904658e3c2459043947c8ac682f282bd92f348a81c76b6b614f5",
            "welcome_cheer_004.png" to "64fb95be9e529b65eafe9e1ca347c376ee696b542b0dc926e0a5dfc5831f4a6a",
            "welcome_cheer_005.png" to "6c91d6cc7d500be4a545cab3eb083119f7db9eb67319255464216b17b8664b87",
            "welcome_cheer_006.png" to "2d5bab9e4b8d6c3c816de4beba90ecb683483d48dd139558b4c9717f92a5276b",
            "welcome_cheer_007.png" to "92f54c666593097f775f60ebe918cfc8265ff93484389f5e9ab3aadf79c80a9c",
            "welcome_cheer_008.png" to "21595df87cba69aebf3f6f61fec63ef8b023582ca7dd80b3357d2e5a5e598abc",
            "welcome_cheer_009.png" to "9b67566b46830b4914674ab606c560e39548e6c1964191a0193eb6a82fea494a",
            "welcome_cheer_010.png" to "77449a02560d2d5a375f13df5484a8b25b4416f9af03b56132df5033f36613b1",
            "welcome_cheer_011.png" to "531c1653625d05d093cc7964bedeca5c2ac2b1249df51013ca671fd5d7010de6",
            "welcome_cheer_012.png" to "208e43c001e640912a7d9f0227c19d716cbdbef3d8a0758fc8fbe9fa30ab89b5",
            "welcome_cheer_013.png" to "a3c272218228ce3afd88a1f669bacb4aee18dd2b8a19ef24664fb10e42bd301a",
            "welcome_cheer_014.png" to "a3c272218228ce3afd88a1f669bacb4aee18dd2b8a19ef24664fb10e42bd301a",
            "welcome_cheer_015.png" to "19e8dd0f062df121bcca6067e5e776508903db61db47f25a8f0ae2aea9d7dc5c",
            "welcome_cheer_016.png" to "767e43a9a503721b92718077071cf48b2f830edd4019760293e269aa634f93be",
            "welcome_cheer_017.png" to "3b65a4f2803dcb1942ecd4d4e179db0c673c4f178e108ff00b7c8bba8299da71",
            "welcome_cheer_018.png" to "a5ac9e0738829e4afced51ffa6d792ef8bad0e7ef7a7ea9db6417b8cd7229790",
            "welcome_cheer_019.png" to "84f5a3870700fe3cd0707315b0cf519ebc628d2f02c8a9cb3ab5a650af54ee12",
            "welcome_cheer_020.png" to "179d70cf2ecaa601cb2142613a5312a2c3ee0323a29fd2ae2a230b2a55ec7858",
            "welcome_cheer_021.png" to "e0b7ab014b7c733e5e6abedba0813e225682413b309d2c5f52409110eca36480",
            "welcome_cheer_022.png" to "abcef774efb3451c92f424fecfb9b1a3d920b4c4c29fc6625bad323aa9fce44a",
            "welcome_cheer_023.png" to "5e476bec2516705277a98a09d9a5a1f430cbc6b52bed586afcf6d015f70aea71",
            "welcome_cheer_024.png" to "f66ef1000feb5b08405500cbc0d965a3eca216e02f406a49f96aff70cef57128",
            "welcome_cheer_025.png" to "900d0c8ed3d8c8b6947c8aac8ef0dd54118f0bb3b2dcb46979c06a0ebaae8a0c",
            "welcome_cheer_026.png" to "69bf6ec2acbed8170ceaa8287cfbdc64bce22e679d6cdd3d900ea92f164fdd23",
            "welcome_cheer_027.png" to "8a6d7e0cb4780da70ce13193d9fd16766aafa55b23bb5c0885d698975e5c1607",
            "welcome_cheer_028.png" to "388cf772b60fdb10bfff2535e06a16ab54829d51df6726f611e87f29a190ed08",
            "welcome_cheer_029.png" to "f85f8b436c9d883d2089ea975c3ffa958f34291e39ec4cd78bbc4b3552ef1126",
            "welcome_cheer_030.png" to "ea04efed48ca692ba62cbb5044eacfaff5cb8017cec1ed05b8deecd0889b3b6a",
            "welcome_cheer_031.png" to "ef4aed6cfdb7f7dae8a91a901d04873acbd97bfd7472be4e8b5c3a102c9eb901",
            "welcome_cheer_032.png" to "487059f336b401c2864ff969dc484666c83516678cc1dfe38929f470d5cae089",
            "welcome_cheer_033.png" to "ba1206a74f7b56871ea3dfd534ce9b6814377b7c633ac4b9d8833eb7e4306a8f",
            "welcome_cheer_034.png" to "eab140a94adcd5274f359918e3a5ae03c36cee8d30015fdb2f54eb02138838a9",
            "welcome_cheer_035.png" to "16b7827b42008d72bfb4008df48ee8056fda49420fcd7fa7a82bd90fcc27dba5",
            "welcome_cheer_036.png" to "cadc64953746fde9528d86569ba4679ca0eb6ad9c004ec093e3788e321613b2d",
            "welcome_cheer_037.png" to "cfa95a24ea25e8b0323c6a577e6bb522133bac0120f5229ef4b0cbee47191d94",
            "welcome_cheer_038.png" to "15b7afbb0dcdbdd61893330ca32080aa6d07f064949dba35f0ddad7600938ab0",
            "welcome_cheer_039.png" to "15b7afbb0dcdbdd61893330ca32080aa6d07f064949dba35f0ddad7600938ab0",
            "welcome_cheer_040.png" to "f6689573e123c00b57d33bf1f9816d214cecdc3a3b3e796590d323d63b47b0f1",
            "welcome_cheer_041.png" to "f6fd49f5d11c81239691cc678a67bde406c8744820b27302081a524ecac766a2",
            "welcome_cheer_042.png" to "c96663440fdcd67b75dd260805e6d2c57cefe621a120451a8093bda6a91baadd",
            "welcome_cheer_043.png" to "be178ca731f09b208ccf0680501e6c4adcd896f894e25a57eb88e4485abba1f4",
            "welcome_cheer_044.png" to "22cf0dfc425b25fad61f7c862887af51a3c250d7f157176110de487e7bc1d91f",
            "welcome_cheer_045.png" to "368a7ea98a969cef3fce8b2a1cef0fd5222449f374489babc8877cce28d68951",
            "welcome_cheer_046.png" to "9329a412b625d0e45989f6885e53aa424d822a077ad1ab03e50ec132ebb14dba",
            "welcome_cheer_047.png" to "9ce6331413e988b3205ee140655c08d6911778508f6110ceae56adc564040498",
            "welcome_cheer_048.png" to "b0555df32dcbe6b4de26fed9e6c7c08b9f4f1ad7c75486143787ff203ca05da0",
            "welcome_cheer_049.png" to "3aeb024e4a137cd13ff5f2bbda211779c5d32f5f58adcfa448cb7903fa6e8155",
            "welcome_cheer_050.png" to "47cd0eade519bc19e53ad5ff5d9779e102a0407b0435d05d20df0968768287d0",
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
        assertTrue(gradle.contains("versionCode = 5"))
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
