package dev.qlutch.app

object ApiConfig {
    // Local builds keep the emulator route to the host API.
    // -Pqlutch.e1.api=true selects https://qlutch.strict-rabbit.com.
    const val BASE_URL: String = BuildConfig.API_BASE_URL
}
