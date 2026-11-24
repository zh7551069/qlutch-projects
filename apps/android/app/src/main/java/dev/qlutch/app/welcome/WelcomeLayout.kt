package dev.qlutch.app.welcome

/**
 * P02 welcome board, taken from the Bridge export in
 * design/bridge-export/p02-welcome. Sizes are the 390×844 artboard's CSS pixels.
 */
internal object WelcomeLayout {
    const val ARTBOARD_WIDTH_DP = 390
    const val ARTBOARD_HEIGHT_DP = 844
    const val CANVAS = 0xFFF2F5F7

    /** Gap between the cheer slot and the wordmark. This preview sets it to 0. */
    const val SLOT_TO_WORDMARK_GAP_DP = 0

    /** Gap between the wordmark and the tagline. Bridge export is 20; this preview keeps 12. */
    const val WORDMARK_TO_TAGLINE_GAP_DP = 12
    const val CONTENT_PADDING_HORIZONTAL_DP = 16
    const val CONTENT_PADDING_TOP_DP = 24
    /** Moves the centered slot + wordmark + tagline group up from the content area's vertical center. */
    const val CONTENT_GROUP_SHIFT_UP_DP = 24

    const val SLOT_WIDTH_DP = 358
    const val SLOT_HEIGHT_DP = 240
    const val SLOT_RADIUS_DP = 22
    const val SLOT_COLOR = 0xFFF1F4F9

    /**
     * The cheer frames are 716×480 crops of a jumping squirrel. At rest (frames
     * 1-5, 16-17, 35, 40-46 and 50, about 40% of the cycle) its ink centre sits
     * 14 artboard px left of the canvas centre, so on screen it reads 7dp off
     * centre to the left. Shifting the artwork right by that much centres the
     * pose the eye actually rests on; the airborne frames drift a little right,
     * which reads as the squirrel leaping that way and is not noticeable at
     * 40ms per frame.
     */
    const val SLOT_CONTENT_SHIFT_RIGHT_DP = -7

    /**
     * The frames reserve headroom for the jump, which leaves the squirrel
     * sitting near the bottom of the slot: at rest its ink stops 19.4dp above
     * the slot edge, and 40dp of empty space separates it from the wordmark.
     * Moving the artwork down tightens that to about 30dp while keeping the
     * squirrel the same size and still clear of the slot's bottom edge.
     */
    const val SLOT_CONTENT_SHIFT_DOWN_DP = 20

    const val WORDMARK = "Qlutch。"
    const val WORDMARK_LATIN = "Qlutch"
    const val WORDMARK_PERIOD = "。"

    /**
     * The wordmark is the same baked artwork the splash draws
     * (res/drawable-nodpi/splash_g4_wordmark.png), so both screens cannot drift
     * apart in typeface. The glyphs, the trailing "。" and the #6C5CE7 fill are
     * already inside the PNG; nothing here is rendered from a font.
     */
    const val WORDMARK_ART_WIDTH_PX = 941
    const val WORDMARK_ART_HEIGHT_PX = 238

    /** Display width on the 390 artboard. Height follows the artwork aspect (about 44dp). */
    const val WORDMARK_IMAGE_WIDTH_DP = 174

    /**
     * The wordmark keeps the box the old 50sp text line occupied so that
     * SLOT_TO_WORDMARK_GAP_DP and WORDMARK_TO_TAGLINE_GAP_DP still measure the
     * same edges. That line was 72dp tall, not 50dp: the trailing "。" was set
     * in Noto Sans SC, whose line metrics run about 1.45em, and it stretched
     * the shared row. Measured on the emulator at 420dpi (190px).
     */
    const val WORDMARK_BOX_HEIGHT_DP = 72

    /**
     * A 50sp text line puts its ink a little below the box centre, because
     * Inter's ascent is taller than its descent, while the baked PNG is
     * geometrically centred in its own box. Shifting the image down by this
     * keeps the ink where the text had it, so neither the slot gap above nor
     * the tagline gap below moves.
     */
    const val WORDMARK_INK_SHIFT_DOWN_DP = 5

    /** Baked into the shared artwork; kept here so the palette stays in one place. */
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

    const val CHEER_FRAME_COUNT = 50
    const val CHEER_FRAME_MILLIS = 40L
    /** One pass is 50×40ms = 2.00s. The next pass starts on this boundary. */
    const val CHEER_PLAY_MILLIS = CHEER_FRAME_COUNT * CHEER_FRAME_MILLIS
    const val CHEER_CYCLE_MILLIS = CHEER_PLAY_MILLIS

    const val CHEER_FRAME_WIDTH_PX = 716
    const val CHEER_FRAME_HEIGHT_PX = 480
}

/** Zero-based frame in a continuous loop. Index 0 is frame 1. */
internal fun welcomeCheerFrameIndex(elapsedMillis: Long): Int {
    val count = WelcomeLayout.CHEER_FRAME_COUNT.toLong()
    val frame = WelcomeLayout.CHEER_FRAME_MILLIS
    return elapsedMillis.floorDiv(frame).mod(count).toInt()
}

/** Millis until the next frame boundary. The cycle restarts immediately. */
internal fun welcomeCheerDelayUntilNextBoundary(elapsedMillis: Long): Long {
    val frame = WelcomeLayout.CHEER_FRAME_MILLIS
    val into = elapsedMillis.mod(frame)
    return if (into == 0L) frame else frame - into
}
