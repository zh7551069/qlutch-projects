package dev.qlutch.app.flow

import android.Manifest
import android.content.Intent
import android.content.pm.PackageManager
import android.graphics.ImageDecoder
import android.graphics.drawable.Animatable
import android.graphics.drawable.AnimatedImageDrawable
import android.os.Build
import android.os.Bundle
import android.widget.ImageView
import androidx.activity.ComponentActivity
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.Image
import androidx.compose.foundation.clickable
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawing
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.layout.windowInsetsPadding
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableLongStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.core.content.ContextCompat
import dev.qlutch.app.ApiConfig
import dev.qlutch.app.MainActivity
import dev.qlutch.app.R
import dev.qlutch.app.ui.AlipayMark
import dev.qlutch.app.ui.BackButton
import dev.qlutch.app.ui.ChoiceRow
import dev.qlutch.app.ui.ClosedChannel
import dev.qlutch.app.ui.ClosedToast
import dev.qlutch.app.ui.FlagMark
import dev.qlutch.app.ui.FlowTopBar
import dev.qlutch.app.ui.FooterNote
import dev.qlutch.app.ui.HairlineLabel
import dev.qlutch.app.ui.ImportChip
import dev.qlutch.app.ui.LanguageChip
import dev.qlutch.app.ui.LineField
import dev.qlutch.app.ui.NotesField
import dev.qlutch.app.ui.OptionSheet
import dev.qlutch.app.ui.OtpCells
import dev.qlutch.app.ui.PermissionCheck
import dev.qlutch.app.ui.PhoneField
import dev.qlutch.app.ui.PrimaryButton
import dev.qlutch.app.ui.PromptRow
import dev.qlutch.app.ui.QlutchTone
import dev.qlutch.app.ui.ScreenTitle
import dev.qlutch.app.ui.SecondaryButton
import dev.qlutch.app.ui.SheetRow
import dev.qlutch.app.ui.WaitMeter
import dev.qlutch.app.ui.WeChatMark
import dev.qlutch.app.ui.qlutchMedium
import dev.qlutch.app.ui.qlutchText
import dev.qlutch.app.welcome.WelcomeType
import kotlinx.coroutines.delay
import androidx.compose.material3.Text

class FlowActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        val purpose = intent.getStringExtra(EXTRA_PURPOSE) ?: FlowLogic.PURPOSE_REGISTER
        setContent {
            NewUserFlow(
                purpose = purpose,
                api = HttpIdentityApi(ApiConfig.BASE_URL),
                onHome = {
                    startActivity(Intent(this, MainActivity::class.java))
                    finish()
                },
                onLeave = { finish() },
            )
        }
    }

    companion object {
        const val EXTRA_PURPOSE = "purpose"

        fun intent(context: android.content.Context, purpose: String): Intent {
            return Intent(context, FlowActivity::class.java).putExtra(EXTRA_PURPOSE, purpose)
        }
    }
}

@Composable
private fun NewUserFlow(
    purpose: String,
    api: IdentityApi,
    onHome: () -> Unit,
    onLeave: () -> Unit,
) {
    var place by remember { mutableStateOf(FlowPlace(FlowRoute.AUTH)) }
    var backStack by remember { mutableStateOf(listOf<FlowPlace>()) }
    var mode by remember { mutableStateOf(purpose) }
    var region by remember { mutableStateOf(FlowLogic.REGION_MAINLAND) }
    var locale by remember { mutableStateOf(FlowLogic.LOCALE_HANS) }
    var phone by remember { mutableStateOf("") }
    var code by remember { mutableStateOf("") }
    var hint by remember { mutableStateOf<String?>(null) }
    var nickname by remember { mutableStateOf("") }
    var selected by remember { mutableStateOf("") }
    var notes by remember {
        mutableStateOf("光合作用：绿色植物利用光能，把二氧化碳和水变成有机物，并放出氧气。\n这是本章要掌握的核心过程。")
    }
    var contacts by remember { mutableStateOf(false) }
    var notices by remember { mutableStateOf(true) }
    var permissionsSeen by remember { mutableStateOf(false) }
    var bearer by remember { mutableStateOf("") }
    var error by remember { mutableStateOf("") }
    var busy by remember { mutableStateOf(false) }
    var sheet by remember { mutableStateOf("") }
    var search by remember { mutableStateOf("") }
    var toast by remember { mutableStateOf(false) }
    var waitStarted by remember { mutableLongStateOf(0L) }
    var now by remember { mutableLongStateOf(0L) }

    fun push(next: FlowPlace) {
        backStack = backStack + place
        if (next.route == FlowRoute.WAIT) waitStarted = 0L
        place = next
        error = ""
        toast = false
    }

    fun go(next: FlowPlace) {
        place = next
        error = ""
        toast = false
    }

    fun back() {
        toast = false
        if (place.route == FlowRoute.WAIT) waitStarted = 0L
        if (backStack.isEmpty()) {
            onLeave()
        } else {
            place = backStack.last()
            backStack = backStack.dropLast(1)
        }
    }

    fun openHome() {
        onHome()
    }

    fun afterSession(nextStep: String) {
        val next = FlowLogic.placeAfterSession(nextStep, permissionsSeen)
        if (next.route == FlowRoute.HOME) {
            openHome()
        } else {
            backStack = emptyList()
            go(next)
        }
    }

    LaunchedEffect(toast) {
        if (toast) {
            delay(2000)
            toast = false
        }
    }

    Board {
        when (place.route) {
            FlowRoute.AUTH -> AuthScreen(
                mode = mode,
                region = region,
                locale = locale,
                phone = phone,
                error = error,
                busy = busy,
                onMode = { mode = it },
                onPhone = { phone = FlowLogic.phoneDigits(it) },
                onRegionSheet = { search = ""; sheet = "region" },
                onLanguageSheet = { search = ""; sheet = "language" },
                onClosed = { toast = true },
                onContinue = {
                    if (!FlowLogic.phoneContinueEnabled(phone) || busy) return@AuthScreen
                    busy = true
                    error = ""
                    Thread {
                        val outcome = api.send(FlowLogic.newIdempotencyKey(), region, phone, mode, locale)
                        runOnUi {
                            busy = false
                            when (outcome) {
                                is ApiOutcome.Ok -> {
                                    hint = outcome.value.destination_hint
                                    code = ""
                                    push(FlowPlace(FlowRoute.OTP))
                                }
                                is ApiOutcome.Err -> when (outcome.code) {
                                    "AUTH_ACCOUNT_NOT_FOUND" -> push(FlowPlace(FlowRoute.NO_ACCOUNT))
                                    "AUTH_ACCOUNT_EXISTS" -> mode = FlowLogic.PURPOSE_LOGIN
                                    else -> error = outcome.code
                                }
                                is ApiOutcome.Transport -> error = outcome.reason
                            }
                        }
                    }.start()
                },
            )
            FlowRoute.OTP -> OtpScreen(
                code = code,
                line = FlowLogic.sentToLine(hint, region, phone),
                error = error,
                busy = busy,
                onBack = { back() },
                onCode = { code = FlowLogic.phoneDigits(it).take(6) },
                onContinue = {
                    if (!FlowLogic.otpContinueEnabled(code) || busy) return@OtpScreen
                    busy = true
                    error = ""
                    Thread {
                        val outcome = api.verify(FlowLogic.newIdempotencyKey(), region, phone, mode, code, locale)
                        runOnUi {
                            busy = false
                            when (outcome) {
                                is ApiOutcome.Ok -> {
                                    bearer = outcome.value.access_token
                                    if (outcome.value.next_step == FlowLogic.NEXT_ONBOARDING) {
                                        Thread {
                                            val state = api.onboarding(outcome.value.access_token)
                                            runOnUi {
                                                val step = (state as? ApiOutcome.Ok)?.value?.current_step_id
                                                backStack = emptyList()
                                                go(FlowLogic.placeForOnboarding(step))
                                            }
                                        }.start()
                                    } else {
                                        afterSession(outcome.value.next_step)
                                    }
                                }
                                is ApiOutcome.Err -> {
                                    if (outcome.code == "AUTH_OTP_INVALID" || outcome.code == "AUTH_OTP_EXPIRED" || outcome.code == "AUTH_OTP_MAX_ATTEMPTS") {
                                        code = ""
                                    }
                                    error = outcome.code
                                }
                                is ApiOutcome.Transport -> error = outcome.reason
                            }
                        }
                    }.start()
                },
                onResend = {
                    if (busy) return@OtpScreen
                    busy = true
                    Thread {
                        val outcome = api.send(FlowLogic.newIdempotencyKey(), region, phone, mode, locale)
                        runOnUi {
                            busy = false
                            when (outcome) {
                                is ApiOutcome.Ok -> {
                                    hint = outcome.value.destination_hint
                                    code = ""
                                    error = ""
                                }
                                is ApiOutcome.Err -> error = outcome.code
                                is ApiOutcome.Transport -> error = outcome.reason
                            }
                        }
                    }.start()
                },
            )
            FlowRoute.NO_ACCOUNT -> NoAccountScreen(
                region = region,
                phone = phone,
                busy = busy,
                error = error,
                onCreate = {
                    mode = FlowLogic.PURPOSE_REGISTER
                    busy = true
                    Thread {
                        val outcome = api.send(FlowLogic.newIdempotencyKey(), region, phone, FlowLogic.PURPOSE_REGISTER, locale)
                        runOnUi {
                            busy = false
                            when (outcome) {
                                is ApiOutcome.Ok -> {
                                    hint = outcome.value.destination_hint
                                    code = ""
                                    push(FlowPlace(FlowRoute.OTP))
                                }
                                is ApiOutcome.Err -> error = outcome.code
                                is ApiOutcome.Transport -> error = outcome.reason
                            }
                        }
                    }.start()
                },
                onOther = {
                    mode = FlowLogic.PURPOSE_LOGIN
                    go(FlowPlace(FlowRoute.AUTH))
                },
            )
            FlowRoute.NICKNAME -> NicknameScreen(
                nickname = nickname,
                error = error,
                busy = busy,
                onBack = { back() },
                onNickname = { nickname = it },
                onContinue = {
                    if (!FlowLogic.nicknameContinueEnabled(nickname) || busy) return@NicknameScreen
                    busy = true
                    Thread {
                        val outcome = api.writeOnboarding(bearer, "nickname", nickname.trim(), null)
                        runOnUi {
                            busy = false
                            when (outcome) {
                                is ApiOutcome.Ok -> {
                                    selected = ""
                                    push(FlowPlace(FlowRoute.QUESTION, stepIndex = 2))
                                }
                                is ApiOutcome.Err -> error = outcome.code
                                is ApiOutcome.Transport -> error = outcome.reason
                            }
                        }
                    }.start()
                },
            )
            FlowRoute.QUESTION -> {
                val step = FlowLogic.question(place.stepIndex)
                QuestionScreen(
                    step = step,
                    selected = selected,
                    error = error,
                    busy = busy,
                    onBack = { back() },
                    onSelect = { label ->
                        selected = label
                        if (step.finishLabel == null && !busy) {
                            busy = true
                            Thread {
                                val outcome = api.writeOnboarding(bearer, step.stepId, null, FlowLogic.answerKey(label))
                                runOnUi {
                                    busy = false
                                    when (outcome) {
                                        is ApiOutcome.Ok -> {
                                            selected = ""
                                            val nextIndex = step.index + 1
                                            if (nextIndex <= 7) push(FlowPlace(FlowRoute.QUESTION, stepIndex = nextIndex))
                                        }
                                        is ApiOutcome.Err -> error = outcome.code
                                        is ApiOutcome.Transport -> error = outcome.reason
                                    }
                                }
                            }.start()
                        }
                    },
                    onFinish = {
                        if (selected.isEmpty() || busy) return@QuestionScreen
                        busy = true
                        Thread {
                            val outcome = api.writeOnboarding(bearer, step.stepId, null, FlowLogic.answerKey(selected))
                            runOnUi {
                                busy = false
                                when (outcome) {
                                    is ApiOutcome.Ok -> afterSession(outcome.value.next_step)
                                    is ApiOutcome.Err -> error = outcome.code
                                    is ApiOutcome.Transport -> error = outcome.reason
                                }
                            }
                        }.start()
                    },
                )
            }
            FlowRoute.PERMISSIONS_CONTACTS -> PermissionContactsScreen(
                contacts = contacts,
                notices = notices,
                onContacts = { contacts = it },
                onNotices = { notices = it },
                onBack = { back() },
                onContinue = {
                    push(FlowPlace(FlowRoute.PERMISSIONS_LOCATION, stepIndex = 7))
                },
                onSkip = { push(FlowPlace(FlowRoute.PERMISSIONS_LOCATION, stepIndex = 7)) },
            )
            FlowRoute.PERMISSIONS_LOCATION -> PermissionLocationScreen(
                onBack = { back() },
                onContinue = {
                    permissionsSeen = true
                    push(FlowPlace(FlowRoute.PACK, stepIndex = 7))
                },
                onSkip = {
                    permissionsSeen = true
                    push(FlowPlace(FlowRoute.PACK, stepIndex = 7))
                },
            )
            FlowRoute.PACK -> PackScreen(
                onBack = { back() },
                onPhoto = { push(FlowPlace(FlowRoute.WAIT, stepIndex = 7, waitSource = FlowRoute.PACK)) },
                onMore = { push(FlowPlace(FlowRoute.MORE, stepIndex = 7)) },
            )
            FlowRoute.MORE -> MoreScreen(
                toast = toast,
                onBack = { back() },
                onUpload = { push(FlowPlace(FlowRoute.WAIT, stepIndex = 7, waitSource = FlowRoute.MORE)) },
                onClosed = { toast = true },
                onPaste = { push(FlowPlace(FlowRoute.PASTE, stepIndex = 7)) },
                onNoMaterial = { openHome() },
            )
            FlowRoute.PASTE -> PasteScreen(
                notes = notes,
                onNotes = { notes = it },
                onBack = { back() },
                onContinue = {
                    if (!FlowLogic.notesContinueEnabled(notes)) return@PasteScreen
                    push(FlowPlace(FlowRoute.WAIT, stepIndex = 7, waitSource = FlowRoute.PASTE))
                },
            )
            FlowRoute.WAIT -> WaitScreen(
                startedAt = waitStarted,
                now = now,
                onBack = { back() },
                onArm = {
                    val stamp = android.os.SystemClock.elapsedRealtime()
                    waitStarted = stamp
                    now = stamp
                },
                onTick = { now = android.os.SystemClock.elapsedRealtime() },
                onFinished = { openHome() },
            )
            FlowRoute.HOME -> Unit
        }
        if (sheet == "region") {
            OptionSheet(
                title = "选择国家或地区",
                searchHint = "搜索国家或区号",
                query = search,
                onQuery = { search = it },
                onClose = { sheet = "" },
            ) {
                FlowLogic.filterRegions(search).forEach { choice ->
                    SheetRow(
                        title = choice.name,
                        subtitle = choice.code,
                        selected = choice.code == region,
                        leading = { FlagMark(choice.code == FlowLogic.REGION_HONG_KONG, large = true) },
                        onClick = {
                            region = choice.code
                            sheet = ""
                        },
                    )
                }
            }
        }
        if (sheet == "language") {
            OptionSheet(
                title = "选择语言",
                searchHint = "搜索语言",
                query = search,
                onQuery = { search = it },
                onClose = { sheet = "" },
            ) {
                FlowLogic.filterLanguages(search).forEach { choice ->
                    SheetRow(
                        title = choice.label,
                        subtitle = null,
                        selected = choice.locale == locale,
                        onClick = {
                            locale = choice.locale
                            sheet = ""
                        },
                    )
                }
            }
        }
    }
}

private fun runOnUi(block: () -> Unit) {
    android.os.Handler(android.os.Looper.getMainLooper()).post(block)
}

@Composable
private fun Board(content: @Composable () -> Unit) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(QlutchTone.page),
        contentAlignment = Alignment.TopCenter,
    ) {
        Column(
            modifier = Modifier
                .widthIn(max = 390.dp)
                .fillMaxWidth()
                .fillMaxHeight()
                .windowInsetsPadding(WindowInsets.safeDrawing),
        ) {
            content()
        }
    }
}

@Composable
private fun AuthScreen(
    mode: String,
    region: String,
    locale: String,
    phone: String,
    error: String,
    busy: Boolean,
    onMode: (String) -> Unit,
    onPhone: (String) -> Unit,
    onRegionSheet: () -> Unit,
    onLanguageSheet: () -> Unit,
    onClosed: () -> Unit,
    onContinue: () -> Unit,
) {
    val login = mode == FlowLogic.PURPOSE_LOGIN
    Column(Modifier.fillMaxSize(), horizontalAlignment = Alignment.CenterHorizontally) {
        Column(
            Modifier
                .fillMaxWidth()
                .padding(start = 16.dp, end = 16.dp, top = 48.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Image(
                painter = painterResource(R.drawable.flow_wordmark_small),
                contentDescription = "Qlutch",
                modifier = Modifier.width(111.dp).height(28.dp),
                contentScale = ContentScale.Fit,
            )
            Spacer(Modifier.height(20.dp))
            ScreenTitle("开始吧！")
            Spacer(Modifier.height(32.dp))
            Row(
                modifier = Modifier
                    .clip(RoundedCornerShape(24.dp))
                    .background(QlutchTone.line)
                    .padding(4.dp),
            ) {
                TabChip("注册", !login) { onMode(FlowLogic.PURPOSE_REGISTER) }
                TabChip("登录", login) { onMode(FlowLogic.PURPOSE_LOGIN) }
            }
            Spacer(Modifier.height(36.dp))
            Text(
                if (login) "用手机号登录" else "用手机号注册",
                style = qlutchText(16, QlutchTone.ink, WelcomeType.notoSemiBold, FontWeight.SemiBold, TextAlign.Start),
            )
            Spacer(Modifier.height(16.dp))
            PhoneField(
                region = region,
                hongKong = region == FlowLogic.REGION_HONG_KONG,
                number = FlowLogic.formatPhone(region, phone),
                onRegion = onRegionSheet,
                onNumber = onPhone,
            )
            Spacer(Modifier.height(16.dp))
            PrimaryButton("继续", enabled = FlowLogic.phoneContinueEnabled(phone) && !busy, onClick = onContinue)
            if (error.isNotEmpty()) {
                Spacer(Modifier.height(8.dp))
                Text(error, style = qlutchText(13, QlutchTone.footer, WelcomeType.notoRegular, FontWeight.Normal))
            }
        }
        Spacer(Modifier.weight(1f))
        Column(
            Modifier
                .fillMaxWidth()
                .padding(start = 16.dp, end = 16.dp, top = 8.dp, bottom = 28.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            HairlineLabel("其他方式")
            ClosedChannel("使用微信登录", { WeChatMark() }, onClosed)
            ClosedChannel("使用支付宝登录", { AlipayMark() }, onClosed)
            Spacer(Modifier.height(8.dp))
            LanguageChip(if (locale == FlowLogic.LOCALE_HANT) "繁体中文" else "简体中文", onLanguageSheet)
        }
    }
}

@Composable
private fun TabChip(label: String, selected: Boolean, onClick: () -> Unit) {
    Box(
        modifier = Modifier
            .height(40.dp)
            .clip(RoundedCornerShape(20.dp))
            .background(if (selected) QlutchTone.white else Color.Transparent)
            .clickable(onClick = onClick)
            .padding(horizontal = 22.dp),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            label,
            style = qlutchText(
                16,
                if (selected) QlutchTone.ink else QlutchTone.slate,
                WelcomeType.notoSemiBold,
                FontWeight.SemiBold,
            ),
        )
    }
}

@Composable
private fun OtpScreen(
    code: String,
    line: String,
    error: String,
    busy: Boolean,
    onBack: () -> Unit,
    onCode: (String) -> Unit,
    onContinue: () -> Unit,
    onResend: () -> Unit,
) {
    Column(Modifier.fillMaxSize()) {
        Box(Modifier.padding(start = 16.dp, end = 16.dp, top = 16.dp)) { BackButton(onBack) }
        Column(
            Modifier
                .weight(1f)
                .fillMaxWidth()
                .padding(start = 16.dp, end = 16.dp, top = 88.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(20.dp),
        ) {
            Box(Modifier.size(80.dp), contentAlignment = Alignment.Center) {
                Text("✈", style = qlutchText(36, QlutchTone.brand, WelcomeType.notoRegular, FontWeight.Normal))
            }
            ScreenTitle("查看你的短信")
            OtpCells(code)
            androidx.compose.foundation.text.BasicTextField(
                value = code,
                onValueChange = onCode,
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                modifier = Modifier.size(1.dp),
            )
            Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(6.dp)) {
                Text(line, style = qlutchText(14, QlutchTone.slate, WelcomeType.notoRegular, FontWeight.Normal))
                Text("请查看短信", style = qlutchText(14, QlutchTone.ink, WelcomeType.notoSemiBold, FontWeight.SemiBold))
                if (error.isNotEmpty()) {
                    Text(error, style = qlutchText(13, QlutchTone.footer, WelcomeType.notoRegular, FontWeight.Normal))
                }
            }
        }
        Column(
            Modifier
                .fillMaxWidth()
                .padding(start = 16.dp, end = 16.dp, bottom = 28.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            PrimaryButton("继续", enabled = FlowLogic.otpContinueEnabled(code) && !busy, onClick = onContinue)
            SecondaryButton("没有收到验证码", enabled = !busy, onClick = onResend)
        }
    }
}

@Composable
private fun NoAccountScreen(
    region: String,
    phone: String,
    busy: Boolean,
    error: String,
    onCreate: () -> Unit,
    onOther: () -> Unit,
) {
    Column(Modifier.fillMaxSize()) {
        Column(
            Modifier
                .weight(1f)
                .fillMaxWidth()
                .padding(horizontal = 16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp, Alignment.CenterVertically),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Text("该手机号还没有账号", style = qlutchText(22, QlutchTone.ink, WelcomeType.notoExtraBold, FontWeight.Bold))
            Text(
                "用 $region ${FlowLogic.formatPhone(region, phone)} 创建账号，\n或换一种方式登录。",
                style = qlutchText(16, QlutchTone.ink, WelcomeType.notoRegular, FontWeight.Normal),
            )
            if (error.isNotEmpty()) {
                Text(error, style = qlutchText(13, QlutchTone.footer, WelcomeType.notoRegular, FontWeight.Normal))
            }
        }
        Column(
            Modifier.padding(start = 16.dp, end = 16.dp, bottom = 28.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            PrimaryButton("创建账号", enabled = !busy, onClick = onCreate)
            SecondaryButton("用其他方式登录", onClick = onOther)
        }
    }
}

@Composable
private fun NicknameScreen(
    nickname: String,
    error: String,
    busy: Boolean,
    onBack: () -> Unit,
    onNickname: (String) -> Unit,
    onContinue: () -> Unit,
) {
    Column(Modifier.fillMaxSize()) {
        FlowTopBar(1, onBack)
        PromptRow("先给自己起个昵称")
        Column(
            Modifier
                .fillMaxWidth()
                .padding(start = 16.dp, end = 16.dp, top = 16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            LineField(nickname, onNickname, "小辉")
            PrimaryButton("继续", enabled = FlowLogic.nicknameContinueEnabled(nickname) && !busy, onClick = onContinue)
            if (error.isNotEmpty()) {
                Text(error, style = qlutchText(13, QlutchTone.footer, WelcomeType.notoRegular, FontWeight.Normal))
            }
        }
    }
}

@Composable
private fun QuestionScreen(
    step: QuestionStep,
    selected: String,
    error: String,
    busy: Boolean,
    onBack: () -> Unit,
    onSelect: (String) -> Unit,
    onFinish: () -> Unit,
) {
    Column(Modifier.fillMaxSize()) {
        FlowTopBar(step.index, onBack)
        PromptRow(step.title)
        Spacer(Modifier.height(152.dp))
        Column(
            Modifier.padding(horizontal = 16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            step.choices.forEach { label ->
                ChoiceRow(label, selected = label == selected, onClick = { if (!busy) onSelect(label) })
            }
            if (error.isNotEmpty()) {
                Text(error, style = qlutchText(13, QlutchTone.footer, WelcomeType.notoRegular, FontWeight.Normal))
            }
        }
        Spacer(Modifier.weight(1f))
        if (step.finishLabel != null) {
            Box(Modifier.padding(start = 16.dp, end = 16.dp, top = 8.dp, bottom = 34.dp)) {
                PrimaryButton(step.finishLabel, enabled = selected.isNotEmpty() && !busy, onClick = onFinish)
            }
        }
    }
}

@Composable
private fun PermissionContactsScreen(
    contacts: Boolean,
    notices: Boolean,
    onContacts: (Boolean) -> Unit,
    onNotices: (Boolean) -> Unit,
    onBack: () -> Unit,
    onContinue: () -> Unit,
    onSkip: () -> Unit,
) {
    val context = LocalContext.current
    val launcher = rememberLauncherForActivityResult(ActivityResultContracts.RequestMultiplePermissions()) {
        onContinue()
    }
    Column(Modifier.fillMaxSize()) {
        FlowTopBar(7, onBack)
        PromptRow("先开几个权限", "开了学习更顺手，跳过也没问题")
        Column(
            Modifier
                .weight(1f)
                .padding(horizontal = 16.dp, vertical = 8.dp),
            verticalArrangement = Arrangement.Center,
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Column(verticalArrangement = Arrangement.spacedBy(28.dp), horizontalAlignment = Alignment.CenterHorizontally) {
                Image(
                    painter = painterResource(R.drawable.flow_perm_contacts),
                    contentDescription = null,
                    modifier = Modifier.fillMaxWidth().height(186.dp),
                    contentScale = ContentScale.Fit,
                )
                Column(
                    Modifier
                        .fillMaxWidth()
                        .clip(RoundedCornerShape(22.dp))
                        .background(QlutchTone.white),
                ) {
                    PermissionLine("开启通讯录", "帮你找到也在用的同学", contacts) { onContacts(!contacts) }
                    Box(Modifier.fillMaxWidth().height(1.dp).background(QlutchTone.line))
                    PermissionLine("开启消息通知", "到点提醒你学习，别把计划忘了", notices) { onNotices(!notices) }
                }
            }
        }
        Column(
            Modifier.padding(start = 16.dp, end = 16.dp, bottom = 34.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            PrimaryButton("继续") {
                val wants = mutableListOf<String>()
                if (contacts) wants += Manifest.permission.READ_CONTACTS
                if (notices && Build.VERSION.SDK_INT >= 33) wants += Manifest.permission.POST_NOTIFICATIONS
                val missing = wants.filter {
                    ContextCompat.checkSelfPermission(context, it) != PackageManager.PERMISSION_GRANTED
                }
                if (missing.isEmpty()) onContinue() else launcher.launch(missing.toTypedArray())
            }
            SecondaryButton("跳过", onClick = onSkip)
        }
    }
}

@Composable
private fun PermissionLine(title: String, body: String, checked: Boolean, onToggle: () -> Unit) {
    Row(
        Modifier.fillMaxWidth().height(100.dp).padding(horizontal = 16.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Box(
            Modifier.size(40.dp).clip(CircleShape).background(QlutchTone.slot),
            contentAlignment = Alignment.Center,
        ) {
            Text(if (title.contains("通讯录")) "人" else "铃", style = qlutchText(14, QlutchTone.ink, WelcomeType.notoSemiBold, FontWeight.SemiBold))
        }
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Text(title, style = qlutchText(16, QlutchTone.ink, WelcomeType.notoSemiBold, FontWeight.SemiBold, TextAlign.Start, 22))
            Text(body, style = qlutchText(13, QlutchTone.slate, qlutchMedium, FontWeight.Medium, TextAlign.Start, 18))
        }
        PermissionCheck(checked, onToggle)
    }
}

@Composable
private fun PermissionLocationScreen(onBack: () -> Unit, onContinue: () -> Unit, onSkip: () -> Unit) {
    val context = LocalContext.current
    val launcher = rememberLauncherForActivityResult(ActivityResultContracts.RequestMultiplePermissions()) {
        onContinue()
    }
    Column(Modifier.fillMaxSize()) {
        FlowTopBar(7, onBack)
        PromptRow("找附近的学校或同学", "开一下定位就能看到，跳过也没问题")
        Column(
            Modifier.weight(1f).padding(horizontal = 16.dp),
            verticalArrangement = Arrangement.Center,
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Column(verticalArrangement = Arrangement.spacedBy(20.dp)) {
                Image(
                    painter = painterResource(R.drawable.flow_perm_location),
                    contentDescription = null,
                    modifier = Modifier.fillMaxWidth().height(186.dp),
                    contentScale = ContentScale.Fit,
                )
                Column(
                    Modifier
                        .fillMaxWidth()
                        .clip(RoundedCornerShape(22.dp))
                        .background(QlutchTone.white)
                        .padding(16.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    Text("查找附近的学校或同学", style = qlutchText(16, QlutchTone.ink, WelcomeType.notoSemiBold, FontWeight.SemiBold, TextAlign.Start, 22))
                }
                Text(
                    "定位只用来找附近的学校或同学",
                    style = qlutchText(13, QlutchTone.footer, WelcomeType.notoSemiBold, FontWeight.SemiBold),
                )
            }
        }
        Column(
            Modifier.padding(start = 16.dp, end = 16.dp, bottom = 34.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            PrimaryButton("继续") {
                val perms = arrayOf(Manifest.permission.ACCESS_COARSE_LOCATION, Manifest.permission.ACCESS_FINE_LOCATION)
                val missing = perms.filter {
                    ContextCompat.checkSelfPermission(context, it) != PackageManager.PERMISSION_GRANTED
                }
                if (missing.isEmpty()) onContinue() else launcher.launch(missing.toTypedArray())
            }
            SecondaryButton("跳过", onClick = onSkip)
        }
    }
}

@Composable
private fun PackScreen(onBack: () -> Unit, onPhoto: () -> Unit, onMore: () -> Unit) {
    val context = LocalContext.current
    val launcher = rememberLauncherForActivityResult(ActivityResultContracts.RequestPermission()) { granted ->
        if (granted) onPhoto()
    }
    Column(Modifier.fillMaxSize(), horizontalAlignment = Alignment.CenterHorizontally) {
        FlowTopBar(7, onBack)
        PromptRow("创建第一个知识包", "先把一份教材或课件交进来")
        Spacer(Modifier.weight(1f))
        Column(
            Modifier
                .width(358.dp)
                .height(190.dp)
                .clip(RoundedCornerShape(24.dp)),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(16.dp, Alignment.CenterVertically),
        ) {
            Box(Modifier.size(64.dp).clip(CircleShape).background(QlutchTone.slot), contentAlignment = Alignment.Center) {
                Text("拍", style = qlutchText(16, QlutchTone.ink, WelcomeType.notoSemiBold, FontWeight.SemiBold))
            }
            Text("拍摄教材或课件", style = qlutchText(16, QlutchTone.ink, WelcomeType.notoSemiBold, FontWeight.SemiBold, lineSp = 22))
        }
        Spacer(Modifier.weight(1f))
        Column(
            Modifier.padding(start = 16.dp, end = 16.dp, bottom = 34.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            FooterNote("教材不用于训练模型")
            PrimaryButton("拍照") {
                val granted = ContextCompat.checkSelfPermission(context, Manifest.permission.CAMERA) == PackageManager.PERMISSION_GRANTED
                if (granted) onPhoto() else launcher.launch(Manifest.permission.CAMERA)
            }
            SecondaryButton("更多导入方式", onClick = onMore)
        }
    }
}

@Composable
private fun MoreScreen(
    toast: Boolean,
    onBack: () -> Unit,
    onUpload: () -> Unit,
    onClosed: () -> Unit,
    onPaste: () -> Unit,
    onNoMaterial: () -> Unit,
) {
    val picker = rememberLauncherForActivityResult(ActivityResultContracts.GetContent()) { uri ->
        if (uri != null) onUpload()
    }
    Box(Modifier.fillMaxSize()) {
        Column(Modifier.fillMaxSize(), horizontalAlignment = Alignment.CenterHorizontally) {
            FlowTopBar(7, onBack)
            PromptRow("用其他方式导入", "把教材或课件交进来")
            Column(
                Modifier
                    .weight(1f)
                    .padding(start = 16.dp, end = 16.dp, top = 24.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Column(
                    Modifier
                        .width(358.dp)
                        .height(190.dp)
                        .clickable { picker.launch("*/*") },
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.spacedBy(16.dp, Alignment.CenterVertically),
                ) {
                    Box(Modifier.size(64.dp).clip(CircleShape).background(QlutchTone.slot), contentAlignment = Alignment.Center) {
                        Text("传", style = qlutchText(16, QlutchTone.ink, WelcomeType.notoSemiBold, FontWeight.SemiBold))
                    }
                    Text("上传文件", style = qlutchText(16, QlutchTone.ink, WelcomeType.notoSemiBold, FontWeight.SemiBold))
                    Text("PDF、PPT、Word、图片或音频", style = qlutchText(13, QlutchTone.footer, WelcomeType.notoRegular, FontWeight.Normal))
                }
                HairlineLabel("或")
                Row(Modifier.width(358.dp), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                    ImportChip("抖音", Modifier.weight(1f), onClosed)
                    ImportChip("B站", Modifier.weight(1f), onClosed)
                }
                ImportChip("粘贴笔记", Modifier.width(358.dp), onPaste)
            }
            Box(Modifier.padding(start = 16.dp, end = 16.dp, bottom = 34.dp)) {
                SecondaryButton("我没有材料", onClick = onNoMaterial)
            }
        }
        Box(Modifier.fillMaxSize().padding(bottom = 75.dp), contentAlignment = Alignment.BottomCenter) {
            ClosedToast(toast)
        }
    }
}

@Composable
private fun PasteScreen(notes: String, onNotes: (String) -> Unit, onBack: () -> Unit, onContinue: () -> Unit) {
    Column(Modifier.fillMaxSize(), horizontalAlignment = Alignment.CenterHorizontally) {
        FlowTopBar(7, onBack)
        PromptRow("把笔记粘过来", "我会从里面挖掘知识点")
        Column(Modifier.weight(1f).padding(start = 16.dp, end = 16.dp, top = 24.dp)) {
            NotesField(notes, onNotes)
        }
        Box(Modifier.padding(start = 16.dp, end = 16.dp, bottom = 34.dp)) {
            PrimaryButton("继续", enabled = FlowLogic.notesContinueEnabled(notes), onClick = onContinue)
        }
    }
}

@Composable
private fun WaitScreen(
    startedAt: Long,
    now: Long,
    onBack: () -> Unit,
    onArm: () -> Unit,
    onTick: () -> Unit,
    onFinished: () -> Unit,
) {
    var armed by remember { mutableIntStateOf(0) }
    LaunchedEffect(Unit) {
        if (armed == 0) {
            armed = 1
            onArm()
        }
    }
    LaunchedEffect(startedAt) {
        if (startedAt == 0L) return@LaunchedEffect
        while (!FlowLogic.mapGateOpen(android.os.SystemClock.elapsedRealtime() - startedAt)) {
            onTick()
            delay(50)
        }
        onTick()
        onFinished()
    }
    val elapsed = if (startedAt == 0L) 0L else (now - startedAt).coerceAtLeast(0L)
    Column(Modifier.fillMaxSize(), horizontalAlignment = Alignment.CenterHorizontally) {
        FlowTopBar(7, onBack)
        PromptRow(FlowLogic.WAIT_TITLE, FlowLogic.WAIT_CAPTION)
        Column(
            Modifier
                .weight(1f)
                .padding(horizontal = 16.dp),
            verticalArrangement = Arrangement.spacedBy(20.dp, Alignment.CenterVertically),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Box(
                Modifier
                    .width(358.dp)
                    .height(626.dp)
                    .clip(RoundedCornerShape(0.dp)),
            ) {
                Image(
                    painter = painterResource(R.drawable.flow_reading_still),
                    contentDescription = null,
                    modifier = Modifier.fillMaxSize(),
                    contentScale = ContentScale.Fit,
                )
                ReadingGif(Modifier.fillMaxSize())
            }
            Text(
                FlowLogic.WAIT_FOOTNOTE,
                modifier = Modifier.width(358.dp),
                style = qlutchText(16, QlutchTone.slate, WelcomeType.notoSemiBold, FontWeight.SemiBold, lineSp = 22),
            )
            WaitMeter(FlowLogic.mapFraction(elapsed))
        }
    }
}

@Composable
private fun ReadingGif(modifier: Modifier) {
    val context = LocalContext.current
    AndroidView(
        modifier = modifier,
        factory = { viewContext ->
            ImageView(viewContext).apply {
                scaleType = ImageView.ScaleType.FIT_CENTER
                val source = ImageDecoder.createSource(context.resources, R.raw.reading_2s)
                val drawable = ImageDecoder.decodeDrawable(source)
                if (drawable is AnimatedImageDrawable) {
                    drawable.repeatCount = AnimatedImageDrawable.REPEAT_INFINITE
                }
                setImageDrawable(drawable)
                (drawable as? Animatable)?.start()
            }
        },
    )
}
