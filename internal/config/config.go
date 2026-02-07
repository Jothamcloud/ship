package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	configDir  = ".shipfe"
	configFile = "config.json"
)

// Config represents the deployment configuration
type Config struct {
	Region         string `json:"region"`
	BucketName     string `json:"bucket_name"`
	DistributionID string `json:"distribution_id"`
	CloudFrontURL  string `json:"cloudfront_url"`
	OACID          string `json:"oac_id"`
	Framework      string `json:"framework"`
	BuildOutputDir string `json:"build_output_dir"`
	CreatedAt      string `json:"created_at"`
}

// configPath returns the full path to the config file
func configPath(projectDir string) string {
	return filepath.Join(projectDir, configDir, configFile)
}

// Load reads the config from .shipfe/config.json
func Load(projectDir string) (*Config, error) {
	path := configPath(projectDir)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config not found: %w", err)
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// Save writes the config to .shipfe/config.json
func Save(projectDir string, cfg *Config) error {
	dir := filepath.Join(projectDir, configDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set creation time if not already set
	if cfg.CreatedAt == "" {
		cfg.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	path := configPath(projectDir)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Remove deletes the .shipfe directory
func Remove(projectDir string) error {
	dir := filepath.Join(projectDir, configDir)
	return os.RemoveAll(dir)
}

// Exists checks if a config exists for the project
func Exists(projectDir string) bool {
	path := configPath(projectDir)
	_, err := os.Stat(path)
	return err == nil
}
