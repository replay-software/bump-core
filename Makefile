.PHONY: all build-linux-amd64 build-linux-arm64 build-macos-amd64 build-macos-arm64

all: build-linux-amd64 build-linux-arm64 build-macos-amd64 build-macos-arm64

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o bin/bump-linux-amd64 bump.go

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o bin/bump-linux-arm64 bump.go

build-macos-amd64:
	GOOS=darwin GOARCH=amd64 go build -o bin/bump-darwin-amd64 bump.go

build-macos-arm64:
	GOOS=darwin GOARCH=arm64 go build -o bin/bump-darwin-arm64 bump.go

clean:
	rm -rf bin
