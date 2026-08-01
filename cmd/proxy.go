package cmd

import (
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// newHTTPClient returns an http.Client with the given timeout that honors
// proxies: explicit HTTP_PROXY/HTTPS_PROXY env vars first (Go default), then
// the Windows system proxy (registry-backed) — matching what PowerShell's
// Invoke-WebRequest uses. This is why irm|iex works in PowerShell but Go
// binaries hang on blocked networks.
func newHTTPClient(timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = proxyFunc()
	return &http.Client{Timeout: timeout, Transport: transport}
}

// proxyFunc resolves the proxy for a request. Falls back to the Windows
// system proxy when no HTTP_PROXY/HTTPS_PROXY env vars are set.
func proxyFunc() func(*http.Request) (*url.URL, error) {
	envProxy := http.ProxyFromEnvironment
	return func(req *http.Request) (*url.URL, error) {
		// env vars win (Go default behavior)
		if u, err := envProxy(req); err == nil && u != nil {
			return u, nil
		}
		// Windows system proxy (IE settings) — only when env proxy is absent
		if runtime.GOOS == "windows" {
			if p := windowsSystemProxy(); p != "" {
				return url.Parse(p)
			}
		}
		return nil, nil
	}
}

// windowsSystemProxy reads the WinINET proxy from the registry
// (HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings).
// Returns "" when disabled or unset. Uses `reg query` (no external deps).
func windowsSystemProxy() string {
	out, err := exec.Command("reg", "query",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		"/v", "ProxyEnable").CombinedOutput()
	if err != nil || !strings.Contains(string(out), "0x1") {
		return ""
	}
	out, err = exec.Command("reg", "query",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		"/v", "ProxyServer").CombinedOutput()
	if err != nil {
		return ""
	}
	return parseProxyServerLine(string(out))
}

// parseProxyServerLine extracts a usable proxy URL from a `reg query` output
// line. Handles: plain "127.0.0.1:7890", "http://127.0.0.1:7890", and
// per-protocol lists "http=...;https=..." (prefers the https entry).
func parseProxyServerLine(line string) string {
	idx := strings.LastIndex(line, "REG_SZ")
	if idx == -1 {
		return ""
	}
	proxy := strings.TrimSpace(line[idx+len("REG_SZ"):])
	if proxy == "" {
		return ""
	}
	// If it's a per-protocol list, prefer the https entry
	for _, part := range strings.Split(proxy, ";") {
		if strings.HasPrefix(part, "https=") {
			addr := strings.TrimPrefix(part, "https=")
			if !strings.Contains(addr, "://") {
				addr = "http://" + addr
			}
			return addr
		}
	}
	if !strings.Contains(proxy, "://") {
		proxy = "http://" + proxy
	}
	return proxy
}
