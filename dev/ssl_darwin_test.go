package dev

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/puma/puma-dev/homedir"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	supportPath      = homedir.MustExpand(supportDir)
	expectedCertPath = filepath.Join(supportPath, "cert.pem")
)

func makeTestCert(t *testing.T) func() {
	os.RemoveAll(expectedCertPath)
	err := SetupOurCert()
	assert.Nil(t, err)

	return func() {
		os.RemoveAll(expectedCertPath)
	}
}

func TestSetupOurCert_ensureNotWorldReadable(t *testing.T) {
	t.Skip("not implemented yet - https://github.com/puma/puma-dev/issues/215")
}

func TestTrustCert_newCert(t *testing.T) {
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		t.Skip("No TTY available; can't exercise TLS cert setup")
	}

	defer makeTestCert(t)()

	stdOut := WithStdoutCaptured(func() {
		err := TrustCert(expectedCertPath)
		assert.Nil(t, err)
	})

	assert.Regexp(t, "^* Adding certification to login keychain as trusted", stdOut)
	assert.Regexp(t, "! There is probably a dialog open that requires you to authenticate\\n$", stdOut)
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
