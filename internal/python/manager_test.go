package python

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPipPresets(t *testing.T) {
	want := []PipMirror{
		{Name: "pypi", URL: "https://pypi.org/simple/"},
		{Name: "tsinghua", URL: "https://pypi.tuna.tsinghua.edu.cn/simple/"},
		{Name: "aliyun", URL: "https://mirrors.aliyun.com/pypi/simple/"},
		{Name: "huawei", URL: "https://mirrors.huaweicloud.com/pypi/simple/"},
		{Name: "tencent", URL: "https://mirrors.tencent.com/pypi/simple/"},
		{Name: "ustc", URL: "https://pypi.mirrors.ustc.edu.cn/simple/"},
	}

	got := PipPresets()
	if len(got) != len(want) {
		t.Fatalf("PipPresets() returned %d mirrors, want %d", len(got), len(want))
	}
	for i, wantMirror := range want {
		if got[i] != wantMirror {
			t.Errorf("PipPresets()[%d] = %+v, want %+v", i, got[i], wantMirror)
		}
	}
}

func TestFindPipMirror(t *testing.T) {
	tests := []struct {
		name string
		want *PipMirror
		ok   bool
	}{
		{name: "tsinghua", want: &PipMirror{Name: "tsinghua", URL: "https://pypi.tuna.tsinghua.edu.cn/simple/"}, ok: true},
		{name: "missing", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := FindPipMirror(tt.name)
			if ok != tt.ok {
				t.Fatalf("FindPipMirror(%q) found = %v, want %v", tt.name, ok, tt.ok)
			}
			if tt.want == nil {
				if got != nil {
					t.Errorf("FindPipMirror(%q) = %+v, want nil", tt.name, got)
				}
			} else if got == nil || *got != *tt.want {
				t.Errorf("FindPipMirror(%q) = %+v, want %+v", tt.name, got, tt.want)
			}
		})
	}
}

func TestGoProxyPresets(t *testing.T) {
	want := []GoProxy{
		{Name: "official", URL: "https://proxy.golang.org,direct"},
		{Name: "goproxy.cn", URL: "https://goproxy.cn,direct"},
		{Name: "goproxy.io", URL: "https://goproxy.io,direct"},
	}

	got := GoProxyPresets()
	if len(got) != len(want) {
		t.Fatalf("GoProxyPresets() returned %d proxies, want %d", len(got), len(want))
	}
	for i, wantProxy := range want {
		if got[i] != wantProxy {
			t.Errorf("GoProxyPresets()[%d] = %+v, want %+v", i, got[i], wantProxy)
		}
	}
}

func TestFindGoProxy(t *testing.T) {
	tests := []struct {
		name string
		want *GoProxy
		ok   bool
	}{
		{name: "goproxy.cn", want: &GoProxy{Name: "goproxy.cn", URL: "https://goproxy.cn,direct"}, ok: true},
		{name: "missing", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := FindGoProxy(tt.name)
			if ok != tt.ok {
				t.Fatalf("FindGoProxy(%q) found = %v, want %v", tt.name, ok, tt.ok)
			}
			if tt.want == nil {
				if got != nil {
					t.Errorf("FindGoProxy(%q) = %+v, want nil", tt.name, got)
				}
			} else if got == nil || *got != *tt.want {
				t.Errorf("FindGoProxy(%q) = %+v, want %+v", tt.name, got, tt.want)
			}
		})
	}
}

func TestGoMirrorPresets(t *testing.T) {
	want := []GoMirror{
		{Name: "official", URL: "https://go.dev/dl/"},
		{Name: "google-cn", URL: "https://golang.google.cn/dl/"},
	}

	got := GoMirrorPresets()
	if len(got) != len(want) {
		t.Fatalf("GoMirrorPresets() returned %d mirrors, want %d", len(got), len(want))
	}
	for i, wantMirror := range want {
		if got[i] != wantMirror {
			t.Errorf("GoMirrorPresets()[%d] = %+v, want %+v", i, got[i], wantMirror)
		}
	}
}

func TestFindGoMirror(t *testing.T) {
	tests := []struct {
		name string
		want *GoMirror
		ok   bool
	}{
		{name: "google-cn", want: &GoMirror{Name: "google-cn", URL: "https://golang.google.cn/dl/"}, ok: true},
		{name: "missing", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := FindGoMirror(tt.name)
			if ok != tt.ok {
				t.Fatalf("FindGoMirror(%q) found = %v, want %v", tt.name, ok, tt.ok)
			}
			if tt.want == nil {
				if got != nil {
					t.Errorf("FindGoMirror(%q) = %+v, want nil", tt.name, got)
				}
			} else if got == nil || *got != *tt.want {
				t.Errorf("FindGoMirror(%q) = %+v, want %+v", tt.name, got, tt.want)
			}
		})
	}
}

func TestEnsureDir(t *testing.T) {
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	if err := EnsureDir(); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(tmpHome, ".wade", "go", "versions")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Errorf("EnsureDir() created %q, which is not a directory", path)
	}
}

func TestUseUnknownMirrorOrProxy(t *testing.T) {
	tests := []struct {
		name string
		use  func(string) error
	}{
		{name: "pip mirror", use: UsePipMirror},
		{name: "Go proxy", use: UseGoProxy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.use("missing"); err == nil {
				t.Errorf("%s unknown name error = nil, want non-nil", tt.name)
			}
		})
	}
}
