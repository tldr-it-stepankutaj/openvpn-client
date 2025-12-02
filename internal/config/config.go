package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigPath = "/etc/openvpn/client/config.yaml"
	DefaultTimeout    = 10 * time.Second
	DefaultSessionDir = "/var/run/openvpn"
	DefaultFirewall   = "nftables"

	EnvConfigPath   = "OPENVPN_CLIENT_CONFIG"
	EnvAPIBaseURL   = "OPENVPN_API_BASE_URL"
	EnvAPIToken     = "OPENVPN_API_TOKEN"
	EnvAPIUsername  = "OPENVPN_API_USERNAME"
	EnvAPIPassword  = "OPENVPN_API_PASSWORD"
	EnvAPITimeout   = "OPENVPN_API_TIMEOUT"
	EnvSessionDir   = "OPENVPN_SESSION_DIR"
	EnvFirewallType = "OPENVPN_FIREWALL_TYPE"
)

type Config struct {
	API      APIConfig      `yaml:"api"`
	OpenVPN  OpenVPNConfig  `yaml:"openvpn"`
	Firewall FirewallConfig `yaml:"firewall"`
}

type APIConfig struct {
	BaseURL  string        `yaml:"base_url"`
	Token    string        `yaml:"token"`
	Username string        `yaml:"username"`
	Password string        `yaml:"password"`
	Timeout  time.Duration `yaml:"timeout"`
}

type OpenVPNConfig struct {
	SessionDir string `yaml:"session_dir"`
}

type FirewallConfig struct {
	Type     string         `yaml:"type"`
	NFTables NFTablesConfig `yaml:"nftables"`
	IPTables IPTablesConfig `yaml:"iptables"`
}

type NFTablesConfig struct {
	RulesFile     string `yaml:"rules_file"`
	ReloadCommand string `yaml:"reload_command"`
}

type IPTablesConfig struct {
	ChainName     string `yaml:"chain_name"`
	RulesFile     string `yaml:"rules_file"`
	ReloadCommand string `yaml:"reload_command"`
}

// UseToken returns true if API token authentication should be used
func (c *APIConfig) UseToken() bool {
	return c.Token != ""
}

// Load loads configuration from file with environment variable overrides.
// Priority: CLI argument > environment variable > config file > defaults
func Load(configPath string) (*Config, error) {
	// Determine config path
	path := resolveConfigPath(configPath)

	// Read and parse config file
	cfg, err := loadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
	}

	// Apply environment variable overrides
	applyEnvOverrides(cfg)

	// Apply defaults
	applyDefaults(cfg)

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func resolveConfigPath(cliPath string) string {
	if cliPath != "" {
		return cliPath
	}
	if envPath := os.Getenv(EnvConfigPath); envPath != "" {
		return envPath
	}
	return DefaultConfigPath
}

func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv(EnvAPIBaseURL); v != "" {
		cfg.API.BaseURL = v
	}
	if v := os.Getenv(EnvAPIToken); v != "" {
		cfg.API.Token = v
	}
	if v := os.Getenv(EnvAPIUsername); v != "" {
		cfg.API.Username = v
	}
	if v := os.Getenv(EnvAPIPassword); v != "" {
		cfg.API.Password = v
	}
	if v := os.Getenv(EnvAPITimeout); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.API.Timeout = d
		}
	}
	if v := os.Getenv(EnvSessionDir); v != "" {
		cfg.OpenVPN.SessionDir = v
	}
	if v := os.Getenv(EnvFirewallType); v != "" {
		cfg.Firewall.Type = v
	}
}

func applyDefaults(cfg *Config) {
	if cfg.API.Timeout == 0 {
		cfg.API.Timeout = DefaultTimeout
	}
	if cfg.OpenVPN.SessionDir == "" {
		cfg.OpenVPN.SessionDir = DefaultSessionDir
	}
	if cfg.Firewall.Type == "" {
		cfg.Firewall.Type = DefaultFirewall
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.API.BaseURL == "" {
		return fmt.Errorf("api.base_url is required")
	}

	// Must have either token or username/password
	if c.API.Token == "" && (c.API.Username == "" || c.API.Password == "") {
		return fmt.Errorf("api.token or api.username/password is required")
	}

	if c.Firewall.Type != "nftables" && c.Firewall.Type != "iptables" {
		return fmt.Errorf("firewall.type must be 'nftables' or 'iptables'")
	}

	return nil
}
