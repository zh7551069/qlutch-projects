package dev.qlutch.app.flow

import dev.qlutch.contract.FirstPackSkipResponse
import dev.qlutch.contract.ImportRequest
import dev.qlutch.contract.KnowledgeJob
import dev.qlutch.contract.KnowledgePack
import dev.qlutch.contract.MaterialUploadUrlRequest
import dev.qlutch.contract.MaterialUploadUrlResponse
import dev.qlutch.contract.OnboardingState
import dev.qlutch.contract.OnboardingWriteRequest
import dev.qlutch.contract.OtpSendRequest
import dev.qlutch.contract.OtpSendResponse
import dev.qlutch.contract.OtpVerifyRequest
import dev.qlutch.contract.SessionResponse
import dev.qlutch.contract.parseErrorBody
import dev.qlutch.contract.parseFirstPackSkipResponse
import dev.qlutch.contract.parseKnowledgeJob
import dev.qlutch.contract.parseKnowledgePack
import dev.qlutch.contract.parseMaterialUploadUrlResponse
import dev.qlutch.contract.parseOnboardingState
import dev.qlutch.contract.parseOtpSendResponse
import dev.qlutch.contract.parseSessionResponse
import java.net.HttpURLConnection
import java.net.URI

internal sealed class ApiOutcome<out T> {
    data class Ok<T>(val value: T) : ApiOutcome<T>()
    data class Err(val status: Int, val code: String) : ApiOutcome<Nothing>()
    data class Transport(val reason: String) : ApiOutcome<Nothing>()
}

/**
 * Calls the generated A1 parsers. OTP send and verify both carry
 * `Idempotency-Key`. The generated verify method does not attach that header,
 * and the service currently ignores it on verify; the header is still sent.
 */
internal interface IdentityApi {
    fun send(idempotencyKey: String, region: String, phone: String, purpose: String, locale: String): ApiOutcome<OtpSendResponse>
    fun verify(idempotencyKey: String, region: String, phone: String, purpose: String, code: String, locale: String): ApiOutcome<SessionResponse>
    fun onboarding(bearer: String): ApiOutcome<OnboardingState>
    fun writeOnboarding(bearer: String, stepId: String, nickname: String?, answer: String?): ApiOutcome<OnboardingState>
    fun skipFirstPack(idempotencyKey: String, bearer: String): ApiOutcome<FirstPackSkipResponse>
    fun createPack(bearer: String): ApiOutcome<KnowledgePack>
    fun createUploadUrl(
        bearer: String,
        packId: String,
        source: String,
        contentType: String,
        byteSize: Long,
        filename: String?,
    ): ApiOutcome<MaterialUploadUrlResponse>
    fun putMaterial(url: String, method: String, headers: Map<String, String>, bytes: ByteArray): ApiOutcome<Unit>
    fun createImport(idempotencyKey: String, bearer: String, packId: String, materialId: String): ApiOutcome<KnowledgeJob>
    fun getJob(bearer: String, jobId: String): ApiOutcome<KnowledgeJob>
}

internal class HttpIdentityApi(private val baseUrl: String) : IdentityApi {
    override fun send(idempotencyKey: String, region: String, phone: String, purpose: String, locale: String): ApiOutcome<OtpSendResponse> {
        val body = OtpSendRequest(region, phone, purpose, locale)
        return post("/v1/identity/otp/send", body.toJSON().toString(), idempotencyKey, null, ::parseOtpSendResponse)
    }

    override fun verify(idempotencyKey: String, region: String, phone: String, purpose: String, code: String, locale: String): ApiOutcome<SessionResponse> {
        val body = OtpVerifyRequest(region, phone, purpose, code, locale)
        return post("/v1/identity/otp/verify", body.toJSON().toString(), idempotencyKey, null, ::parseSessionResponse)
    }

    override fun onboarding(bearer: String): ApiOutcome<OnboardingState> {
        return call("GET", "/v1/identity/onboarding", null, "", bearer, ::parseOnboardingState)
    }

    override fun writeOnboarding(bearer: String, stepId: String, nickname: String?, answer: String?): ApiOutcome<OnboardingState> {
        val body = OnboardingWriteRequest(stepId, nickname, answer)
        return call("PUT", "/v1/identity/onboarding", body.toJSON().toString(), "", bearer, ::parseOnboardingState)
    }

    override fun skipFirstPack(idempotencyKey: String, bearer: String): ApiOutcome<FirstPackSkipResponse> {
        return call("POST", "/v1/knowledge/first-pack/skip", "{}", idempotencyKey, bearer, ::parseFirstPackSkipResponse)
    }

    override fun createPack(bearer: String): ApiOutcome<KnowledgePack> {
        return call("POST", "/v1/knowledge/packs", "{}", "", bearer, ::parseKnowledgePack)
    }

    override fun createUploadUrl(
        bearer: String,
        packId: String,
        source: String,
        contentType: String,
        byteSize: Long,
        filename: String?,
    ): ApiOutcome<MaterialUploadUrlResponse> {
        val body = MaterialUploadUrlRequest(packId, source, contentType, byteSize, filename)
        return call("POST", "/v1/knowledge/materials/upload-url", body.toJSON().toString(), "", bearer, ::parseMaterialUploadUrlResponse)
    }

    override fun putMaterial(url: String, method: String, headers: Map<String, String>, bytes: ByteArray): ApiOutcome<Unit> {
        return try {
            val connection = (URI(url).toURL().openConnection() as HttpURLConnection)
            try {
                connection.requestMethod = if (method.isBlank()) "PUT" else method
                connection.connectTimeout = 15000
                connection.readTimeout = 15000
                for ((name, value) in headers) {
                    if (value.isNotEmpty()) connection.setRequestProperty(name, value)
                }
                connection.doOutput = true
                connection.outputStream.use { it.write(bytes) }
                val code = connection.responseCode
                if (code in 200..299) ApiOutcome.Ok(Unit) else ApiOutcome.Err(code, "VALIDATION_MATERIAL_NOT_UPLOADED")
            } finally {
                connection.disconnect()
            }
        } catch (error: Exception) {
            ApiOutcome.Transport(error.javaClass.simpleName)
        }
    }

    override fun createImport(idempotencyKey: String, bearer: String, packId: String, materialId: String): ApiOutcome<KnowledgeJob> {
        val body = ImportRequest(packId, materialId)
        return call("POST", "/v1/knowledge/imports", body.toJSON().toString(), idempotencyKey, bearer, ::parseKnowledgeJob)
    }

    override fun getJob(bearer: String, jobId: String): ApiOutcome<KnowledgeJob> {
        val encoded = java.net.URLEncoder.encode(jobId, Charsets.UTF_8).replace("+", "%20")
        return call("GET", "/v1/knowledge/jobs/$encoded", null, "", bearer, ::parseKnowledgeJob)
    }

    private fun <T> post(path: String, json: String, idempotencyKey: String, bearer: String?, parse: (String) -> T): ApiOutcome<T> {
        return call("POST", path, json, idempotencyKey, bearer ?: "", parse)
    }

    private fun <T> call(
        method: String,
        path: String,
        json: String?,
        idempotencyKey: String,
        bearer: String,
        parse: (String) -> T,
    ): ApiOutcome<T> {
        return try {
            val connection = (URI(baseUrl.trimEnd('/') + path).toURL().openConnection() as HttpURLConnection)
            try {
                connection.requestMethod = method
                connection.setRequestProperty("Accept", "application/json")
                if (idempotencyKey.isNotEmpty()) connection.setRequestProperty("Idempotency-Key", idempotencyKey)
                if (bearer.isNotEmpty()) connection.setRequestProperty("Authorization", "Bearer $bearer")
                connection.connectTimeout = 5000
                connection.readTimeout = 5000
                if (json != null) {
                    connection.doOutput = true
                    connection.setRequestProperty("Content-Type", "application/json")
                    connection.outputStream.use { it.write(json.toByteArray(Charsets.UTF_8)) }
                }
                val code = connection.responseCode
                val stream = if (code >= 400) connection.errorStream else connection.inputStream
                val text = stream?.bufferedReader()?.use { it.readText() } ?: ""
                if (code in 200..299) {
                    ApiOutcome.Ok(parse(text))
                } else {
                    ApiOutcome.Err(code, errorCode(text))
                }
            } finally {
                connection.disconnect()
            }
        } catch (error: Exception) {
            ApiOutcome.Transport(error.javaClass.simpleName)
        }
    }

    private fun errorCode(text: String): String {
        return try {
            parseErrorBody(text).error.code
        } catch (_: Exception) {
            Regex("\"code\"\\s*:\\s*\"([A-Z0-9_]+)\"").find(text)?.groupValues?.getOrNull(1) ?: "INTERNAL"
        }
    }
}
