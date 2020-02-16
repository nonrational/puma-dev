package dev

import (
	"crypto/tls"
	"path/filepath"
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/puma/puma-dev/homedir"
	"github.com/stretchr/testify/assert"
)

var (
	supportPath      = homedir.MustExpand(supportDir)
	expectedCertPath = filepath.Join(supportPath, "cert.pem")
	expectedKeyPath  = filepath.Join(supportPath, "key.pem")
)

func TestGeneratePumaDevCertificateAuthority(t *testing.T) {
	tmpPath := "tmp"

	defer MakeDirectoryOrFail(t, tmpPath)()

	testKeyPath := filepath.Join(tmpPath, "testkey.pem")
	testCertPath := filepath.Join(tmpPath, "testcert.pem")

	if err := GeneratePumaDevCertificateAuthority(testCertPath, testKeyPath); err != nil {
		assert.Fail(t, err.Error())
	}

	tlsCert, err := tls.LoadX509KeyPair(testCertPath, testKeyPath)

	assert.Nil(t, err)
	assert.NotNil(t, tlsCert)
}
