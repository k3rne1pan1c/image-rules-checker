package rules

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the rule configuration
type Config struct {
	Rules RuleConfigs `yaml:"rules"`
}

// RuleConfigs contains all rule configurations
type RuleConfigs struct {
	MinDimensions    *DimensionRule `yaml:"min_dimensions,omitempty"`
	MaxDimensions    *DimensionRule `yaml:"max_dimensions,omitempty"`
	RequireTimestamp *BoolRule      `yaml:"require_timestamp,omitempty"`
	GPSAllowed       *GPSRule       `yaml:"gps_allowed,omitempty"`
	RequireOrientation *BoolRule    `yaml:"require_orientation,omitempty"`
}

// DimensionRule checks image dimensions
type DimensionRule struct {
	Enabled bool `yaml:"enabled"`
	Width   int  `yaml:"width"`
	Height  int  `yaml:"height"`
}

// BoolRule is a simple boolean check
type BoolRule struct {
	Enabled bool `yaml:"enabled"`
}

// GPSRule checks GPS data presence
type GPSRule struct {
	Enabled bool `yaml:"enabled"`
	Allowed bool `yaml:"allowed"` // true = GPS allowed, false = GPS not allowed
}

// LoadConfig loads rule configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Rules: RuleConfigs{
			MinDimensions: &DimensionRule{
				Enabled: true,
				Width:   800,
				Height:  600,
			},
			MaxDimensions: &DimensionRule{
				Enabled: false,
				Width:   10000,
				Height:  10000,
			},
			RequireTimestamp: &BoolRule{
				Enabled: true,
			},
			GPSAllowed: &GPSRule{
				Enabled: true,
				Allowed: false, // GPS not allowed by default
			},
			RequireOrientation: &BoolRule{
				Enabled: false,
			},
		},
	}
}

