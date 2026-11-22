package dev.qlutch.app.splash

import android.view.View
import android.view.ViewTreeObserver
import androidx.compose.runtime.staticCompositionLocalOf
import kotlin.math.sqrt

/**
 * Square crop of the welcome clip's first frame used as the Android 12+ splash icon.
 *
 * The system draws that icon through the same inset as an adaptive icon: a circle
 * whose diameter is 2/3 of the bitmap, and OEM masks can trim that circle to the
 * 66/108 safe zone. [scale] fits the whole squirrel, tail included, inside that
 * circle. The 9:16 frame itself is not the icon.
 */
internal object SplashSystemIcon {
    const val CANVAS_PX = 1080
    const val SAFE_ZONE_NUMERATOR = 66
    const val SAFE_ZONE_DENOMINATOR = 108

    /** Inclusive bounds of the non-background pixels in splash_first_frame.png. */
    const val CONTENT_LEFT = 96
    const val CONTENT_TOP = 498
    const val CONTENT_RIGHT = 855
    const val CONTENT_BOTTOM = 1449

    /** Content pixel farthest from the content-box center. */
    const val FARTHEST_X = 791
    const val FARTHEST_Y = 1383

    fun contentCenterX(): Float = (CONTENT_LEFT + CONTENT_RIGHT) / 2f

    fun contentCenterY(): Float = (CONTENT_TOP + CONTENT_BOTTOM) / 2f

    fun safeRadiusPx(): Float =
        CANVAS_PX * SAFE_ZONE_NUMERATOR / (2f * SAFE_ZONE_DENOMINATOR)

    fun contentRadiusPx(): Float {
        val dx = FARTHEST_X - contentCenterX()
        val dy = FARTHEST_Y - contentCenterY()
        return sqrt(dx * dx + dy * dy)
    }

    /** Icon pixels per source pixel. Below 1, so the squirrel is scaled down, not stretched. */
    fun scale(): Float = safeRadiusPx() / contentRadiusPx()
}

/**
 * Latches once the in-app still, the G4 wordmark, and the slogan have each been
 * given a non-zero size. Layout is the last thing that happens before the draw
 * that paints them, so this is the signal to let that draw through and take the
 * system splash down with it. A later zero-size pass does not clear the latch.
 */
internal class SplashLaunchHandoff {
    private var stillReady = false
    private var wordmarkReady = false
    private var sloganReady = false

    val contentReady: Boolean
        get() = stillReady && wordmarkReady && sloganReady

    fun onStillLaidOut(width: Int, height: Int) {
        if (width > 0 && height > 0) stillReady = true
    }

    fun onWordmarkLaidOut(width: Int, height: Int) {
        if (width > 0 && height > 0) wordmarkReady = true
    }

    fun onSloganLaidOut(width: Int, height: Int) {
        if (width > 0 && height > 0) sloganReady = true
    }
}

internal val LocalSplashLaunchHandoff = staticCompositionLocalOf { SplashLaunchHandoff() }

/**
 * Cancels window draws until [contentReady] is true, which keeps the starting
 * window (the system splash on Android 12+, the themed window background before
 * that) on screen. The first draw that is allowed is the one that paints the
 * in-app still, so the splash does not come down onto an empty frame.
 * Once released, later frames are not blocked again.
 */
internal class SplashDrawGate(
    private val contentReady: () -> Boolean,
) {
    var released: Boolean = false
        private set

    fun onPreDraw(): Boolean {
        if (released) return true
        if (!contentReady()) return false
        released = true
        return true
    }
}

internal fun View.holdSplashUntilContentReady(contentReady: () -> Boolean) {
    val gate = SplashDrawGate(contentReady)
    val listener = object : ViewTreeObserver.OnPreDrawListener {
        override fun onPreDraw(): Boolean {
            val proceed = gate.onPreDraw()
            if (proceed) {
                viewTreeObserver.removeOnPreDrawListener(this)
            }
            return proceed
        }
    }
    viewTreeObserver.addOnPreDrawListener(listener)
}
