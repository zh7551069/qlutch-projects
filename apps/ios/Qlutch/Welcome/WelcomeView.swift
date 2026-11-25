import SwiftUI

struct WelcomeView: View {
    let onStart: () -> Void
    let onExisting: () -> Void
    @State private var frame = 0
    private let timer = Timer.publish(every: 0.04, on: .main, in: .common).autoconnect()

    var body: some View {
        VStack(spacing: 0) {
            VStack(spacing: 20) {
                cheer
                wordmark
                Text("沉迷于学习，对学习上瘾")
                    .font(qlutchFont("NotoSansSC-SemiBold", 16))
                    .foregroundStyle(QlutchTone.ink)
                    .frame(width: 358)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .padding(.horizontal, 16)
            .padding(.top, 24)
            VStack(spacing: 16) {
                PrimaryButton(label: "开始", action: onStart)
                SecondaryButton(label: "已有账号", action: onExisting)
                Text("教材不用于训练模型")
                    .font(qlutchFont("NotoSansSC-Regular", 13))
                    .foregroundStyle(QlutchTone.footer)
                Text("继续即表示同意《用户协议》和《隐私政策》")
                    .font(qlutchFont("NotoSansSC-Regular", 13))
                    .foregroundStyle(QlutchTone.footer)
                    .frame(width: 358)
                    .multilineTextAlignment(.center)
            }
            .padding(.horizontal, 16)
            .padding(.bottom, 28)
        }
        .frame(maxWidth: 390)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(QlutchTone.page)
        .onReceive(timer) { _ in
            frame = (frame + 1) % 50
        }
    }

    private var cheer: some View {
        let name = String(format: "welcome_cheer_%03d", frame + 1)
        return ZStack {
            QlutchTone.slot
            if let image = bundled(name, "png") {
                Image(uiImage: image)
                    .resizable()
                    .scaledToFit()
                    .offset(x: -7, y: 20)
            }
        }
        .frame(width: 358, height: 240)
        .clipShape(RoundedRectangle(cornerRadius: 22))
    }

    private var wordmark: some View {
        Group {
            if let image = bundled("wordmark", "png") {
                Image(uiImage: image).resizable().scaledToFit()
            }
        }
        .frame(width: 174, height: 44)
    }
}
