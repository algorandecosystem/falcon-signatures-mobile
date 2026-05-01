import java.util.Base64

plugins {
    id("com.android.library")
    id("org.jetbrains.kotlin.android")
    id("maven-publish")
    id("signing")
    id("tech.yanand.maven-central-publish") version "1.3.0"
}

android {
    namespace = "io.github.algorandecosystem"
    publishing {
        singleVariant("release") {
            withSourcesJar()
            withJavadocJar()
        }
    }
    compileSdk = libs.versions.android.compileSdk.get().toInt()
}

tasks.register<Jar>("javadocJar") {
    archiveClassifier.set("javadoc")
    // You can add actual javadoc generation here if needed
}

tasks.register<Jar>("sourcesJar") {
    archiveClassifier.set("sources")
    from(android.sourceSets.getByName("main").java.srcDirs)
}

afterEvaluate {
    val versionTag = System.getenv("VERSION_TAG") ?: "0.1.0"
    publishing {
        publications {
            create<MavenPublication>("AlgoSDKRelease") {
                artifact(file("FalconAlgoSDK.aar")) {
                    extension = "aar"
                }

                artifact(tasks.named("sourcesJar"))
                artifact(tasks.named("javadocJar"))

                groupId = "io.github.algorandecosystem"
                artifactId = "falcon-signatures-mobile-sdk"
                version = versionTag
                setupPom("FalconAlgoSDK")
            }
        }

        repositories {
            maven {
                name = "Local"
                url = uri(layout.buildDirectory.dir("repos/bundles").get().asFile.toURI())
            }
        }
    }
}

signing {
    // About GPG signing, please refer to https://central.sonatype.org/publish/requirements/gpg/
    val signingKey = System.getenv("GPG_PRIVATE_KEY") ?: ""
    val signingPassword = System.getenv("GPG_PASSPHRASE") ?: ""
    useInMemoryPgpKeys(signingKey, signingPassword)
    sign(publishing.publications)
}

val username = System.getenv("CENTRAL_USERNAME") ?: ""
val password = System.getenv("CENTRAL_PASSWORD") ?: ""

mavenCentral {
    repoDir = layout.buildDirectory.dir("repos/bundles")
    authToken = if (username.isNotEmpty() && password.isNotEmpty()) {
        Base64.getEncoder().encodeToString("$username:$password".toByteArray())
    } else {
        System.getenv("CENTRAL_TOKEN") ?: ""
    }

    publishingType = "USER_MANAGED"
    maxWait = 500
}

tasks.register("publishAllToMavenLocal") {
    dependsOn("publishAlgoSDKReleasePublicationToMavenLocal")
}

// Helper function to configure POM metadata
fun MavenPublication.setupPom(libName: String) {
    pom {
        packaging = "aar"
        this.name.set(libName)
        this.description.set("$libName: Falcon Signatures Mobile SDK")
        this.url.set("https://github.com/algorandecosystem/falcon-signatures-mobile")
        this.inceptionYear.set("2026")

        licenses {
            license {
                this.name.set("GNU General Public License v3.0")
                this.url.set("https://github.com/algorandecosystem/falcon-signatures-mobile/blob/main/LICENSE")
            }
        }

        developers {
            developer {
                this.id.set("algorandecosystem")
                this.name.set("Algorand Ecosystem")
                this.email.set("contact@algorand.eco")
            }
        }

        scm {
            this.connection.set("scm:git:git://github.com/algorandecosystem/falcon-signatures-mobile.git")
            this.developerConnection.set("scm:git:ssh://git@github.com/algorandecosystem/falcon-signatures-mobile.git")
            this.url.set("https://github.com/algorandecosystem/falcon-signatures-mobile")
        }
    }
}
