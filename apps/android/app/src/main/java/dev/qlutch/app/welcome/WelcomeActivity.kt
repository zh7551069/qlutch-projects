package dev.qlutch.app.welcome

import android.os.Build
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.SystemBarStyle
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import dev.qlutch.app.flow.FlowActivity
import dev.qlutch.app.flow.FlowLogic

/** P02 welcome. No session check: splash always lands here, and the buttons stay. */
class WelcomeActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
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
            WelcomeScreen(
                onStart = {
                    startActivity(FlowActivity.intent(this, FlowLogic.PURPOSE_REGISTER))
                },
                onExistingAccount = {
                    startActivity(FlowActivity.intent(this, FlowLogic.PURPOSE_LOGIN))
                },
            )
        }
    }
}
