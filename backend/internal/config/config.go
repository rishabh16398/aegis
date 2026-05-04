package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	Scanner     ScannerConfig     `yaml:"scanner"`
	Network     NetworkConfig     `yaml:"network"`
	ThreatIntel ThreatIntelConfig `yaml:"threat_intel"`
	LogLevel    string            `yaml:"log_level"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type ScannerConfig struct {
	Workers        int      `yaml:"workers"`
	WatchPaths     []string `yaml:"watch_paths"`
	QuarantinePath string   `yaml:"quarantine_path"`
}

type NetworkConfig struct {
	Interface string `yaml:"interface"`
	Enabled   bool   `yaml:"enabled"`
}

type ThreatIntelConfig struct {
	VirusTotalKey string `yaml:"virustotal_key"`
	AbuseIPDBKey  string `yaml:"abuseipdb_key"`
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg.applyDefaults()
	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Server.Host == "" {
		c.Server.Host = "127.0.0.1"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Database.Path == "" {
		c.Database.Path = "./aegis.db"
	}
	if c.Scanner.Workers == 0 {
		c.Scanner.Workers = 4
	}
	if c.Scanner.QuarantinePath == "" {
		c.Scanner.QuarantinePath = "./quarantine"
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
}
