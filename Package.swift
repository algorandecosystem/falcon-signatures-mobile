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
            url: "https://github.com/algorandecosystem/falcon-signatures-mobile/releases/download/v0.0.13/falcon-signatures-mobile-sdk-v0.0.13.xcframework.zip",
            checksum: "c50f826e594690ac34c1f39f79d6c28c8d91627145be92e81ff5e2390a46fb08"
        )
    ]
)
