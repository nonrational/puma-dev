package dev

import (
	"testing"

	. "github.com/puma/puma-dev/dev/devtest"
	"github.com/stretchr/testify/assert"
)

func TestTrustCert_Linux_noCertProvided(t *testing.T) {
	resetAndRead := CaptureStdout()

	err := TrustCert("/does/not/exist")
	assert.Nil(t, err)

	stdOut, _ := resetAndRead()

	assert.Regexp(t, "^! Add /does/not/exist to your browser to trust CA\\n$", stdOut)
}
