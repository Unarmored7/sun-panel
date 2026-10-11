package remoteimage

import (
	"context"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"testing"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestPublicAddressPolicy(t *testing.T) {
	for _, raw := range []string{"127.0.0.1", "10.0.0.1", "172.16.0.1", "192.168.1.1", "169.254.169.254", "100.64.0.1", "192.0.2.1", "198.18.0.1", "::1", "fc00::1", "fe80::1", "::ffff:127.0.0.1", "64:ff9b::a00:1", "2001:db8::1", "2002:a00:1::"} {
		if publicAddress(netip.MustParseAddr(raw)) {
			t.Fatal(raw)
		}
	}
	for _, raw := range []string{"1.1.1.1", "8.8.8.8", "2606:4700:4700::1111"} {
		if !publicAddress(netip.MustParseAddr(raw)) {
			t.Fatal(raw)
		}
	}
}
func TestURLAndDownloadLimits(t *testing.T) {
	for _, raw := range []string{"file:///etc/passwd", "ftp://example.com/a", "https://user:password@example.com", "http://example.com:3002/a"} {
		if _, err := ValidateURL(raw); err == nil {
			t.Fatal(raw)
		}
	}
	client := &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("oversized body")), Request: r, Header: http.Header{}, ContentLength: -1}, nil
	})}
	if _, _, _, err := Get(context.Background(), client, "https://example.com/image", 4); err == nil {
		t.Fatal("unbounded chunked body")
	}
	realClient := NewClient()
	if _, _, _, err := Get(context.Background(), realClient, "http://127.0.0.1/a", 100); err == nil {
		t.Fatal("private host fetched")
	}
	if err := realClient.CheckRedirect(&http.Request{URL: mustURL(t, "http://example.com:3002")}, nil); err == nil {
		t.Fatal("redirect port allowed")
	}
}
func mustURL(t *testing.T, s string) *url.URL {
	t.Helper()
	u, err := url.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return u
}
