package dev

import (
	"log"
	"net"
	"testing"
	"time"

	"github.com/avast/retry-go"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

var tDNSResponder DNSResponder

func TestServe(t *testing.T) {
	tDNSResponder.Address = "localhost:31337"
	errChan := make(chan error)

	go func() {
		if err := tDNSResponder.Serve([]string{"test"}); err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	shortTimeout := time.Duration(5 * time.Second)
	protocols := map[string](func() *dns.Server){
		"tcp": func() *dns.Server { return tDNSResponder.tcpServer },
		"udp": func() *dns.Server { return tDNSResponder.udpServer },
	}

	for protocol, serverLookup := range protocols {
		dialError := retry.Do(
			func() error {
				if _, err := net.DialTimeout(protocol, "localhost:31337", shortTimeout); err != nil {
					return err
				}
				serverLookup().Shutdown()
				return nil
			},
			retry.OnRetry(func(n uint, err error) {
				log.Printf("#%d: %s\n", n, err)
			}),
		)

		assert.NoError(t, dialError)
	}
}
