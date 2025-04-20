package dev

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/puma/puma-dev/homedir"
	"github.com/vektra/errors"
)

var CACert *tls.Certificate

func GeneratePumaDevCertificateAuthority(certPath string, keyPath string, domains []string) error {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.Context(err, "generating new RSA key")
	}

	// create certificate structure with proper values
	notBefore := time.Now()
	notAfter := notBefore.Add(9999 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return errors.Context(err, "generating serial number")
	}

	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Developer Certificate"},
			CommonName:   fmt.Sprintf("Puma-dev CA (%v)", serialNumber),
		},
		NotBefore:                   notBefore,
		NotAfter:                    notAfter,
		KeyUsage:                    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:                 []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid:       true,
		IsCA:                        true,
		PermittedDNSDomainsCritical: true,
		PermittedDNSDomains:         domains,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, priv.Public(), priv)

	if err != nil {
		return errors.Context(err, "creating CA cert")
	}

	certOut, err := os.Create(certPath)
	if err != nil {
		return errors.Context(err, "writing cert.pem")
	}

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Context(err, "writing key.pem")
	}

	pem.Encode(
		keyOut,
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		},
	)

	keyOut.Close()

	return nil
}

func EnsurePermittedDNSDomains(cert *tls.Certificate, domains []string) error {
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %v", err)
	}

	if !x509Cert.PermittedDNSDomainsCritical {
		// If this certificate was generated prior to introducing PermittedDNSDomains support,
		// inform the user that they should reinstall to take advantage of the new feature.
		log.Println("Your puma-dev CA is outdated and can sign arbitrary domains.\nFor your security, use `-uninstall` to remove it and `-install` to generate and trust a new CA.")
		return nil
	}

	// Create a map of existing permitted domains for quick lookup
	existingDomains := make(map[string]struct{}, len(x509Cert.PermittedDNSDomains))
	for _, domain := range x509Cert.PermittedDNSDomains {
		existingDomains[domain] = struct{}{}
	}

	// Check if all topLevelDomains are already in the certificate
	missingDomains := []string{}
	for _, domain := range domains {
		if _, found := existingDomains[domain]; !found {
			missingDomains = append(missingDomains, domain)
		}
	}

	// log.Println("Existing domains: ", x509Cert.PermittedDNSDomains)
	// log.Println("Missing domains:", missingDomains)

	// If no domains are missing, return early
	if len(missingDomains) == 0 {
		return nil
	}

	return fmt.Errorf("CA does not support domains %v", missingDomains)
}

// SetupOurCert sets up a certificate authority (CA) for the given DNS domains.
// It ensures that the necessary directory structure exists, checks for an existing
// key and certificate pair, validates the domains against the existing certificate,
// and generates a new CA if needed. If a new CA is generated, it also attempts to
// trust the certificate on the system.
//
// Parameters:
//   - dnsDomains: A slice of strings representing the DNS domains for which the
//     certificate authority should be set up.
//
// Returns:
//   - An error if any step in the process fails, such as directory creation, loading
//     the key pair, domain validation, certificate generation, or trusting the certificate.
func SetupOurCert(dnsDomains []string) error {
	dir := homedir.MustExpand(SupportDir)

	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	keyPath := filepath.Join(dir, "key.pem")
	certPath := filepath.Join(dir, "cert.pem")

	tlsCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err == nil {
		if domainErr := EnsurePermittedDNSDomains(&tlsCert, dnsDomains); domainErr != nil {
			log.Fatal("Existing puma-dev CA is invalid. Please `-uninstall` and try again: ", domainErr)
			return domainErr
		}

		log.Printf("Existing puma-dev CA keypair found for domain(s) %v. Assuming trusted.", dnsDomains)
		CACert = &tlsCert
		return nil
	}

	if certGenErr := GeneratePumaDevCertificateAuthority(certPath, keyPath, dnsDomains); certGenErr != nil {
		return certGenErr
	}

	if trustCertErr := TrustCert(certPath); trustCertErr != nil {
		return trustCertErr
	}

	return nil
}

type certCache struct {
	lock  sync.Mutex
	cache *lru.ARCCache
}

func NewCertCache() *certCache {
	cache, err := lru.NewARC(1024)
	if err != nil {
		panic(err)
	}

	return &certCache{
		cache: cache,
	}
}

func (c *certCache) GetCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	name := clientHello.ServerName

	if val, ok := c.cache.Get(name); ok {
		return val.(*tls.Certificate), nil
	}

	cert, err := makeCert(CACert, name)
	if err != nil {
		return nil, err
	}

	c.cache.Add(name, cert)

	return cert, nil
}

// Generate a specific certificate for a given application domain name, e.g., "foo.test"
func makeCert(parent *tls.Certificate, name string) (*tls.Certificate, error) {
	// start by generating private key
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// create certificate structure with proper values
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %v", err)
	}

	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Puma-dev Signed"},
			CommonName:   name,
		},
		NotBefore:   notBefore,
		NotAfter:    notAfter,
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	cert.DNSNames = append(cert.DNSNames, name)

	x509parent, err := x509.ParseCertificate(parent.Certificate[0])
	if err != nil {
		return nil, err
	}

	derBytes, err := x509.CreateCertificate(
		rand.Reader, cert, x509parent, privKey.Public(), parent.PrivateKey)

	if err != nil {
		return nil, fmt.Errorf("could not create certificate: %v", err)
	}

	tlsCert := &tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  privKey,
		Leaf:        cert,
	}

	return tlsCert, nil
}
