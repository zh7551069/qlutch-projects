package dev.qlutch.app.splash

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawing
import androidx.compose.foundation.layout.windowInsetsPadding
import androidx.compose.foundation.text.BasicText
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.colorResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.em
import androidx.compose.ui.unit.sp
import dev.qlutch.app.R
import kotlinx.coroutines.delay

@Composable
internal fun StaticBrandSplash(onContinue: () -> Unit) {
    val canvas = colorResource(R.color.canvas)
    val ink = colorResource(R.color.ink)
    val spark = colorResource(R.color.spark)
    val wordmark = stringResource(R.string.splash_wordmark)
    val period = stringResource(R.string.splash_wordmark_period)
    val slogan = stringResource(R.string.splash_slogan)
    val continueLabel = stringResource(R.string.splash_continue)
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
    val mark = buildAnnotatedString {
        withStyle(
            SpanStyle(
                color = ink,
                fontWeight = FontWeight.Bold,
                letterSpacing = (-0.04).em,
            ),
        ) {
            append(wordmark)
        }
        withStyle(SpanStyle(color = spark, fontWeight = FontWeight.Bold)) {
            append(period)
        }
    }
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(canvas)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClickLabel = continueLabel,
                onClick = ::go,
            )
            .windowInsetsPadding(WindowInsets.safeDrawing)
            .padding(horizontal = 40.dp, vertical = 24.dp),
        contentAlignment = Alignment.Center,
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            BasicText(
                text = mark,
                modifier = Modifier.fillMaxWidth(),
                style = TextStyle(
                    fontFamily = FontFamily.SansSerif,
                    fontSize = 48.sp,
                    lineHeight = 56.sp,
                    textAlign = TextAlign.Center,
                ),
                maxLines = 2,
            )
            Spacer(modifier = Modifier.height(28.dp))
            BasicText(
                text = slogan,
                modifier = Modifier.fillMaxWidth(),
                style = TextStyle(
                    color = ink,
                    fontFamily = FontFamily.SansSerif,
                    fontWeight = FontWeight.Normal,
                    fontSize = 18.sp,
                    lineHeight = 28.sp,
                    textAlign = TextAlign.Center,
                ),
                maxLines = 3,
            )
        }
    }
}
