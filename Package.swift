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
            url: "https://github.com/algorandecosystem/falcon-signatures-mobile/releases/download/v0.0.14/falcon-signatures-mobile-sdk-v0.0.14.xcframework.zip",
            checksum: "9fcdd5b1c7b7dbac2c2f45182200dd228e9c5083740591e53d5c10c93b3425e5"
        )
    ]
)
