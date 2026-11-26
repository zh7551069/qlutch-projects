package dev.qlutch.app

object ApiConfig {
    // Local builds keep the emulator route to the host API.
    // -Pqlutch.e1.api=true selects the E1 root on HTTP port 80.
    const val BASE_URL: String = BuildConfig.API_BASE_URL
}
