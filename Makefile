VERSION = -X main.Version=$$(git --no-pager tag -n1 --sort version:refname --format=%\(refname\))
VERSION_NUEMRIC = $$(git --no-pager tag -n1 --sort version:refname --format=%\(refname\))
build:
	@echo "Building for version ${VERSION}"
	CGO_ENABLED=1 go build -ldflags="${VERSION}" -o ./builds/loadtest-"${VERSION_NUEMRIC}"


VERSIONLATEST = -X main.Version=$$(git rev-parse HEAD)
buildlatest:
	@echo "Building for version ${VERSIONLATEST}"
	CGO_ENABLED=1 go build -ldflags="${VERSION}" -o ./builds/loadtest-latest