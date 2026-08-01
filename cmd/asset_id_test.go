package cmd

import "testing"

// TestAssetIDByName uses a real GitHub-style release JSON body to verify the
// backwards id lookup (id appears BEFORE name in GitHub asset objects).
func TestAssetIDByName(t *testing.T) {
	body := `{"assets":[
{"url":"https://api.github.com/repos/wadefengx/wade/releases/assets/111","id":111,"node_id":"x","name":"wade-windows-amd64.sha256","label":"","uploader":{"id":41898282}},
{"url":"https://api.github.com/repos/wadefengx/wade/releases/assets/222","id":222,"node_id":"y","name":"wade-windows-amd64.zip","label":"","uploader":{"id":41898282}},
{"url":"https://api.github.com/repos/wadefengx/wade/releases/assets/333","id":333,"node_id":"z","name":"wade-darwin-arm64.tar.gz","label":"","uploader":{"id":41898282}}
]}`

	cases := []struct {
		name, want string
	}{
		{"wade-windows-amd64.zip", "222"},
		{"wade-darwin-arm64.tar.gz", "333"},
		{"wade-windows-amd64.sha256", "111"},
	}
	for _, c := range cases {
		got, err := assetIDByName(body, c.name)
		if err != nil {
			t.Errorf("assetIDByName(%q): %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("assetIDByName(%q) = %s, want %s", c.name, got, c.want)
		}
	}

	// missing asset
	if _, err := assetIDByName(body, "wade-linux-amd64.zip"); err == nil {
		t.Error("expected error for missing asset")
	}
}
