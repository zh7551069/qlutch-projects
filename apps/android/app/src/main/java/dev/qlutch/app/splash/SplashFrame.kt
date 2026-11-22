package dev.qlutch.app.splash

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxScope
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.width
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp

internal object SplashLayout {
    /** Welcome clip width as a fraction of the splash window width. */
    const val WIDTH_FRACTION = 1f / 3f

    /** 9:16 frame. Height = width * 16/9. */
    const val HEIGHT_OVER_WIDTH = 16f / 9f

    fun frameSize(screenWidth: Float): Pair<Float, Float> {
        val width = screenWidth * WIDTH_FRACTION
        val height = width * HEIGHT_OVER_WIDTH
        return width to height
    }
}

/**
 * Extra scale on top of a TextureView's stretch-to-fill, pivoted at the view center.
 * One axis stays at 1 and the other is at most 1, so the clip stays fully inside
 * the view with its aspect ratio intact.
 */
internal fun containScale(
    viewWidth: Int,
    viewHeight: Int,
    videoWidth: Int,
    videoHeight: Int,
): Pair<Float, Float> {
    if (viewWidth <= 0 || viewHeight <= 0 || videoWidth <= 0 || videoHeight <= 0) {
        return 1f to 1f
    }
    val viewAspect = viewWidth.toFloat() / viewHeight.toFloat()
    val videoAspect = videoWidth.toFloat() / videoHeight.toFloat()
    return if (videoAspect > viewAspect) {
        1f to (viewAspect / videoAspect)
    } else {
        (videoAspect / viewAspect) to 1f
    }
}

/**
 * Full-window #F2F5F7 stage. The content slot is the centered 9:16 frame at one
 * third of the window width. A transparent layer above it takes taps anywhere.
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
