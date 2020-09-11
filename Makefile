VERSION = -X main.Version=$$(git describe --tags --candidates=1)
build:
	@echo "Building for version ${VERSION}"
	CGO_ENABLED=1 go build -ldflags="${VERSION}" -o ./builds/loadtest


buildlatest:
	@echo "Building for version ${VERSIONLATEST}"
	CGO_ENABLED=1 go build -ldflags="${VERSION}" -o ./builds/loadtest-latest