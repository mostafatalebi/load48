build:
	VERSION=$(git describe --tags)
	CGO_ENABLED=1 go build -ldflags="-X 'main.Version=$(VERSION)'" -o ./loadtest