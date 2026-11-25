import SwiftUI

enum ShellTab: Hashable {
    case home
    case practice
    case pk
    case friends
    case profile
}

struct RootView: View {
    @State private var health = "尚未检查"
    @State private var phase = "splash"
    @State private var purpose = FlowLogic.purposeRegister

    var body: some View {
        Group {
            if phase == "splash" {
                SplashView { phase = "welcome" }
            } else if phase == "welcome" {
                WelcomeView(
                    onStart: { purpose = FlowLogic.purposeRegister; phase = "flow" },
                    onExisting: { purpose = FlowLogic.purposeLogin; phase = "flow" }
                )
            } else if phase == "flow" {
                FlowView(
                    purpose: purpose,
                    onHome: { phase = "home" },
                    onLeave: { phase = "welcome" }
                )
            } else {
                shell
            }
        }
    }

    private var shell: some View {
        TabView {
            PlaceholderPage(title: "首页", health: health)
                .tabItem { Label("首页", systemImage: "house") }
                .tag(ShellTab.home)
            PlaceholderPage(title: "做题", health: health)
                .tabItem { Label("做题", systemImage: "square.and.pencil") }
                .tag(ShellTab.practice)
            PlaceholderPage(title: "PK", health: health)
                .tabItem { Label("PK", systemImage: "flag") }
                .tag(ShellTab.pk)
            PlaceholderPage(title: "好友", health: health)
                .tabItem { Label("好友", systemImage: "person.2") }
                .tag(ShellTab.friends)
            PlaceholderPage(title: "我的", health: health)
                .tabItem { Label("我的", systemImage: "person") }
                .tag(ShellTab.profile)
        }
        .task {
            await loadHealth()
        }
    }

    private func loadHealth() async {
        do {
            let result = try await QlutchClient(baseURL: ApiConfig.baseURL).getHealth()
            health = "\(result.statusCode) \(result.body.status)"
        } catch {
            health = "不可达"
        }
    }
}

struct PlaceholderPage: View {
    let title: String
    let health: String

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text(title)
                .font(.largeTitle)
            Text("占位")
                .foregroundStyle(.secondary)
            Text("指向 \(ApiConfig.baseURL.absoluteString)")
                .font(.footnote.monospaced())
            Text("探针 \(health)")
                .font(.footnote.monospaced())
            Text("这里没有进度，也没有业务数据。")
                .font(.footnote)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .padding(24)
    }
}
