.PHONY: build test clean install release

# Build variables
BINARY = wade
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
LDFLAGS = -ldflags="-s -w -X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BINARY) .

install: build
	cp $(BINARY) ~/.local/bin/
	@echo "Installed to ~/.local/bin/$(BINARY)"

test:
	go test ./... -v -count=1

clean:
	rm -f $(BINARY)
	go clean

# Cross-platform builds
build-all:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe .

release: build-all
	cd dist && for f in *-amd64 *-arm64; do \
		tar czf $$f.tar.gz $$f && shasum -a 256 $$f.tar.gz > $$f.tar.gz.sha256; \
	done
	cd dist && for f in *.exe; do \
		zip $$f.zip $$f && shasum -a 256 $$f.zip > $$f.zip.sha256; \
	done
	@echo "Release artifacts ready in dist/"
