package registry

// Registry represents a package registry with metadata
type Registry struct {
	Name      string
	URL       string
	IsBuiltIn bool
}

// Presets returns all built-in registries
func Presets() []Registry {
	return []Registry{
		{Name: "npm", URL: "https://registry.npmjs.org/", IsBuiltIn: true},
		{Name: "taobao", URL: "https://registry.npmmirror.com/", IsBuiltIn: true},
		{Name: "tencent", URL: "https://mirrors.tencent.com/npm/", IsBuiltIn: true},
		{Name: "huawei", URL: "https://repo.huaweicloud.com/repository/npm/", IsBuiltIn: true},
		{Name: "cnpm", URL: "http://r.cnpmjs.org/", IsBuiltIn: true},
	}
}

// All returns all registries (presets + custom from config)
func All(custom []Registry) []Registry {
	all := Presets()
	all = append(all, custom...)
	return all
}

// Find searches for a registry by name in presets + custom
func Find(name string, custom []Registry) (*Registry, bool) {
	for _, r := range Presets() {
		if r.Name == name {
			return &r, true
		}
	}
	for _, r := range custom {
		if r.Name == name {
			r.IsBuiltIn = false
			return &r, true
		}
	}
	return nil, false
}

// IsBuiltIn checks if a registry name is a built-in preset
func IsBuiltIn(name string) bool {
	for _, r := range Presets() {
		if r.Name == name {
			return true
		}
	}
	return false
}
