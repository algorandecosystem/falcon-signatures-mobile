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
            url: "https://github.com/algorandecosystem/falcon-signatures-mobile/releases/download/v0.0.11/falcon-signatures-mobile-sdk-v0.0.11.xcframework.zip",
            checksum: "86b0e2a0ffddac9c0f80f7e9ff6b08e6abb7bf8ef9c552ee8f1657609e919890"
        )
    ]
)
