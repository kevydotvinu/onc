VERSION=$(shell git describe --tags --long --dirty 2>/dev/null)

all:
	mkdir -p netlify/functions
	go build -ldflags "-X main.version=${VERSION}" -o netlify/functions/onc