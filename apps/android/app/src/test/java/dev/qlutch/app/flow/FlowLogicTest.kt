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
