package cmd

import "testing"

// TestWindowsSystemProxyParse covers the registry output parsing
// (REG_SZ value extraction + per-protocol https preference + scheme default).
func TestWindowsSystemProxyParse(t *testing.T) {
	cases := []struct {
		name, input, want string
	}{
		{"plain address", `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Internet Settings
    ProxyServer    REG_SZ    127.0.0.1:7890`, "http://127.0.0.1:7890"},
		{"with http scheme", `    ProxyServer    REG_SZ    http://127.0.0.1:7890`, "http://127.0.0.1:7890"},
		{"per-protocol list", `    ProxyServer    REG_SZ    http=127.0.0.1:7890;https=127.0.0.1:7890`, "http://127.0.0.1:7890"},
		{"empty", `    ProxyServer    REG_SZ    `, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// monkey-patch the exec via a small indirection is overkill;
			// instead we exercise the parsing helper extracted from the raw line.
			got := parseProxyServerLine(c.input)
			if got != c.want {
				t.Errorf("parseProxyServerLine(%q) = %q, want %q", c.input, got, c.want)
			}
		})
	}
}
