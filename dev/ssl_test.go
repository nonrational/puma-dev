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

	dnsName := "rack-hi-puma.localhost"
	parent, err := tls.LoadX509KeyPair(testCertPath, testKeyPath)
	assert.NoError(t, err)

	tlsCert, err := makeCert(&parent, dnsName)
	assert.NoError(t, err)

	derBytes := tlsCert.Certificate[0]

	x509Cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		assert.FailNowf(t, "failed to parse certificate: %", err.Error())
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

	if _, err := x509Cert.Verify(opts); err != nil {
		assert.FailNowf(t, "failed to verify certificate: %s", err.Error())
	}
}
