package dev.qlutch.app.splash

internal fun scaleIsOff(raw: String?): Boolean {
    if (raw.isNullOrBlank()) return false
    val normalized = raw.trim().removeSuffix("f").removeSuffix("F").replace(',', '.')
    return normalized.toFloatOrNull() == 0f
}

internal fun isReduceMotion(
    animatorScaleOff: Boolean,
    transitionScaleOff: Boolean,
    platformReduceMotion: Boolean,
): Boolean = animatorScaleOff || transitionScaleOff || platformReduceMotion
