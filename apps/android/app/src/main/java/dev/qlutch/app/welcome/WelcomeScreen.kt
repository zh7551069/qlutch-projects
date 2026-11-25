package dev.qlutch.app.welcome

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawing
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.layout.windowInsetsPadding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.clipToBounds
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.PlatformTextStyle
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import dev.qlutch.app.R
import dev.qlutch.app.ui.QlutchButton

@Composable
internal fun WelcomeScreen(
    onStart: () -> Unit = {},
    onExistingAccount: () -> Unit = {},
) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(WelcomeLayout.CANVAS)),
        contentAlignment = Alignment.TopCenter,
    ) {
        Column(
            modifier = Modifier
                .widthIn(max = WelcomeLayout.ARTBOARD_WIDTH_DP.dp)
                .fillMaxWidth()
                .fillMaxHeight()
                .windowInsetsPadding(WindowInsets.safeDrawing)
                .clipToBounds(),
        ) {
            Column(
                modifier = Modifier
                    .weight(1f)
                    .fillMaxWidth()
                    .padding(
                        start = WelcomeLayout.CONTENT_PADDING_HORIZONTAL_DP.dp,
                        end = WelcomeLayout.CONTENT_PADDING_HORIZONTAL_DP.dp,
                        top = WelcomeLayout.CONTENT_PADDING_TOP_DP.dp,
                    )
                    .offset(y = (-WelcomeLayout.CONTENT_GROUP_SHIFT_UP_DP).dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(
                    WelcomeLayout.SLOT_TO_WORDMARK_GAP_DP.dp,
                    Alignment.CenterVertically,
                ),
            ) {
                WelcomeCheerSlot(
                    modifier = Modifier
                        .widthIn(max = WelcomeLayout.SLOT_WIDTH_DP.dp)
                        .fillMaxWidth()
                        .height(WelcomeLayout.SLOT_HEIGHT_DP.dp),
                )
                Column(
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.spacedBy(
                        WelcomeLayout.WORDMARK_TO_TAGLINE_GAP_DP.dp,
                    ),
                ) {
                    Wordmark()
                    CopyText(
                        text = stringResource(R.string.welcome_tagline),
                        widthDp = WelcomeLayout.TAGLINE_WIDTH_DP,
                        sizeSp = WelcomeLayout.TAGLINE_SIZE_SP,
                        color = WelcomeLayout.TAGLINE_COLOR,
                        fontFamily = WelcomeType.notoSemiBold,
                        fontWeight = FontWeight.SemiBold,
                    )
                }
            }
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(
                        start = WelcomeLayout.FOOTER_PADDING_HORIZONTAL_DP.dp,
                        end = WelcomeLayout.FOOTER_PADDING_HORIZONTAL_DP.dp,
                        bottom = WelcomeLayout.FOOTER_PADDING_BOTTOM_DP.dp,
                    ),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(WelcomeLayout.FOOTER_GAP_DP.dp),
            ) {
                WelcomeButton(
                    label = stringResource(R.string.welcome_start),
                    container = WelcomeLayout.PRIMARY_COLOR,
                    labelColor = WelcomeLayout.PRIMARY_TEXT_COLOR,
                    borderColor = null,
                    onClick = onStart,
                )
                WelcomeButton(
                    label = stringResource(R.string.welcome_existing_account),
                    container = WelcomeLayout.SECONDARY_COLOR,
                    labelColor = WelcomeLayout.SECONDARY_TEXT_COLOR,
                    borderColor = WelcomeLayout.SECONDARY_BORDER_COLOR,
                    onClick = onExistingAccount,
                )
                CopyText(
                    text = stringResource(R.string.welcome_training_note),
                    widthDp = WelcomeLayout.LEGAL_WIDTH_DP,
                    sizeSp = WelcomeLayout.FOOTER_TEXT_SIZE_SP,
                    color = WelcomeLayout.FOOTER_TEXT_COLOR,
                    fontFamily = WelcomeType.notoRegular,
                    fontWeight = FontWeight.Normal,
                )
                CopyText(
                    text = stringResource(R.string.welcome_legal),
                    widthDp = WelcomeLayout.LEGAL_WIDTH_DP,
                    sizeSp = WelcomeLayout.FOOTER_TEXT_SIZE_SP,
                    color = WelcomeLayout.FOOTER_TEXT_COLOR,
                    fontFamily = WelcomeType.notoRegular,
                    fontWeight = FontWeight.Normal,
                )
            }
        }
    }
}

/**
 * The splash's baked wordmark PNG, drawn at the size the 50sp text line used to
 * take. One image means the two screens cannot show different glyphs, and the
 * trailing "。" rides along inside the artwork instead of a second typeface.
 */
@Composable
private fun Wordmark() {
    val description = stringResource(R.string.welcome_wordmark_latin) +
        stringResource(R.string.welcome_wordmark_period)
    Box(
        modifier = Modifier.height(WelcomeLayout.WORDMARK_BOX_HEIGHT_DP.dp),
        contentAlignment = Alignment.Center,
    ) {
        Image(
            painter = painterResource(R.drawable.splash_g4_wordmark),
            contentDescription = description,
            modifier = Modifier
                .offset(y = WelcomeLayout.WORDMARK_INK_SHIFT_DOWN_DP.dp)
                .width(WelcomeLayout.WORDMARK_IMAGE_WIDTH_DP.dp)
                .aspectRatio(
                    WelcomeLayout.WORDMARK_ART_WIDTH_PX.toFloat() /
                        WelcomeLayout.WORDMARK_ART_HEIGHT_PX.toFloat(),
                ),
            contentScale = ContentScale.Fit,
        )
    }
}

@Composable
private fun CopyText(
    text: String,
    widthDp: Int,
    sizeSp: Int,
    color: Long,
    fontFamily: FontFamily,
    fontWeight: FontWeight,
) {
    Text(
        text = text,
        modifier = Modifier
            .widthIn(max = widthDp.dp)
            .fillMaxWidth(),
        style = welcomeTextStyle(sizeSp, color, fontFamily, fontWeight),
    )
}

@Composable
private fun WelcomeButton(
    label: String,
    container: Long,
    labelColor: Long,
    borderColor: Long?,
    onClick: () -> Unit,
) {
    QlutchButton(
        label = label,
        container = container,
        labelColor = labelColor,
        borderColor = borderColor,
        borderDp = if (borderColor == null) 0 else WelcomeLayout.SECONDARY_BORDER_DP,
        widthDp = WelcomeLayout.BUTTON_WIDTH_DP,
        heightDp = WelcomeLayout.BUTTON_HEIGHT_DP,
        radiusDp = WelcomeLayout.BUTTON_RADIUS_DP,
        textSizeSp = WelcomeLayout.BUTTON_TEXT_SIZE_SP,
        fontFamily = WelcomeType.notoSemiBold,
        fontWeight = FontWeight.SemiBold,
        onClick = onClick,
    )
}

private fun welcomeTextStyle(
    sizeSp: Int,
    color: Long,
    fontFamily: FontFamily,
    fontWeight: FontWeight,
): TextStyle {
    return TextStyle(
        color = Color(color),
        fontFamily = fontFamily,
        fontWeight = fontWeight,
        fontSize = sizeSp.sp,
        lineHeight = sizeSp.sp,
        textAlign = TextAlign.Center,
        letterSpacing = 0.sp,
        platformStyle = PlatformTextStyle(includeFontPadding = false),
    )
}
