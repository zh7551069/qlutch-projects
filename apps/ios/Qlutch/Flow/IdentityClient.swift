import Foundation

enum CallResult<T> {
    case ok(T)
    case err(Int, String)
    case transport(String)
}

/// OTP send and verify both send `Idempotency-Key`. The generated verify
/// method does not, and the service ignores the header on verify.
struct IdentityClient {
    let baseURL: URL

    func send(key: String, region: String, phone: String, purpose: String, locale: String) async -> CallResult<OtpSendResponse> {
        let body = OtpSendRequest(region_code: region, phone_national: phone, purpose: purpose, locale: locale)
        return await post("/v1/identity/otp/send", body: body, key: key, bearer: nil)
    }

    func verify(key: String, region: String, phone: String, purpose: String, code: String, locale: String) async -> CallResult<SessionResponse> {
        let body = OtpVerifyRequest(region_code: region, phone_national: phone, purpose: purpose, code: code, locale: locale)
        return await post("/v1/identity/otp/verify", body: body, key: key, bearer: nil)
    }

    func onboarding(bearer: String) async -> CallResult<OnboardingState> {
        await call("GET", "/v1/identity/onboarding", body: nil, key: "", bearer: bearer)
    }

    func writeOnboarding(bearer: String, stepId: String, nickname: String?, answer: String?) async -> CallResult<OnboardingState> {
        let body = OnboardingWriteRequest(step_id: stepId, nickname: nickname, answer: answer)
        let data = try? JSONEncoder().encode(body)
        return await call("PUT", "/v1/identity/onboarding", body: data, key: "", bearer: bearer)
    }

    func skipFirstPack(key: String, bearer: String) async -> CallResult<FirstPackSkipResponse> {
        let body = FirstPackSkipRequest()
        return await post("/v1/knowledge/first-pack/skip", body: body, key: key, bearer: bearer)
    }

    func createPack(bearer: String) async -> CallResult<KnowledgePack> {
        await post("/v1/knowledge/packs", body: KnowledgePackCreateRequest(), key: "", bearer: bearer)
    }

    func createUploadURL(bearer: String, packId: String, source: String, contentType: String, byteSize: Int, filename: String?) async -> CallResult<MaterialUploadUrlResponse> {
        let body = MaterialUploadUrlRequest(pack_id: packId, source: source, content_type: contentType, byte_size: byteSize, filename: filename)
        return await post("/v1/knowledge/materials/upload-url", body: body, key: "", bearer: bearer)
    }

    func putMaterial(url: URL, method: String, headers: [String: String], bytes: Data) async -> CallResult<Empty> {
        var request = URLRequest(url: url)
        request.httpMethod = method.isEmpty ? "PUT" : method
        request.httpBody = bytes
        for (name, value) in headers where !value.isEmpty {
            request.setValue(value, forHTTPHeaderField: name)
        }
        do {
            let (_, response) = try await URLSession.shared.data(for: request)
            let status = (response as? HTTPURLResponse)?.statusCode ?? 0
            if (200..<300).contains(status) { return .ok(Empty()) }
            return .err(status, "VALIDATION_MATERIAL_NOT_UPLOADED")
        } catch {
            return .transport("transport")
        }
    }

    func createImport(key: String, bearer: String, packId: String, materialId: String) async -> CallResult<KnowledgeJob> {
        let body = ImportRequest(pack_id: packId, material_id: materialId)
        return await post("/v1/knowledge/imports", body: body, key: key, bearer: bearer)
    }

    func getJob(bearer: String, jobId: String) async -> CallResult<KnowledgeJob> {
        let encoded = jobId.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? jobId
        return await call("GET", "/v1/knowledge/jobs/\(encoded)", body: nil, key: "", bearer: bearer)
    }

    private func post<B: Encodable, T: Decodable>(_ path: String, body: B, key: String, bearer: String?) async -> CallResult<T> {
        let data = try? JSONEncoder().encode(body)
        return await call("POST", path, body: data, key: key, bearer: bearer)
    }

    private func call<T: Decodable>(_ method: String, _ path: String, body: Data?, key: String, bearer: String?) async -> CallResult<T> {
        let root = baseURL.absoluteString.hasSuffix("/") ? String(baseURL.absoluteString.dropLast()) : baseURL.absoluteString
        guard let url = URL(string: root + path) else { return .transport("badURL") }
        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        if !key.isEmpty { request.setValue(key, forHTTPHeaderField: "Idempotency-Key") }
        if let bearer, !bearer.isEmpty { request.setValue("Bearer \(bearer)", forHTTPHeaderField: "Authorization") }
        if let body {
            request.httpBody = body
            request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        }
        do {
            let (data, response) = try await URLSession.shared.data(for: request)
            let status = (response as? HTTPURLResponse)?.statusCode ?? 0
            if (200..<300).contains(status) {
                let decoded = try JSONDecoder().decode(T.self, from: data)
                return .ok(decoded)
            }
            let code = (try? JSONDecoder().decode(ErrorBody.self, from: data))?.error.code ?? "INTERNAL"
            return .err(status, code)
        } catch {
            return .transport("transport")
        }
    }
}
