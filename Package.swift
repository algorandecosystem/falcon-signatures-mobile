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
            url: "https://github.com/algorandecosystem/falcon-signatures-mobile/releases/download/v0.0.9/falcon-signatures-mobile-sdk-v0.0.9.xcframework.zip",
            checksum: "d3dce137a0f6c35609c86855ffb59c8040220b6e91f369042b0ff2a62dc5352d"
        )
    ]
)
