security -v create-keychain -p 'puma-dev-test-keychain-password' ~/Library/puma-dev.keychain-db
security -v set-keychain-settings -lut 72000 ~/Library/puma-dev.keychain-db
security    list-keychains | xargs security -v list-keychains -s
security -v default-keychain -s ~/Library/puma-dev.keychain-db
security -v unlock-keychain -p 'puma-dev-test-keychain-password' ~/Library/puma-dev.keychain-db
