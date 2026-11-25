import SwiftUI
import AVFoundation

struct SplashView: View {
    let onDone: () -> Void
    @State private var holding = false
    @State private var left = false
    @State private var player: AVPlayer?
    var body: some View {
        ZStack {
            QlutchTone.splash.ignoresSafeArea()
            VStack(spacing: 0) {
                Spacer()
                VStack(spacing: 16) {
                    ZStack {
                        RoundedRectangle(cornerRadius: 22).fill(QlutchTone.slot)
                        if let player {
                            VideoFill(player: player)
                                .clipShape(RoundedRectangle(cornerRadius: 22))
                        }
                    }
                    .frame(width: 260, height: 260)
                    wordmark
                    Text("让记知识，变得轻松有趣")
                        .font(qlutchFont("NotoSansSC-Regular", 15))
                        .foregroundStyle(QlutchTone.brand)
                }
                Spacer()
                if holding {
                    Text("正在擦亮你的知识地图")
                        .font(qlutchFont("NotoSansSC-Regular", 15))
                        .foregroundStyle(QlutchTone.footer)
                        .padding(.bottom, 40)
                } else {
                    Color.clear.frame(height: 32)
                }
            }
        }
        .contentShape(Rectangle())
        .onTapGesture { finish() }
        .gesture(DragGesture(minimumDistance: 24).onEnded { value in
            if value.startLocation.x < 32 && value.translation.width > 48 { finish() }
        })
        .onAppear(perform: start)
        .onDisappear { player?.pause() }
    }

    private var wordmark: some View {
        Group {
            if let image = bundled("wordmark", "png") {
                Image(uiImage: image).resizable().scaledToFit()
            }
        }
        .frame(width: 143, height: 36)
    }

    private func start() {
        guard let url = bundleURL("splash", "mp4") else {
            holdThenLeave()
            return
        }
        let item = AVPlayerItem(url: url)
        let av = AVPlayer(playerItem: item)
        player = av
        NotificationCenter.default.addObserver(forName: .AVPlayerItemDidPlayToEndTime, object: item, queue: .main) { _ in
            holdThenLeave()
        }
        av.play()
    }

    private func holdThenLeave() {
        holding = true
        DispatchQueue.main.asyncAfter(deadline: .now() + 1) { finish() }
    }

    private func finish() {
        if left { return }
        left = true
        player?.pause()
        onDone()
    }
}

private struct VideoFill: UIViewRepresentable {
    let player: AVPlayer
    func makeUIView(context: Context) -> PlayerView {
        let view = PlayerView()
        view.playerLayer.player = player
        view.playerLayer.videoGravity = .resizeAspect
        return view
    }
    func updateUIView(_ uiView: PlayerView, context: Context) {
        uiView.playerLayer.player = player
    }
}

final class PlayerView: UIView {
    override class var layerClass: AnyClass { AVPlayerLayer.self }
    var playerLayer: AVPlayerLayer { layer as! AVPlayerLayer }
}
