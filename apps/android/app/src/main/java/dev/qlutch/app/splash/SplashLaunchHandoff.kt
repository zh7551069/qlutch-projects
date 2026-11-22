package dev.qlutch.app.splash

import android.view.View
import android.view.ViewTreeObserver
import androidx.compose.runtime.staticCompositionLocalOf

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
