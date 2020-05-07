#!/usr/bin/env bash

user=runner
keychain_home="/Users/$user/Library/Keychains"
keychain_file="$keychain_home/puma-dev.keychain-db"

mkdir -p "$keychain_home"
ls -la "$keychain_home"

print_keychains() {
  echo '> Listing keychains...'
  security list-keychains

  echo -n '> Login: '
  security login-keychain 

  echo -n '> Default: ' 
  security default-keychain 
}

print_keychains()

echo '> Create puma-dev.keychain-db'
security create-keychain -p 'puma-dev-test-keychain-password' "$keychain_file"
security set-keychain-settings -lut 72000 "$keychain_file"

echo '> Add puma-dev.keychain-db to the list of keychains'
security list-keychains | xargs security -v list-keychains -s "$keychain_file"

echo '> Set puma-dev as default keychain'
security -v default-keychain -s "$keychain_file"

echo '> Set puma-dev as login keychain'
security -v login-keychain -d user -s "$keychain_file"

echo '> Unlock puma-dev keychain'
security -v unlock-keychain -p 'puma-dev-test-keychain-password' "$keychain_file"

print_keychains()