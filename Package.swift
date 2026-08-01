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
            url: "https://github.com/algorandecosystem/falcon-signatures-mobile/releases/download/v0.0.15/falcon-signatures-mobile-sdk-v0.0.15.xcframework.zip",
            checksum: "524e0916bd60ca88f77aea4da17c09b299beb1a3099efef732b282058479e0e6"
        )
    ]
)
