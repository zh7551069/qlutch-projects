import Foundation

enum ApiConfig {
    #if QLUTCH_E1_API
    /// E1 Staging configuration. HTTP port 80.
    static let baseURL = URL(string: "http://47.93.180.37")!
    #else
    /// Simulator shares the host loopback. Debug and Release keep this root.
    static let baseURL = URL(string: "http://127.0.0.1:4710")!
    #endif
}
