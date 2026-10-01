package oidc

import (
	"errors"
	"net"
	"net/http"
	"syscall"
	"time"
)

// maxOutboundRedirects caps how far a federation request will follow a
// provider's redirects before giving up.
const maxOutboundRedirects = 5

// errBlockedAddress is a federation request aimed at an address that is not
// out on the public internet.
var errBlockedAddress = errors.New("oidc: refusing to connect to a non-public address")

// newFederationClient builds the HTTP client every outbound federation call
// uses — social sign-in (discovery, token, userinfo), SSO (discovery, JWKS,
// userinfo, token, SAML metadata). It refuses to connect to a private,
// link-local, unspecified or multicast address, both at the first dial and
// at every redirect, so a connection an administrator sets up, or a social
// sign-in anybody can start, cannot be turned into a request against this
// server's own network: the cloud metadata service at 169.254.169.254, a
// database on 10.x, another service on the same host.
//
// The check runs in the dialer, on the address DNS actually resolved to, so a
// name that resolves to a private address is caught too. Loopback is left
// allowed on purpose: a provider on this machine over plain http is how one
// is tried out (see model/sso.go and internal/api/sso/validation.go).
func newFederationClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{Timeout: timeout}
	dialer.Control = func(_, address string, _ syscall.RawConn) error {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return err
		}
		if ip := net.ParseIP(host); ip == nil || blockedAddress(ip) {
			return errBlockedAddress
		}
		return nil
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: &http.Transport{DialContext: dialer.DialContext},
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			if len(via) >= maxOutboundRedirects {
				return errors.New("oidc: too many redirects")
			}
			return nil
		},
	}
}

// blockedAddress reports whether a resolved address is one no federation
// request may reach. Loopback is deliberately not blocked.
func blockedAddress(ip net.IP) bool {
	return ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified()
}
