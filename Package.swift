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
            url: "https://github.com/algorandecosystem/falcon-signatures-mobile/releases/download/v0.0.17/falcon-signatures-mobile-sdk-v0.0.17.xcframework.zip",
            checksum: "5f14d94383f670fe71b512a982ff969a890d4ce73d1b9bf26dda9fcbb3976da2"
        )
    ]
)
