import Foundation

enum ApiConfig {
    #if QLUTCH_E1_API
    /// E1 staging root. HTTPS only, no path prefix.
    static let baseURL = URL(string: "https://qlutch.strict-rabbit.com")!
    #else
    /// Simulator shares the host loopback. Debug and Release keep this root.
    static let baseURL = URL(string: "http://127.0.0.1:4710")!
    #endif
}
