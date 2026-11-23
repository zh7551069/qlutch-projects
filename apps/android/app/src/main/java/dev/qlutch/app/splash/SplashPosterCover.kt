package dev.qlutch.app.splash

/**
 * The packed welcome-clip frame stays on screen until the player has actually
 * painted a frame. Preparing the player or attaching a surface still leaves a
 * blank plate, so those events do not retire the still.
 */
internal class SplashPosterCover {
    private var videoFrameOnScreen = false

    val showingStill: Boolean get() = !videoFrameOnScreen

    fun onPlayerPrepared() = Unit

    fun onSurfaceAttached() = Unit

    fun onFirstFrameRendered() {
        videoFrameOnScreen = true
    }
}
