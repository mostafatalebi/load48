VERSION = -X main.Version=$$(git describe --tags)
build:
	@echo "Building for version ${VERSION}"
	CGO_ENABLED=1 go build -ldflags="${VERSION}" -o ./loadtest

VERSION = -X main.Version=$$(git rev-parse HEAD)
buildlatest:
	@echo "Building for version ${VERSION}"
	CGO_ENABLED=1 go build -ldflags="${VERSION}" -o ./loadtest