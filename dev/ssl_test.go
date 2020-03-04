package dev

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"path/filepath"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/stretchr/testify/assert"
)

var (
	tmpPath      = "tmp"
	testKeyPath  = filepath.Join(tmpPath, "testkey.pem")
	testCertPath = filepath.Join(tmpPath, "testcert.pem")
)

func TestGeneratePumaDevCertificateAuthority(t *testing.T) {
	defer MakeDirectoryOrFail(t, tmpPath)()

	if err := GeneratePumaDevCertificateAuthority(testCertPath, testKeyPath); err != nil {
		assert.Fail(t, err.Error())
	}

	_, err := tls.LoadX509KeyPair(testCertPath, testKeyPath)

	assert.NoError(t, err)
}

func TestMakeCert(t *testing.T) {
	defer MakeDirectoryOrFail(t, tmpPath)()

	if err := GeneratePumaDevCertificateAuthority(testCertPath, testKeyPath); err != nil {
		assert.FailNow(t, err.Error())
	}

	parent, err := tls.LoadX509KeyPair(testCertPath, testKeyPath)
	assert.NoError(t, err)

	domainNameValidityTestCases := map[string]bool{
		"mail.google.com": false,
		"golang.org":      false,
		"gmail.com":       false,
		"cdc.gov":         false,
		"unc.edu":         false,
		"nip.io":          false,

		"app.loc.al":            true,
		"a.b.c.d.e.pdev":        true,
		"something.local":       true,
		"rack-hi-puma.test":     true,
		"something.localhost":   true,
		"rack-hi-puma.puma.dev": true,
	}

	for dnsName, expectedValid := range domainNameValidityTestCases {
		t.Run(dnsName, func(t *testing.T) {
			tlsCert, err := makeCert(&parent, dnsName)
			assert.NoError(t, err)

			// (tls.Certificate [][]byte) is a list of (x509.Certificate []byte)
			derBytes := tlsCert.Certificate[0]

			x509Cert, err := x509.ParseCertificate(derBytes)
			if err != nil {
				assert.FailNowf(t, "failed to parse certificate", "err: %s", err.Error())
			}

			rootPEM, err := ioutil.ReadFile(testCertPath)
			assert.NoError(t, err)

			roots := x509.NewCertPool()
			ok := roots.AppendCertsFromPEM([]byte(rootPEM))
			if !ok {
				assert.FailNow(t, "failed to append CA root")
			}

			opts := x509.VerifyOptions{
				Roots:         roots,
				DNSName:       dnsName,
				Intermediates: x509.NewCertPool(),
			}

			if _, err := x509Cert.Verify(opts); (err == nil) != expectedValid {
				assert.FailNowf(t, "certificate failed validity check", "%s was valid=%v, should be valid=%v", dnsName, (err == nil), expectedValid)
			}
		})
	}
}
