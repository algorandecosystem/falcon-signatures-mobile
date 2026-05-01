default: tidy test

fmt:
	go fmt ./...

tidy:
	go mod tidy

test:
	go test ./... -race

install-go-mobile:
	go install golang.org/x/mobile/cmd/gomobile@latest
	go install golang.org/x/mobile/cmd/gobind@latest
	go get golang.org/x/mobile/bind
	gomobile init

android:
	mkdir -p output
	CGO_LDFLAGS="-Wl,-z,max-page-size=16384" gomobile bind -target=android -androidapi=21 -o=output/FalconAlgoSDK.aar -javapkg=io.github.algorandecosystem github.com/algorand/go-mobile-algorand-sdk/v2/sdk

ios:
	mkdir -p output
	gomobile bind -target=ios -o=output/FalconAlgoSDK.xcframework -prefix=Algo github.com/algorand/go-mobile-algorand-sdk/v2/sdk

.PHONY: default fmt tidy test install-go-mobile android ios
