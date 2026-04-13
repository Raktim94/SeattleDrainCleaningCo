package httpapi

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/nodedr/submify/apps/api/internal/config"
)

// OriginAllowed reports whether the browser Origin header is permitted for this request.
// Order: explicit ALLOWED_ORIGINS → host suffixes → same-host as request (tunnel / reverse proxy)
// → optional private-LAN relaxation.
func OriginAllowed(origin string, r *http.Request, cfg config.Config) bool {
	if origin == "" {
		return true
	}
	for _, o := range cfg.AllowedOrigins {
		if o == origin {
			return true
		}
	}
	if originMatchesHostSuffix(origin, cfg.CorsOriginHostSuffixes) {
		return true
	}
	if cfg.CorsAllowSameHostOrigin && originMatchesRequestHost(origin, r) {
		return true
	}
	if !cfg.CorsRelaxPrivateNetworks {
		return false
	}
	return isRelaxedHTTPOrigin(origin)
}

// originMatchesRequestHost is true when Origin's host:port equals the public host the client used
// (Host / X-Forwarded-Host + X-Forwarded-Proto), e.g. Cloudflare Tunnel or nginx in front of the API.
func originMatchesRequestHost(origin string, r *http.Request) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	oh := strings.ToLower(u.Hostname())
	op := u.Port()
	if op == "" {
		if u.Scheme == "https" {
			op = "443"
		} else {
			op = "80"
		}
	}

	raw := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if i := strings.Index(raw, ","); i >= 0 {
		raw = strings.TrimSpace(raw[:i])
	}
	if raw == "" {
		raw = r.Host
	}
	rh, rp, err := net.SplitHostPort(raw)
	if err != nil {
		rh = strings.ToLower(strings.TrimSpace(raw))
		rp = ""
	} else {
		rh = strings.ToLower(rh)
	}
	if rp == "" {
		proto := forwardedProto(r)
		if proto == "https" {
			rp = "443"
		} else {
			rp = "80"
		}
	}

	return oh == rh && op == rp
}

func forwardedProto(r *http.Request) string {
	p := strings.ToLower(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")))
	switch p {
	case "https", "http":
		return p
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

// originMatchesHostSuffix allows https://api.example.com when suffix is "example.com".
func originMatchesHostSuffix(origin string, suffixes []string) bool {
	if len(suffixes) == 0 {
		return false
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for _, suf := range suffixes {
		suf = strings.ToLower(strings.TrimSpace(suf))
		if suf == "" {
			continue
		}
		if host == suf {
			return true
		}
		if strings.HasSuffix(host, "."+suf) {
			return true
		}
	}
	return false
}

func isRelaxedHTTPOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	if ip != nil {
		return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
	}
	if strings.HasSuffix(host, ".local") {
		return true
	}
	return false
}
