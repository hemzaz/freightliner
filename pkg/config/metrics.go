package config

// MetricsConfig holds metrics-specific configuration
type MetricsConfig struct {
	Enabled   bool   `yaml:"enabled" env:"METRICS_ENABLED" default:"true"`
	Port      int    `yaml:"port" env:"METRICS_PORT" default:"2112"`
	Path      string `yaml:"path" env:"METRICS_PATH" default:"/metrics"`
	Namespace string `yaml:"namespace" env:"METRICS_NAMESPACE" default:"freightliner"`
}
