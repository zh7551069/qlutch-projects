import SwiftUI
import UIKit

enum QlutchTone {
    static let page = Color(red: 242 / 255, green: 245 / 255, blue: 247 / 255)
    static let splash = Color(red: 242 / 255, green: 245 / 255, blue: 248 / 255)
    static let ink = Color(red: 15 / 255, green: 23 / 255, blue: 42 / 255)
    static let white = Color.white
    static let slate = Color(red: 100 / 255, green: 116 / 255, blue: 139 / 255)
    static let footer = Color(red: 124 / 255, green: 132 / 255, blue: 143 / 255)
    static let line = Color(red: 225 / 255, green: 232 / 255, blue: 240 / 255)
    static let backRing = Color(red: 226 / 255, green: 227 / 255, blue: 231 / 255)
    static let slot = Color(red: 241 / 255, green: 244 / 255, blue: 249 / 255)
    static let brand = Color(red: 108 / 255, green: 92 / 255, blue: 231 / 255)
    static let focus = Color(red: 67 / 255, green: 129 / 255, blue: 230 / 255)
    static let selected = Color(red: 232 / 255, green: 228 / 255, blue: 255 / 255)
    static let alipay = Color(red: 22 / 255, green: 119 / 255, blue: 255 / 255)
    static let flag = Color(red: 222 / 255, green: 41 / 255, blue: 16 / 255)
    static let dim = Color.black.opacity(0.4)
}

func qlutchFont(_ name: String, _ size: CGFloat) -> Font {
    .custom(name, size: size)
}

struct PrimaryButton: View {
    let label: String
    var enabled: Bool = true
    let action: () -> Void
    var body: some View {
        Button(action: { if enabled { action() } }) {
            Text(label)
                .font(qlutchFont("NotoSansSC-SemiBold", 16))
                .foregroundStyle(QlutchTone.white)
                .frame(maxWidth: .infinity)
                .frame(height: 50)
                .background(enabled ? QlutchTone.ink : QlutchTone.slate)
                .clipShape(Capsule())
        }
        .buttonStyle(.plain)
        .disabled(!enabled)
    }
}

struct SecondaryButton: View {
    let label: String
    var enabled: Bool = true
    let action: () -> Void
    var body: some View {
        Button(action: { if enabled { action() } }) {
            Text(label)
                .font(qlutchFont("NotoSansSC-SemiBold", 16))
                .foregroundStyle(QlutchTone.ink)
                .frame(maxWidth: .infinity)
                .frame(height: 50)
                .background(QlutchTone.white)
                .clipShape(Capsule())
                .overlay(Capsule().stroke(QlutchTone.line, lineWidth: 1))
        }
        .buttonStyle(.plain)
        .disabled(!enabled)
    }
}

struct BackButton: View {
    let action: () -> Void
    var body: some View {
        Button(action: action) {
            Image(systemName: "chevron.left")
                .font(.system(size: 14, weight: .semibold))
                .foregroundStyle(QlutchTone.ink)
                .frame(width: 40, height: 40)
                .background(QlutchTone.white)
                .clipShape(Circle())
                .overlay(Circle().stroke(QlutchTone.backRing, lineWidth: 1))
        }
        .buttonStyle(.plain)
    }
}

struct StepProgress: View {
    let step: Int
    var body: some View {
        GeometryReader { geo in
            ZStack(alignment: .leading) {
                Capsule().fill(QlutchTone.slot)
                Capsule()
                    .fill(LinearGradient(colors: [
                        Color(red: 69 / 255, green: 55 / 255, blue: 202 / 255),
                        Color(red: 105 / 255, green: 41 / 255, blue: 216 / 255),
                        Color(red: 210 / 255, green: 68 / 255, blue: 238 / 255),
                    ], startPoint: .leading, endPoint: .trailing))
                    .frame(width: max(0, geo.size.width * CGFloat(min(7, max(0, step))) / 7), height: 8)
            }
        }
        .frame(height: 16)
    }
}

struct FlowTopBar: View {
    let step: Int
    let onBack: () -> Void
    var body: some View {
        HStack(spacing: 12) {
            BackButton(action: onBack)
            StepProgress(step: step)
        }
        .padding(.horizontal, 16)
        .padding(.top, 16)
    }
}

struct PromptRow: View {
    let title: String
    var subtitle: String? = nil
    var body: some View {
        HStack(alignment: .center, spacing: 12) {
            avatar
            VStack(alignment: .leading, spacing: 4) {
                Text(title)
                    .font(qlutchFont("NotoSansSC-ExtraBold", 22))
                    .foregroundStyle(QlutchTone.ink)
                    .frame(maxWidth: .infinity, alignment: .leading)
                if let subtitle {
                    Text(subtitle)
                        .font(qlutchFont("NotoSansSC-SemiBold", 16))
                        .foregroundStyle(QlutchTone.slate)
                        .frame(maxWidth: .infinity, alignment: .leading)
                }
            }
        }
        .padding(.horizontal, 16)
        .padding(.top, 16)
    }

    private var avatar: some View {
        Group {
            if let image = UIImage(named: "avatar") ?? bundled("avatar", "png") {
                Image(uiImage: image).resizable().scaledToFill()
            } else {
                QlutchTone.slot
            }
        }
        .frame(width: 48, height: 48)
        .clipShape(Circle())
    }
}

struct ChoiceRow: View {
    let label: String
    let selected: Bool
    let action: () -> Void
    var body: some View {
        Button(action: action) {
            HStack(spacing: 8) {
                if selected {
                    Image(systemName: "checkmark")
                        .font(.system(size: 14, weight: .bold))
                        .foregroundStyle(QlutchTone.ink)
                }
                Text(label)
                    .font(qlutchFont("NotoSansSC-SemiBold", 16))
                    .foregroundStyle(QlutchTone.ink)
            }
            .frame(maxWidth: .infinity)
            .frame(height: 50)
            .background(selected ? QlutchTone.selected : QlutchTone.white)
            .clipShape(Capsule())
            .shadow(color: QlutchTone.ink.opacity(0.08), radius: 4, y: 2)
        }
        .buttonStyle(.plain)
    }
}

struct PermissionCheck: View {
    let checked: Bool
    let action: () -> Void
    var body: some View {
        Button(action: action) {
            ZStack {
                RoundedRectangle(cornerRadius: 6)
                    .fill(checked ? QlutchTone.ink : QlutchTone.white)
                    .overlay(RoundedRectangle(cornerRadius: 6).stroke(QlutchTone.line, lineWidth: checked ? 0 : 1.5))
                if checked {
                    Image(systemName: "checkmark")
                        .font(.system(size: 11, weight: .bold))
                        .foregroundStyle(QlutchTone.white)
                }
            }
            .frame(width: 24, height: 24)
        }
        .buttonStyle(.plain)
    }
}

struct OtpCells: View {
    let code: String
    var body: some View {
        HStack(spacing: 8) {
            ForEach(0..<6, id: \.self) { index in
                let char = index < code.count ? String(code[code.index(code.startIndex, offsetBy: index)]) : ""
                let focused = code.count == index
                Text(char)
                    .font(qlutchFont("NotoSansSC-SemiBold", 22))
                    .foregroundStyle(QlutchTone.ink)
                    .frame(width: 48, height: 56)
                    .background(QlutchTone.white)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                    .overlay(RoundedRectangle(cornerRadius: 12).stroke(focused ? QlutchTone.focus : QlutchTone.line, lineWidth: focused ? 2 : 1))
            }
        }
        .accessibilityHidden(true)
    }
}

/// Clear field laid over the six cells. Register and login share it.
/// It becomes first responder when the page appears, and cell taps land on it.
struct OtpCaptureField: UIViewRepresentable {
    @Binding var code: String
    var focused: Bool

    func makeUIView(context: Context) -> OtpTextField {
        let field = OtpTextField()
        field.keyboardType = .numberPad
        field.textContentType = .oneTimeCode
        field.textColor = .clear
        field.tintColor = .clear
        field.backgroundColor = .clear
        field.autocorrectionType = .no
        field.spellCheckingType = .no
        field.borderStyle = .none
        field.accessibilityLabel = "验证码"
        field.setContentHuggingPriority(.defaultLow, for: .horizontal)
        field.setContentHuggingPriority(.defaultLow, for: .vertical)
        field.setContentCompressionResistancePriority(.defaultLow, for: .horizontal)
        field.setContentCompressionResistancePriority(.defaultLow, for: .vertical)
        field.delegate = context.coordinator
        field.onDigits = { digits in
            if context.coordinator.code.wrappedValue != digits {
                context.coordinator.code.wrappedValue = digits
            }
        }
        return field
    }

    func updateUIView(_ field: OtpTextField, context: Context) {
        context.coordinator.code = $code
        field.wantsFocus = focused
        let next = String(FlowLogic.phoneDigits(code).prefix(6))
        field.suppress = true
        if field.text != next { field.text = next }
        field.suppress = false
        if focused && field.window != nil && !field.isFirstResponder {
            field.becomeFirstResponder()
        }
    }

    func makeCoordinator() -> Coordinator { Coordinator(code: $code) }

    final class Coordinator: NSObject, UITextFieldDelegate {
        var code: Binding<String>

        init(code: Binding<String>) { self.code = code }

        func textField(_ textField: UITextField, shouldChangeCharactersIn range: NSRange, replacementString string: String) -> Bool {
            let current = textField.text ?? ""
            guard let textRange = Range(range, in: current) else { return false }
            textField.text = current.replacingCharacters(in: textRange, with: string)
            return false
        }
    }
}

final class OtpTextField: UITextField {
    var onDigits: ((String) -> Void)?
    var suppress = false
    var wantsFocus = false
    private var writing = false

    override var text: String? {
        get { super.text }
        set {
            if writing { return }
            writing = true
            let digits = String(FlowLogic.phoneDigits(newValue ?? "").prefix(6))
            let changed = super.text != digits
            super.text = digits
            writing = false
            if changed && !suppress { onDigits?(digits) }
        }
    }

    override func didMoveToWindow() {
        super.didMoveToWindow()
        if wantsFocus && window != nil && !isFirstResponder {
            becomeFirstResponder()
        }
    }
}

struct ClosedToast: View {
    let visible: Bool
    var body: some View {
        if visible {
            Text(FlowLogic.toastClosed)
                .font(qlutchFont("NotoSansSC-SemiBold", 14))
                .foregroundStyle(QlutchTone.white)
                .padding(.horizontal, 20)
                .padding(.vertical, 12)
                .background(QlutchTone.ink)
                .clipShape(Capsule())
        }
    }
}

struct ClosedChannel: View {
    let label: String
    let mark: String
    let action: () -> Void
    var body: some View {
        Button(action: action) {
            HStack(spacing: 8) {
                if mark == "支" {
                    Text("支")
                        .font(qlutchFont("NotoSansSC-ExtraBold", 14))
                        .foregroundStyle(QlutchTone.white)
                        .frame(width: 24, height: 24)
                        .background(QlutchTone.alipay)
                        .clipShape(RoundedRectangle(cornerRadius: 5.5))
                } else {
                    Circle().fill(Color(red: 0.13, green: 0.77, blue: 0.37)).frame(width: 16, height: 16)
                }
                Text(label)
                    .font(qlutchFont("NotoSansSC-SemiBold", 16))
                    .foregroundStyle(QlutchTone.ink)
            }
            .frame(maxWidth: .infinity)
            .frame(minHeight: 50, maxHeight: 50)
            .background(QlutchTone.white)
            .clipShape(Capsule())
            .overlay(Capsule().stroke(QlutchTone.line, lineWidth: 1))
        }
        .buttonStyle(.plain)
    }
}

struct FlagMark: View {
    let hongKong: Bool
    var large: Bool = false
    var body: some View {
        ZStack {
            RoundedRectangle(cornerRadius: large ? 8 : 2).fill(QlutchTone.flag)
            Circle()
                .fill(hongKong ? QlutchTone.white : Color(red: 1, green: 222 / 255, blue: 0))
                .frame(width: large ? 10 : 4, height: large ? 10 : 4)
                .offset(x: large ? -6 : -4, y: large ? -4 : -2)
        }
        .frame(width: large ? 40 : 22, height: large ? 40 : 16)
        .clipShape(large ? AnyShape(Circle()) : AnyShape(RoundedRectangle(cornerRadius: 2)))
    }
}

struct PhoneNumberField: UIViewRepresentable {
    let region: String
    @Binding var digits: String

    func makeUIView(context: Context) -> UITextField {
        let field = UITextField()
        field.borderStyle = .none
        field.backgroundColor = .clear
        field.font = UIFont(name: "NotoSansSC-Regular", size: 16) ?? .systemFont(ofSize: 16)
        field.keyboardType = .numberPad
        field.placeholder = "138 0013 8000"
        field.autocorrectionType = .no
        field.spellCheckingType = .no
        field.smartInsertDeleteType = .no
        field.adjustsFontSizeToFitWidth = false
        field.setContentHuggingPriority(.defaultLow, for: .horizontal)
        field.setContentCompressionResistancePriority(.defaultLow, for: .horizontal)
        field.delegate = context.coordinator
        field.text = FlowLogic.formatPhone(region: region, digits: digits)
        return field
    }

    func updateUIView(_ field: UITextField, context: Context) {
        context.coordinator.region = region
        context.coordinator.digits = $digits
        let formatted = FlowLogic.formatPhone(region: region, digits: digits)
        guard field.text != formatted else { return }
        let current = field.text ?? ""
        let caret: Int
        if let range = field.selectedTextRange {
            caret = field.offset(from: field.beginningOfDocument, to: range.start)
        } else {
            caret = formatted.utf16.count
        }
        let digitCount = FlowLogic.phoneDigits(current) == digits
            ? FlowLogic.digitsBeforeCursor(current, caret)
            : digits.utf16.count
        field.text = formatted
        let offset = FlowLogic.phoneCursor(formatted, digitCount)
        if let position = field.position(from: field.beginningOfDocument, offset: offset) {
            field.selectedTextRange = field.textRange(from: position, to: position)
        }
    }

    func makeCoordinator() -> Coordinator {
        Coordinator(region: region, digits: $digits)
    }

    final class Coordinator: NSObject, UITextFieldDelegate {
        var region: String
        var digits: Binding<String>

        init(region: String, digits: Binding<String>) {
            self.region = region
            self.digits = digits
        }

        func textField(_ textField: UITextField, shouldChangeCharactersIn range: NSRange, replacementString string: String) -> Bool {
            let current = textField.text ?? ""
            guard let textRange = Range(range, in: current) else { return false }
            let proposed = current.replacingCharacters(in: textRange, with: string)
            let cursor = range.location + (string as NSString).length
            let edit = FlowLogic.editPhoneField(region: region, proposedText: proposed, proposedCursor: cursor)
            textField.text = edit.text
            let offset = min(edit.cursor, (textField.text ?? "").utf16.count)
            if let position = textField.position(from: textField.beginningOfDocument, offset: offset) {
                textField.selectedTextRange = textField.textRange(from: position, to: position)
            }
            digits.wrappedValue = edit.digits
            return false
        }
    }
}

struct PhoneField: View {
    let region: String
    let hongKong: Bool
    @Binding var digits: String
    let onRegion: () -> Void
    var body: some View {
        HStack(spacing: 10) {
            Button(action: onRegion) {
                HStack(spacing: 6) {
                    FlagMark(hongKong: hongKong)
                    Text(region).font(qlutchFont("NotoSansSC-SemiBold", 16)).foregroundStyle(QlutchTone.ink)
                    Image(systemName: "chevron.down").font(.system(size: 10)).foregroundStyle(QlutchTone.ink)
                }
            }
            .buttonStyle(.plain)
            Rectangle().fill(QlutchTone.line).frame(width: 1, height: 20)
            PhoneNumberField(region: region, digits: $digits)
                .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding(.horizontal, 16)
        .frame(height: 50)
        .background(QlutchTone.white)
        .clipShape(Capsule())
        .overlay(Capsule().stroke(QlutchTone.line, lineWidth: 1))
    }
}

struct LanguageChip: View {
    let label: String
    let action: () -> Void
    var body: some View {
        Button(action: action) {
            HStack(spacing: 6) {
                Image(systemName: "globe").font(.system(size: 14)).foregroundStyle(QlutchTone.ink)
                Text(label).font(qlutchFont("NotoSansSC-SemiBold", 14)).foregroundStyle(QlutchTone.ink)
                Image(systemName: "chevron.down").font(.system(size: 10)).foregroundStyle(QlutchTone.ink)
            }
            .padding(.horizontal, 14)
            .padding(.vertical, 8)
            .background(QlutchTone.white)
            .clipShape(Capsule())
            .overlay(Capsule().stroke(QlutchTone.line, lineWidth: 1))
        }
        .buttonStyle(.plain)
    }
}

struct OptionSheet<Rows: View>: View {
    let title: String
    let searchHint: String
    @Binding var query: String
    let onClose: () -> Void
    @ViewBuilder var rows: () -> Rows
    var body: some View {
        ZStack(alignment: .bottom) {
            QlutchTone.dim.ignoresSafeArea().onTapGesture(perform: onClose)
            VStack(spacing: 8) {
                HStack {
                    Color.clear.frame(width: 40, height: 40)
                    Text(title)
                        .font(qlutchFont("NotoSansSC-SemiBold", 17))
                        .foregroundStyle(QlutchTone.ink)
                        .frame(maxWidth: .infinity)
                    Button(action: onClose) {
                        Image(systemName: "xmark")
                            .frame(width: 40, height: 40)
                            .background(QlutchTone.page)
                            .clipShape(Circle())
                    }
                    .buttonStyle(.plain)
                }
                TextField(searchHint, text: $query)
                    .font(qlutchFont("NotoSansSC-Regular", 15))
                    .padding(.horizontal, 16)
                    .frame(height: 44)
                    .background(QlutchTone.page)
                    .clipShape(Capsule())
                rows()
            }
            .padding(.horizontal, 16)
            .padding(.top, 16)
            .padding(.bottom, 28)
            .frame(maxWidth: 390)
            .background(QlutchTone.white)
            .clipShape(RoundedRectangle(cornerRadius: 28, style: .continuous))
        }
    }
}

func bundled(_ name: String, _ ext: String) -> UIImage? {
    let urls = [
        Bundle.main.url(forResource: name, withExtension: ext),
        Bundle.main.url(forResource: name, withExtension: ext, subdirectory: "Resources"),
        Bundle.main.url(forResource: name, withExtension: ext, subdirectory: "cheer"),
        Bundle.main.url(forResource: name, withExtension: ext, subdirectory: "Resources/cheer"),
    ]
    for url in urls {
        if let url, let image = UIImage(contentsOfFile: url.path) { return image }
    }
    return nil
}

func bundleURL(_ name: String, _ ext: String) -> URL? {
    let candidates = [
        Bundle.main.url(forResource: name, withExtension: ext),
        Bundle.main.url(forResource: name, withExtension: ext, subdirectory: "Resources"),
        Bundle.main.url(forResource: name, withExtension: ext, subdirectory: "Resources/cheer"),
        Bundle.main.url(forResource: name, withExtension: ext, subdirectory: "cheer"),
    ]
    return candidates.compactMap { $0 }.first
}
