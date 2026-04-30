pluginManagement {
    repositories {
        google()
        gradlePluginPortal()
        mavenCentral()
    }
}

dependencyResolutionManagement {
    repositories {
        google()
        mavenCentral()
    }
}

rootProject.name = "falcon-signatures-mobile-sdk"

include(":falcon-android-sdk")
project(":falcon-android-sdk").projectDir = file("output")
