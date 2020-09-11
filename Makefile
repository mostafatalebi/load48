VERSION = -X main.Version=$$(git describe --tags --candidates=1)
VERSION_NUEMRIC = $$(git describe --tags --candidates=1)
build:
	@echo "Building for version ${VERSION}"
	CGO_ENABLED=1 go build -ldflags="${VERSION}" -o ./builds/loadtest-"${VERSION_NUEMRIC}"


VERSIONLATEST = -X main.Version=$$(git rev-parse HEAD)
buildlatest:
	@echo "Building for version ${VERSIONLATEST}"
	CGO_ENABLED=1 go build -ldflags="${VERSION}" -o ./builds/loadtest-latest