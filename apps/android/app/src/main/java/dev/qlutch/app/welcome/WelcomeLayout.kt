package dev.qlutch.app.welcome

/**
 * P02 welcome board, taken from the Bridge export in
 * design/bridge-export/p02-welcome. Sizes are the 390×844 artboard's CSS pixels.
 */
internal object WelcomeLayout {
    const val ARTBOARD_WIDTH_DP = 390
    const val ARTBOARD_HEIGHT_DP = 844
    const val CANVAS = 0xFFFFFFFF

    const val CONTENT_GAP_DP = 20
    const val CONTENT_PADDING_HORIZONTAL_DP = 16
    const val CONTENT_PADDING_TOP_DP = 24

    const val SLOT_WIDTH_DP = 358
    const val SLOT_HEIGHT_DP = 240
    const val SLOT_RADIUS_DP = 22
    const val SLOT_COLOR = 0xFFF1F4F9

    const val WORDMARK = "Qlutch。"
    const val WORDMARK_LATIN = "Qlutch"
    const val WORDMARK_PERIOD = "。"
    const val WORDMARK_SIZE_SP = 50
    const val WORDMARK_WEIGHT = 800
    const val WORDMARK_COLOR = 0xFF6C5CE7

    const val TAGLINE = "沉迷于学习，对学习上瘾"
    const val TAGLINE_SIZE_SP = 16
    const val TAGLINE_WEIGHT = 600
    const val TAGLINE_COLOR = 0xFF0F172A
    const val TAGLINE_WIDTH_DP = 358

    const val FOOTER_GAP_DP = 16
    const val FOOTER_PADDING_HORIZONTAL_DP = 16
    const val FOOTER_PADDING_BOTTOM_DP = 28

    const val BUTTON_WIDTH_DP = 358
    const val BUTTON_HEIGHT_DP = 50
    const val BUTTON_RADIUS_DP = 25
    const val BUTTON_TEXT_SIZE_SP = 16
    const val BUTTON_TEXT_WEIGHT = 600
    const val PRIMARY_LABEL = "开始"
    const val PRIMARY_COLOR = 0xFF0F172A
    const val PRIMARY_TEXT_COLOR = 0xFFFFFFFF
    const val SECONDARY_LABEL = "已有账号"
    const val SECONDARY_COLOR = 0xFFFFFFFF
    const val SECONDARY_TEXT_COLOR = 0xFF0F172A
    const val SECONDARY_BORDER_COLOR = 0xFFE1E8F0
    const val SECONDARY_BORDER_DP = 1

    const val TRAINING_NOTE = "教材不用于训练模型"
    const val LEGAL = "继续即表示同意《用户协议》和《隐私政策》"
    const val FOOTER_TEXT_SIZE_SP = 13
    const val FOOTER_TEXT_WEIGHT = 400
    const val FOOTER_TEXT_COLOR = 0xFF7C848F
    const val LEGAL_WIDTH_DP = 358

    const val CHEER_FRAME_COUNT = 25
    const val CHEER_FRAME_MILLIS = 40L
    const val CHEER_HOLD_MILLIS = 1_000L
    const val CHEER_PLAY_MILLIS = CHEER_FRAME_COUNT * CHEER_FRAME_MILLIS
    const val CHEER_CYCLE_MILLIS = CHEER_PLAY_MILLIS + CHEER_HOLD_MILLIS

    const val CHEER_FRAME_WIDTH_PX = 716
    const val CHEER_FRAME_HEIGHT_PX = 480
}

/** Zero-based frame during a play-then-hold cycle. Index 0 is frame 1. */
internal fun welcomeCheerFrameIndex(elapsedMillis: Long): Int {
    val t = elapsedMillis.mod(WelcomeLayout.CHEER_CYCLE_MILLIS)
    val played = t / WelcomeLayout.CHEER_FRAME_MILLIS
    return if (played >= WelcomeLayout.CHEER_FRAME_COUNT) {
        WelcomeLayout.CHEER_FRAME_COUNT - 1
    } else {
        played.toInt()
    }
}

/** Millis until the next frame boundary, or until the hold ends. */
internal fun welcomeCheerDelayUntilNextBoundary(elapsedMillis: Long): Long {
    val cycle = WelcomeLayout.CHEER_CYCLE_MILLIS
    val t = elapsedMillis.mod(cycle)
    val next = if (t >= WelcomeLayout.CHEER_PLAY_MILLIS) {
        cycle
    } else {
        val frame = WelcomeLayout.CHEER_FRAME_MILLIS
        (t / frame + 1) * frame
    }
    return (next - t).coerceAtLeast(1L)
}
