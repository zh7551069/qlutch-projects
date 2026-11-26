package dev.qlutch.app.flow

import java.util.UUID
import kotlin.math.roundToInt

/**
 * Navigation and copy for the new-user flow. Option keys are the Bridge frame
 * labels. H3 (`NEED_ONBOARDING` / `NEED_FIRST_PACK` / `READY_HOME`) is only the
 * routing hypothesis already implemented by the identity service.
 */
internal object FlowLogic {
    /** Length of one pass of the reading GIF. Looping uses this; it does not open home. */
    const val READING_LOOP_MILLIS = 2_000L

    const val JOB_ACCEPTED = "accepted"
    const val JOB_RUNNING = "running"
    const val JOB_MAP_READY = "map_ready"
    const val JOB_FAILED_RETRYABLE = "failed_retryable"
    const val JOB_FAILED_TERMINAL = "failed_terminal"

    /**
     * How often the wait page reads GET /v1/knowledge/jobs/{id}.
     * A4-O2 is still open, so this is an implementation default, not a locked product interval.
     */
    const val JOB_POLL_INTERVAL_MILLIS = 3_000L
    const val PROGRESS_TRACK_DP = 318
    /** Bridge wait meter is 248 wide with a 99-wide hold. It is not a countdown. */
    const val WAIT_TRACK_DP = 248
    const val WAIT_METER_HOLD_DP = 99

    const val REGION_MAINLAND = "+86"
    const val REGION_HONG_KONG = "+852"
    const val LOCALE_HANS = "zh-Hans"
    const val LOCALE_HANT = "zh-Hant"

    const val PURPOSE_REGISTER = "register"
    const val PURPOSE_LOGIN = "login"

    const val NEXT_ONBOARDING = "NEED_ONBOARDING"
    const val NEXT_FIRST_PACK = "NEED_FIRST_PACK"
    const val NEXT_HOME = "READY_HOME"

    const val TOAST_CLOSED = "功能尚未开放"
    const val WAIT_TITLE = "正在挖掘知识点"
    const val WAIT_CAPTION = "生成知识地图需要一点时间，完成后才会进入首页"
    const val WAIT_FOOTNOTE = "请先在这一页等完，不能提前进入首页"

    val regions = listOf(
        RegionChoice("中国大陆", REGION_MAINLAND),
        RegionChoice("香港", REGION_HONG_KONG),
    )

    val languages = listOf(
        LanguageChoice("简体中文", LOCALE_HANS),
        LanguageChoice("繁体中文", LOCALE_HANT),
    )

    val questions = listOf(
        QuestionStep(
            stepId = "nickname",
            index = 1,
            title = "先给自己起个昵称",
            choices = emptyList(),
            finishLabel = null,
        ),
        QuestionStep(
            stepId = "age",
            index = 2,
            title = "你多大啦？",
            choices = listOf("18 岁以下", "18–25 岁", "26–35 岁", "35 岁以上"),
            finishLabel = null,
        ),
        QuestionStep(
            stepId = "study_or_work",
            index = 3,
            title = "现在是学生还是在职？",
            choices = listOf("在校学生", "在职备考", "自由职业", "待业备考"),
            finishLabel = null,
        ),
        QuestionStep(
            stepId = "identity",
            index = 4,
            title = "你的身份是？",
            choices = listOf("本科生", "研究生", "上班族", "其他"),
            finishLabel = null,
        ),
        QuestionStep(
            stepId = "how_to_address",
            index = 5,
            title = "怎么称呼你？",
            choices = listOf("叫我同学", "叫我老师", "直接开始吧", "随便"),
            finishLabel = null,
        ),
        QuestionStep(
            stepId = "exam_track",
            index = 6,
            title = "备考方向是？",
            choices = listOf("自考专升本", "考研", "公考", "职业资格"),
            finishLabel = null,
        ),
        QuestionStep(
            stepId = "daily_time",
            index = 7,
            title = "每天能投入多久？",
            choices = listOf("15 分钟", "30 分钟", "1 小时", "越多越好"),
            finishLabel = "建档完成，上传第一份材料",
        ),
    )

    fun question(index: Int): QuestionStep = questions.first { it.index == index }

    fun answerKey(label: String): String = label

    fun newIdempotencyKey(): String = UUID.randomUUID().toString()

    fun phoneDigits(raw: String): String = raw.filter { it in '0'..'9' }

    fun formatPhone(region: String, digits: String): String {
        val body = phoneDigits(digits)
        if (region == REGION_MAINLAND) {
            return formatMainlandPhone(body)
        }
        return body.chunked(4).joinToString(" ")
    }

    /** +86 stays 3-4-4 for the whole entry, matching the 138 0013 8000 placeholder. */
    private fun formatMainlandPhone(digits: String): String {
        if (digits.length <= 3) return digits
        if (digits.length <= 7) return digits.substring(0, 3) + " " + digits.substring(3)
        return digits.substring(0, 3) + " " + digits.substring(3, 7) + " " + digits.substring(7)
    }

    fun digitsBeforeCursor(text: String, cursor: Int): Int {
        val end = cursor.coerceIn(0, text.length)
        var count = 0
        for (index in 0 until end) {
            if (text[index] in '0'..'9') count += 1
        }
        return count
    }

    /** Offset in [formatted] sitting just after [digitCount] digits. The full count lands after the last digit. */
    fun phoneCursor(formatted: String, digitCount: Int): Int {
        if (digitCount <= 0) return 0
        var seen = 0
        for (index in formatted.indices) {
            if (formatted[index] in '0'..'9') {
                seen += 1
                if (seen == digitCount) return index + 1
            }
        }
        return formatted.length
    }

    fun editPhoneField(region: String, proposedText: String, proposedCursor: Int): PhoneFieldEdit {
        val digits = phoneDigits(proposedText)
        val text = formatPhone(region, digits)
        val cursor = phoneCursor(text, digitsBeforeCursor(proposedText, proposedCursor))
        return PhoneFieldEdit(digits, text, cursor)
    }

    /** Caret in [after] after one contiguous insert, delete, or replace of [before]. */
    fun cursorAfterReplacement(before: String, after: String): Int {
        var prefix = 0
        val shared = minOf(before.length, after.length)
        while (prefix < shared && before[prefix] == after[prefix]) prefix += 1
        var suffix = 0
        while (
            suffix < before.length - prefix &&
            suffix < after.length - prefix &&
            before[before.length - 1 - suffix] == after[after.length - 1 - suffix]
        ) {
            suffix += 1
        }
        return after.length - suffix
    }

    fun otpContinueEnabled(code: String): Boolean = code.length == 6 && code.all { it in '0'..'9' }

    fun phoneContinueEnabled(digits: String): Boolean = digits.isNotEmpty() && digits.length <= 32

    fun nicknameContinueEnabled(nickname: String): Boolean {
        val trimmed = nickname.trim()
        return trimmed.isNotEmpty() && trimmed.codePointCount(0, trimmed.length) <= 32
    }

    fun notesContinueEnabled(notes: String): Boolean = notes.trim().isNotEmpty()

    fun progressFillDp(stepIndex: Int, trackDp: Int = PROGRESS_TRACK_DP): Int {
        val clamped = stepIndex.coerceIn(0, 7)
        return (trackDp * clamped / 7f).roundToInt()
    }

    /** Home opens only when the job status is map_ready. Elapsed time is not that signal. */
    fun waitReleasesHome(status: String): Boolean = status == JOB_MAP_READY

    /** Retry is a new import of the same material, and only for a retryable failure. */
    fun waitAllowsRetry(status: String, retryable: Boolean): Boolean {
        return status == JOB_FAILED_RETRYABLE && retryable
    }

    fun waitMeterFraction(): Float = WAIT_METER_HOLD_DP.toFloat() / WAIT_TRACK_DP.toFloat()

    /** Frame inside the looping 2s reading clip. The index wraps; it does not finish the wait. */
    fun readingLoopFrame(elapsedMillis: Long, frameCount: Int = 50, frameMillis: Long = 40L): Int {
        val count = frameCount.coerceAtLeast(1).toLong()
        val frame = frameMillis.coerceAtLeast(1L)
        return (elapsedMillis.coerceAtLeast(0L) / frame).mod(count).toInt()
    }

    fun placeAfterSession(nextStep: String, permissionsSeen: Boolean): FlowPlace {
        return when (nextStep) {
            NEXT_ONBOARDING -> FlowPlace(FlowRoute.NICKNAME, stepIndex = 1)
            NEXT_FIRST_PACK -> if (permissionsSeen) {
                FlowPlace(FlowRoute.PACK, stepIndex = 7)
            } else {
                FlowPlace(FlowRoute.PERMISSIONS_CONTACTS, stepIndex = 7)
            }
            NEXT_HOME -> FlowPlace(FlowRoute.HOME, stepIndex = 7)
            else -> FlowPlace(FlowRoute.NICKNAME, stepIndex = 1)
        }
    }

    fun placeForOnboarding(stepId: String?): FlowPlace {
        val step = questions.firstOrNull { it.stepId == stepId } ?: questions.first()
        return if (step.index == 1) {
            FlowPlace(FlowRoute.NICKNAME, stepIndex = 1)
        } else {
            FlowPlace(FlowRoute.QUESTION, stepIndex = step.index)
        }
    }

    fun filterRegions(query: String): List<RegionChoice> {
        val needle = query.trim()
        if (needle.isEmpty()) return regions
        return regions.filter { choice ->
            choice.name.contains(needle) || choice.code.contains(needle)
        }
    }

    fun filterLanguages(query: String): List<LanguageChoice> {
        val needle = query.trim()
        if (needle.isEmpty()) return languages
        return languages.filter { it.label.contains(needle) }
    }

    fun sentToLine(hint: String?, region: String, digits: String): String {
        val target = if (!hint.isNullOrBlank()) hint else region + " " + formatPhone(region, digits)
        return "验证码已发送至 $target"
    }
}

internal data class PhoneFieldEdit(
    val digits: String,
    val text: String,
    val cursor: Int,
)

internal data class RegionChoice(val name: String, val code: String)

internal data class LanguageChoice(val label: String, val locale: String)

internal data class QuestionStep(
    val stepId: String,
    val index: Int,
    val title: String,
    val choices: List<String>,
    val finishLabel: String?,
)

internal enum class FlowRoute {
    AUTH,
    OTP,
    NO_ACCOUNT,
    NICKNAME,
    QUESTION,
    PERMISSIONS_CONTACTS,
    PERMISSIONS_LOCATION,
    PACK,
    MORE,
    PASTE,
    WAIT,
    HOME,
}

internal data class FlowPlace(
    val route: FlowRoute,
    val stepIndex: Int = 1,
    val waitSource: FlowRoute = FlowRoute.PACK,
)
