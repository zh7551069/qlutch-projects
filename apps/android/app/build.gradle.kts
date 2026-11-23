plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.compose.compiler)
}

androidComponents {
    // AGP 9 registers Compose Preview tasks for every variant. Release unit tests
    // are off unless this is set, and the Kotlin script sees enableUnitTest as a
    // read-only Boolean, so the Java setter is called directly.
    beforeVariants { builder ->
        builder.javaClass
            .getMethod("setEnableUnitTest", Boolean::class.javaPrimitiveType)
            .invoke(builder, true)
    }
}

android {
    namespace = "dev.qlutch.app"
    compileSdk = libs.versions.compileSdk.get().toInt()
    compileSdkMinor = libs.versions.compileSdkMinor.get().toInt()

    defaultConfig {
        applicationId = "dev.qlutch.app"
        minSdk = libs.versions.minSdk.get().toInt()
        targetSdk = libs.versions.targetSdk.get().toInt()
        // Earlier preview packages already shipped as 0.0.0 (1). Keep the name and step the code.
        versionCode = 2
        versionName = "0.0.0"
    }

    buildTypes {
        release {
            isMinifyEnabled = false
        }
    }

    buildFeatures {
        compose = true
    }

    androidResources {
        noCompress += "mp4"
    }

    testOptions {
        unitTests.isIncludeAndroidResources = true
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
}

dependencies {
    implementation(platform(libs.compose.bom))
    implementation(libs.compose.ui)
    implementation(libs.compose.foundation)
    implementation(libs.compose.material3)
    implementation(libs.activity.compose)
    testImplementation(libs.junit)
}
