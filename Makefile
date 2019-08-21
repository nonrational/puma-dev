all:
	go build ./cmd/puma-dev

install:
	go install ./cmd/puma-dev

release:
	gox -os="darwin linux" -arch="amd64" -ldflags "-X main.Version=$$RELEASE" ./cmd/puma-dev
	mv puma-dev_linux_amd64 puma-dev
	tar czvf puma-dev-$$RELEASE-linux-amd64.tar.gz puma-dev
	mv puma-dev_darwin_amd64 puma-dev
	zip puma-dev-$$RELEASE-darwin-amd64.zip puma-dev

RICHGOBIN := $(shell command -v richgo 2> /dev/null)
test:
ifdef RICHGOBIN
	richgo test -v ./...
else
	go test -v ./...
endif

.PHONY: all release
