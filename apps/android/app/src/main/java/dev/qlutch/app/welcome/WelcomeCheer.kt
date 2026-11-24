package dev.qlutch.app.welcome

import android.graphics.BitmapFactory
import android.os.SystemClock
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive

/**
 * Plays the 50 slot frames at 40ms and starts again at frame 1 on the 2.00s
 * boundary. The clock is elapsed time, so a late wake does not stretch the
 * cycle. The frames are already the 716×480 slot crop; Fit keeps the
 * character and the contact shadow inside the slot.
 * The offset below nudges the character to where the eye reads it as centred
 * and places it closer to the wordmark; see SLOT_CONTENT_SHIFT_RIGHT_DP and
 * SLOT_CONTENT_SHIFT_DOWN_DP.
 */
@Composable
internal fun WelcomeCheerSlot(modifier: Modifier = Modifier) {
    val context = LocalContext.current
    val frames = remember {
        WelcomeCheerFrames.ids.map { id ->
            val bitmap = BitmapFactory.decodeResource(context.resources, id)
            checkNotNull(bitmap) { "missing cheer frame $id" }.asImageBitmap()
        }
    }
    var index by remember { mutableIntStateOf(welcomeCheerFrameIndex(0L)) }
    LaunchedEffect(Unit) {
        val start = SystemClock.elapsedRealtime()
        while (isActive) {
            val elapsed = SystemClock.elapsedRealtime() - start
            val next = welcomeCheerFrameIndex(elapsed)
            if (next != index) index = next
            delay(welcomeCheerDelayUntilNextBoundary(elapsed))
        }
    }
    Box(
        modifier = modifier
            .clip(RoundedCornerShape(WelcomeLayout.SLOT_RADIUS_DP.dp))
            .background(Color(WelcomeLayout.SLOT_COLOR)),
        contentAlignment = Alignment.Center,
    ) {
        Image(
            bitmap = frames[index],
            contentDescription = null,
            modifier = Modifier
                .fillMaxSize()
                .offset(
                    x = WelcomeLayout.SLOT_CONTENT_SHIFT_RIGHT_DP.dp,
                    y = WelcomeLayout.SLOT_CONTENT_SHIFT_DOWN_DP.dp,
                ),
            contentScale = ContentScale.Fit,
            alignment = Alignment.Center,
        )
    }
}
