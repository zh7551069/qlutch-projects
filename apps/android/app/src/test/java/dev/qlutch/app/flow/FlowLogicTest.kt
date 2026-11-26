package dev.qlutch.app.flow

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import java.io.File
import java.security.MessageDigest

class FlowLogicTest {
    @Test
    fun optionKeysAreTheBridgeLabels() {
        val age = FlowLogic.question(2)
        assertEquals("age", age.stepId)
        assertEquals(listOf("18 岁以下", "18–25 岁", "26–35 岁", "35 岁以上"), age.choices)
        age.choices.forEach { assertEquals(it, FlowLogic.answerKey(it)) }
        assertEquals("建档完成，上传第一份材料", FlowLogic.question(7).finishLabel)
        assertEquals(null, FlowLogic.question(6).finishLabel)
        assertEquals(7, FlowLogic.questions.size)
    }

    @Test
    fun skipFollowsTheServerNextStep() {
        assertEquals(FlowRoute.HOME, FlowLogic.placeAfterSession(FlowLogic.NEXT_HOME, true).route)
        val unfinished = FlowLogic.placeAfterSession(FlowLogic.NEXT_ONBOARDING, true)
        assertEquals(FlowRoute.NICKNAME, unfinished.route)
        assertFalse(unfinished.route == FlowRoute.HOME)
        assertFalse(unfinished.route == FlowRoute.WAIT)
        assertFalse(FlowLogic.waitReleasesHome(FlowLogic.JOB_ACCEPTED))
        assertFalse(FlowLogic.waitReleasesHome(FlowLogic.JOB_RUNNING))
    }

    @Test
    fun h3RoutesStayAnImplementationHypothesis() {
        assertEquals(FlowRoute.NICKNAME, FlowLogic.placeAfterSession(FlowLogic.NEXT_ONBOARDING, false).route)
        assertEquals(FlowRoute.PERMISSIONS_CONTACTS, FlowLogic.placeAfterSession(FlowLogic.NEXT_FIRST_PACK, false).route)
        assertEquals(FlowRoute.PACK, FlowLogic.placeAfterSession(FlowLogic.NEXT_FIRST_PACK, true).route)
        assertEquals(FlowRoute.HOME, FlowLogic.placeAfterSession(FlowLogic.NEXT_HOME, false).route)
        assertEquals(FlowRoute.QUESTION, FlowLogic.placeForOnboarding("exam_track").route)
        assertEquals(6, FlowLogic.placeForOnboarding("exam_track").stepIndex)
    }

    @Test
    fun otpAndMapGate() {
        assertFalse(FlowLogic.otpContinueEnabled("12345"))
        assertTrue(FlowLogic.otpContinueEnabled("123456"))
        assertFalse(FlowLogic.otpContinueEnabled("12345a"))
        assertEquals("138 0013 8000", FlowLogic.formatPhone("+86", "13800138000"))
        assertEquals("1234 5", FlowLogic.formatPhone("+852", "12345"))
        assertEquals("1234 5678", FlowLogic.formatPhone("+852", "12345678"))
        assertTrue(FlowLogic.waitReleasesHome(FlowLogic.JOB_MAP_READY))
        assertFalse(FlowLogic.waitReleasesHome(FlowLogic.JOB_ACCEPTED))
        assertFalse(FlowLogic.waitReleasesHome(FlowLogic.JOB_RUNNING))
        assertFalse(FlowLogic.waitReleasesHome(FlowLogic.JOB_FAILED_RETRYABLE))
        assertFalse(FlowLogic.waitReleasesHome(FlowLogic.JOB_FAILED_TERMINAL))
        assertFalse(FlowLogic.waitReleasesHome(""))
        assertTrue(FlowLogic.waitAllowsRetry(FlowLogic.JOB_FAILED_RETRYABLE, true))
        assertFalse(FlowLogic.waitAllowsRetry(FlowLogic.JOB_FAILED_TERMINAL, false))
        assertFalse(FlowLogic.waitAllowsRetry(FlowLogic.JOB_FAILED_RETRYABLE, false))
        assertFalse(FlowLogic.waitAllowsRetry(FlowLogic.JOB_ACCEPTED, false))
        assertEquals(0, FlowLogic.readingLoopFrame(0L))
        assertEquals(0, FlowLogic.readingLoopFrame(FlowLogic.READING_LOOP_MILLIS))
        assertEquals(1, FlowLogic.readingLoopFrame(FlowLogic.READING_LOOP_MILLIS + 40L))
        assertTrue(FlowLogic.waitMeterFraction() < 1f)
        assertEquals(99f / 248f, FlowLogic.waitMeterFraction(), 0.0001f)
        val key = FlowLogic.newIdempotencyKey()
        assertTrue(key.isNotEmpty())
        assertTrue(key.length <= 200)
        assertTrue(key.all { it.code in 32..126 })
    }

    @Test
    fun mainlandPhoneStaysThreeFourFourAndCaretStaysAfterTheLastDigit() {
        val expect = listOf(
            "1",
            "13",
            "138",
            "138 0",
            "138 00",
            "138 001",
            "138 0013",
            "138 0013 8",
            "138 0013 80",
            "138 0013 800",
            "138 0013 8000",
        )
        var shown = ""
        val typed = "13800138000"
        for (index in typed.indices) {
            val after = shown + typed[index]
            val proposed = FlowLogic.cursorAfterReplacement(shown, after)
            val edit = FlowLogic.editPhoneField(FlowLogic.REGION_MAINLAND, after, proposed)
            assertEquals(expect[index], edit.text)
            assertEquals(edit.text.length, edit.cursor)
            assertEquals(typed.substring(0, index + 1), edit.digits)
            assertFalse(edit.digits.contains(" "))
            shown = edit.text
        }
    }

    @Test
    fun mainlandCaretStaysOnTheEditedDigitAcrossSpaces() {
        val full = "138 0013 8000"
        val withoutMiddleOne = full.removeRange(6, 7)
        val deleted = FlowLogic.editPhoneField(
            "+86",
            withoutMiddleOne,
            FlowLogic.cursorAfterReplacement(full, withoutMiddleOne),
        )
        assertEquals("1380038000", deleted.digits)
        assertEquals("138 0038 000", deleted.text)
        assertEquals(FlowLogic.phoneCursor(deleted.text, 5), deleted.cursor)

        val grouped = "138 4567"
        val withoutSpace = grouped.removeRange(3, 4)
        val crossed = FlowLogic.editPhoneField(
            "+86",
            withoutSpace,
            FlowLogic.cursorAfterReplacement(grouped, withoutSpace),
        )
        assertEquals("1384567", crossed.digits)
        assertEquals("138 4567", crossed.text)
        assertEquals(3, FlowLogic.digitsBeforeCursor(crossed.text, crossed.cursor))

        val inserted = crossed.text.substring(0, crossed.cursor) + "9" + crossed.text.substring(crossed.cursor)
        val typed = FlowLogic.editPhoneField("+86", inserted, crossed.cursor + 1)
        assertEquals("13894567", typed.digits)
        assertEquals("138 9456 7", typed.text)
        assertEquals(4, FlowLogic.digitsBeforeCursor(typed.text, typed.cursor))
        assertFalse(typed.digits.contains(" "))
    }

    @Test
    fun authChannelsKeepFiftyDpAndScrollInsteadOfCompressing() {
        val ui = repoFile("apps/android/app/src/main/java/dev/qlutch/app/ui/QlutchUi.kt").readText()
        val channel = ui.substringAfter("fun ClosedChannel(").substringBefore("fun WeChatMark(")
        assertTrue(channel.contains("requiredHeight(50.dp)"))
        assertFalse(channel.contains(".height(50.dp)"))
        val activity = repoFile("apps/android/app/src/main/java/dev/qlutch/app/flow/FlowActivity.kt").readText()
        val auth = activity.substringAfter("private fun AuthScreen(").substringBefore("private fun TabChip(")
        assertTrue(auth.contains("verticalScroll(rememberScrollState())"))
        assertTrue(auth.contains("AuthBody("))
        val body = activity.substringAfter("private fun AuthBody(").substringBefore("private fun AuthScreen(")
        assertTrue(body.contains("Constraints.Infinity"))
        assertTrue(activity.contains("使用微信登录"))
        assertTrue(activity.contains("使用支付宝登录"))
        assertTrue(activity.contains("FlowLogic.TOAST_CLOSED") || activity.contains("onClosed"))
    }

    @Test
    fun progressStaysOnStepSevenAfterOnboarding() {
        assertEquals(45, FlowLogic.progressFillDp(1))
        assertEquals(91, FlowLogic.progressFillDp(2))
        assertEquals(318, FlowLogic.progressFillDp(7))
    }

    @Test
    fun waitCopyAndGifBytes() {
        assertEquals("正在挖掘知识点", FlowLogic.WAIT_TITLE)
        assertEquals("生成知识地图需要一点时间，完成后才会进入首页", FlowLogic.WAIT_CAPTION)
        assertEquals("请先在这一页等完，不能提前进入首页", FlowLogic.WAIT_FOOTNOTE)
        assertFalse(FlowLogic.WAIT_TITLE == "正在挖掘")
        val pack = repoFile("design/bridge-export/onboarding-20261125/index.html")
        assertEquals(403695, pack.length())
        assertEquals("e6aae9f3ffd6cf52253725b2279d31c5", md5(pack.readBytes()))
        val gif = repoFile("apps/android/app/src/main/res/raw/reading_2s.gif")
        val bytes = gif.readBytes()
        assertEquals("GIF89a", bytes.copyOfRange(0, 6).toString(Charsets.US_ASCII))
        assertEquals(3390848, bytes.size)
        assertEquals("71525c5f1e068c4c5ca262af359ea3e1", md5(bytes))
        val source = repoFile("design/bridge-export/onboarding-20261125/images/f97e58e5a17c396b905d0cc0c45a0fc5f0e2e72c.png")
        assertEquals(md5(bytes), md5(source.readBytes()))
    }

    @Test
    fun sheetsOnlyListTheLockedChoices() {
        assertEquals(listOf("+86", "+852"), FlowLogic.regions.map { it.code })
        assertEquals(listOf("zh-Hans", "zh-Hant"), FlowLogic.languages.map { it.locale })
        assertEquals(listOf("香港"), FlowLogic.filterRegions("852").map { it.name })
        assertEquals(listOf("繁体中文"), FlowLogic.filterLanguages("繁").map { it.label })
        assertEquals("功能尚未开放", FlowLogic.TOAST_CLOSED)
    }

    @Test
    fun waitFailureStaysOnTheWaitPageAndHidesRawStatus() {
        assertFalse(FlowLogic.waitFailed(FlowLogic.JOB_ACCEPTED))
        assertFalse(FlowLogic.waitFailed(FlowLogic.JOB_RUNNING))
        assertFalse(FlowLogic.waitFailed(FlowLogic.JOB_MAP_READY))
        assertTrue(FlowLogic.waitFailed(FlowLogic.JOB_FAILED_RETRYABLE))
        assertTrue(FlowLogic.waitFailed(FlowLogic.JOB_FAILED_TERMINAL))
        assertFalse(FlowLogic.waitReleasesHome(FlowLogic.JOB_FAILED_RETRYABLE))
        assertFalse(FlowLogic.waitReleasesHome(FlowLogic.JOB_FAILED_TERMINAL))
        assertTrue(FlowLogic.waitAllowsRetry(FlowLogic.JOB_FAILED_RETRYABLE, true))
        assertFalse(FlowLogic.waitAllowsRetry(FlowLogic.JOB_FAILED_TERMINAL, true))

        assertEquals(
            "地图暂时没有生成，可以用同一份材料再试一次",
            FlowLogic.waitFailureMessage(FlowLogic.JOB_FAILED_RETRYABLE, null),
        )
        assertEquals(
            "这份材料已经无法读取，请重新上传",
            FlowLogic.waitFailureMessage(FlowLogic.JOB_FAILED_TERMINAL, "  "),
        )
        assertEquals(
            "地图暂时没有生成，可以用同一份材料再试一次",
            FlowLogic.waitFailureMessage(FlowLogic.JOB_FAILED_RETRYABLE, "failed_retryable"),
        )
        assertEquals("服务暂时不可用", FlowLogic.waitFailureMessage(FlowLogic.JOB_FAILED_TERMINAL, "服务暂时不可用"))
        assertEquals("", FlowLogic.waitFailureMessage(FlowLogic.JOB_RUNNING, null))
        assertFalse(FlowLogic.WAIT_FAIL_TITLE == FlowLogic.WAIT_TITLE)
        assertEquals("", FlowLogic.sourcePageError("failed_retryable"))
        assertEquals("", FlowLogic.sourcePageError(" failed_terminal "))
        assertEquals("", FlowLogic.sourcePageError("map_ready"))
        assertEquals("VALIDATION_MATERIAL_NOT_UPLOADED", FlowLogic.sourcePageError("VALIDATION_MATERIAL_NOT_UPLOADED"))

        val activity = repoFile("apps/android/app/src/main/java/dev/qlutch/app/flow/FlowActivity.kt").readText()
        val wait = activity.substringAfter("private fun WaitScreen(").substringBefore("private fun ReadingGif(")
        assertTrue(wait.contains("FlowLogic.WAIT_FAIL_TITLE"))
        assertTrue(wait.contains("failure_message"))
        assertTrue(wait.contains("if (failed)"))
        assertTrue(wait.contains("if (retryable)"))
        assertTrue(wait.contains("再试一次"))
        assertFalse(wait.contains("failure_code"))
        assertFalse(wait.contains("JOB_FAILED_RETRYABLE"))
        val paste = activity.substringAfter("private fun PasteScreen(").substringBefore("private fun WaitScreen(")
        val more = activity.substringAfter("private fun MoreScreen(").substringBefore("private fun PasteScreen(")
        assertTrue(paste.contains("FlowLogic.sourcePageError"))
        assertTrue(more.contains("FlowLogic.sourcePageError"))
        val retry = activity.substringAfter("fun retryImport()").substringBefore("fun skipFirstPack()")
        assertTrue(retry.contains("FlowLogic.newIdempotencyKey()"))
        assertTrue(retry.contains("packId, materialId"))
        assertFalse(retry.contains("createUploadUrl"))
        val overlay = activity.substringAfter("overlay = {").substringBefore("when (place.route)")
        assertTrue(overlay.contains("ClosedToast(toast)"))
        val auth = activity.substringAfter("FlowRoute.AUTH ->").substringBefore("FlowRoute.OTP ->")
        assertTrue(auth.contains("onClosed = { toast = true }"))

        val ios = repoFile("apps/ios/Qlutch/Flow/FlowView.swift").readText()
        val iosLogic = repoFile("apps/ios/Qlutch/Flow/FlowLogic.swift").readText()
        val iosControls = repoFile("apps/ios/Qlutch/UI/Controls.swift").readText()
        assertTrue(iosLogic.contains("func waitFailureMessage"))
        assertTrue(iosLogic.contains("func sourcePageError"))
        assertTrue(ios.contains("OtpCaptureField"))
        assertTrue(ios.contains("otpFocused = true"))
        assertTrue(ios.contains("FlowLogic.waitFailTitle"))
        assertTrue(ios.contains("FlowLogic.sourcePageError"))
        assertTrue(ios.contains("ClosedToast(visible: true)"))
        assertTrue(ios.contains("使用微信登录"))
        assertTrue(ios.contains("使用支付宝登录"))
        assertFalse(ios.substringAfter("private var wait:").substringBefore("private var regionSheet:").contains("failure_code"))
        val iosRetry = ios.substringAfter("private func retryImport()").substringBefore("private func callError")
        assertTrue(iosRetry.contains("FlowLogic.newIdempotencyKey()"))
        assertTrue(iosRetry.contains("packId: packId, materialId: materialId"))
        assertFalse(iosRetry.contains("createUploadURL"))
        assertTrue(iosControls.contains("becomeFirstResponder()"))
        assertTrue(iosControls.contains("keyboardType = .numberPad"))
        assertFalse(ios.contains(".frame(width: 1, height: 1)"))

        val harmony = repoFile("apps/harmony/entry/src/main/ets/flow/FlowHost.ets").readText()
        val harmonyLogic = repoFile("apps/harmony/entry/src/main/ets/flow/FlowLogic.ets").readText()
        assertTrue(harmonyLogic.contains("static waitFailureMessage"))
        assertTrue(harmonyLogic.contains("static sourcePageError"))
        assertTrue(harmony.contains("FlowLogic.waitFailTitle"))
        assertTrue(harmony.contains("FlowLogic.sourcePageError"))
        assertTrue(harmony.contains("failure_message"))
        assertTrue(harmony.contains("ClosedToast({ visible: true })"))
        assertTrue(harmony.contains("使用微信登录"))
        assertTrue(harmony.contains("使用支付宝登录"))
        val harmonyWait = harmony.substringAfter("waitPage()").substringBefore("top(stepIndex: number)")
        assertFalse(harmonyWait.contains("failure_code"))
        assertTrue(harmonyWait.contains("this.waitFailed"))
        assertTrue(harmonyWait.contains("再试一次"))
        val harmonyRetry = harmony.substringAfter("private retryImport()").substringBefore("private pickFile()")
        assertTrue(harmonyRetry.contains("FlowLogic.newIdempotencyKey()"))
        assertTrue(harmonyRetry.contains("this.packId, this.materialId"))
        assertFalse(harmonyRetry.contains("createUploadUrl"))
    }

    @Test
    fun authSheetsOverlayThePageAndOnlyTouchLocalChoice() {
        val source = repoFile("apps/android/app/src/main/java/dev/qlutch/app/flow/FlowActivity.kt").readText()
        val overlay = source.substringAfter("overlay = {").substringBefore("when (place.route)")
        assertTrue(overlay.contains("if (sheet == \"region\")"))
        assertTrue(overlay.contains("if (sheet == \"language\")"))
        assertTrue(overlay.contains("region = choice.code"))
        assertTrue(overlay.contains("locale = choice.locale"))
        assertTrue(overlay.contains("sheet = \"\""))
        assertFalse(overlay.contains("api."))
        val page = source.substringAfter("when (place.route)").substringBefore("private fun runOnUi")
        assertFalse(page.contains("OptionSheet("))
        val board = source.substringAfter("private fun Board(").substringBefore("private fun AuthScreen")
        val column = board.substringAfter("Column(").substringBefore("overlay()")
        assertTrue(column.contains("content()"))
        assertFalse(column.contains("OptionSheet("))
    }

    private fun md5(bytes: ByteArray): String {
        return MessageDigest.getInstance("MD5").digest(bytes).joinToString("") { "%02x".format(it) }
    }

    private fun repoFile(relative: String): File {
        var dir = File(System.getProperty("user.dir") ?: ".").absoluteFile
        repeat(8) {
            val candidate = File(dir, relative)
            if (candidate.exists()) return candidate
            dir = dir.parentFile ?: return candidate
        }
        error("Missing $relative")
    }
}
