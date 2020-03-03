package dev

import (
	"bytes"
	"fmt"
	"os/exec"
)

const SupportDir = "~/Library/Application Support/io.puma.dev"

func LoginKeyChain() (string, error) {
	// TODO: Clean up this command a bit
	discoverLoginKeychainCmd := exec.Command("sh", "-c", `security login-keychain | xargs | tr -d '"' | tr -d '\n'`)

	var stdout, stderr bytes.Buffer
	discoverLoginKeychainCmd.Stdout = &stdout
	discoverLoginKeychainCmd.Stderr = &stderr

	if err := discoverLoginKeychainCmd.Run(); err != nil {
		return "", fmt.Errorf("could not find login keychain. security login-keychain had %s, %s", err.Error(), stderr.Bytes())
	}

	// quotedKeychainPath := string(stdout.Bytes())

	return string(stdout.Bytes()), nil
}

func TrustCert(cert string) error {
	fmt.Printf("* Adding certification to login keychain as trusted\n")
	fmt.Printf("! There is probably a dialog open that requires you to authenticate\n")

	login, keychainError := LoginKeyChain()

	if keychainError != nil {
		return keychainError
	}

	addTrustedCertCommand := exec.Command("sh", "-c", fmt.Sprintf(`security add-trusted-cert -k '%s' '%s'`, login, cert))

	var stdout, stderr bytes.Buffer
	addTrustedCertCommand.Stdout = &stdout
	addTrustedCertCommand.Stderr = &stderr

	if err := addTrustedCertCommand.Run(); err != nil {
		return fmt.Errorf("add-trusted-cert had %s. %s", err.Error(), stderr.Bytes())
	}

	fmt.Printf("* Certificates setup, ready for https operations!\n")

	return nil
}
