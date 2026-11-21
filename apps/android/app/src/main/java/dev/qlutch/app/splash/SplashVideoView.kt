package dev.qlutch.app.splash

import android.content.Context
import android.graphics.Matrix
import android.graphics.SurfaceTexture
import android.media.MediaPlayer
import android.os.Handler
import android.os.Looper
import android.view.Surface
import android.view.TextureView
import android.view.View
import dev.qlutch.app.R

/**
 * Plays the silent cold-start video, center-cropped to the view.
 * Completion and decode failure call [onFinished]. Taps are handled by the overlay.
 */
internal class SplashVideoView(context: Context) : TextureView(context), TextureView.SurfaceTextureListener {
    var onFinished: (() -> Unit)? = null

    private val mainHandler = Handler(Looper.getMainLooper())
    private val player = MediaPlayer()
    private var surface: Surface? = null
    private var prepared = false
    private var finished = false
    private var released = false
    private var started = false
    private var inForeground = true

    init {
        isOpaque = false
        keepScreenOn = true
        importantForAccessibility = View.IMPORTANT_FOR_ACCESSIBILITY_NO
        surfaceTextureListener = this
        player.setOnPreparedListener { mp ->
            if (released || finished) return@setOnPreparedListener
            prepared = true
            mp.setVolume(0f, 0f)
            applyCover()
            if (surface != null && inForeground) {
                mp.start()
            }
        }
        player.setOnVideoSizeChangedListener { _, _, _ -> applyCover() }
        player.setOnCompletionListener { dispatchFinished() }
        player.setOnErrorListener { _, _, _ ->
            dispatchFinished()
            true
        }
    }

    fun startPlayback() {
        if (started || released || finished) return
        started = true
        try {
            val afd = resources.openRawResourceFd(R.raw.splash)
            try {
                player.setDataSource(afd.fileDescriptor, afd.startOffset, afd.length)
            } finally {
                afd.close()
            }
            player.isLooping = false
            player.prepareAsync()
        } catch (_: Exception) {
            dispatchFinished()
        }
    }

    fun pausePlayback() {
        inForeground = false
        if (released || !prepared || finished) return
        try {
            if (player.isPlaying) player.pause()
        } catch (_: IllegalStateException) {
            dispatchFinished()
        }
    }

    fun resumePlayback() {
        inForeground = true
        if (released || !prepared || finished || surface == null) return
        try {
            if (!player.isPlaying) {
                player.setVolume(0f, 0f)
                player.start()
            }
        } catch (_: IllegalStateException) {
            dispatchFinished()
        }
    }

    fun releasePlayer() {
        if (released) return
        released = true
        player.setOnPreparedListener(null)
        player.setOnCompletionListener(null)
        player.setOnErrorListener(null)
        player.setOnVideoSizeChangedListener(null)
        try {
            player.setSurface(null)
        } catch (_: Exception) {
        }
        surface?.release()
        surface = null
        try {
            player.release()
        } catch (_: Exception) {
        }
    }

    override fun onDetachedFromWindow() {
        releasePlayer()
        super.onDetachedFromWindow()
    }

    override fun onSurfaceTextureAvailable(surfaceTexture: SurfaceTexture, width: Int, height: Int) {
        if (released) return
        surface?.release()
        val next = Surface(surfaceTexture)
        surface = next
        try {
            player.setSurface(next)
        } catch (_: Exception) {
            dispatchFinished()
            return
        }
        applyCover()
        if (prepared && !finished && inForeground) {
            try {
                player.setVolume(0f, 0f)
                player.start()
            } catch (_: IllegalStateException) {
                dispatchFinished()
            }
        }
    }

    override fun onSurfaceTextureSizeChanged(surfaceTexture: SurfaceTexture, width: Int, height: Int) {
        applyCover()
    }

    override fun onSurfaceTextureDestroyed(surfaceTexture: SurfaceTexture): Boolean {
        try {
            if (!released) player.setSurface(null)
        } catch (_: Exception) {
        }
        surface?.release()
        surface = null
        return true
    }

    override fun onSurfaceTextureUpdated(surfaceTexture: SurfaceTexture) = Unit

    override fun onSizeChanged(w: Int, h: Int, oldw: Int, oldh: Int) {
        super.onSizeChanged(w, h, oldw, oldh)
        applyCover()
    }

    private fun applyCover() {
        if (released) return
        val videoWidth: Int
        val videoHeight: Int
        try {
            videoWidth = player.videoWidth
            videoHeight = player.videoHeight
        } catch (_: IllegalStateException) {
            return
        }
        val (scaleX, scaleY) = coverScale(width, height, videoWidth, videoHeight)
        val matrix = Matrix()
        matrix.setScale(scaleX, scaleY, width / 2f, height / 2f)
        setTransform(matrix)
    }

    private fun dispatchFinished() {
        if (finished || released) return
        finished = true
        val callback = onFinished
        mainHandler.post { callback?.invoke() }
    }
}
