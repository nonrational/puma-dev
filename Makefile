build:
	go build ./cmd/puma-dev

clean:
	rm -f ./puma-dev

install:
	go install ./cmd/puma-dev

lint:
	golangci-lint run

# List all of the modules that are dependencies of your current module, along with the latest version available for each:
deps:
	go list -m -u all

release:
	rm -rf ./rel
	mkdir ./rel

	rm -rf ./pkg
	mkdir ./pkg

	SDKROOT=$$(xcrun --sdk macosx --show-sdk-path) GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags "-X main.Version=$$RELEASE" -o puma-dev_darwin_amd64 ./cmd/puma-dev
	SDKROOT=$$(xcrun --sdk macosx --show-sdk-path) GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -ldflags "-X main.Version=$$RELEASE" -o puma-dev_darwin_arm64 ./cmd/puma-dev
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.Version=$$RELEASE" -o puma-dev_linux_amd64 ./cmd/puma-dev

	mkdir rel/linux_amd64
	mv -v puma-dev_linux_amd64 rel/linux_amd64/puma-dev
	tar -C rel/linux_amd64 -cvzf "pkg/puma-dev-$$RELEASE-linux-amd64.tar.gz" puma-dev

	mkdir rel/darwin_amd64
	mv -v puma-dev_darwin_amd64 rel/darwin_amd64/puma-dev
	zip -j -v "pkg/puma-dev-$$RELEASE-darwin-amd64.zip" rel/darwin_amd64/puma-dev

	mkdir rel/darwin_arm64
	mv -v puma-dev_darwin_arm64 rel/darwin_arm64/puma-dev
	zip -j -v "pkg/puma-dev-$$RELEASE-darwin-arm64.zip" rel/darwin_arm64/puma-dev

check-test-env:
	go version > /dev/null
	ruby -v > /dev/null

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

.PHONY: release
