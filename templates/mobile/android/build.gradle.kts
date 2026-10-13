plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android") version "1.9.22" apply false
}

android {
    namespace = "com.pphlx.app"
    compileSdk = 34

    defaultConfig {
        applicationId = "com.pphlx.app"
        minSdk = 26
        targetSdk = 34
        versionCode = 1
        versionName = "1.0.0"

        ndk {
            abiFilters.addAll(listOf("arm64-v8a", "x86_64"))
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = false
            proguardFiles(getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro")
        }
    }
}

dependencies {
    implementation("com.pphlx:pphlx-android:1.2.0")
}
