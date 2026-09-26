package jose

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"errors"
	"fmt"
)

// Sealer encrypts private keys before they are stored, with AES-256-GCM under
// a key derived from the server's secret. A database dump without the secret
// holds nothing that can sign a token.
type Sealer struct {
	aead cipher.AEAD
}

// NewSealer derives the encryption key from `secret`. The secret is already
// long and random, so a single SHA-256 is enough to shape it into a key.
func NewSealer(secret string) (*Sealer, error) {
	sum := sha256.Sum256([]byte(secret))

	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &Sealer{aead: aead}, nil
}

// Seal encodes a private key and encrypts it.
func (s *Sealer) Seal(key crypto.Signer) ([]byte, error) {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("jose: encode private key: %w", err)
	}

	return s.SealBytes(der)
}

// SealBytes encrypts any secret the server has to be able to read back — a
// signing key, an administrator's TOTP seed: a fresh nonce, then the
// ciphertext.
func (s *Sealer) SealBytes(plain []byte) ([]byte, error) {
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return s.aead.Seal(nonce, nonce, plain, nil), nil
}

// OpenBytes decrypts what SealBytes made.
func (s *Sealer) OpenBytes(sealed []byte) ([]byte, error) {
	size := s.aead.NonceSize()
	if len(sealed) < size {
		return nil, ErrWrongSecret
	}

	plain, err := s.aead.Open(nil, sealed[:size], sealed[size:], nil)
	if err != nil {
		return nil, ErrWrongSecret
	}

	return plain, nil
}

// ErrWrongSecret is returned when a stored key cannot be decrypted: almost
// always because LOGINER_SECRET_KEY is not the one it was sealed with.
var ErrWrongSecret = errors.New("jose: a signing key could not be decrypted; is LOGINER_SECRET_KEY the one it was made with?")

// Open decrypts and decodes a key Seal made.
func (s *Sealer) Open(sealed []byte) (crypto.Signer, error) {
	der, err := s.OpenBytes(sealed)
	if err != nil {
		return nil, err
	}

	key, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, fmt.Errorf("jose: decode private key: %w", err)
	}

	signer, ok := key.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("jose: stored key of type %T cannot sign", key)
	}

	return signer, nil
}
