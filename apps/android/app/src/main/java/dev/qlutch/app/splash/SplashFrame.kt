package dev.qlutch.app.splash

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxScope
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.absoluteOffset
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.width
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import dev.qlutch.app.R

internal object SplashLayout {
    /** Welcome clip width as a fraction of the splash window width. */
    const val WIDTH_FRACTION = 2f / 3f

    /** 9:16 frame. Height = width * 16/9. */
    const val HEIGHT_OVER_WIDTH = 16f / 9f

    /** Packed welcome clip. The frame matches it, so contain does not crop. */
    const val SOURCE_WIDTH = 1080f
    const val SOURCE_HEIGHT = 1920f

    /**
     * Bottom of the landing contact shadow in the 1080×1920 source.
     * Frame 42 is the largest, most solid shadow, at y=1462. Standing frames
     * sit higher, at y=1454. The wordmark stays on this landing edge for the
     * whole splash, including while the squirrel is in the air.
     */
    const val LANDING_SHADOW_BOTTOM_Y = 1462f

    /**
     * Gap from that shadow bottom to the wordmark top, on a 375-wide design.
     * Scaled by screenWidth/375. The design value sits in the 20–32 range.
     */
    const val WORDMARK_GAP_AT_375 = 26f
    const val DESIGN_WIDTH = 375f

    /**
     * Wordmark width over the contained clip.
     * 11/20 of the 720-wide frame on a 1080-wide screen is 396,
     * about the squirrel body excluding the tail.
     */
    const val WORDMARK_WIDTH_NUMERATOR = 11f
    const val WORDMARK_WIDTH_DENOMINATOR = 20f

    /** Packed PNG is 941×238. */
    const val WORDMARK_ASPECT_WIDTH = 941f
    const val WORDMARK_ASPECT_HEIGHT = 238f

    /**
     * Inclusive alpha bounds of the packed wordmark. The PNG is 238px tall
     * and the ink, including the faint edge, runs from y=8 through y=229.
     * Slogan size is a fraction of this visual height, not the clear padding.
     */
    const val WORDMARK_INK_TOP = 8f
    const val WORDMARK_INK_BOTTOM = 229f
    const val WORDMARK_INK_HEIGHT = WORDMARK_INK_BOTTOM - WORDMARK_INK_TOP + 1f

    /**
     * Gap from the wordmark box bottom to the slogan ink top, on a 375-wide
     * design. Scaled by screenWidth/375, same as the wordmark gap. 10 sits
     * in the 8–12dp band and lands at about 0.2–0.3 of the wordmark height.
     */
    const val SLOGAN_GAP_AT_375 = 10f

    /** 42% of the wordmark ink height. Inside the 40%–45% band, and lighter than the mark. */
    const val SLOGAN_SIZE_OVER_INK_HEIGHT = 0.42f

    /**
     * Advance, in ems, of the splash slogan in Noto Sans SC Medium.
     * Ten fullwidth glyphs at 1000 units each. The line is capped to the wordmark width.
     */
    const val SLOGAN_ADVANCE_EMS = 10f

    fun frameSize(screenWidth: Float): Pair<Float, Float> {
        val width = screenWidth * WIDTH_FRACTION
        val height = width * HEIGHT_OVER_WIDTH
        return width to height
    }

    /**
     * Wordmark box inside the centered 9:16 frame. [frameWidth] and [frameHeight]
     * are in one unit (dp in Compose, px in the playback view). The result uses
     * that same unit. There is no playback-time input, so the mark does not
     * follow the jump.
     */
    fun wordmarkInFrame(frameWidth: Float, frameHeight: Float): FrameWordmark {
        val (scaleX, scaleY) = containScale(frameWidth, frameHeight, SOURCE_WIDTH, SOURCE_HEIGHT)
        val displayedWidth = frameWidth * scaleX
        val displayedHeight = frameHeight * scaleY
        val videoLeft = (frameWidth - displayedWidth) / 2f
        val videoTop = (frameHeight - displayedHeight) / 2f
        val shadowBottom = videoTop + LANDING_SHADOW_BOTTOM_Y * (displayedHeight / SOURCE_HEIGHT)
        val screenWidth = frameWidth / WIDTH_FRACTION
        val gap = WORDMARK_GAP_AT_375 * (screenWidth / DESIGN_WIDTH)
        val width = displayedWidth * WORDMARK_WIDTH_NUMERATOR / WORDMARK_WIDTH_DENOMINATOR
        val height = width * WORDMARK_ASPECT_HEIGHT / WORDMARK_ASPECT_WIDTH
        val left = videoLeft + (displayedWidth - width) / 2f
        return FrameWordmark(
            left = left,
            top = shadowBottom + gap,
            width = width,
            height = height,
            shadowBottom = shadowBottom,
            gap = gap,
        )
    }

    /**
     * Slogan slot inside the centered 9:16 frame. [top] is the slogan ink top,
     * one fixed gap under the wordmark box. There is no playback-time input,
     * so the line stays put while the squirrel jumps. [width] is the wordmark
     * width; the glyphs are centered in that slot and [textWidth] is not wider.
     */
    fun sloganInFrame(frameWidth: Float, frameHeight: Float): FrameSlogan {
        val mark = wordmarkInFrame(frameWidth, frameHeight)
        val screenWidth = frameWidth / WIDTH_FRACTION
        val gap = SLOGAN_GAP_AT_375 * (screenWidth / DESIGN_WIDTH)
        val inkHeight = mark.height * WORDMARK_INK_HEIGHT / WORDMARK_ASPECT_HEIGHT
        var fontSize = inkHeight * SLOGAN_SIZE_OVER_INK_HEIGHT
        var textWidth = fontSize * SLOGAN_ADVANCE_EMS
        if (textWidth > mark.width && SLOGAN_ADVANCE_EMS > 0f) {
            fontSize = mark.width / SLOGAN_ADVANCE_EMS
            textWidth = mark.width
        }
        return FrameSlogan(
            left = mark.left,
            top = mark.top + mark.height + gap,
            width = mark.width,
            fontSize = fontSize,
            textWidth = textWidth,
            gap = gap,
        )
    }

    /** Screen-space frame, wordmark, and slogan. Same anchors as the frame helpers. */
    fun placement(screenWidth: Float, screenHeight: Float): SplashPlacement {
        val (frameWidth, frameHeight) = frameSize(screenWidth)
        val frameLeft = (screenWidth - frameWidth) / 2f
        val frameTop = (screenHeight - frameHeight) / 2f
        val mark = wordmarkInFrame(frameWidth, frameHeight)
        val slogan = sloganInFrame(frameWidth, frameHeight)
        return SplashPlacement(
            frameLeft = frameLeft,
            frameTop = frameTop,
            frameWidth = frameWidth,
            frameHeight = frameHeight,
            wordmarkLeft = frameLeft + mark.left,
            wordmarkTop = frameTop + mark.top,
            wordmarkWidth = mark.width,
            wordmarkHeight = mark.height,
            shadowBottom = frameTop + mark.shadowBottom,
            wordmarkGap = mark.gap,
            sloganLeft = frameLeft + slogan.left,
            sloganTop = frameTop + slogan.top,
            sloganWidth = slogan.width,
            sloganTextWidth = slogan.textWidth,
            sloganFontSize = slogan.fontSize,
            sloganGap = slogan.gap,
        )
    }
}

internal data class FrameWordmark(
    val left: Float,
    val top: Float,
    val width: Float,
    val height: Float,
    val shadowBottom: Float,
    val gap: Float,
)

internal data class FrameSlogan(
    val left: Float,
    val top: Float,
    val width: Float,
    val fontSize: Float,
    val textWidth: Float,
    val gap: Float,
)

internal data class SplashPlacement(
    val frameLeft: Float,
    val frameTop: Float,
    val frameWidth: Float,
    val frameHeight: Float,
    val wordmarkLeft: Float,
    val wordmarkTop: Float,
    val wordmarkWidth: Float,
    val wordmarkHeight: Float,
    val shadowBottom: Float,
    val wordmarkGap: Float,
    val sloganLeft: Float,
    val sloganTop: Float,
    val sloganWidth: Float,
    val sloganTextWidth: Float,
    val sloganFontSize: Float,
    val sloganGap: Float,
)

/**
 * Extra scale on top of a TextureView's stretch-to-fill, pivoted at the view center.
 * One axis stays at 1 and the other is at most 1, so the clip stays fully inside
 * the view with its aspect ratio intact.
 */
internal fun containScale(
    viewWidth: Float,
    viewHeight: Float,
    videoWidth: Float,
    videoHeight: Float,
): Pair<Float, Float> {
    if (viewWidth <= 0f || viewHeight <= 0f || videoWidth <= 0f || videoHeight <= 0f) {
        return 1f to 1f
    }
    val viewAspect = viewWidth / viewHeight
    val videoAspect = videoWidth / videoHeight
    return if (videoAspect > viewAspect) {
        1f to (viewAspect / videoAspect)
    } else {
        (videoAspect / viewAspect) to 1f
    }
}

internal fun containScale(
    viewWidth: Int,
    viewHeight: Int,
    videoWidth: Int,
    videoHeight: Int,
): Pair<Float, Float> = containScale(
    viewWidth.toFloat(),
    viewHeight.toFloat(),
    videoWidth.toFloat(),
    videoHeight.toFloat(),
)

/**
 * Full-window #F2F5F7 stage. The content slot is the centered 9:16 frame at two
 * thirds of the window width. A transparent layer above it takes taps anywhere.
 */
@Composable
internal fun SplashStage(
    background: Color,
    onTap: () -> Unit,
    tapLabel: String,
    content: @Composable BoxScope.() -> Unit,
) {
    Box(modifier = Modifier.fillMaxSize().background(background)) {
        BoxWithConstraints(
            modifier = Modifier.fillMaxSize(),
            contentAlignment = Alignment.Center,
        ) {
            val (frameWidth, frameHeight) = SplashLayout.frameSize(maxWidth.value)
            Box(
                modifier = Modifier
                    .width(frameWidth.dp)
                    .height(frameHeight.dp),
                content = content,
            )
        }
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Transparent)
                .clickable(
                    interactionSource = remember { MutableInteractionSource() },
                    indication = null,
                    onClickLabel = tapLabel,
                    onClick = onTap,
                ),
        )
    }
}

/**
 * Purple G4 wordmark inside the 9:16 frame. The PNG is already #6C5CE7 and
 * already transparent; this draws it with Fit and does not add a plate or a tint.
 * Position comes only from the frame size, so it stays put while the clip plays.
 */
@Composable
internal fun G4Wordmark() {
    val description = stringResource(R.string.splash_wordmark) +
        stringResource(R.string.splash_wordmark_period)
    BoxWithConstraints(modifier = Modifier.fillMaxSize()) {
        val mark = SplashLayout.wordmarkInFrame(maxWidth.value, maxHeight.value)
        Image(
            painter = painterResource(R.drawable.splash_g4_wordmark),
            contentDescription = description,
            modifier = Modifier
                .absoluteOffset(x = mark.left.dp, y = mark.top.dp)
                .width(mark.width.dp)
                .height(mark.height.dp),
            contentScale = ContentScale.Fit,
        )
    }
}
