.PHONY: build clean install lint release-clean release-build release-package-darwin release-package-linux release test clean-test test-macos-interactive test-macos-filesystem-setup coverage test-macos-interactive test-macos-manual-setup-install devel-uninstall devel-setup-install

build:
	go build ./cmd/puma-dev

clean:
	rm -f ./puma-dev

install:
	go install ./cmd/puma-dev

lint:
	golangci-lint run

release-clean:
	rm -rf ./rel
	rm -rf ./pkg

release-build:
	mkdir ./rel
	mkdir ./pkg

	SDKROOT=$$(xcrun --sdk macosx --show-sdk-path) gox -cgo -os="darwin" -arch="amd64 arm64" -ldflags "-X main.Version=$$RELEASE" ./cmd/puma-dev
	gox -os="linux" -arch="amd64" -ldflags "-X main.Version=$$RELEASE" ./cmd/puma-dev

	mkdir rel/linux_amd64
	mv -v puma-dev_linux_amd64 rel/linux_amd64/puma-dev

	mkdir rel/darwin_amd64
	mv -v puma-dev_darwin_amd64 rel/darwin_amd64/puma-dev

	mkdir rel/darwin_arm64
	mv -v puma-dev_darwin_arm64 rel/darwin_arm64/puma-dev

release-package-darwin:
	gon -log-level=debug -log-json ./gon_amd64.json
	mv pkg/puma-dev-darwin-amd64.zip "pkg/puma-dev-$$RELEASE-darwin-amd64.zip"

	gon -log-level=debug -log-json ./gon_arm64.json
	mv pkg/puma-dev-darwin-arm64.zip "pkg/puma-dev-$$RELEASE-darwin-arm64.zip"

release-package-linux:
	tar -C rel/linux_amd64 -cvzf "pkg/puma-dev-$$RELEASE-linux-amd64.tar.gz" puma-dev

release: release-clean release-build release-package-darwin release-package-linux
	openssl dgst -sha256 pkg/*

test: clean-test
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

clean-test:
	rm -rf $$HOME/.puma-dev-test_*

test-macos-filesystem-setup:
	sudo mkdir -p /etc/resolver;
	sudo chmod 0775 /etc/resolver;
	sudo chown :staff /etc/resolver;

coverage: test
	go tool cover -html=coverage.out -o coverage.html

test-macos-interactive:
	@echo "This will break your existing puma-dev setup. You'll need to run setup/install again. Cool? Cool."
	@echo "Also, prepare to provide your system password several times."
	@read -p "Press [return] to continue..."
	rm -rf "$$HOME/Library/Application\ Support/io.puma.dev"
	go test ./... -v -test.run=DarwinInteractive -count=1
	rm -rf "$$HOME/Library/Application\ Support/io.puma.dev"

test-macos-manual-setup-install: clean build
	sudo launchctl unload "$$HOME/Library/LaunchAgents/io.puma.dev.plist"
	rm -rf "$$HOME/Library/Application\ Support/io.puma.dev"
	rm -f "$$HOME/Library/LaunchAgents/io.puma.dev.plist"
	rm -f "$$HOME/Library/Logs/puma-dev.log"

	sudo ./puma-dev -d 'test:localhost:loc.al:puma' -setup
	./puma-dev -d 'test:localhost:loc.al:puma' -install

	test -f "$$HOME/Library/LaunchAgents/io.puma.dev.plist"
	launchctl list io.puma.dev > /dev/null
	test -f "$$HOME/Library/Logs/puma-dev.log"
	test 'Hi Puma!' == "$$(curl -s https://rack-hi-puma.puma)" && echo "PASS"

devel-setup-install: build
	sudo ./puma-dev -d 'test:puma:puma.dev:localhost' -setup
	./puma-dev -d 'test:puma:puma.dev:localhost' -install

devel-uninstall: build
	./puma-dev -uninstall -d 'test:puma:puma.dev:localhost'
