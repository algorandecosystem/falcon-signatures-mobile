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
            url: "https://github.com/algorandecosystem/falcon-signatures-mobile/releases/download/v0.0.12/falcon-signatures-mobile-sdk-v0.0.12.xcframework.zip",
            checksum: "a1dafc1d333d6542616d6e50cfe56246f6f9627ac77096f415b1ed9aef0eb0a9"
        )
    ]
)
