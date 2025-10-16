package config

type OutputType string

const (
	OutputStatic OutputType = "static"
	OutputServer OutputType = "server"
	OutputHybrid OutputType = "hybrid"
)

type AdapterName string

const (
	AdapterStandalone AdapterName = "standalone"
	AdapterCloudflare AdapterName = "cloudflare"
	AdapterNetlify    AdapterName = "netlify"
	AdapterVercel     AdapterName = "vercel"
)

type Config struct {
	Site           string          `toml:"site"`
	Base           string          `toml:"base"`
	OutDir         string          `toml:"outDir"`
	SrcDir         string          `toml:"srcDir"`
	PackageManager string          `toml:"packageManager"`
	Output         OutputConfig    `toml:"output"`
	Server         ServerConfig    `toml:"server"`
	Adapter        AdapterConfig   `toml:"adapter"`
	Lifecycle      LifecycleConfig `toml:"lifecycle"`
	Plugins        []PluginConfig  `toml:"plugins"`
}

type OutputConfig struct {
	Type OutputType `toml:"type"`
}

type ServerConfig struct {
	Port int    `toml:"port"`
	Host string `toml:"host"`
}

type AdapterConfig struct {
	Name   AdapterName            `toml:"name"`
	Config map[string]interface{} `toml:"config"`
}

type LifecycleConfig struct {
	Enabled         bool `toml:"enabled"`
	StartupTimeout  int  `toml:"startupTimeout"`
	ShutdownTimeout int  `toml:"shutdownTimeout"`
}

type PluginConfig struct {
	Name   string                 `toml:"name"`
	Config map[string]interface{} `toml:"config"`
}

func DefaultConfig() *Config {
	return &Config{
		Site:           "",
		Base:           "/",
		OutDir:         "./dist",
		SrcDir:         "./src",
		PackageManager: "npm",
		Output: OutputConfig{
			Type: OutputStatic,
		},
		Server: ServerConfig{
			Port: 4322,
			Host: "localhost",
		},
		Adapter: AdapterConfig{
			Name:   AdapterStandalone,
			Config: make(map[string]interface{}),
		},
		Lifecycle: LifecycleConfig{
			Enabled:         true,
			StartupTimeout:  30,
			ShutdownTimeout: 10,
		},
	}
}
