package dev.qlutch.app.splash

import android.content.Intent
import android.os.Build
import android.os.Bundle
import android.view.View
import androidx.activity.ComponentActivity
import androidx.activity.OnBackPressedCallback
import androidx.activity.SystemBarStyle
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.runtime.CompositionLocalProvider
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import dev.qlutch.app.welcome.WelcomeActivity
import java.util.concurrent.atomic.AtomicBoolean

class SplashActivity : ComponentActivity() {
    private val handedOff = AtomicBoolean(false)
    private val launchHandoff = SplashLaunchHandoff()
    private var reduceMotion = false

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        // Android 12+ has already painted a solid plate. Drop it with no scale
        // or fade once our first frame is allowed, so the plate does not zoom
        // into the still.
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            splashScreen.setOnExitAnimationListener { splashScreenView ->
                splashScreenView.remove()
            }
        }
        // Same still, wordmark, and slogan as playback, in this window's pixels.
        // Installed before the first draw. The theme background stays a flat
        // color because the system process cannot run this drawable.
        window.setBackgroundDrawable(SplashLaunchArtworkDrawable(this))
        // Before setContent, so the first traversal cannot paint an empty frame
        // under the system splash. The splash stays until the still is laid out.
        findViewById<View>(android.R.id.content).holdSplashUntilContentReady {
            launchHandoff.contentReady
        }
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
            CompositionLocalProvider(LocalSplashLaunchHandoff provides launchHandoff) {
                if (reduceMotion) {
                    StaticWelcomeStill(onContinue = ::openNext)
                } else {
                    ColdStartVideo(onFinished = ::openNext)
                }
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
        val next = Intent(this, WelcomeActivity::class.java).addFlags(Intent.FLAG_ACTIVITY_NO_ANIMATION)
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
