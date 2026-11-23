package dev.qlutch.app.splash

import androidx.compose.foundation.Image
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.res.colorResource
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.viewinterop.AndroidView
import dev.qlutch.app.R
import kotlinx.coroutines.delay

/**
 * Reduce-motion hold. Shows the packed first frame of the welcome clip inside the
 * same centered frame as playback, with the same fixed purple wordmark and
 * slogan, then continues after [SplashTiming.STATIC_HOLD_MILLIS] or on tap.
 * Does not play the clip.
 */
@Composable
internal fun StaticWelcomeStill(onContinue: () -> Unit) {
    val background = colorResource(R.color.splash_background)
    val continueLabel = stringResource(R.string.splash_continue)
    val launchHandoff = LocalSplashLaunchHandoff.current
    val handedOff = remember { mutableStateOf(false) }
    fun go() {
        if (handedOff.value) return
        handedOff.value = true
        onContinue()
    }
    LaunchedEffect(Unit) {
        delay(SplashTiming.STATIC_HOLD_MILLIS)
        go()
    }
    SplashStage(background = background, onTap = ::go, tapLabel = continueLabel) {
        Image(
            painter = painterResource(R.drawable.splash_first_frame),
            contentDescription = null,
            modifier = Modifier.fillMaxSize().onSizeChanged { size ->
                launchHandoff.onStillLaidOut(size.width, size.height)
            },
            contentScale = ContentScale.Fit,
        )
        G4Wordmark()
        AndroidView(
            modifier = Modifier.fillMaxSize(),
            factory = { context ->
                SplashSloganHost(context) { width, height ->
                    launchHandoff.onSloganLaidOut(width, height)
                }
            },
        )
    }
}
