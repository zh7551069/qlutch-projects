import Foundation

indirect enum FlowRoute: Equatable {
    case auth
    case otp
    case noAccount
    case nickname
    case question(Int)
    case permissionsContacts
    case permissionsLocation
    case pack
    case more
    case paste
    case wait(FlowRoute)
    case home
}

struct RegionChoice: Equatable {
    let name: String
    let code: String
}

struct LanguageChoice: Equatable {
    let label: String
    let locale: String
}

struct QuestionStep: Equatable {
    let stepId: String
    let index: Int
    let title: String
    let choices: [String]
    let finishLabel: String?
}

enum FlowLogic {
    /// One pass of the reading GIF. The clip loops; this is not a release gate.
    static let readingLoopMillis: Double = 2_000
    static let jobAccepted = "accepted"
    static let jobRunning = "running"
    static let jobMapReady = "map_ready"
    static let jobFailedRetryable = "failed_retryable"
    static let jobFailedTerminal = "failed_terminal"
    /// A4-O2 is still open. This poll interval is an implementation default, not a locked product number.
    static let jobPollIntervalMillis: Double = 3_000
    /// Bridge wait meter holds at 99 of 248. It does not fill to the end on a timer.
    static let waitMeterHold: Double = 99
    static let waitMeterTrack: Double = 248
    static let regionMainland = "+86"
    static let regionHongKong = "+852"
    static let localeHans = "zh-Hans"
    static let localeHant = "zh-Hant"
    static let purposeRegister = "register"
    static let purposeLogin = "login"
    static let nextOnboarding = "NEED_ONBOARDING"
    static let nextFirstPack = "NEED_FIRST_PACK"
    static let nextHome = "READY_HOME"
    static let toastClosed = "功能尚未开放"
    static let waitTitle = "正在挖掘知识点"
    static let waitCaption = "生成知识地图需要一点时间，完成后才会进入首页"
    static let waitFootnote = "请先在这一页等完，不能提前进入首页"

    static let regions = [
        RegionChoice(name: "中国大陆", code: regionMainland),
        RegionChoice(name: "香港", code: regionHongKong),
    ]
    static let languages = [
        LanguageChoice(label: "简体中文", locale: localeHans),
        LanguageChoice(label: "繁体中文", locale: localeHant),
    ]
    static let questions: [QuestionStep] = [
        QuestionStep(stepId: "nickname", index: 1, title: "先给自己起个昵称", choices: [], finishLabel: nil),
        QuestionStep(stepId: "age", index: 2, title: "你多大啦？", choices: ["18 岁以下", "18–25 岁", "26–35 岁", "35 岁以上"], finishLabel: nil),
        QuestionStep(stepId: "study_or_work", index: 3, title: "现在是学生还是在职？", choices: ["在校学生", "在职备考", "自由职业", "待业备考"], finishLabel: nil),
        QuestionStep(stepId: "identity", index: 4, title: "你的身份是？", choices: ["本科生", "研究生", "上班族", "其他"], finishLabel: nil),
        QuestionStep(stepId: "how_to_address", index: 5, title: "怎么称呼你？", choices: ["叫我同学", "叫我老师", "直接开始吧", "随便"], finishLabel: nil),
        QuestionStep(stepId: "exam_track", index: 6, title: "备考方向是？", choices: ["自考专升本", "考研", "公考", "职业资格"], finishLabel: nil),
        QuestionStep(stepId: "daily_time", index: 7, title: "每天能投入多久？", choices: ["15 分钟", "30 分钟", "1 小时", "越多越好"], finishLabel: "建档完成，上传第一份材料"),
    ]

    static func question(_ index: Int) -> QuestionStep {
        questions.first { $0.index == index } ?? questions[0]
    }

    static func newIdempotencyKey() -> String { UUID().uuidString }

    static func phoneDigits(_ raw: String) -> String { raw.filter(\.isNumber) }

    static func formatPhone(region: String, digits: String) -> String {
        if region == regionMainland && digits.count == 11 {
            let chars = Array(digits)
            return String(chars[0..<3]) + " " + String(chars[3..<7]) + " " + String(chars[7..<11])
        }
        var parts: [String] = []
        var rest = digits
        while !rest.isEmpty {
            parts.append(String(rest.prefix(4)))
            rest = String(rest.dropFirst(4))
        }
        return parts.joined(separator: " ")
    }

    static func otpContinueEnabled(_ code: String) -> Bool {
        code.count == 6 && code.allSatisfy(\.isNumber)
    }

    static func phoneContinueEnabled(_ digits: String) -> Bool {
        !digits.isEmpty && digits.count <= 32
    }

    static func nicknameContinueEnabled(_ nickname: String) -> Bool {
        let trimmed = nickname.trimmingCharacters(in: .whitespacesAndNewlines)
        return !trimmed.isEmpty && trimmed.count <= 32
    }

    static func notesContinueEnabled(_ notes: String) -> Bool {
        !notes.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    /// Home opens only when the job status is map_ready. Elapsed time is not that signal.
    static func waitReleasesHome(_ status: String) -> Bool { status == jobMapReady }

    static func waitAllowsRetry(_ status: String, retryable: Bool) -> Bool {
        status == jobFailedRetryable && retryable
    }

    static func waitMeterFraction() -> Double { waitMeterHold / waitMeterTrack }

    static func placeAfterSession(_ nextStep: String, permissionsSeen: Bool) -> FlowRoute {
        switch nextStep {
        case nextOnboarding: return .nickname
        case nextFirstPack: return permissionsSeen ? .pack : .permissionsContacts
        case nextHome: return .home
        default: return .nickname
        }
    }

    static func placeForOnboarding(_ stepId: String?) -> FlowRoute {
        let step = questions.first { $0.stepId == stepId } ?? questions[0]
        return step.index == 1 ? .nickname : .question(step.index)
    }

    static func filterRegions(_ query: String) -> [RegionChoice] {
        let needle = query.trimmingCharacters(in: .whitespacesAndNewlines)
        if needle.isEmpty { return regions }
        return regions.filter { $0.name.contains(needle) || $0.code.contains(needle) }
    }

    static func filterLanguages(_ query: String) -> [LanguageChoice] {
        let needle = query.trimmingCharacters(in: .whitespacesAndNewlines)
        if needle.isEmpty { return languages }
        return languages.filter { $0.label.contains(needle) }
    }

    static func sentToLine(hint: String?, region: String, digits: String) -> String {
        let target = (hint?.isEmpty == false) ? hint! : region + " " + formatPhone(region: region, digits: digits)
        return "验证码已发送至 \(target)"
    }
}
