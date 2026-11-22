package dev.qlutch.app.splash

import android.content.Context
import android.graphics.Paint
import android.graphics.Rect
import android.graphics.Typeface
import android.os.Build
import android.util.TypedValue
import android.view.Gravity
import android.view.MotionEvent
import android.view.View
import android.widget.FrameLayout
import android.widget.TextView
import dev.qlutch.app.R
import java.io.File
import java.util.Locale
import kotlin.math.roundToInt

internal enum class SloganTypefaceSource {
    PINGFANG_SC_MEDIUM,
    BUNDLED_NOTO_SANS_SC_MEDIUM,
}

/** PingFang SC Medium when the device has it. Otherwise the bundled Noto Sans SC Medium. */
internal fun chooseSloganTypefaceSource(pingFangMediumAvailable: Boolean): SloganTypefaceSource {
    return if (pingFangMediumAvailable) {
        SloganTypefaceSource.PINGFANG_SC_MEDIUM
    } else {
        SloganTypefaceSource.BUNDLED_NOTO_SANS_SC_MEDIUM
    }
}

internal object SplashTypefaces {
    fun slogan(context: Context): Typeface {
        val pingFang = pingFangMediumOrNull()
        return when (chooseSloganTypefaceSource(pingFang != null)) {
            SloganTypefaceSource.PINGFANG_SC_MEDIUM -> checkNotNull(pingFang)
            SloganTypefaceSource.BUNDLED_NOTO_SANS_SC_MEDIUM ->
                context.resources.getFont(R.font.noto_sans_sc_medium)
        }
    }
}

internal fun fittedSloganFontSize(preferred: Float, measuredWidth: Float, maxWidth: Float): Float {
    if (preferred <= 0f || maxWidth <= 0f) return preferred
    if (measuredWidth <= maxWidth || measuredWidth <= 0f) return preferred
    return preferred * maxWidth / measuredWidth
}

/**
 * Space between the text layout top and the slogan ink top, in pixels.
 * Both splash paths shift the line up by this amount so the visible glyphs,
 * not the font's extra ascent, sit on [SplashLayout.sloganInFrame]'s top.
 * [paint] is the slogan view's paint, so the inset matches what that view draws.
 */
internal fun sloganInkInsetPx(paint: Paint, text: String): Float {
    if (text.isEmpty() || paint.textSize <= 0f) return 0f
    val bounds = Rect()
    paint.getTextBounds(text, 0, text.length, bounds)
    if (bounds.top >= 0) return 0f
    return -paint.fontMetrics.ascent + bounds.top
}

internal fun createSloganView(context: Context): TextView {
    return PassThroughTextView(context).apply {
        setText(R.string.splash_slogan)
        setTextColor(context.getColor(R.color.splash_brand))
        typeface = SplashTypefaces.slogan(context)
        @Suppress("DEPRECATION")
        isElegantTextHeight = false
        includeFontPadding = false
        gravity = Gravity.CENTER_HORIZONTAL or Gravity.TOP
        maxLines = 1
        ellipsize = null
        letterSpacing = 0f
        setPadding(0, 0, 0, 0)
        setBackgroundColor(android.graphics.Color.TRANSPARENT)
        textLocale = Locale.SIMPLIFIED_CHINESE
        isClickable = false
        isFocusable = false
        isFocusableInTouchMode = false
        importantForAccessibility = View.IMPORTANT_FOR_ACCESSIBILITY_YES
    }
}

/**
 * Places [view] from [SplashLayout.sloganInFrame]. The slot is the wordmark
 * width, the glyphs are centered in it, and the ink top is the fixed gap
 * under the wordmark. Playback time is not an input.
 */
internal fun layoutSplashSlogan(view: TextView, frameWidth: Int, frameHeight: Int) {
    if (frameWidth <= 0 || frameHeight <= 0) {
        view.layout(0, 0, 0, 0)
        return
    }
    val slogan = SplashLayout.sloganInFrame(frameWidth.toFloat(), frameHeight.toFloat())
    val text = view.text?.toString().orEmpty()
    view.setTextSize(TypedValue.COMPLEX_UNIT_PX, slogan.fontSize)
    val measured = view.paint.measureText(text)
    val fitted = fittedSloganFontSize(slogan.fontSize, measured, slogan.width)
    if (fitted != slogan.fontSize) {
        view.setTextSize(TypedValue.COMPLEX_UNIT_PX, fitted)
    }
    val inset = sloganInkInsetPx(view.paint, text)
    val slotWidth = slogan.width.roundToInt().coerceAtLeast(0)
    val left = slogan.left.roundToInt()
    val top = (slogan.top - inset).roundToInt()
    view.measure(
        View.MeasureSpec.makeMeasureSpec(slotWidth, View.MeasureSpec.EXACTLY),
        View.MeasureSpec.makeMeasureSpec(0, View.MeasureSpec.UNSPECIFIED),
    )
    view.layout(left, top, left + slotWidth, top + view.measuredHeight.coerceAtLeast(0))
}

/** Reduced-motion host. Fills the 9:16 frame and draws only the slogan line. */
internal class SplashSloganHost(
    context: Context,
    private val onLaidOut: (width: Int, height: Int) -> Unit = { _, _ -> },
) : FrameLayout(context) {
    private val slogan = createSloganView(context)

    init {
        importantForAccessibility = View.IMPORTANT_FOR_ACCESSIBILITY_NO
        isClickable = false
        addView(slogan, LayoutParams(LayoutParams.WRAP_CONTENT, LayoutParams.WRAP_CONTENT))
    }

    override fun onLayout(changed: Boolean, left: Int, top: Int, right: Int, bottom: Int) {
        layoutSplashSlogan(slogan, right - left, bottom - top)
        onLaidOut(slogan.width, slogan.height)
    }

    override fun onInterceptTouchEvent(event: MotionEvent): Boolean = false

    override fun onTouchEvent(event: MotionEvent): Boolean = false
}

private class PassThroughTextView(context: Context) : TextView(context) {
    override fun onTouchEvent(event: MotionEvent): Boolean = false
}

private val PINGFANG_MEDIUM_FILES = listOf(
    "/system/fonts/PingFangSC-Medium.otf",
    "/system/fonts/PingFangSC-Medium.ttf",
    "/system/fonts/PingFangSC-Medium.ttc",
    "/system/fonts/PingFang-Medium.otf",
    "/system/fonts/PingFang-Medium.ttf",
    "/system/fonts/PingFang.ttc",
)

private fun pingFangMediumOrNull(): Typeface? {
    for (path in PINGFANG_MEDIUM_FILES) {
        val file = File(path)
        if (!file.isFile) continue
        loadPingFang(file)?.let { return it }
    }
    return pingFangFromSystemFonts()
}

private fun loadPingFang(file: File): Typeface? {
    return try {
        if (file.name.endsWith(".ttc", ignoreCase = true)) {
            Typeface.Builder(file).setWeight(500).build()
        } else {
            Typeface.createFromFile(file)
        }
    } catch (_: RuntimeException) {
        null
    }
}

private fun pingFangFromSystemFonts(): Typeface? {
    if (Build.VERSION.SDK_INT < 29) return null
    return try {
        val matches = android.graphics.fonts.SystemFonts.getAvailableFonts().mapNotNull { font ->
            val file = font.file ?: return@mapNotNull null
            if (!file.name.contains("PingFang", ignoreCase = true)) return@mapNotNull null
            font
        }
        val medium = matches.minByOrNull { kotlin.math.abs(it.style.weight - 500) } ?: return null
        val file = medium.file ?: return null
        Typeface.Builder(file).setTtcIndex(medium.ttcIndex).setWeight(500).build()
    } catch (_: RuntimeException) {
        null
    }
}
