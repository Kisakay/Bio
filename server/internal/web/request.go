package web

import (
	"net"
	"net/http"
	"strings"
)

func clientIP(r *http.Request) string {
	for _, header := range []string{"CF-Connecting-IP", "X-Forwarded-For", "X-Real-IP"} {
		value := strings.TrimSpace(r.Header.Get(header))
		if value == "" {
			continue
		}

		if header == "X-Forwarded-For" {
			value = strings.TrimSpace(strings.Split(value, ",")[0])
		}

		if parsed := normalizeIP(value); parsed != "" {
			return parsed
		}
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return normalizeIP(host)
	}

	return normalizeIP(r.RemoteAddr)
}

func normalizeIP(value string) string {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil {
		return ""
	}

	return ip.String()
}
