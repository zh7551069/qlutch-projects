package dev.qlutch.app.splash

import android.content.Context
import android.graphics.Bitmap
import android.graphics.Canvas
import android.graphics.ColorFilter
import android.graphics.Paint
import android.graphics.PixelFormat
import android.graphics.RectF
import android.graphics.drawable.BitmapDrawable
import android.graphics.drawable.Drawable
import android.text.TextPaint
import dev.qlutch.app.R
import java.util.Locale
import kotlin.math.roundToInt

/**
 * Window-sized copy of the in-app first frame: the centered 9:16 still at two
 * thirds of the window width, the G4 wordmark, and the slogan. Drawn from the
 * same assets and the same layout math as playback. Nothing here is baked into
 * the mp4, and nothing is a full-screen 9:16 plate.
 *
 * The system starting window cannot run this class. [Theme.Qlutch.Splash] keeps
 * a plain color there. The activity installs this drawable before the first
 * draw, so the activity window matches the still that then covers the video.
 */
internal class SplashLaunchArtworkDrawable(context: Context) : Drawable() {
    private val resources = context.resources
    private val backgroundColor = context.getColor(R.color.splash_background)
    private val sloganText = context.getString(R.string.splash_slogan)
    private val firstFrame = bitmap(context, R.drawable.splash_first_frame)
    private val wordmark = bitmap(context, R.drawable.splash_g4_wordmark)
    private val bitmapPaint = Paint(Paint.ANTI_ALIAS_FLAG or Paint.FILTER_BITMAP_FLAG)
    private val fillPaint = Paint().apply { color = backgroundColor }
    private val sloganPaint = TextPaint(Paint.ANTI_ALIAS_FLAG).apply {
        density = resources.displayMetrics.density
        isSubpixelText = true
        color = context.getColor(R.color.splash_brand)
        typeface = SplashTypefaces.slogan(context)
        @Suppress("DEPRECATION")
        isElegantTextHeight = false
        letterSpacing = 0f
        textLocale = Locale.SIMPLIFIED_CHINESE
    }

    override fun draw(canvas: Canvas) {
        val bounds = bounds
        if (bounds.isEmpty) return
        canvas.drawRect(bounds, fillPaint)
        val density = resources.displayMetrics.density
        val frame = SplashLaunchGeometry.framePx(bounds.width(), bounds.height(), density)
        if (frame.width <= 0 || frame.height <= 0) return
        val frameLeft = bounds.left + frame.left
        val frameTop = bounds.top + frame.top
        drawFitCenter(canvas, firstFrame, frameLeft, frameTop, frame.width, frame.height)
        val mark = SplashLaunchGeometry.wordmarkInFramePx(frame.width, frame.height)
        drawFitCenter(
            canvas,
            wordmark,
            frameLeft + mark.left,
            frameTop + mark.top,
            mark.width,
            mark.height,
        )
        drawSlogan(canvas, frame, frameLeft, frameTop)
    }

    override fun setAlpha(alpha: Int) = Unit

    override fun setColorFilter(colorFilter: ColorFilter?) = Unit

    @Deprecated("Deprecated in Android")
    override fun getOpacity(): Int = PixelFormat.OPAQUE

    private fun drawSlogan(canvas: Canvas, frame: PxRect, frameLeft: Int, frameTop: Int) {
        if (sloganText.isEmpty()) return
        val placement = sloganDrawPlacement(
            frame.width.toFloat(),
            frame.height.toFloat(),
            measuredTextWidth = { fontSizePx ->
                sloganPaint.textSize = fontSizePx
                sloganPaint.measureText(sloganText)
            },
            inkInsetPx = { fontSizePx ->
                sloganPaint.textSize = fontSizePx
                sloganInkInsetPx(sloganPaint, sloganText)
            },
        )
        sloganPaint.textSize = placement.fontSizePx
        val slotLeft = frameLeft + placement.slotLeft.roundToInt()
        val slotWidth = placement.slotWidth.roundToInt().coerceAtLeast(0)
        val baseline = frameTop + placement.viewTop.roundToInt() + (-sloganPaint.fontMetricsInt.ascent)
        val textWidth = sloganPaint.measureText(sloganText)
        val textLeft = slotLeft + (slotWidth - textWidth) / 2f
        canvas.drawText(sloganText, textLeft, baseline.toFloat(), sloganPaint)
    }

    private fun drawFitCenter(
        canvas: Canvas,
        bitmap: Bitmap?,
        left: Int,
        top: Int,
        width: Int,
        height: Int,
    ) {
        if (bitmap == null || width <= 0 || height <= 0) return
        val (scaleX, scaleY) = containScale(width, height, bitmap.width, bitmap.height)
        val drawnWidth = width * scaleX
        val drawnHeight = height * scaleY
        val dx = left + (width - drawnWidth) / 2f
        val dy = top + (height - drawnHeight) / 2f
        canvas.drawBitmap(
            bitmap,
            null,
            RectF(dx, dy, dx + drawnWidth, dy + drawnHeight),
            bitmapPaint,
        )
    }

    private fun bitmap(context: Context, resId: Int): Bitmap? {
        val drawable = context.getDrawable(resId) as? BitmapDrawable ?: return null
        return drawable.bitmap
    }
}
