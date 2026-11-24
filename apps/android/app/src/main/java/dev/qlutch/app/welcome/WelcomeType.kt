package dev.qlutch.app.welcome

import androidx.compose.ui.text.font.Font
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import dev.qlutch.app.R

/**
 * Chinese copy falls back to a Noto Sans SC subset at the same weight, because
 * Inter has no CJK glyphs.
 *
 * The wordmark itself is no longer set in Inter: both the splash and the
 * welcome screen draw the same baked PNG (see WelcomeLayout.WORDMARK_ART_*),
 * so the extra-bold entries below stay only as the brand typeface reference.
 */
internal object WelcomeType {
    val interExtraBold = FontFamily(
        Font(R.font.inter_extrabold, FontWeight.ExtraBold),
    )
    val notoExtraBold = FontFamily(
        Font(R.font.noto_sans_sc_extrabold, FontWeight.ExtraBold),
    )
    val notoSemiBold = FontFamily(
        Font(R.font.noto_sans_sc_semibold, FontWeight.SemiBold),
    )
    val notoRegular = FontFamily(
        Font(R.font.noto_sans_sc_regular, FontWeight.Normal),
    )
}
