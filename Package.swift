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
            url: "https://github.com/algorandecosystem/falcon-signatures-mobile/releases/download/v0.0.10/falcon-signatures-mobile-sdk-v0.0.10.xcframework.zip",
            checksum: "91299ca77132f4615a81591b944dca065a51a55df9e18b444428d29e0d6f49c4"
        )
    ]
)
