SOURCES=$(wildcard *.go cmd/*/*.go)
VERSION=$(shell git describe --tags --long --dirty 2>/dev/null)

onc : $(SOURCES)
	mkdir -p netlify/functions
	go build -ldflags "-X main.version=${VERSION}" -o netlify/functions/$@ ./cmd/onc