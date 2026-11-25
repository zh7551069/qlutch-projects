import SwiftUI
import ImageIO
import UIKit

final class LoopGate {
    var on = false
}

struct ReadingStage: View {
    @State private var image = UIImage()
    @State private var source: CGImageSource?
    @State private var count = 1
    @State private var started = Date()
    @State private var gate = LoopGate()

    var body: some View {
        ZStack {
            if let still = bundled("reading_still", "png") {
                Image(uiImage: still).resizable().scaledToFit()
            }
            Image(uiImage: image).resizable().scaledToFit()
        }
        .frame(width: 358, height: 626)
        .onAppear {
            if let url = bundleURL("reading_2s", "gif"),
               let src = CGImageSourceCreateWithURL(url as CFURL, nil) {
                source = src
                count = max(1, CGImageSourceGetCount(src))
                started = Date()
                gate.on = true
                tick()
            }
        }
        .onDisappear { gate.on = false }
    }

    private func tick() {
        guard gate.on, let source else { return }
        let elapsed = Date().timeIntervalSince(started)
        let index = Int(elapsed / 0.04) % count
        if let cg = CGImageSourceCreateImageAtIndex(source, index, nil) {
            image = UIImage(cgImage: cg)
        }
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.04) { tick() }
    }
}
