package dev

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/vektra/neko"
)

func TestApp(t *testing.T) {
  n := neko.Start(t)

	n.It("verifies sanity", func() {
		assert.Equal(t, 1, 1)
	})

  n.Meow()
}
