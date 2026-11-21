package dev.qlutch.app.splash

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.colorResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.viewinterop.AndroidView
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import dev.qlutch.app.R

@Composable
internal fun ColdStartVideo(onFinished: () -> Unit) {
    val canvas = colorResource(R.color.canvas)
    val skipLabel = stringResource(R.string.splash_skip)
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
    Box(modifier = Modifier.fillMaxSize().background(canvas)) {
        AndroidView(
            modifier = Modifier.fillMaxSize(),
            factory = { context ->
                SplashVideoView(context).apply {
                    this.onFinished = onFinished
                    startPlayback()
                    videoView.value = this
                }
            },
            onRelease = { it.releasePlayer() },
        )
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Transparent)
                .clickable(
                    interactionSource = remember { MutableInteractionSource() },
                    indication = null,
                    onClickLabel = skipLabel,
                    onClick = onFinished,
                ),
        )
    }
}
