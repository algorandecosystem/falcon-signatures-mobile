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
            url: "https://github.com/algorandecosystem/falcon-signatures-mobile/releases/download/v0.0.18/falcon-signatures-mobile-sdk-v0.0.18.xcframework.zip",
            checksum: "56bd73fd95fc5cf9b5ed886804c40974db7b4aa848d57cf333c83d34510c1652"
        )
    ]
)
