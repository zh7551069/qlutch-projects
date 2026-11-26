package dev.qlutch.app

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Test

class ApiConfigTest {
    @Test
    fun apiRootMatchesTheSelectedBuild() {
        val url = ApiConfig.BASE_URL
        if (BuildConfig.QLUTCH_E1_API) {
            assertEquals("http://47.93.180.37", url)
            assertFalse(url.contains("4710"))
        } else {
            assertEquals("http://10.0.2.2:4710", url)
            assertFalse(url.contains("47.93.180.37"))
        }
    }
}
