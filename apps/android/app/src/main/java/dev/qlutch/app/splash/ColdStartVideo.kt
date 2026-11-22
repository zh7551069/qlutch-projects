package dev.qlutch.app.splash

import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.colorResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.viewinterop.AndroidView
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import dev.qlutch.app.R

@Composable
internal fun ColdStartVideo(onFinished: () -> Unit) {
    val background = colorResource(R.color.splash_background)
    val skipLabel = stringResource(R.string.splash_skip)
    val launchHandoff = LocalSplashLaunchHandoff.current
    val videoView = remember { mutableStateOf<SplashVideoView?>(null) }
    val lifecycleOwner = LocalLifecycleOwner.current
    DisposableEffect(lifecycleOwner) {
        val observer = LifecycleEventObserver { _, event ->
            val view = videoView.value ?: return@LifecycleEventObserver
            when (event) {
                Lifecycle.Event.ON_STOP -> view.pausePlayback()
                Lifecycle.Event.ON_START -> view.resumePlayback()
                else -> Unit
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose { lifecycleOwner.lifecycle.removeObserver(observer) }
    }
    SplashStage(
        background = background,
        onTap = { videoView.value?.skip() ?: onFinished() },
        tapLabel = skipLabel,
    ) {
        AndroidView(
            modifier = Modifier.fillMaxSize(),
            factory = { context ->
                SplashPlaybackView(context).apply {
                    this.launchHandoff = launchHandoff
                    video.onFinished = onFinished
                    video.startPlayback()
                    videoView.value = video
                }
            },
            onRelease = { it.video.releasePlayer() },
        )
    }
}
