package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the schema for .costctl.yaml
type Config struct {
	Clouds      []string            `yaml:"clouds"`
	TagMappings map[string][]string `yaml:"tag_mappings"`
	Thresholds  Thresholds          `yaml:"thresholds"`
	Budget      Budget              `yaml:"budget"`
}

// Thresholds defines waste detection limits
type Thresholds struct {
	CPUUtilizationPercent float64 `yaml:"cpu_utilization_percent"`
	IdleDays              int     `yaml:"idle_days"`
	SnapshotRetentionDays int     `yaml:"snapshot_retention_days"`
	DiskUnattachedDays    int     `yaml:"disk_unattached_days"`
}

// Budget defines budget rules, useful for CI/CD checks
type Budget struct {
	TotalMonthly          float64 `yaml:"total_monthly"`
	FailOnWasteThreshold  float64 `yaml:"fail_on_waste_threshold"`
}

// DefaultConfig returns the default configurations for costctl
func DefaultConfig() *Config {
	return &Config{
		Clouds: []string{"aws", "azure", "gcp"},
		TagMappings: map[string][]string{
			"owner":       {"owner", "Owner", "owner-contact", "creator", "CreatedBy"},
			"environment": {"env", "environment", "Environment", "stage", "Stage"},
			"project":     {"project", "Project", "app", "application", "Application"},
		},
		Thresholds: Thresholds{
			CPUUtilizationPercent: 5.0,
			IdleDays:              7,
			SnapshotRetentionDays: 30,
			DiskUnattachedDays:    3,
		},
		Budget: Budget{
			TotalMonthly:         10000.0,
			FailOnWasteThreshold: 500.0,
		},
	}
}

// LoadConfig loads configuration from a file or falls back to default
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		// Try default locations: current dir, then home dir
		path = ".costctl.yaml"
		if _, err := os.Stat(path); os.IsNotExist(err) {
			home, err := os.UserHomeDir()
			if err == nil {
				path = filepath.Join(home, ".costctl.yaml")
			}
		}
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml config: %w", err)
	}

	return cfg, nil
}

// NormalizeTags maps cloud-specific tags to canonical dimensions (owner, environment, project)
func (c *Config) NormalizeTags(rawTags map[string]string) map[string]string {
	normalized := make(map[string]string)
	
	// Pre-fill canonical tags with "unknown"
	for canonical := range c.TagMappings {
		normalized[canonical] = "unknown"
	}

	for key, val := range rawTags {
		matched := false
		for canonical, variations := range c.TagMappings {
			for _, variation := range variations {
				if strings.EqualFold(key, variation) {
					normalized[canonical] = val
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		// Preserve other tags as-is
		if !matched {
			normalized[strings.ToLower(key)] = val
		}
	}
	
	return normalized
}
