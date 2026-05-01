// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "FalconMobileSDK",
    platforms: [.iOS(.v13)],
    products: [
        .library(name: "FalconMobileSDK", targets: ["FalconMobileSDK"])
    ],
    targets: [
        .binaryTarget(
            name: "FalconMobileSDK",
            url: "https://github.com/algorandecosystem/falcon-signatures-mobile/releases/download/v0.0.8/falcon-signatures-mobile-sdk-v0.0.8.xcframework.zip",
            checksum: "18234998daef645684f4a2c2ad9f01f7a446bfab0949b04d5eb113cbc072158a"
        )
    ]
)
