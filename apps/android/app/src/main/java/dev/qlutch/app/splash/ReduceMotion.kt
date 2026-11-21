package dev.qlutch.app.splash

import android.content.Context
import android.provider.Settings

internal object ReduceMotion {
    fun isEnabled(context: Context): Boolean = runCatching {
        val resolver = context.contentResolver
        val animator = Settings.Global.getString(resolver, Settings.Global.ANIMATOR_DURATION_SCALE)
        val transition = Settings.Global.getString(resolver, Settings.Global.TRANSITION_ANIMATION_SCALE)
        val platform = Settings.Global.getInt(resolver, PLATFORM_REDUCE_MOTION, 0) == 1 ||
            Settings.Secure.getInt(resolver, PLATFORM_REDUCE_MOTION, 0) == 1
        isReduceMotion(scaleIsOff(animator), scaleIsOff(transition), platform)
    }.getOrDefault(false)

    private const val PLATFORM_REDUCE_MOTION = "reduce_motion"
}
