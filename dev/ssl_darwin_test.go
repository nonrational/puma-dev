package dev

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/stretchr/testify/assert"
)

func deleteAllPumaDevCAFromDefaultKeychain() {
	exec.Command("sh", "-c", fmt.Sprintf(`for sha in $(security find-certificate -a -c "Puma-dev CA" -Z | awk '/SHA-1/ {print $3}'); do security delete-certificate -t -Z $sha; done`)).Run()
	exec.Command("sh", "-c", fmt.Sprintf(`for sha in $(security find-certificate -a -c "Puma-dev CA" -Z | awk '/SHA-1/ {print $3}'); do security delete-certificate -Z $sha; done`)).Run()

	log.Println("! NOTICE - REMOVED ALL CERTS LIKE \"Puma-dev CA\" FROM THE DEFAULT macOS KEYCHAIN")
}

func TestSetupOurCert_ensureNotWorldReadable(t *testing.T) {
	t.Skip("not implemented yet - https://github.com/puma/puma-dev/issues/215")
}

func TestSetupOurCert_InteractiveCertificateInstall(t *testing.T) {
	if flag.Lookup("test.run").Value.String() != t.Name() {
		t.Skipf("interactive test must be specified with -test.run=%s", t.Name())
	}

	os.Remove(expectedCertPath)
	os.Remove(expectedKeyPath)

	assert.False(t, FileExists(expectedCertPath))
	assert.False(t, FileExists(expectedKeyPath))

	certInstallStdOut := WithStdoutCaptured(func() {
		err := SetupOurCert()
		assert.Nil(t, err)

		assert.True(t, FileExists(expectedCertPath))
		assert.True(t, FileExists(expectedKeyPath))
	})

	assert.Regexp(t, "^\\* Adding certification to login keychain as trusted\\n", certInstallStdOut)
	assert.Regexp(t, "! There is probably a dialog open that requires you to authenticate\\n", certInstallStdOut)
	assert.Regexp(t, "\\* Certificates setup, ready for https operations!\\n$", certInstallStdOut)

	defer func() {
		deleteAllPumaDevCAFromDefaultKeychain()
		os.Remove(expectedCertPath)
		os.Remove(expectedKeyPath)
	}()
}

func TestTrustCert_noCertProvided(t *testing.T) {
	stdOut := WithStdoutCaptured(func() {
		err := TrustCert("/does/not/exist")
		assert.NotNil(t, err)
		assert.Regexp(t, "Error reading file /does/not/exist\\n$", err)
	})

	assert.Regexp(t, "^* Adding certification to login keychain as trusted", stdOut)
	assert.Regexp(t, "! There is probably a dialog open that requires you to authenticate\\n$", stdOut)
}
