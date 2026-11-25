import SwiftUI
import UIKit
import Contacts
import UserNotifications
import CoreLocation
import AVFoundation
import UniformTypeIdentifiers

struct FlowView: View {
    let purpose: String
    let onHome: () -> Void
    let onLeave: () -> Void

    @State private var route: FlowRoute = .auth
    @State private var stack: [FlowRoute] = []
    @State private var mode: String
    @State private var region = FlowLogic.regionMainland
    @State private var locale = FlowLogic.localeHans
    @State private var phone = ""
    @State private var code = ""
    @State private var hint: String?
    @State private var nickname = ""
    @State private var selected = ""
    @State private var notes = "光合作用：绿色植物利用光能，把二氧化碳和水变成有机物，并放出氧气。\n这是本章要掌握的核心过程。"
    @State private var contacts = false
    @State private var notices = true
    @State private var permissionsSeen = false
    @State private var bearer = ""
    @State private var error = ""
    @State private var busy = false
    @State private var sheet = ""
    @State private var search = ""
    @State private var toast = false
    @State private var waitStart: Date?
    @State private var releasedHome = false
    @State private var waitEntry = 0
    @State private var jobId = ""
    @State private var packId = ""
    @State private var materialId = ""
    @State private var showCamera = false
    @State private var waitRetryable = false
    @State private var showImporter = false
    @StateObject private var location = LocationAsk()

    private let api = IdentityClient(baseURL: ApiConfig.baseURL)

    init(purpose: String, onHome: @escaping () -> Void, onLeave: @escaping () -> Void) {
        self.purpose = purpose
        self.onHome = onHome
        self.onLeave = onLeave
        _mode = State(initialValue: purpose)
    }

    var body: some View {
        ZStack {
            QlutchTone.page.ignoresSafeArea()
            VStack(spacing: 0) {
                screen
            }
            .frame(maxWidth: 390)
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            if sheet == "region" { regionSheet }
            if sheet == "language" { languageSheet }
        }
        .fileImporter(isPresented: $showImporter, allowedContentTypes: [.item]) { result in
            if case .success(let url) = result {
                let started = url.startAccessingSecurityScopedResource()
                defer { if started { url.stopAccessingSecurityScopedResource() } }
                if let data = try? Data(contentsOf: url) {
                    let type = url.pathExtension.isEmpty ? "application/octet-stream" : mimeType(url.pathExtension)
                    submitMaterial(source: "file", contentType: type, bytes: data, filename: url.lastPathComponent, waitSource: .more)
                }
            }
        }
        .sheet(isPresented: $showCamera) {
            CameraCapture { data in
                showCamera = false
                submitMaterial(source: "camera", contentType: "image/jpeg", bytes: data, filename: nil, waitSource: .pack)
            }
        }
    }

    @ViewBuilder private var screen: some View {
        switch route {
        case .auth: auth
        case .otp: otp
        case .noAccount: noAccount
        case .nickname: nicknameScreen
        case .question(let index): question(FlowLogic.question(index))
        case .permissionsContacts: contactsScreen
        case .permissionsLocation: locationScreen
        case .pack: pack
        case .more: more
        case .paste: paste
        case .wait: wait
        case .home: EmptyView()
        }
    }

    private var auth: some View {
        let login = mode == FlowLogic.purposeLogin
        return VStack(spacing: 0) {
            VStack(spacing: 0) {
                bundledImage("wordmark_small", 111, 28).padding(.top, 48)
                Text("把课件或教材上传给AI，挖掘知识点生成题卡，每天完成一组题卡或与好友PK抢答题卡，强化知识点记忆")
                    .font(qlutchFont("NotoSansSC-Regular", 13))
                    .foregroundStyle(QlutchTone.footer)
                    .multilineTextAlignment(.center)
                    .frame(maxWidth: 358)
                    .padding(.top, 12)
                Text("开始吧！").font(qlutchFont("NotoSansSC-ExtraBold", 32)).foregroundStyle(QlutchTone.ink).padding(.top, 20)
                HStack(spacing: 0) {
                    tab("注册", !login) { mode = FlowLogic.purposeRegister }
                    tab("登录", login) { mode = FlowLogic.purposeLogin }
                }
                .padding(4)
                .background(QlutchTone.line)
                .clipShape(Capsule())
                .padding(.top, 32)
                Text(login ? "用手机号登录" : "用手机号注册")
                    .font(qlutchFont("NotoSansSC-SemiBold", 16))
                    .padding(.top, 36)
                PhoneField(region: region, hongKong: region == FlowLogic.regionHongKong, digits: $phone) {
                    search = ""; sheet = "region"
                }
                .padding(.top, 16)
                PrimaryButton(label: "继续", enabled: FlowLogic.phoneContinueEnabled(phone) && !busy) { sendCode(pushOtp: true) }
                    .padding(.top, 16)
                if !error.isEmpty { Text(error).font(qlutchFont("NotoSansSC-Regular", 13)).foregroundStyle(QlutchTone.footer).padding(.top, 8) }
            }
            .padding(.horizontal, 16)
            Spacer()
            VStack(spacing: 12) {
                Text("其他方式").font(qlutchFont("NotoSansSC-Regular", 13)).foregroundStyle(QlutchTone.slate)
                ClosedChannel(label: "使用微信登录", mark: "微") { showToast() }
                ClosedChannel(label: "使用支付宝登录", mark: "支") { showToast() }
                LanguageChip(label: locale == FlowLogic.localeHant ? "繁体中文" : "简体中文") {
                    search = ""; sheet = "language"
                }
                .padding(.top, 8)
            }
            .padding(.horizontal, 16)
            .padding(.bottom, 28)
        }
    }

    private var otp: some View {
        VStack {
            HStack { BackButton { back() }; Spacer() }.padding(.horizontal, 16).padding(.top, 16)
            VStack(spacing: 20) {
                Text("✈").font(.system(size: 36)).foregroundStyle(QlutchTone.brand).frame(width: 80, height: 80)
                Text("查看你的短信").font(qlutchFont("NotoSansSC-ExtraBold", 32))
                OtpCells(code: code)
                TextField("", text: Binding(get: { code }, set: { code = String(FlowLogic.phoneDigits($0).prefix(6)) }))
                    .keyboardType(.numberPad)
                    .frame(width: 1, height: 1)
                    .opacity(0.01)
                Text(FlowLogic.sentToLine(hint: hint, region: region, digits: phone))
                    .font(qlutchFont("NotoSansSC-Regular", 14)).foregroundStyle(QlutchTone.slate)
                Text("请查看短信").font(qlutchFont("NotoSansSC-SemiBold", 14))
                if !error.isEmpty { Text(error).font(qlutchFont("NotoSansSC-Regular", 13)).foregroundStyle(QlutchTone.footer) }
            }
            .padding(.top, 88)
            Spacer()
            VStack(spacing: 16) {
                PrimaryButton(label: "继续", enabled: FlowLogic.otpContinueEnabled(code) && !busy, action: verify)
                SecondaryButton(label: "没有收到验证码", enabled: !busy) { sendCode(pushOtp: false) }
            }
            .padding(.horizontal, 16).padding(.bottom, 28)
        }
    }

    private var noAccount: some View {
        VStack {
            Spacer()
            Text("该手机号还没有账号").font(qlutchFont("NotoSansSC-ExtraBold", 22))
            Text("用 \(region) \(FlowLogic.formatPhone(region: region, digits: phone)) 创建账号，\n或换一种方式登录。")
                .font(qlutchFont("NotoSansSC-Regular", 16))
                .multilineTextAlignment(.center)
                .padding(.top, 12)
            if !error.isEmpty { Text(error).foregroundStyle(QlutchTone.footer).padding(.top, 8) }
            Spacer()
            VStack(spacing: 16) {
                PrimaryButton(label: "创建账号", enabled: !busy) {
                    mode = FlowLogic.purposeRegister
                    sendCode(pushOtp: true)
                }
                SecondaryButton(label: "用其他方式登录") {
                    mode = FlowLogic.purposeLogin
                    route = .auth
                }
            }
            .padding(.horizontal, 16).padding(.bottom, 28)
        }
    }

    private var nicknameScreen: some View {
        VStack(alignment: .leading, spacing: 0) {
            FlowTopBar(step: 1, onBack: back)
            PromptRow(title: "先给自己起个昵称")
            TextField("小辉", text: $nickname)
                .font(qlutchFont("NotoSansSC-Regular", 15))
                .padding(.horizontal, 20)
                .frame(height: 50)
                .background(QlutchTone.white)
                .clipShape(Capsule())
                .overlay(Capsule().stroke(QlutchTone.line, lineWidth: 1))
                .padding(.horizontal, 16)
                .padding(.top, 16)
            PrimaryButton(label: "继续", enabled: FlowLogic.nicknameContinueEnabled(nickname) && !busy) {
                Task { await write(stepId: "nickname", nickname: nickname.trimmingCharacters(in: .whitespacesAndNewlines), answer: nil) { push(.question(2)) } }
            }
            .padding(.horizontal, 16).padding(.top, 16)
            if !error.isEmpty { Text(error).padding(.horizontal, 16).padding(.top, 8) }
            Spacer()
        }
    }

    private func question(_ step: QuestionStep) -> some View {
        VStack(spacing: 0) {
            FlowTopBar(step: step.index, onBack: back)
            PromptRow(title: step.title)
            Spacer().frame(height: 152)
            VStack(spacing: 12) {
                ForEach(step.choices, id: \.self) { label in
                    ChoiceRow(label: label, selected: selected == label) {
                        selected = label
                        if step.finishLabel == nil && !busy {
                            Task { await write(stepId: step.stepId, nickname: nil, answer: label) {
                                selected = ""
                                push(.question(step.index + 1))
                            } }
                        }
                    }
                }
                if !error.isEmpty { Text(error).font(qlutchFont("NotoSansSC-Regular", 13)).foregroundStyle(QlutchTone.footer) }
            }
            .padding(.horizontal, 16)
            Spacer()
            if let finish = step.finishLabel {
                PrimaryButton(label: finish, enabled: !selected.isEmpty && !busy) {
                    Task { await write(stepId: step.stepId, nickname: nil, answer: selected) {} }
                }
                .padding(.horizontal, 16).padding(.bottom, 34)
            }
        }
    }

    private var contactsScreen: some View {
        VStack(spacing: 0) {
            FlowTopBar(step: 7, onBack: back)
            PromptRow(title: "先开几个权限", subtitle: "开了学习更顺手，跳过也没问题")
            Spacer()
            bundledImage("perm_contacts", 358, 186)
            permissionCard
            Spacer()
            VStack(spacing: 16) {
                PrimaryButton(label: "继续", action: requestChecked)
                SecondaryButton(label: "跳过") { push(.permissionsLocation) }
            }
            .padding(.horizontal, 16).padding(.bottom, 34)
        }
    }

    private var permissionCard: some View {
        VStack(spacing: 0) {
            permLine("开启通讯录", "帮你找到也在用的同学", contacts) { contacts.toggle() }
            Rectangle().fill(QlutchTone.line).frame(height: 1)
            permLine("开启消息通知", "到点提醒你学习，别把计划忘了", notices) { notices.toggle() }
        }
        .background(QlutchTone.white)
        .clipShape(RoundedRectangle(cornerRadius: 22))
        .overlay(RoundedRectangle(cornerRadius: 22).stroke(QlutchTone.line, lineWidth: 1))
        .padding(.horizontal, 16)
        .padding(.top, 28)
    }

    private func permLine(_ title: String, _ body: String, _ checked: Bool, _ toggle: @escaping () -> Void) -> some View {
        HStack(spacing: 12) {
            Text(title.contains("通讯录") ? "人" : "铃")
                .frame(width: 40, height: 40)
                .background(QlutchTone.slot)
                .clipShape(Circle())
            VStack(alignment: .leading, spacing: 2) {
                Text(title).font(qlutchFont("NotoSansSC-SemiBold", 16))
                Text(body).font(qlutchFont("NotoSansSC-Medium", 13)).foregroundStyle(QlutchTone.slate)
            }
            Spacer()
            PermissionCheck(checked: checked, action: toggle)
        }
        .padding(.horizontal, 16)
        .frame(height: 100)
    }

    private var locationScreen: some View {
        VStack(spacing: 0) {
            FlowTopBar(step: 7, onBack: back)
            PromptRow(title: "找附近的学校或同学", subtitle: "开一下定位就能看到，跳过也没问题")
            Spacer()
            bundledImage("perm_location", 358, 186)
            Text("查找附近的学校或同学")
                .font(qlutchFont("NotoSansSC-SemiBold", 16))
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(16)
                .background(QlutchTone.white)
                .clipShape(RoundedRectangle(cornerRadius: 22))
                .padding(.horizontal, 16)
                .padding(.top, 20)
            Text("定位只用来找附近的学校或同学")
                .font(qlutchFont("NotoSansSC-SemiBold", 13))
                .foregroundStyle(QlutchTone.footer)
                .padding(.top, 12)
            Spacer()
            VStack(spacing: 16) {
                PrimaryButton(label: "继续") {
                    location.ask { permissionsSeen = true; push(.pack) }
                }
                SecondaryButton(label: "跳过") { permissionsSeen = true; push(.pack) }
            }
            .padding(.horizontal, 16).padding(.bottom, 34)
        }
    }

    private var pack: some View {
        VStack(spacing: 0) {
            FlowTopBar(step: 7, onBack: back)
            PromptRow(title: "创建第一个知识包", subtitle: "先把一份教材或课件交进来")
            Spacer()
            VStack(spacing: 16) {
                Text("拍").frame(width: 64, height: 64).background(QlutchTone.slot).clipShape(Circle())
                Text("拍摄教材或课件").font(qlutchFont("NotoSansSC-SemiBold", 16))
            }
            .frame(width: 358, height: 190)
            Spacer()
            Text("教材不用于训练模型").font(qlutchFont("NotoSansSC-Regular", 13)).foregroundStyle(QlutchTone.footer)
            VStack(spacing: 16) {
                PrimaryButton(label: "拍照") {
                    AVCaptureDevice.requestAccess(for: .video) { granted in
                        if granted { DispatchQueue.main.async { showCamera = true } }
                    }
                }
                SecondaryButton(label: "更多导入方式") { push(.more) }
            }
            .padding(.horizontal, 16).padding(.bottom, 34).padding(.top, 12)
        }
    }

    private var more: some View {
        ZStack(alignment: .bottom) {
            VStack(spacing: 0) {
                FlowTopBar(step: 7, onBack: back)
                PromptRow(title: "用其他方式导入", subtitle: "把教材或课件交进来")
                VStack(spacing: 16) {
                    Button { showImporter = true } label: {
                        VStack(spacing: 8) {
                            Text("传").frame(width: 64, height: 64).background(QlutchTone.slot).clipShape(Circle())
                            Text("上传文件").font(qlutchFont("NotoSansSC-SemiBold", 16)).foregroundStyle(QlutchTone.ink)
                            Text("PDF、PPT、Word、图片或音频").font(qlutchFont("NotoSansSC-Regular", 13)).foregroundStyle(QlutchTone.footer)
                        }
                        .frame(width: 358, height: 190)
                    }
                    .buttonStyle(.plain)
                    Text("或").font(qlutchFont("NotoSansSC-Regular", 13)).foregroundStyle(QlutchTone.footer)
                    HStack(spacing: 12) {
                        chip("抖音") { showToast() }
                        chip("B站") { showToast() }
                    }
                    chip("粘贴笔记") { push(.paste) }.frame(width: 358)
                }
                .padding(.top, 24)
                Spacer()
                if !error.isEmpty {
                    Text(error)
                        .font(qlutchFont("NotoSansSC-Regular", 13))
                        .foregroundStyle(QlutchTone.footer)
                        .padding(.horizontal, 16)
                }
                SecondaryButton(label: "我没有材料", enabled: !busy, action: skipNoMaterial)
                    .padding(.horizontal, 16).padding(.bottom, 34)
            }
            ClosedToast(visible: toast).padding(.bottom, 75)
        }
    }

    private var paste: some View {
        VStack(spacing: 0) {
            FlowTopBar(step: 7, onBack: back)
            PromptRow(title: "把笔记粘过来", subtitle: "我会从里面挖掘知识点")
            TextEditor(text: $notes)
                .font(qlutchFont("NotoSansSC-Regular", 15))
                .scrollContentBackground(.hidden)
                .padding(12)
                .frame(height: 220)
                .background(QlutchTone.white)
                .clipShape(RoundedRectangle(cornerRadius: 22))
                .overlay(RoundedRectangle(cornerRadius: 22).stroke(QlutchTone.line, lineWidth: 1))
                .padding(.horizontal, 16)
                .padding(.top, 24)
            Spacer()
            if !error.isEmpty {
                Text(error)
                    .font(qlutchFont("NotoSansSC-Regular", 13))
                    .foregroundStyle(QlutchTone.footer)
            }
            PrimaryButton(label: "继续", enabled: FlowLogic.notesContinueEnabled(notes) && !busy) {
                submitMaterial(source: "paste_note", contentType: "text/plain", bytes: Data(notes.utf8), filename: nil, waitSource: .paste)
            }
                .padding(.horizontal, 16).padding(.bottom, 34)
        }
    }

    private var wait: some View {
        VStack(spacing: 0) {
            FlowTopBar(step: 7, onBack: back)
            PromptRow(title: FlowLogic.waitTitle, subtitle: FlowLogic.waitCaption)
            Spacer()
            ReadingStage()
                .id(waitEntry)
            Text(FlowLogic.waitFootnote)
                .font(qlutchFont("NotoSansSC-SemiBold", 16))
                .foregroundStyle(QlutchTone.slate)
                .multilineTextAlignment(.center)
                .frame(width: 358)
                .padding(.top, 20)
            Capsule()
                .fill(QlutchTone.line)
                .frame(width: 248, height: 8)
                .overlay(alignment: .leading) {
                    Capsule()
                        .fill(LinearGradient(colors: [QlutchTone.brand, Color(red: 210 / 255, green: 68 / 255, blue: 238 / 255)], startPoint: .leading, endPoint: .trailing))
                        .frame(width: 248 * FlowLogic.waitMeterFraction(), height: 8)
                }
                .padding(.top, 20)
            if !error.isEmpty {
                Text(error)
                    .font(qlutchFont("NotoSansSC-Regular", 13))
                    .foregroundStyle(QlutchTone.footer)
                    .padding(.top, 12)
            }
            if waitRetryable {
                SecondaryButton(label: "再试一次", enabled: !busy, action: retryImport)
                    .padding(.horizontal, 16)
                    .padding(.top, 12)
            }
            Spacer()
        }
        .onAppear {
            waitStart = Date()
            releasedHome = false
        }
        .onDisappear {
            waitStart = nil
            releasedHome = false
            waitRetryable = false
        }
        .task(id: waitEntry) {
            guard !jobId.isEmpty else { return }
            while !Task.isCancelled {
                let outcome = await api.getJob(bearer: bearer, jobId: jobId)
                if Task.isCancelled { return }
                switch outcome {
                case .ok(let job):
                    waitRetryable = FlowLogic.waitAllowsRetry(job.status, retryable: job.retryable)
                    if FlowLogic.waitReleasesHome(job.status) {
                        onHome()
                        return
                    }
                    if job.status == FlowLogic.jobFailedRetryable || job.status == FlowLogic.jobFailedTerminal {
                        error = (job.failure_code?.isEmpty == false) ? job.failure_code! : job.status
                        return
                    }
                    error = ""
                case .err(_, let code):
                    error = code
                case .transport(let reason):
                    error = reason
                }
                try? await Task.sleep(nanoseconds: UInt64(FlowLogic.jobPollIntervalMillis * 1_000_000))
            }
        }
    }

    private var regionSheet: some View {
        OptionSheet(title: "选择国家或地区", searchHint: "搜索国家或区号", query: $search, onClose: { sheet = "" }) {
            ForEach(FlowLogic.filterRegions(search), id: \.code) { choice in
                Button {
                    region = choice.code
                    sheet = ""
                } label: {
                    HStack {
                        FlagMark(hongKong: choice.code == FlowLogic.regionHongKong, large: true)
                        VStack(alignment: .leading) {
                            Text(choice.name).font(qlutchFont("NotoSansSC-SemiBold", 16)).foregroundStyle(QlutchTone.ink)
                            Text(choice.code).font(qlutchFont("NotoSansSC-Regular", 13)).foregroundStyle(QlutchTone.slate)
                        }
                        Spacer()
                        Image(systemName: region == choice.code ? "largecircle.fill.circle" : "circle")
                    }
                    .padding(.vertical, 12)
                }
                .buttonStyle(.plain)
            }
        }
    }

    private var languageSheet: some View {
        OptionSheet(title: "选择语言", searchHint: "搜索语言", query: $search, onClose: { sheet = "" }) {
            ForEach(FlowLogic.filterLanguages(search), id: \.locale) { choice in
                Button {
                    locale = choice.locale
                    sheet = ""
                } label: {
                    HStack {
                        Text(choice.label).font(qlutchFont("NotoSansSC-SemiBold", 16)).foregroundStyle(QlutchTone.ink)
                        Spacer()
                        Image(systemName: locale == choice.locale ? "largecircle.fill.circle" : "circle")
                    }
                    .padding(.vertical, 16)
                }
                .buttonStyle(.plain)
            }
        }
    }

    private func tab(_ label: String, _ selected: Bool, _ action: @escaping () -> Void) -> some View {
        Button(action: action) {
            Text(label)
                .font(qlutchFont("NotoSansSC-SemiBold", 16))
                .foregroundStyle(selected ? QlutchTone.ink : QlutchTone.slate)
                .padding(.horizontal, 22)
                .frame(height: 40)
                .background(selected ? QlutchTone.white : Color.clear)
                .clipShape(Capsule())
        }
        .buttonStyle(.plain)
    }

    private func chip(_ label: String, _ action: @escaping () -> Void) -> some View {
        Button(action: action) {
            Text(label)
                .font(qlutchFont("NotoSansSC-SemiBold", 16))
                .foregroundStyle(QlutchTone.ink)
                .frame(maxWidth: .infinity)
                .frame(height: 50)
                .background(QlutchTone.white)
                .clipShape(Capsule())
                .shadow(color: QlutchTone.ink.opacity(0.08), radius: 8, y: 4)
        }
        .buttonStyle(.plain)
    }

    private func bundledImage(_ name: String, _ width: CGFloat, _ height: CGFloat) -> some View {
        Group {
            if let image = bundled(name, "png") {
                Image(uiImage: image).resizable().scaledToFit()
            }
        }
        .frame(width: width, height: height)
    }

    private func resetWaitSession() {
        waitStart = nil
        releasedHome = false
        waitRetryable = false
        jobId = ""
        packId = ""
        materialId = ""
        waitEntry += 1
    }

    private func openWait(source: FlowRoute, job: KnowledgeJob) {
        resetWaitSession()
        jobId = job.job_id
        packId = job.pack_id
        materialId = job.material_id
        push(.wait(source))
    }

    private func push(_ next: FlowRoute) {
        stack.append(route)
        route = next
        error = ""
        toast = false
    }

    private func back() {
        toast = false
        if case .wait = route { resetWaitSession() }
        if stack.isEmpty { onLeave() } else {
            route = stack.removeLast()
        }
    }

    private func showToast() {
        toast = true
        DispatchQueue.main.asyncAfter(deadline: .now() + 2) { toast = false }
    }

    private func sendCode(pushOtp: Bool) {
        guard !busy else { return }
        busy = true
        error = ""
        let key = FlowLogic.newIdempotencyKey()
        let region = region
        let phone = phone
        let mode = mode
        let locale = locale
        Task {
            let outcome = await api.send(key: key, region: region, phone: phone, purpose: mode, locale: locale)
            await MainActor.run {
                busy = false
                switch outcome {
                case .ok(let body):
                    hint = body.destination_hint
                    code = ""
                    if pushOtp { push(.otp) }
                case .err(_, let code):
                    if code == "AUTH_ACCOUNT_NOT_FOUND" { push(.noAccount) }
                    else if code == "AUTH_ACCOUNT_EXISTS" { self.mode = FlowLogic.purposeLogin }
                    else { error = code }
                case .transport(let reason):
                    error = reason
                }
            }
        }
    }

    private func verify() {
        guard FlowLogic.otpContinueEnabled(code), !busy else { return }
        busy = true
        let key = FlowLogic.newIdempotencyKey()
        Task {
            let outcome = await api.verify(key: key, region: region, phone: phone, purpose: mode, code: code, locale: locale)
            await MainActor.run {
                busy = false
                switch outcome {
                case .ok(let session):
                    bearer = session.access_token
                    if session.next_step == FlowLogic.nextOnboarding {
                        Task {
                            let state = await api.onboarding(bearer: session.access_token)
                            await MainActor.run {
                                let step = (try? state.okStep()) ?? nil
                                stack = []
                                route = FlowLogic.placeForOnboarding(step)
                            }
                        }
                    } else {
                        follow(session.next_step)
                    }
                case .err(_, let codeName):
                    if ["AUTH_OTP_INVALID", "AUTH_OTP_EXPIRED", "AUTH_OTP_MAX_ATTEMPTS"].contains(codeName) { code = "" }
                    error = codeName
                case .transport(let reason):
                    error = reason
                }
            }
        }
    }

    private func write(stepId: String, nickname: String?, answer: String?, then: @escaping () -> Void) async {
        busy = true
        let outcome = await api.writeOnboarding(bearer: bearer, stepId: stepId, nickname: nickname, answer: answer)
        await MainActor.run {
            busy = false
            switch outcome {
            case .ok(let state):
                if stepId == "daily_time" { follow(state.next_step) } else { then() }
            case .err(_, let code): error = code
            case .transport(let reason): error = reason
            }
        }
    }

    private func submitMaterial(source: String, contentType: String, bytes: Data, filename: String?, waitSource: FlowRoute) {
        guard !busy, !bytes.isEmpty else { return }
        busy = true
        error = ""
        Task {
            let pack = await api.createPack(bearer: bearer)
            guard case .ok(let created) = pack else {
                await MainActor.run { busy = false; error = callError(pack) }
                return
            }
            let upload = await api.createUploadURL(bearer: bearer, packId: created.pack_id, source: source, contentType: contentType, byteSize: bytes.count, filename: filename)
            guard case .ok(let ticket) = upload, let uploadURL = URL(string: ticket.upload_url) else {
                await MainActor.run { busy = false; error = callError(upload) }
                return
            }
            let put = await api.putMaterial(url: uploadURL, method: ticket.upload_method, headers: ticket.upload_headers, bytes: bytes)
            guard case .ok = put else {
                await MainActor.run { busy = false; error = callError(put) }
                return
            }
            let job = await api.createImport(key: FlowLogic.newIdempotencyKey(), bearer: bearer, packId: ticket.pack_id, materialId: ticket.material_id)
            await MainActor.run {
                busy = false
                if case .ok(let createdJob) = job {
                    openWait(source: waitSource, job: createdJob)
                } else {
                    error = callError(job)
                }
            }
        }
    }

    private func retryImport() {
        guard !busy, !packId.isEmpty, !materialId.isEmpty else { return }
        busy = true
        error = ""
        let key = FlowLogic.newIdempotencyKey()
        Task {
            let job = await api.createImport(key: key, bearer: bearer, packId: packId, materialId: materialId)
            await MainActor.run {
                busy = false
                if case .ok(let created) = job {
                    jobId = created.job_id
                    packId = created.pack_id
                    materialId = created.material_id
                    waitRetryable = false
                    error = ""
                    waitEntry += 1
                } else {
                    error = callError(job)
                }
            }
        }
    }

    private func callError<T>(_ result: CallResult<T>) -> String {
        switch result {
        case .ok: return ""
        case .err(_, let code): return code
        case .transport(let reason): return reason
        }
    }

    private func mimeType(_ ext: String) -> String {
        switch ext.lowercased() {
        case "jpg", "jpeg": return "image/jpeg"
        case "png": return "image/png"
        case "webp": return "image/webp"
        case "heic": return "image/heic"
        case "pdf": return "application/pdf"
        case "ppt": return "application/vnd.ms-powerpoint"
        case "pptx": return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
        case "doc": return "application/msword"
        case "docx": return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
        case "mp3": return "audio/mpeg"
        case "m4a": return "audio/mp4"
        case "wav": return "audio/wav"
        case "txt": return "text/plain"
        default: return "application/octet-stream"
        }
    }

    private func skipNoMaterial() {
        guard !busy else { return }
        busy = true
        error = ""
        let key = FlowLogic.newIdempotencyKey()
        Task {
            let outcome = await api.skipFirstPack(key: key, bearer: bearer)
            await MainActor.run {
                switch outcome {
                case .ok(let body):
                    applySkip(body.next_step)
                case .err(_, let code):
                    busy = false
                    error = code
                case .transport(let reason):
                    busy = false
                    error = reason
                }
            }
        }
    }

    private func applySkip(_ nextStep: String) {
        if nextStep == FlowLogic.nextOnboarding {
            Task {
                let state = await api.onboarding(bearer: bearer)
                await MainActor.run {
                    busy = false
                    switch state {
                    case .ok(let body):
                        stack = []
                        route = FlowLogic.placeForOnboarding(body.current_step_id)
                    case .err(_, let code):
                        error = code
                    case .transport(let reason):
                        error = reason
                    }
                }
            }
            return
        }
        busy = false
        follow(nextStep)
    }

    private func follow(_ nextStep: String) {
        let next = FlowLogic.placeAfterSession(nextStep, permissionsSeen: permissionsSeen)
        if next == .home { onHome() }
        else {
            stack = []
            route = next
        }
    }

    private func requestChecked() {
        Task {
            if contacts {
                _ = await withCheckedContinuation { (cont: CheckedContinuation<Bool, Never>) in
                    CNContactStore().requestAccess(for: .contacts) { ok, _ in cont.resume(returning: ok) }
                }
            }
            if notices {
                _ = try? await UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .sound, .badge])
            }
            await MainActor.run { push(.permissionsLocation) }
        }
    }
}

private extension CallResult where T == OnboardingState {
    func okStep() throws -> String? {
        if case .ok(let state) = self { return state.current_step_id }
        return nil
    }
}

final class LocationAsk: NSObject, ObservableObject, CLLocationManagerDelegate {
    private let manager = CLLocationManager()
    private var done: (() -> Void)?
    func ask(_ done: @escaping () -> Void) {
        let status = manager.authorizationStatus
        if status != .notDetermined {
            done()
            return
        }
        self.done = done
        manager.delegate = self
        manager.requestWhenInUseAuthorization()
    }
    func locationManagerDidChangeAuthorization(_ manager: CLLocationManager) {
        if manager.authorizationStatus != .notDetermined {
            done?()
            done = nil
        }
    }
}

private struct CameraCapture: UIViewControllerRepresentable {
    let onImage: (Data) -> Void

    func makeCoordinator() -> Coordinator { Coordinator(onImage: onImage) }

    func makeUIViewController(context: Context) -> UIImagePickerController {
        let picker = UIImagePickerController()
        picker.sourceType = .camera
        picker.delegate = context.coordinator
        return picker
    }

    func updateUIViewController(_ uiViewController: UIImagePickerController, context: Context) {}

    final class Coordinator: NSObject, UIImagePickerControllerDelegate, UINavigationControllerDelegate {
        let onImage: (Data) -> Void
        init(onImage: @escaping (Data) -> Void) { self.onImage = onImage }
        func imagePickerController(_ picker: UIImagePickerController, didFinishPickingMediaWithInfo info: [UIImagePickerController.InfoKey: Any]) {
            let image = info[.originalImage] as? UIImage
            let data = image?.jpegData(compressionQuality: 0.9) ?? Data()
            onImage(data)
        }
        func imagePickerControllerDidCancel(_ picker: UIImagePickerController) {
            onImage(Data())
        }
    }
}
