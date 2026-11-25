package dev.qlutch.app.flow

import dev.qlutch.contract.FirstPackSkipResponse
import dev.qlutch.contract.OnboardingState
import dev.qlutch.contract.OnboardingWriteRequest
import dev.qlutch.contract.OtpSendRequest
import dev.qlutch.contract.OtpSendResponse
import dev.qlutch.contract.OtpVerifyRequest
import dev.qlutch.contract.SessionResponse
import dev.qlutch.contract.parseErrorBody
import dev.qlutch.contract.parseFirstPackSkipResponse
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
