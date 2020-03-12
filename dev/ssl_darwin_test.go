package dev

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"os"
	"path/filepath"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/puma/puma-dev/homedir"
	"github.com/stretchr/testify/assert"
)

var (
	liveSupportPath = homedir.MustExpand(SupportDir)
	liveCertPath    = filepath.Join(liveSupportPath, "cert.pem")
	liveKeyPath     = filepath.Join(liveSupportPath, "key.pem")
)

func TestSetupOurCert_ensureNotWorldReadable(t *testing.T) {
	t.Skip("not implemented yet - https://github.com/puma/puma-dev/issues/215")
}

func TestSetupOurCert_InteractiveCertificateInstall(t *testing.T) {
	if flag.Lookup("test.run").Value.String() != t.Name() {
		t.Skipf("interactive test must be specified with -test.run=%s", t.Name())
	}

	os.Remove(liveCertPath)
	os.Remove(liveKeyPath)

	assert.False(t, FileExists(liveCertPath))
	assert.False(t, FileExists(liveKeyPath))

	certInstallStdOut := WithStdoutCaptured(func() {
		err := SetupOurCert()
		assert.Nil(t, err)

		assert.True(t, FileExists(liveCertPath))
		assert.True(t, FileExists(liveKeyPath))
	})

	assert.Regexp(t, "^\\* Adding certification to login keychain as trusted\\n", certInstallStdOut)
	assert.Regexp(t, "! There is probably a dialog open that requires you to authenticate\\n", certInstallStdOut)
	assert.Regexp(t, "\\* Certificates setup, ready for https operations!\\n$", certInstallStdOut)

	defer func() {
		err := DeleteAllPumaDevCAFromDefaultKeychain()
		assert.NoError(t, err)
		os.Remove(liveCertPath)
		os.Remove(liveKeyPath)
	}()

	t.Run("verify CA signed certificate", func(t *testing.T) {
		dnsName := "rack-hi-puma.localhost"

		parent, err := tls.LoadX509KeyPair(liveCertPath, liveKeyPath)
		assert.NoError(t, err)

		tlsCert, err := makeCert(&parent, dnsName)
		assert.NoError(t, err)

		// (tls.Certificate [][]byte) is a list of (x509.Certificate []byte)
		derBytes := tlsCert.Certificate[0]

		x509Cert, err := x509.ParseCertificate(derBytes)
		if err != nil {
			assert.FailNowf(t, "failed to parse certificate", "err: %s", err.Error())
		}

		opts := x509.VerifyOptions{
			DNSName:       dnsName,
			Intermediates: x509.NewCertPool(),
		}

		if _, err := x509Cert.Verify(opts); err != nil {
			assert.FailNowf(t, "failed to verify certificate", "err: %s", err.Error())
		}
	})
}

func TestTrustCert_Darwin_noCertProvided(t *testing.T) {
	stdOut := WithStdoutCaptured(func() {
		err := TrustCert("/does/not/exist")
		assert.NotNil(t, err)
		assert.Regexp(t, "Error reading file /does/not/exist\\n$", err)
	})

	assert.Regexp(t, "^* Adding certification to login keychain as trusted", stdOut)
	assert.Regexp(t, "! There is probably a dialog open that requires you to authenticate\\n$", stdOut)
}

func TestLoginKeychain(t *testing.T) {
	expected := homedir.MustExpand("~/Library/Keychains/login.keychain-db")
	actual, err := loginKeyChain()
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
