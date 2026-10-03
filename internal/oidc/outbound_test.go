package oidc

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBlockedAddress(t *testing.T) {
	for _, tt := range []struct {
		address string
		blocked bool
	}{
		{"169.254.169.254", true},
		{"::ffff:169.254.169.254", true},
		{"10.0.0.5", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"100.100.100.200", true},
		{"fd00::1", true},
		{"fe80::1", true},
		{"0.0.0.0", true},
		{"224.0.0.1", true},
		{"8.8.8.8", false},
		{"2001:4860:4860::8888", false},
		{"100.128.0.1", false},
	} {
		if got := blockedAddress(net.ParseIP(tt.address)); got != tt.blocked {
			t.Errorf("blockedAddress(%s) = %v, want %v", tt.address, got, tt.blocked)
		}
	}
}

func TestOnLoopback(t *testing.T) {
	for _, tt := range []struct {
		issuer string
		want   bool
	}{
		{"http://localhost:5173", true},
		{"http://127.0.0.1:8080", true},
		{"http://[::1]:8080", true},
		{"https://id.localhost", true},
		{"https://id.example.com", false},
		{"https://localhost.example.com", false},
	} {
		if got := onLoopback(tt.issuer); got != tt.want {
			t.Errorf("onLoopback(%s) = %v, want %v", tt.issuer, got, tt.want)
		}
	}
}

// A provider on this machine is reachable from a server in development and
// refused by one in production, where loopback is the admin listener.
func TestFederationClientAndLoopback(t *testing.T) {
	local := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer local.Close()

	for _, tt := range []struct {
		name          string
		allowLoopback bool
		refused       bool
	}{
		{"a server on loopback reaches it", true, false},
		{"a server in production is refused", false, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			res, err := newFederationClient(time.Second, tt.allowLoopback).Get(local.URL)
			if err == nil {
				res.Body.Close()
			}
			if refused := errors.Is(err, errBlockedAddress); refused != tt.refused {
				t.Errorf("err = %v, want refused %v", err, tt.refused)
			}
		})
	}
}
