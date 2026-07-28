package python

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wadefengx/wade/internal/config"
)

// PipMirror represents a pip registry mirror
type PipMirror struct {
	Name string
	URL  string
}

// PipPresets returns built-in pip mirrors
func PipPresets() []PipMirror {
	return []PipMirror{
		{Name: "pypi", URL: "https://pypi.org/simple/"},
		{Name: "tsinghua", URL: "https://pypi.tuna.tsinghua.edu.cn/simple/"},
		{Name: "aliyun", URL: "https://mirrors.aliyun.com/pypi/simple/"},
		{Name: "huawei", URL: "https://mirrors.huaweicloud.com/pypi/simple/"},
		{Name: "tencent", URL: "https://mirrors.tencent.com/pypi/simple/"},
		{Name: "ustc", URL: "https://pypi.mirrors.ustc.edu.cn/simple/"},
	}
}

// FindPipMirror finds a pip mirror by name
func FindPipMirror(name string) (*PipMirror, bool) {
	for _, m := range PipPresets() {
		if m.Name == name {
			return &m, true
		}
	}
	return nil, false
}

// UsePipMirror switches pip to a mirror
func UsePipMirror(name string) error {
	mirror, ok := FindPipMirror(name)
	if !ok {
		return fmt.Errorf("unknown pip mirror: %s", name)
	}

	// pip config set global.index-url <url>
	return exec.Command("pip", "config", "set", "global.index-url", mirror.URL).Run()
}

// GoProxy represents a Go proxy mirror
type GoProxy struct {
	Name string
	URL  string
}

// GoProxyPresets returns built-in Go proxies
func GoProxyPresets() []GoProxy {
	return []GoProxy{
		{Name: "official", URL: "https://proxy.golang.org,direct"},
		{Name: "goproxy.cn", URL: "https://goproxy.cn,direct"},
		{Name: "goproxy.io", URL: "https://goproxy.io,direct"},
	}
}

// FindGoProxy finds a Go proxy by name
func FindGoProxy(name string) (*GoProxy, bool) {
	for _, p := range GoProxyPresets() {
		if p.Name == name {
			return &p, true
		}
	}
	return nil, false
}

// UseGoProxy switches Go proxy
func UseGoProxy(name string) error {
	proxy, ok := FindGoProxy(name)
	if !ok {
		return fmt.Errorf("unknown Go proxy: %s", name)
	}
	return exec.Command("go", "env", "-w", fmt.Sprintf("GOPROXY=%s", proxy.URL)).Run()
}

// DetectSystemPython returns the detected Python version
func DetectSystemPython() []string {
	var versions []string
	for _, name := range []string{"python3", "python"} {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		out, err := exec.Command(path, "--version").Output()
		if err != nil {
			continue
		}
		ver := strings.TrimSpace(string(out))
		versions = append(versions, fmt.Sprintf("%s (system: %s)", path, ver))
	}
	return versions
}

// DetectSystemGo returns the detected Go version
func DetectSystemGo() string {
	path, err := exec.LookPath("go")
	if err != nil {
		return ""
	}
	out, err := exec.Command(path, "version").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// GoMirror represents a Go download mirror
type GoMirror struct {
	Name string
	URL  string
}

// GoMirrorPresets returns built-in Go download mirrors
func GoMirrorPresets() []GoMirror {
	return []GoMirror{
		{Name: "official", URL: "https://go.dev/dl/"},
		{Name: "google-cn", URL: "https://golang.google.cn/dl/"},
		{Name: "npmmirror", URL: "https://npmmirror.com/mirrors/go/"},
		{Name: "aliyun", URL: "https://mirrors.aliyun.com/go/"},
	}
}

// FindGoMirror finds a Go mirror by name
func FindGoMirror(name string) (*GoMirror, bool) {
	for _, m := range GoMirrorPresets() {
		if m.Name == name {
			return &m, true
		}
	}
	return nil, false
}

// EnsureDir creates the go/python directory under ~/.wade/
func EnsureDir() error {
	dir, err := config.WadeDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(dir, "go", "versions"), 0755)
}