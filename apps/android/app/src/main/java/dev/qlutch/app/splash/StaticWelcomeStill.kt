package dev.qlutch.app.splash

import android.content.Context
import android.graphics.Bitmap
import android.media.MediaMetadataRetriever
import androidx.compose.foundation.Image
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.colorResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.viewinterop.AndroidView
import dev.qlutch.app.R
import kotlinx.coroutines.delay

/**
 * Reduce-motion hold. Shows the first frame of the packed welcome clip inside the
 * same centered frame as playback, with the same fixed purple wordmark and
 * slogan, then continues after [SplashTiming.STATIC_HOLD_MILLIS] or on tap.
 * Does not play the clip.
 */
@Composable
internal fun StaticWelcomeStill(onContinue: () -> Unit) {
    val background = colorResource(R.color.splash_background)
    val continueLabel = stringResource(R.string.splash_continue)
    val context = LocalContext.current
    val handedOff = remember { mutableStateOf(false) }
    val still = remember(context) { decodeSplashStill(context)?.asImageBitmap() }
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
        if (still != null) {
            Image(
                bitmap = still,
                contentDescription = null,
                modifier = Modifier.fillMaxSize(),
                contentScale = ContentScale.Fit,
            )
        }
        G4Wordmark()
        AndroidView(
            modifier = Modifier.fillMaxSize(),
            factory = { context -> SplashSloganHost(context) },
        )
    }
}

internal fun decodeSplashStill(context: Context): Bitmap? {
    val retriever = MediaMetadataRetriever()
    try {
        val afd = context.resources.openRawResourceFd(R.raw.splash)
        try {
            retriever.setDataSource(afd.fileDescriptor, afd.startOffset, afd.length)
        } finally {
            afd.close()
        }
        val frame = try {
            retriever.getFrameAtIndex(0)
        } catch (_: Exception) {
            null
        } ?: try {
            retriever.getFrameAtTime(0, MediaMetadataRetriever.OPTION_CLOSEST)
        } catch (_: Exception) {
            null
        }
        return frame?.let { softwareBitmap(it) }
    } catch (_: Exception) {
        return null
    } finally {
        try {
            retriever.release()
        } catch (_: Exception) {
        }
    }
}

private fun softwareBitmap(frame: Bitmap): Bitmap {
    if (frame.config == Bitmap.Config.ARGB_8888) return frame
    return frame.copy(Bitmap.Config.ARGB_8888, false) ?: frame
}
