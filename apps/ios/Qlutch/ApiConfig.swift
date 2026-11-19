import Foundation

enum ApiConfig {
    /// Documented local API root. The iOS Simulator shares the host loopback.
    static let baseURL = URL(string: "http://127.0.0.1:4710")!
}
