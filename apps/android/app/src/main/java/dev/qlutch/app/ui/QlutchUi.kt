package dev.qlutch.app.ui

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.WindowInsetsSides
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.requiredHeight
import androidx.compose.foundation.layout.only
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawing
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.layout.windowInsetsPadding
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.PlatformTextStyle
import androidx.compose.ui.text.TextRange
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.input.TextFieldValue
import androidx.compose.ui.text.font.Font
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import dev.qlutch.app.R
import dev.qlutch.app.flow.FlowLogic
import dev.qlutch.app.welcome.WelcomeType

internal object QlutchTone {
    val page = Color(0xFFF2F5F7)
    val splash = Color(0xFFF2F5F8)
    val ink = Color(0xFF0F172A)
    val white = Color(0xFFFFFFFF)
    val slate = Color(0xFF64748B)
    val footer = Color(0xFF7C848F)
    val line = Color(0xFFE1E8F0)
    val backRing = Color(0xFFE2E3E7)
    val slot = Color(0xFFF1F4F9)
    val brand = Color(0xFF6C5CE7)
    val focus = Color(0xFF4381E6)
    val selected = Color(0xFFE8E4FF)
    val alipay = Color(0xFF1677FF)
    val flag = Color(0xFFDE2910)
    val dim = Color(0x660F172A)
    val progress = Brush.horizontalGradient(listOf(Color(0xFF4537CA), Color(0xFF6929D8), Color(0xFFD244EE)))
    val waitProgress = Brush.horizontalGradient(listOf(Color(0xFF6C5CE7), Color(0xFFD244EE)))
}

internal fun qlutchText(
    sizeSp: Int,
    color: Color,
    fontFamily: FontFamily,
    fontWeight: FontWeight,
    align: TextAlign = TextAlign.Center,
    lineSp: Int = sizeSp,
): TextStyle {
    return TextStyle(
        color = color,
        fontFamily = fontFamily,
        fontWeight = fontWeight,
        fontSize = sizeSp.sp,
        lineHeight = lineSp.sp,
        textAlign = align,
        letterSpacing = 0.sp,
        platformStyle = PlatformTextStyle(includeFontPadding = false),
    )
}

private val semi = WelcomeType.notoSemiBold
private val regular = WelcomeType.notoRegular
private val medium = FontFamily(Font(R.font.noto_sans_sc_medium, FontWeight.Medium))
private val bold = WelcomeType.notoExtraBold

/** U01 primary capsule. Disabled uses the OTP empty-state fill and does not click. */
@Composable
internal fun QlutchButton(
    label: String,
    container: Long,
    labelColor: Long,
    borderColor: Long?,
    borderDp: Int,
    widthDp: Int,
    heightDp: Int,
    radiusDp: Int,
    textSizeSp: Int,
    fontFamily: FontFamily,
    fontWeight: FontWeight,
    enabled: Boolean = true,
    onClick: () -> Unit = {},
) {
    val shape = RoundedCornerShape(radiusDp.dp)
    val fill = if (enabled) Color(container) else QlutchTone.slate
    var modifier = Modifier
        .widthIn(max = widthDp.dp)
        .fillMaxWidth()
        .height(heightDp.dp)
        .clip(shape)
        .background(fill, shape)
    if (borderColor != null && enabled) {
        modifier = modifier.border(borderDp.dp, Color(borderColor), shape)
    }
    Box(
        modifier = modifier.clickable(
            interactionSource = remember { MutableInteractionSource() },
            indication = null,
            enabled = enabled,
            role = Role.Button,
            onClick = onClick,
        ),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            text = label,
            style = qlutchText(textSizeSp, Color(labelColor), fontFamily, fontWeight),
        )
    }
}

@Composable
internal fun PrimaryButton(label: String, enabled: Boolean = true, onClick: () -> Unit) {
    QlutchButton(
        label = label,
        container = 0xFF0F172A,
        labelColor = 0xFFFFFFFF,
        borderColor = null,
        borderDp = 0,
        widthDp = 358,
        heightDp = 50,
        radiusDp = 25,
        textSizeSp = 16,
        fontFamily = semi,
        fontWeight = FontWeight.SemiBold,
        enabled = enabled,
        onClick = onClick,
    )
}

/** U02 outlined capsule. */
@Composable
internal fun SecondaryButton(label: String, enabled: Boolean = true, onClick: () -> Unit) {
    QlutchButton(
        label = label,
        container = 0xFFFFFFFF,
        labelColor = 0xFF0F172A,
        borderColor = 0xFFE1E8F0,
        borderDp = 1,
        widthDp = 358,
        heightDp = 50,
        radiusDp = 25,
        textSizeSp = 16,
        fontFamily = semi,
        fontWeight = FontWeight.SemiBold,
        enabled = enabled,
        onClick = onClick,
    )
}

/** U03 text action. Same metrics as the outlined button when the frame uses a full-width control. */
@Composable
internal fun TextAction(label: String, onClick: () -> Unit) {
    SecondaryButton(label = label, onClick = onClick)
}

/** U04 back control. */
@Composable
internal fun BackButton(onClick: () -> Unit) {
    Box(
        modifier = Modifier
            .size(40.dp)
            .shadow(0.dp, CircleShape)
            .clip(CircleShape)
            .background(QlutchTone.white)
            .border(1.dp, QlutchTone.backRing, CircleShape)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                role = Role.Button,
                onClick = onClick,
            ),
        contentAlignment = Alignment.Center,
    ) {
        Canvas(Modifier.size(20.dp)) {
            val paint = QlutchTone.ink
            val stroke = Stroke(width = 1.6.dp.toPx(), cap = StrokeCap.Round)
            drawLine(paint, Offset(size.width * 0.62f, size.height * 0.22f), Offset(size.width * 0.34f, size.height * 0.5f), strokeWidth = stroke.width, cap = StrokeCap.Round)
            drawLine(paint, Offset(size.width * 0.34f, size.height * 0.5f), Offset(size.width * 0.62f, size.height * 0.78f), strokeWidth = stroke.width, cap = StrokeCap.Round)
        }
    }
}

/** U12 seven-step bar. Permissions and P04 stay on step 7. */
@Composable
internal fun StepProgress(stepIndex: Int, modifier: Modifier = Modifier) {
    val fraction = (stepIndex.coerceIn(0, 7)) / 7f
    Box(
        modifier = modifier
            .height(16.dp)
            .clip(RoundedCornerShape(100.dp))
            .background(QlutchTone.slot),
        contentAlignment = Alignment.CenterStart,
    ) {
        Box(
            modifier = Modifier
                .padding(horizontal = 0.dp)
                .fillMaxWidth(fraction)
                .height(8.dp)
                .clip(RoundedCornerShape(100.dp))
                .background(QlutchTone.progress),
        )
    }
}

@Composable
internal fun FlowTopBar(stepIndex: Int, onBack: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(start = 16.dp, end = 16.dp, top = 16.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        BackButton(onBack)
        StepProgress(stepIndex, Modifier.weight(1f))
    }
}

@Composable
internal fun PromptRow(title: String, subtitle: String? = null) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(start = 16.dp, end = 16.dp, top = 16.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        androidx.compose.foundation.Image(
            painter = androidx.compose.ui.res.painterResource(dev.qlutch.app.R.drawable.flow_avatar),
            contentDescription = null,
            modifier = Modifier
                .size(48.dp)
                .clip(CircleShape),
            contentScale = androidx.compose.ui.layout.ContentScale.Crop,
        )
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(4.dp)) {
            Text(title, style = qlutchText(22, QlutchTone.ink, bold, FontWeight.Bold, TextAlign.Start, 28))
            if (subtitle != null) {
                Text(subtitle, style = qlutchText(16, QlutchTone.slate, semi, FontWeight.SemiBold, TextAlign.Start, 22))
            }
        }
    }
}

/** U13 single-choice row. */
@Composable
internal fun ChoiceRow(label: String, selected: Boolean, onClick: () -> Unit) {
    val shape = RoundedCornerShape(25.dp)
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .height(50.dp)
            .shadow(2.dp, shape, ambientColor = Color(0x140F172A), spotColor = Color(0x140F172A))
            .clip(shape)
            .background(if (selected) QlutchTone.selected else QlutchTone.white)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                role = Role.RadioButton,
                onClick = onClick,
            )
            .padding(horizontal = 16.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = if (selected) Arrangement.spacedBy(8.dp, Alignment.CenterHorizontally) else Arrangement.Center,
    ) {
        if (selected) {
            Canvas(Modifier.size(20.dp)) {
                val stroke = Stroke(width = 2.dp.toPx(), cap = StrokeCap.Round)
                drawLine(QlutchTone.ink, Offset(size.width * 0.2f, size.height * 0.52f), Offset(size.width * 0.42f, size.height * 0.74f), stroke.width, StrokeCap.Round)
                drawLine(QlutchTone.ink, Offset(size.width * 0.42f, size.height * 0.74f), Offset(size.width * 0.82f, size.height * 0.28f), stroke.width, StrokeCap.Round)
            }
        }
        Text(label, style = qlutchText(16, QlutchTone.ink, semi, FontWeight.SemiBold))
    }
}

/** U14 permission checkbox. */
@Composable
internal fun PermissionCheck(checked: Boolean, onClick: () -> Unit) {
    Box(
        modifier = Modifier
            .size(24.dp)
            .clip(RoundedCornerShape(6.dp))
            .background(if (checked) QlutchTone.ink else QlutchTone.white)
            .border(if (checked) 0.dp else 1.5.dp, QlutchTone.line, RoundedCornerShape(6.dp))
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                role = Role.Checkbox,
                onClick = onClick,
            ),
        contentAlignment = Alignment.Center,
    ) {
        if (checked) {
            Canvas(Modifier.size(14.dp)) {
                val stroke = Stroke(width = 1.8.dp.toPx(), cap = StrokeCap.Round)
                drawLine(QlutchTone.white, Offset(size.width * 0.15f, size.height * 0.52f), Offset(size.width * 0.4f, size.height * 0.76f), stroke.width, StrokeCap.Round)
                drawLine(QlutchTone.white, Offset(size.width * 0.4f, size.height * 0.76f), Offset(size.width * 0.86f, size.height * 0.24f), stroke.width, StrokeCap.Round)
            }
        }
    }
}

/** U09 six OTP cells. */
@Composable
internal fun OtpCells(code: String) {
    Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
        repeat(6) { index ->
            val char = code.getOrNull(index)?.toString() ?: ""
            val focused = code.length == index
            val shape = RoundedCornerShape(12.dp)
            Box(
                modifier = Modifier
                    .size(width = 48.dp, height = 56.dp)
                    .clip(shape)
                    .background(QlutchTone.white)
                    .border(if (focused) 2.dp else 1.dp, if (focused) QlutchTone.focus else QlutchTone.line, shape),
                contentAlignment = Alignment.Center,
            ) {
                Text(char, style = qlutchText(22, QlutchTone.ink, semi, FontWeight.SemiBold))
            }
        }
    }
}

/** U10 single-line field. */
@Composable
internal fun LineField(
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String,
    modifier: Modifier = Modifier,
) {
    val shape = RoundedCornerShape(25.dp)
    BasicTextField(
        value = value,
        onValueChange = onValueChange,
        textStyle = qlutchText(15, QlutchTone.ink, regular, FontWeight.Normal, TextAlign.Start),
        cursorBrush = SolidColor(QlutchTone.ink),
        singleLine = true,
        modifier = modifier
            .fillMaxWidth()
            .height(50.dp)
            .clip(shape)
            .background(QlutchTone.white)
            .border(1.dp, QlutchTone.line, shape)
            .padding(horizontal = 20.dp),
        decorationBox = { inner ->
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.CenterStart) {
                if (value.isEmpty()) {
                    Text(placeholder, style = qlutchText(15, QlutchTone.ink, regular, FontWeight.Normal, TextAlign.Start))
                }
                inner()
            }
        },
    )
}

/** U11 multiline notes. */
@Composable
internal fun NotesField(value: String, onValueChange: (String) -> Unit) {
    val shape = RoundedCornerShape(22.dp)
    BasicTextField(
        value = value,
        onValueChange = onValueChange,
        textStyle = qlutchText(15, QlutchTone.ink, regular, FontWeight.Normal, TextAlign.Start, 22),
        cursorBrush = SolidColor(QlutchTone.ink),
        modifier = Modifier
            .fillMaxWidth()
            .height(220.dp)
            .clip(shape)
            .background(QlutchTone.white)
            .border(1.dp, QlutchTone.line, shape)
            .padding(16.dp),
    )
}

/** U15 toast. Stays on the current page. */
@Composable
internal fun ClosedToast(visible: Boolean) {
    if (!visible) return
    Box(
        modifier = Modifier
            .clip(RoundedCornerShape(100.dp))
            .background(QlutchTone.ink)
            .padding(horizontal = 20.dp, vertical = 12.dp),
    ) {
        Text(dev.qlutch.app.flow.FlowLogic.TOAST_CLOSED, style = qlutchText(14, QlutchTone.white, semi, FontWeight.SemiBold))
    }
}

/** U16 channel that is shown and never signs the user in. */
@Composable
internal fun ClosedChannel(label: String, mark: @Composable () -> Unit, onClosed: () -> Unit) {
    Row(
        modifier = Modifier
            .widthIn(max = 358.dp)
            .fillMaxWidth()
            .requiredHeight(50.dp)
            .clip(RoundedCornerShape(25.dp))
            .background(QlutchTone.white)
            .border(1.dp, QlutchTone.line, RoundedCornerShape(25.dp))
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                role = Role.Button,
                onClick = onClosed,
            ),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(8.dp, Alignment.CenterHorizontally),
    ) {
        mark()
        Text(label, style = qlutchText(16, QlutchTone.ink, semi, FontWeight.SemiBold))
    }
}

@Composable
internal fun WeChatMark() {
    Canvas(Modifier.size(24.dp)) {
        drawCircle(Color(0xFF22C55E), radius = size.minDimension * 0.28f, center = Offset(size.width * 0.38f, size.height * 0.42f))
        drawCircle(Color(0xFF16A34A), radius = size.minDimension * 0.22f, center = Offset(size.width * 0.64f, size.height * 0.62f))
    }
}

@Composable
internal fun AlipayMark() {
    Box(
        modifier = Modifier
            .size(24.dp)
            .clip(RoundedCornerShape(5.5.dp))
            .background(QlutchTone.alipay),
        contentAlignment = Alignment.Center,
    ) {
        Text("支", style = qlutchText(14, QlutchTone.white, bold, FontWeight.Bold))
    }
}

@Composable
internal fun FlagMark(hongKong: Boolean, large: Boolean = false) {
    val width = if (large) 40.dp else 22.dp
    val height = if (large) 40.dp else 16.dp
    Canvas(Modifier.size(width, height)) {
        drawRoundRect(QlutchTone.flag, cornerRadius = androidx.compose.ui.geometry.CornerRadius(3.dp.toPx(), 3.dp.toPx()))
        val star = if (hongKong) QlutchTone.white else Color(0xFFFFDE00)
        drawCircle(star, radius = size.minDimension * 0.16f, center = Offset(size.width * 0.32f, size.height * 0.38f))
    }
}

/** U07 phone row: region chip plus number. */
@Composable
internal fun PhoneField(
    region: String,
    hongKong: Boolean,
    number: String,
    onRegion: () -> Unit,
    onNumber: (String) -> Unit,
) {
    val digits = FlowLogic.phoneDigits(number)
    val formatted = FlowLogic.formatPhone(region, digits)
    var field by remember {
        mutableStateOf(TextFieldValue(formatted, TextRange(formatted.length)))
    }
    if (field.text != formatted) {
        val caretDigits = if (FlowLogic.phoneDigits(field.text) == digits) {
            FlowLogic.digitsBeforeCursor(field.text, field.selection.end)
        } else {
            digits.length
        }
        field = TextFieldValue(formatted, TextRange(FlowLogic.phoneCursor(formatted, caretDigits)))
    }
    val shape = RoundedCornerShape(25.dp)
    Row(
        modifier = Modifier
            .widthIn(max = 358.dp)
            .fillMaxWidth()
            .height(50.dp)
            .clip(shape)
            .background(QlutchTone.white)
            .border(1.dp, QlutchTone.line, shape)
            .padding(horizontal = 16.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        Row(
            modifier = Modifier.clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onRegion,
            ),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(6.dp),
        ) {
            FlagMark(hongKong)
            Text(region, style = qlutchText(16, QlutchTone.ink, semi, FontWeight.SemiBold))
            ChevronDown()
        }
        Box(Modifier.size(width = 1.dp, height = 20.dp).background(QlutchTone.line))
        BasicTextField(
            value = field,
            onValueChange = { proposed ->
                if (proposed.text == field.text) {
                    field = proposed
                } else {
                    val edit = FlowLogic.editPhoneField(region, proposed.text, proposed.selection.end)
                    field = TextFieldValue(edit.text, TextRange(edit.cursor))
                    if (edit.digits != digits) onNumber(edit.digits)
                }
            },
            textStyle = qlutchText(16, QlutchTone.ink, regular, FontWeight.Normal, TextAlign.Start),
            cursorBrush = SolidColor(QlutchTone.ink),
            singleLine = true,
            modifier = Modifier.weight(1f),
            decorationBox = { inner ->
                Box {
                    if (digits.isEmpty()) {
                        Text("138 0013 8000", style = qlutchText(16, QlutchTone.ink, regular, FontWeight.Normal, TextAlign.Start))
                    }
                    inner()
                }
            },
        )
    }
}

/** U08 language chip. */
@Composable
internal fun LanguageChip(label: String, onClick: () -> Unit) {
    Row(
        modifier = Modifier
            .clip(RoundedCornerShape(20.dp))
            .background(QlutchTone.white)
            .border(1.dp, QlutchTone.line, RoundedCornerShape(20.dp))
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            )
            .padding(horizontal = 14.dp, vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(6.dp),
    ) {
        Canvas(Modifier.size(16.dp)) {
            drawCircle(color = QlutchTone.ink, style = Stroke(width = 1.2.dp.toPx()))
            drawLine(QlutchTone.ink, Offset(size.width / 2f, 1.2.dp.toPx()), Offset(size.width / 2f, size.height - 1.2.dp.toPx()), strokeWidth = 1.dp.toPx())
            drawLine(QlutchTone.ink, Offset(1.2.dp.toPx(), size.height / 2f), Offset(size.width - 1.2.dp.toPx(), size.height / 2f), strokeWidth = 1.dp.toPx())
        }
        Text(label, style = qlutchText(14, QlutchTone.ink, semi, FontWeight.SemiBold))
        ChevronDown()
    }
}

@Composable
internal fun ChevronDown() {
    Canvas(Modifier.size(16.dp)) {
        val stroke = Stroke(width = 1.4.dp.toPx(), cap = StrokeCap.Round)
        drawLine(QlutchTone.ink, Offset(size.width * 0.22f, size.height * 0.38f), Offset(size.width * 0.5f, size.height * 0.64f), stroke.width, StrokeCap.Round)
        drawLine(QlutchTone.ink, Offset(size.width * 0.5f, size.height * 0.64f), Offset(size.width * 0.78f, size.height * 0.38f), stroke.width, StrokeCap.Round)
    }
}

/** U05 / U06 sheet family. */
@Composable
internal fun OptionSheet(
    title: String,
    searchHint: String,
    query: String,
    onQuery: (String) -> Unit,
    onClose: () -> Unit,
    content: @Composable () -> Unit,
) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(QlutchTone.dim)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClose,
            ),
        contentAlignment = Alignment.BottomCenter,
    ) {
        Column(
            modifier = Modifier
                .widthIn(max = 390.dp)
                .fillMaxWidth()
                .windowInsetsPadding(WindowInsets.safeDrawing.only(WindowInsetsSides.Bottom))
                .clip(RoundedCornerShape(topStart = 28.dp, topEnd = 28.dp))
                .background(QlutchTone.white)
                .clickable(
                    interactionSource = remember { MutableInteractionSource() },
                    indication = null,
                    onClick = {},
                )
                .padding(start = 16.dp, end = 16.dp, top = 16.dp, bottom = 28.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Spacer(Modifier.size(40.dp))
                Text(
                    title,
                    modifier = Modifier.weight(1f),
                    style = qlutchText(17, QlutchTone.ink, semi, FontWeight.SemiBold),
                )
                Box(
                    modifier = Modifier
                        .size(40.dp)
                        .clip(CircleShape)
                        .background(QlutchTone.page)
                        .clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = onClose,
                        ),
                    contentAlignment = Alignment.Center,
                ) {
                    Canvas(Modifier.size(20.dp)) {
                        val stroke = Stroke(width = 1.6.dp.toPx(), cap = StrokeCap.Round)
                        drawLine(QlutchTone.ink, Offset(size.width * 0.25f, size.height * 0.25f), Offset(size.width * 0.75f, size.height * 0.75f), stroke.width, StrokeCap.Round)
                        drawLine(QlutchTone.ink, Offset(size.width * 0.75f, size.height * 0.25f), Offset(size.width * 0.25f, size.height * 0.75f), stroke.width, StrokeCap.Round)
                    }
                }
            }
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(44.dp)
                    .clip(RoundedCornerShape(22.dp))
                    .background(QlutchTone.page)
                    .padding(horizontal = 16.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                BasicTextField(
                    value = query,
                    onValueChange = onQuery,
                    textStyle = qlutchText(15, QlutchTone.ink, regular, FontWeight.Normal, TextAlign.Start),
                    cursorBrush = SolidColor(QlutchTone.ink),
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                    decorationBox = { inner ->
                        Box {
                            if (query.isEmpty()) {
                                Text(searchHint, style = qlutchText(15, QlutchTone.slate, regular, FontWeight.Normal, TextAlign.Start))
                            }
                            inner()
                        }
                    },
                )
            }
            content()
        }
    }
}

@Composable
internal fun SheetRow(
    title: String,
    subtitle: String?,
    selected: Boolean,
    leading: @Composable (() -> Unit)? = null,
    onClick: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            )
            .padding(vertical = 16.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        if (leading != null) leading()
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Text(title, style = qlutchText(16, QlutchTone.ink, semi, FontWeight.SemiBold, TextAlign.Start))
            if (subtitle != null) {
                Text(subtitle, style = qlutchText(13, QlutchTone.slate, regular, FontWeight.Normal, TextAlign.Start))
            }
        }
        Canvas(Modifier.size(24.dp)) {
            if (selected) {
                drawCircle(QlutchTone.ink, style = Stroke(width = 1.6.dp.toPx()))
                drawCircle(QlutchTone.ink, radius = size.minDimension * 0.22f)
            } else {
                drawCircle(QlutchTone.line, style = Stroke(width = 1.6.dp.toPx()))
            }
        }
    }
}

@Composable
internal fun HairlineLabel(label: String) {
    Row(
        modifier = Modifier.widthIn(max = 358.dp).fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Box(Modifier.weight(1f).height(1.dp).background(QlutchTone.line))
        Text(label, style = qlutchText(13, if (label == "或") QlutchTone.footer else QlutchTone.slate, regular, FontWeight.Normal))
        Box(Modifier.weight(1f).height(1.dp).background(QlutchTone.line))
    }
}

@Composable
internal fun ImportChip(label: String, modifier: Modifier = Modifier, onClick: () -> Unit) {
    Row(
        modifier = modifier
            .height(50.dp)
            .clip(RoundedCornerShape(25.dp))
            .background(QlutchTone.white)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(8.dp, Alignment.CenterHorizontally),
    ) {
        Text(label, style = qlutchText(16, QlutchTone.ink, semi, FontWeight.SemiBold))
    }
}

@Composable
internal fun WaitMeter(fraction: Float) {
    Box(
        modifier = Modifier
            .width(248.dp)
            .height(8.dp)
            .clip(RoundedCornerShape(100.dp))
            .background(QlutchTone.line),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth(fraction.coerceIn(0f, 1f))
                .height(8.dp)
                .clip(RoundedCornerShape(100.dp))
                .background(QlutchTone.waitProgress),
        )
    }
}

@Composable
internal fun FooterNote(text: String, width: Dp = 358.dp) {
    Text(
        text,
        modifier = Modifier.widthIn(max = width).fillMaxWidth(),
        style = qlutchText(13, QlutchTone.footer, regular, FontWeight.Normal),
    )
}

@Composable
internal fun ScreenTitle(text: String, size: Int = 32) {
    Text(text, style = qlutchText(size, QlutchTone.ink, bold, FontWeight.Bold))
}

internal val qlutchMedium: FontFamily get() = medium
