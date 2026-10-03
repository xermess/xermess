// Package jose signs and verifies this server's JWTs and publishes its keys. It
// is deliberately small: only RS256, PS256 and ES256 with our own keys, no
// "none", no HMAC, and a token's header never chooses the key.
package jose

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// The algorithms a key can sign with.
const (
	RS256 = "RS256"
	PS256 = "PS256"
	ES256 = "ES256"
)

// Algorithms lists every algorithm, RS256 first: it is what every OpenID
// Connect client has to support, so ID tokens use it.
var Algorithms = []string{RS256, PS256, ES256}

// ErrInvalid covers a malformed token, an unknown key and a bad signature
// alike; callers treat them the same.
var ErrInvalid = errors.New("invalid token")

var b64 = base64.RawURLEncoding

// Key is a private key and what it is published as — or, for a key somebody
// else holds, only its public half, which is enough to verify with.
type Key struct {
	ID        string
	Algorithm string
	Private   crypto.Signer
	// PublicKey is the public half of a key this server does not hold, such
	// as an identity provider's (ParseJWK). It is ignored when Private is set.
	PublicKey crypto.PublicKey
}

// public is the key's public half, whichever way it was given.
func (k Key) public() crypto.PublicKey {
	if k.Private != nil {
		return k.Private.Public()
	}

	return k.PublicKey
}

// Header is the part of a JWS header this package writes and reads.
type Header struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid,omitempty"`
	Type      string `json:"typ,omitempty"`
}

// Generate makes a new private key for an algorithm: a 2048-bit RSA key for
// RS256 and PS256, a P-256 key for ES256.
func Generate(algorithm string) (crypto.Signer, error) {
	switch algorithm {
	case RS256, PS256:
		return rsa.GenerateKey(rand.Reader, 2048)
	case ES256:
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	default:
		return nil, fmt.Errorf("jose: unsupported algorithm %q", algorithm)
	}
}

// Sign returns a compact JWS. `typ` is "at+jwt" for access tokens (RFC 9068) or
// "JWT" for ID tokens.
func Sign(key Key, typ string, claims any) (string, error) {
	header, err := json.Marshal(Header{Algorithm: key.Algorithm, KeyID: key.ID, Type: typ})
	if err != nil {
		return "", err
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	input := b64.EncodeToString(header) + "." + b64.EncodeToString(payload)

	signature, err := sign(key, []byte(input))
	if err != nil {
		return "", err
	}

	return input + "." + b64.EncodeToString(signature), nil
}

func sign(key Key, input []byte) ([]byte, error) {
	digest := sha256.Sum256(input)

	switch key.Algorithm {
	case RS256:
		private, ok := key.Private.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("jose: %s needs an RSA key", key.Algorithm)
		}
		return rsa.SignPKCS1v15(rand.Reader, private, crypto.SHA256, digest[:])
	case PS256:
		private, ok := key.Private.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("jose: %s needs an RSA key", key.Algorithm)
		}
		return rsa.SignPSS(rand.Reader, private, crypto.SHA256, digest[:], &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash})
	case ES256:
		private, ok := key.Private.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("jose: %s needs an EC key", key.Algorithm)
		}
		r, s, err := ecdsa.Sign(rand.Reader, private, digest[:])
		if err != nil {
			return nil, err
		}
		// JWS wants r and s as two fixed 32-byte numbers side by side, not
		// the ASN.1 structure Go's SignASN1 would make.
		out := make([]byte, 64)
		r.FillBytes(out[:32])
		s.FillBytes(out[32:])
		return out, nil
	default:
		return nil, fmt.Errorf("jose: unsupported algorithm %q", key.Algorithm)
	}
}

// Verify checks a compact JWS against the key `lookup` returns for its kid and
// decodes the claims. The key decides the algorithm, so a header cannot
// downgrade the check. Claims (issuer, expiry, audience) are the caller's to
// check.
func Verify(token string, lookup func(kid string) (Key, bool), claims any) (Header, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Header{}, ErrInvalid
	}

	rawHeader, err := b64.DecodeString(parts[0])
	if err != nil {
		return Header{}, ErrInvalid
	}

	var header Header
	if err := json.Unmarshal(rawHeader, &header); err != nil {
		return Header{}, ErrInvalid
	}

	key, ok := lookup(header.KeyID)
	if !ok || key.Algorithm != header.Algorithm {
		return Header{}, ErrInvalid
	}

	signature, err := b64.DecodeString(parts[2])
	if err != nil {
		return Header{}, ErrInvalid
	}

	if !verify(key, []byte(parts[0]+"."+parts[1]), signature) {
		return Header{}, ErrInvalid
	}

	payload, err := b64.DecodeString(parts[1])
	if err != nil {
		return Header{}, ErrInvalid
	}

	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(claims); err != nil {
		return Header{}, ErrInvalid
	}

	return header, nil
}

func verify(key Key, input, signature []byte) bool {
	digest := sha256.Sum256(input)

	switch public := key.public().(type) {
	case *rsa.PublicKey:
		switch key.Algorithm {
		case RS256:
			return rsa.VerifyPKCS1v15(public, crypto.SHA256, digest[:], signature) == nil
		case PS256:
			return rsa.VerifyPSS(public, crypto.SHA256, digest[:], signature, &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash}) == nil
		}
	case *ecdsa.PublicKey:
		if key.Algorithm != ES256 || len(signature) != 64 {
			return false
		}
		r := new(big.Int).SetBytes(signature[:32])
		s := new(big.Int).SetBytes(signature[32:])
		return ecdsa.Verify(public, digest[:], r, s)
	}

	return false
}

// JWK is a public key as a JSON Web Key (RFC 7517).
type JWK struct {
	KeyType   string `json:"kty"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`

	// RSA
	N string `json:"n,omitempty"`
	E string `json:"e,omitempty"`

	// EC
	Curve string `json:"crv,omitempty"`
	X     string `json:"x,omitempty"`
	Y     string `json:"y,omitempty"`
}

// Public returns the key's public half as a JWK, which is what the JWKS
// endpoint publishes.
func (k Key) Public() (JWK, error) {
	jwk := JWK{Use: "sig", Algorithm: k.Algorithm, KeyID: k.ID}

	switch public := k.Private.Public().(type) {
	case *rsa.PublicKey:
		jwk.KeyType = "RSA"
		jwk.N = b64.EncodeToString(public.N.Bytes())
		jwk.E = b64.EncodeToString(big.NewInt(int64(public.E)).Bytes())
	case *ecdsa.PublicKey:
		point, err := public.Bytes()
		if err != nil {
			return JWK{}, err
		}
		// Uncompressed point: 0x04, then X and Y, 32 bytes each on P-256.
		jwk.KeyType = "EC"
		jwk.Curve = "P-256"
		jwk.X = b64.EncodeToString(point[1:33])
		jwk.Y = b64.EncodeToString(point[33:65])
	default:
		return JWK{}, fmt.Errorf("jose: unsupported key type %T", public)
	}

	return jwk, nil
}

// ParseJWK reads another party's published public key. RSA keys verify RS256 or
// PS256, P-256 keys ES256; a key without alg gets its type's usual one.
func ParseJWK(jwk JWK) (Key, error) {
	key := Key{ID: jwk.KeyID, Algorithm: jwk.Algorithm}

	switch jwk.KeyType {
	case "RSA":
		n, err := b64.DecodeString(jwk.N)
		if err != nil {
			return Key{}, fmt.Errorf("jose: the key's modulus is not base64url: %w", err)
		}
		e, err := b64.DecodeString(jwk.E)
		if err != nil || len(e) == 0 || len(e) > 4 {
			return Key{}, fmt.Errorf("jose: the key's exponent is not usable")
		}
		if len(n) < 256 {
			return Key{}, fmt.Errorf("jose: an RSA key shorter than 2048 bits is not trusted")
		}

		key.PublicKey = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(new(big.Int).SetBytes(e).Int64())}
		if key.Algorithm == "" {
			key.Algorithm = RS256
		}
	case "EC":
		if jwk.Curve != "P-256" {
			return Key{}, fmt.Errorf("jose: the curve %q is not supported", jwk.Curve)
		}
		x, errX := b64.DecodeString(jwk.X)
		y, errY := b64.DecodeString(jwk.Y)
		if errX != nil || errY != nil || len(x) != 32 || len(y) != 32 {
			return Key{}, fmt.Errorf("jose: the key's point is not usable")
		}

		public, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), append(append([]byte{4}, x...), y...))
		if err != nil {
			return Key{}, fmt.Errorf("jose: the key's point is not on the curve: %w", err)
		}
		key.PublicKey = public
		if key.Algorithm == "" {
			key.Algorithm = ES256
		}
	default:
		return Key{}, fmt.Errorf("jose: the key type %q is not supported", jwk.KeyType)
	}

	return key, nil
}

// HalfHash is at_hash (OpenID Connect Core 3.1.3.6): the base64url left half of
// a SHA-256.
func HalfHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return b64.EncodeToString(sum[:len(sum)/2])
}
