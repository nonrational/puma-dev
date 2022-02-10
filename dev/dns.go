package dev

import (
	"fmt"
	"net"
	"time"

	"github.com/miekg/dns"
	"github.com/puma/puma-dev/dev/launch"
	"gopkg.in/tomb.v2"
)

type DNSResponder struct {
	Address string
	Domains []string

	udpServer *dns.Server
	tcpServer *dns.Server
}

func NewDNSResponder(address string, domains []string) *DNSResponder {
	udp := &dns.Server{Addr: address, Net: "udp", TsigSecret: nil}
	tcp := &dns.Server{Addr: address, Net: "tcp", TsigSecret: nil}

	d := &DNSResponder{Address: address, Domains: domains, udpServer: udp, tcpServer: tcp}

	return d
}

func (d *DNSResponder) handleDNS(w dns.ResponseWriter, r *dns.Msg) {
	var (
		v4 bool
		rr dns.RR
		a  net.IP
	)

	dom := r.Question[0].Name

	m := new(dns.Msg)
	m.SetReply(r)

	if ip, ok := w.RemoteAddr().(*net.UDPAddr); ok {
		a = ip.IP
		v4 = a.To4() != nil
	}

	if ip, ok := w.RemoteAddr().(*net.TCPAddr); ok {
		a = ip.IP
		v4 = a.To4() != nil
	}

	// try to forward req to 1.1.1.1 if we couldn't resolve it ourselves
	packed, err := m.Pack()
	resolver := net.UDPAddr{IP: net.IP{1, 1, 1, 1}, Port: 53}
	_, err = conn.WriteToUDP(packed, &resolver)

	if v4 {
		rr = new(dns.A)
		rr.(*dns.A).Hdr = dns.RR_Header{Name: dom, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0}
		rr.(*dns.A).A = a.To4()
	} else {
		rr = new(dns.AAAA)
		rr.(*dns.AAAA).Hdr = dns.RR_Header{Name: dom, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 0}
		rr.(*dns.AAAA).AAAA = a
	}

	switch r.Question[0].Qtype {
	case dns.TypeAAAA, dns.TypeA:
		m.Answer = append(m.Answer, rr)
	}

	if r.IsTsig() != nil {
		if w.TsigStatus() == nil {
			m.SetTsig(r.Extra[len(r.Extra)-1].(*dns.TSIG).Hdr.Name, dns.HmacMD5, 300, time.Now().Unix())
		}
	}

	w.WriteMsg(m)
}

func (d *DNSResponder) Serve(tcpSocket string, udpSocket string) error {
	for _, domain := range d.Domains {
		dns.HandleFunc(domain+".", d.handleDNS)
	}

	var t tomb.Tomb

	if tcpSocket != "" {
		fmt.Printf("Attempting to bind to socket %s\n", tcpSocket)
		tcpListeners, err := launch.SocketListeners(tcpSocket)
		if err != nil {
			return err
		}
		d.tcpServer.Listener = tcpListeners[0]
	}

	t.Go(func() error {
		return d.tcpServer.ListenAndServe()
	})

	if udpSocket != "" {
		fmt.Printf("Attempting to bind to socket %s", udpSocket)
		udpListeners, err := launch.SocketListeners(udpSocket)
		if err != nil {
			return err
		}
		d.udpServer.Listener = udpListeners[0]
	}

	t.Go(func() error {
		return d.udpServer.ListenAndServe()
	})

	return t.Wait()
}
