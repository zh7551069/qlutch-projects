package dev.qlutch.app.splash

import android.content.Context
import android.graphics.Matrix
import android.graphics.SurfaceTexture
import android.media.MediaPlayer
import android.os.Handler
import android.os.Looper
import android.view.MotionEvent
import android.view.Surface
import android.view.TextureView
import android.view.View
import android.widget.FrameLayout
import android.widget.ImageView
import android.widget.TextView
import dev.qlutch.app.R
import kotlin.math.roundToInt

/**
 * Plays the silent welcome clip fitted inside this view.
 * The view is the centered 9:16 frame, and [containScale] keeps the whole picture
 * visible. Natural completion keeps the last frame for
 * [SplashTiming.POST_COMPLETION_HOLD_MILLIS], then calls [onFinished].
 * Decode failure calls [onFinished] immediately. [skip] leaves immediately,
 * including during that hold, and does not replay the clip.
 * Taps are handled by the overlay.
 */
internal class SplashVideoView(context: Context) : TextureView(context), TextureView.SurfaceTextureListener {
    var onFinished: (() -> Unit)? = null

    private val mainHandler = Handler(Looper.getMainLooper())
    private val completionHold = SplashCompletionHold({ delayMillis, action ->
        val runnable = Runnable { action() }
        mainHandler.postDelayed(runnable, delayMillis)
        SplashCompletionHold.Cancellable { mainHandler.removeCallbacks(runnable) }
    })
    private val player = MediaPlayer()
    private var surface: Surface? = null
    private var prepared = false
    private var finished = false
    private var navigated = false
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
            applyContain()
            if (surface != null && inForeground) {
                mp.start()
            }
        }
        player.setOnVideoSizeChangedListener { _, _, _ -> applyContain() }
        // Natural end only. Looping stays off, so the last frame remains for the hold.
        // Do not start the player again from here.
        player.setOnCompletionListener {
            if (released || finished || navigated) return@setOnCompletionListener
            finished = true
            val callback = onFinished
            completionHold.onNaturalCompletion { deliverNow(callback) }
        }
        player.setOnErrorListener { _, _, _ ->
            dispatchFinished()
            true
        }
    }

    /** Immediate leave for a tap during playback or during the last-frame hold. */
    fun skip() {
        if (navigated) return
        finished = true
        val callback = onFinished
        completionHold.onSkip { deliverNow(callback) }
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
        completionHold.cancelPending()
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
        applyContain()
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
        applyContain()
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
        applyContain()
    }

    private fun applyContain() {
        if (released) return
        val videoWidth: Int
        val videoHeight: Int
        try {
            videoWidth = player.videoWidth
            videoHeight = player.videoHeight
        } catch (_: IllegalStateException) {
            return
        }
        val (scaleX, scaleY) = containScale(width, height, videoWidth, videoHeight)
        val matrix = Matrix()
        matrix.setScale(scaleX, scaleY, width / 2f, height / 2f)
        setTransform(matrix)
    }

    private fun deliverNow(callback: (() -> Unit)?) {
        if (navigated) return
        navigated = true
        callback?.invoke()
    }

    private fun dispatchFinished() {
        if (navigated || released) return
        finished = true
        completionHold.cancelPending()
        val callback = onFinished
        mainHandler.post { deliverNow(callback) }
    }
}

/**
 * Centered welcome clip with the G4 wordmark and the slogan drawn after the texture.
 * Both are laid out from the landing shadow, not from the player position,
 * so they stay put when the squirrel jumps. Neither is part of the mp4.
 */
internal class SplashPlaybackView(context: Context) : FrameLayout(context) {
    val video: SplashVideoView = SplashVideoView(context)
    private val wordmark: ImageView = PassThroughImageView(context).apply {
        setImageResource(R.drawable.splash_g4_wordmark)
        scaleType = ImageView.ScaleType.FIT_CENTER
        imageTintList = null
        setBackgroundColor(android.graphics.Color.TRANSPARENT)
        contentDescription = context.getString(R.string.splash_wordmark) +
            context.getString(R.string.splash_wordmark_period)
        importantForAccessibility = View.IMPORTANT_FOR_ACCESSIBILITY_YES
        isClickable = false
        isFocusable = false
        isFocusableInTouchMode = false
    }
    private val slogan: TextView = createSloganView(context)

    init {
        importantForAccessibility = View.IMPORTANT_FOR_ACCESSIBILITY_NO
        addView(
            video,
            LayoutParams(LayoutParams.MATCH_PARENT, LayoutParams.MATCH_PARENT),
        )
        addView(
            wordmark,
            LayoutParams(LayoutParams.WRAP_CONTENT, LayoutParams.WRAP_CONTENT),
        )
        addView(
            slogan,
            LayoutParams(LayoutParams.WRAP_CONTENT, LayoutParams.WRAP_CONTENT),
        )
    }

    override fun onLayout(changed: Boolean, left: Int, top: Int, right: Int, bottom: Int) {
        val frameWidth = right - left
        val frameHeight = bottom - top
        video.layout(0, 0, frameWidth, frameHeight)
        if (frameWidth <= 0 || frameHeight <= 0) {
            wordmark.layout(0, 0, 0, 0)
            slogan.layout(0, 0, 0, 0)
            return
        }
        val mark = SplashLayout.wordmarkInFrame(frameWidth.toFloat(), frameHeight.toFloat())
        val markWidth = mark.width.roundToInt().coerceAtLeast(0)
        val markHeight = mark.height.roundToInt().coerceAtLeast(0)
        val markLeft = mark.left.roundToInt()
        val markTop = mark.top.roundToInt()
        wordmark.measure(
            View.MeasureSpec.makeMeasureSpec(markWidth, View.MeasureSpec.EXACTLY),
            View.MeasureSpec.makeMeasureSpec(markHeight, View.MeasureSpec.EXACTLY),
        )
        wordmark.layout(markLeft, markTop, markLeft + markWidth, markTop + markHeight)
        layoutSplashSlogan(slogan, frameWidth, frameHeight)
    }
}

private class PassThroughImageView(context: Context) : ImageView(context) {
    override fun onTouchEvent(event: MotionEvent): Boolean = false
}
