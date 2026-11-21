package dev.qlutch.app.splash

/**
 * Extra scale applied on top of a TextureView's default stretch-to-fill.
 * The pivot is the view center, so the result is center-crop: the video
 * covers the view and keeps its aspect ratio.
 */
internal fun coverScale(
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
        (videoAspect / viewAspect) to 1f
    } else {
        1f to (viewAspect / videoAspect)
    }
}
