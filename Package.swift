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
            url: "https://github.com/algorandecosystem/falcon-signatures-mobile/releases/download/v0.0.16/falcon-signatures-mobile-sdk-v0.0.16.xcframework.zip",
            checksum: "d68ed607e6dc179ed65340301b1bd2e470bf30c13aadb78612e388ea78d677ba"
        )
    ]
)
