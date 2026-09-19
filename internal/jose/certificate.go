package jose

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

// SelfSigned makes a self-signed certificate for a key, as PEM.
//
// It is what a SAML service provider hands an identity provider: nobody
// checks it against an authority — the provider is given this exact
// certificate in our metadata and trusts it by being given it — so there is no
// authority to ask for one.
func SelfSigned(key crypto.Signer, commonName string, validFor time.Duration) (string, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 126))
	if err != nil {
		return "", err
	}

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(validFor),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, key.Public(), key)
	if err != nil {
		return "", fmt.Errorf("jose: make a certificate: %w", err)
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})), nil
}

// ParseCertificate reads a certificate from PEM.
func ParseCertificate(certificate string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certificate))
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("jose: not a PEM certificate")
	}

	return x509.ParseCertificate(block.Bytes)
}
