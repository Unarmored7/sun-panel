package remoteimage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"syscall"
	"time"
)

func ValidateURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u == nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || len(raw) > 2000 {
		return nil, errors.New("a public HTTP(S) URL is required")
	}
	if p := u.Port(); p != "" && p != "80" && p != "443" {
		return nil, errors.New("unsupported remote port")
	}
	return u, nil
}

var reserved = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"), netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"), netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"), netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"), netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("2001:db8::/32"), netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("2002::/16"), netip.MustParsePrefix("2001::/32"),
}

func publicAddress(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	for _, block := range reserved {
		if block.Contains(ip) {
			return false
		}
	}
	// Restrict IPv6 to currently allocated global unicast space.
	return !ip.Is6() || netip.MustParsePrefix("2000::/3").Contains(ip)
}

var ErrNotPublic = errors.New("remote address is not public")

// StatusError reports a non-200 response, so callers can explain refusals.
type StatusError struct{ Code int }

func (e StatusError) Error() string { return fmt.Sprintf("remote status %d", e.Code) }

func NewClient() *http.Client {
	// The standard dialer resolves the host and races IPv6 and IPv4 (Happy Eyeballs),
	// so a host without a working IPv6 route still connects over IPv4. Control runs for
	// every address actually dialled, which also covers DNS answers that change.
	dialer := &net.Dialer{Timeout: 5 * time.Second, Control: func(network, address string, _ syscall.RawConn) error {
		host, port, err := net.SplitHostPort(address)
		if err != nil || (port != "80" && port != "443") {
			return errors.New("unsupported remote address")
		}
		ip, err := netip.ParseAddr(host)
		if err != nil || !publicAddress(ip) {
			return ErrNotPublic
		}
		return nil
	}}
	transport := &http.Transport{
		ForceAttemptHTTP2: true, TLSHandshakeTimeout: 6 * time.Second, ResponseHeaderTimeout: 8 * time.Second,
		IdleConnTimeout: 60 * time.Second, MaxIdleConns: 32, MaxConnsPerHost: 4,
		DialContext: dialer.DialContext,
	}
	return &http.Client{Transport: transport, Timeout: 10 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return errors.New("too many redirects")
		}
		_, err := ValidateURL(req.URL.String())
		return err
	}}
}

func Get(ctx context.Context, client *http.Client, raw string, limit int64) ([]byte, string, *url.URL, error) {
	u, err := ValidateURL(raw)
	if err != nil {
		return nil, "", nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, "", nil, err
	}
	// Many sites refuse requests that do not look like a browser.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	response, err := client.Do(req)
	if err != nil {
		return nil, "", nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, "", nil, StatusError{response.StatusCode}
	}
	if response.ContentLength > limit {
		return nil, "", nil, errors.New("remote file is too large")
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, "", nil, err
	}
	if int64(len(data)) > limit {
		return nil, "", nil, errors.New("remote file is too large")
	}
	return data, response.Header.Get("Content-Type"), response.Request.URL, nil
}
