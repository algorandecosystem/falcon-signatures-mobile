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
            url: "https://github.com/algorandecosystem/falcon-signatures-mobile/releases/download/v0.1.0/falcon-signatures-mobile-sdk-v0.1.0.xcframework.zip",
            checksum: "24b227b466931366b6a72ffb5947843ae129e910ed8bf950bb29d7c13d1956d8"
        )
    ]
)
