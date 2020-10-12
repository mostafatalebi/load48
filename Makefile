VERSION = -X main.Version=$$(git describe --tags --candidates=1)
build-linux:
	@echo "Building for version ${VERSION}"
	CGO_ENABLED=1 go build -ldflags="${VERSION}" -o ./builds/loadtest && cp ./builds/loadtest /usr/bin/load48


buildlatest:
	@echo "Building for version ${VERSIONLATEST}"
	CGO_ENABLED=1 go build -ldflags="${VERSION}" -o ./builds/loadtest-latest