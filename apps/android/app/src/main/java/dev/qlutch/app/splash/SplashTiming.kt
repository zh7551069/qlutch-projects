package dev.qlutch.app.splash

internal object SplashTiming {
    /** Static brand hold when motion is reduced. Inside the 1.2–2.0s entrance budget. */
    const val STATIC_HOLD_MILLIS = 1_600L

    /**
     * Last-frame hold after the welcome clip reaches its natural end.
     * Tap and back still leave immediately. The reduced-motion path does not use this.
     */
    const val POST_COMPLETION_HOLD_MILLIS = 1_000L
}

/**
 * After the welcome clip finishes, wait [SplashTiming.POST_COMPLETION_HOLD_MILLIS]
 * before leaving. A skip during playback or during that hold leaves at once and
 * drops the pending leave, so the clip is not played again.
 */
internal class SplashCompletionHold(
    private val schedule: (delayMillis: Long, action: () -> Unit) -> Cancellable,
    private val holdMillis: Long = SplashTiming.POST_COMPLETION_HOLD_MILLIS,
) {
    private var pending: Cancellable? = null
    private var left = false

    fun onNaturalCompletion(leave: () -> Unit) {
        if (left || pending != null) return
        pending = schedule(holdMillis) {
            pending = null
            leaveNow(leave)
        }
    }

    fun onSkip(leave: () -> Unit) {
        pending?.cancel()
        pending = null
        leaveNow(leave)
    }

    /** Drops a pending last-frame hold without leaving. Used when the player is released. */
    fun cancelPending() {
        pending?.cancel()
        pending = null
    }

    private fun leaveNow(leave: () -> Unit) {
        if (left) return
        left = true
        leave()
    }

    fun interface Cancellable {
        fun cancel()
    }
}
