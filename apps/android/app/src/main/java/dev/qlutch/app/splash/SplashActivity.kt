package dev.qlutch.app.splash

import android.content.Intent
import android.os.Build
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.OnBackPressedCallback
import androidx.activity.SystemBarStyle
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import dev.qlutch.app.MainActivity
import java.util.concurrent.atomic.AtomicBoolean

class SplashActivity : ComponentActivity() {
    private val handedOff = AtomicBoolean(false)
    private var reduceMotion = false

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        reduceMotion = ReduceMotion.isEnabled(this)
        onBackPressedDispatcher.addCallback(
            this,
            object : OnBackPressedCallback(true) {
                override fun handleOnBackPressed() {
                    openNext()
                }
            },
        )
        enableEdgeToEdge(
            statusBarStyle = SystemBarStyle.light(
                scrim = android.graphics.Color.TRANSPARENT,
                darkScrim = android.graphics.Color.TRANSPARENT,
            ),
            navigationBarStyle = SystemBarStyle.light(
                scrim = android.graphics.Color.TRANSPARENT,
                darkScrim = android.graphics.Color.TRANSPARENT,
            ),
        )
        if (Build.VERSION.SDK_INT >= 29) {
            window.isNavigationBarContrastEnforced = false
        }
        setContent {
            if (reduceMotion) {
                StaticWelcomeStill(onContinue = ::openNext)
            } else {
                ColdStartVideo(onFinished = ::openNext)
            }
        }
        if (!reduceMotion) hideSystemBars()
    }

    override fun onResume() {
        super.onResume()
        if (!reduceMotion) hideSystemBars()
    }

    private fun hideSystemBars() {
        val controller = WindowCompat.getInsetsController(window, window.decorView)
        controller.systemBarsBehavior = WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
        controller.hide(WindowInsetsCompat.Type.systemBars())
    }

    private fun openNext() {
        if (isFinishing || isDestroyed) return
        if (!handedOff.compareAndSet(false, true)) return
        val next = Intent(this, MainActivity::class.java).addFlags(Intent.FLAG_ACTIVITY_NO_ANIMATION)
        if (Build.VERSION.SDK_INT >= 34) {
            overrideActivityTransition(OVERRIDE_TRANSITION_OPEN, 0, 0)
            overrideActivityTransition(OVERRIDE_TRANSITION_CLOSE, 0, 0)
        }
        startActivity(next)
        finish()
        if (Build.VERSION.SDK_INT < 34) {
            @Suppress("DEPRECATION")
            overridePendingTransition(0, 0)
        }
    }
}
