package dev

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testHTTP HTTPServer

type PumaDevRequestCtx struct {
	URL        string
	HostHeader string
}

func TestHTTP_appForRequest(t *testing.T) {
	var r io.Reader

	testCases := map[PumaDevRequestCtx]string{
		PumaDevRequestCtx{"http://asdf.puma", "qwerty.puma"}:      "asdf",
		PumaDevRequestCtx{"http://127.0.0.1:8080", "qwerty.puma"}: "qwerty",
		PumaDevRequestCtx{"https://127.0.0.1:", "qwerty.puma"}:    "qwerty",
		PumaDevRequestCtx{"https://my.app.puma", "qwerty.puma"}:   "my.app",
	}

	for ctx, expectedAppName := range testCases {
		t.Run(ctx.URL, func(t *testing.T) {
			req, _ := http.NewRequest("GET", ctx.URL, r)
			req.Header.Add("Host", ctx.HostHeader)

			assert.Equal(t, expectedAppName, testHTTP.appForRequest(req))
		})
	}
}

func TestHTTP_removeTLD_test(t *testing.T) {
	str := testHTTP.removeTLD("psychic-octo-guide.test")

	assert.Equal(t, "psychic-octo-guide", str)
}

func TestHTTP_removeTLD_noTld(t *testing.T) {
	str := testHTTP.removeTLD("shiny-train")

	assert.Equal(t, "shiny-train", str)
}

func TestHTTP_removeTLD_mutlipartDomain(t *testing.T) {
	str := testHTTP.removeTLD("expert-eureka.loc.al")

	assert.Equal(t, "expert-eureka.loc", str)
}

func TestHTTP_removeTLD_dev(t *testing.T) {
	str := testHTTP.removeTLD("bookish-giggle.dev:8080")

	assert.Equal(t, "bookish-giggle", str)
}

func TestHTTP_removeTLD_xipIoMalformed(t *testing.T) {
	str := testHTTP.removeTLD("legendary-meme.0.0.xip.io")

	assert.Equal(t, "", str)
}

func TestHTTP_removeTLD_xipIoDots(t *testing.T) {
	str := testHTTP.removeTLD("legendary-meme.0.0.0.0.xip.io")

	assert.Equal(t, "legendary-meme", str)
}

func TestHTTP_removeTLD_nipIoDots(t *testing.T) {
	str := testHTTP.removeTLD("effective-invention.255.255.255.255.nip.io")

	assert.Equal(t, "effective-invention", str)
}
