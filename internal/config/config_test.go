package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shipfe-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &Config{
		Region:         "us-west-2",
		BucketName:     "test-bucket",
		DistributionID: "E123456789",
		CloudFrontURL:  "d123456789.cloudfront.net",
		OACID:          "E987654321",
		Framework:      "vite",
		BuildOutputDir: "dist",
	}

	// Save
	if err := Save(tmpDir, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tmpDir, ".shipfe", "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Load
	loaded, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify values
	if loaded.Region != cfg.Region {
		t.Errorf("expected region %s, got %s", cfg.Region, loaded.Region)
	}
	if loaded.BucketName != cfg.BucketName {
		t.Errorf("expected bucket %s, got %s", cfg.BucketName, loaded.BucketName)
	}
	if loaded.DistributionID != cfg.DistributionID {
		t.Errorf("expected distribution ID %s, got %s", cfg.DistributionID, loaded.DistributionID)
	}
	if loaded.CreatedAt == "" {
		t.Error("expected CreatedAt to be set")
	}
}

func TestLoadNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shipfe-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = Load(tmpDir)
	if err == nil {
		t.Error("expected error for non-existent config")
	}
}

func TestExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shipfe-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Should not exist initially
	if Exists(tmpDir) {
		t.Error("config should not exist initially")
	}

	// Create config
	cfg := &Config{BucketName: "test"}
	if err := Save(tmpDir, cfg); err != nil {
		t.Fatal(err)
	}

	// Should exist now
	if !Exists(tmpDir) {
		t.Error("config should exist after save")
	}
}

func TestRemove(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shipfe-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config
	cfg := &Config{BucketName: "test"}
	if err := Save(tmpDir, cfg); err != nil {
		t.Fatal(err)
	}

	// Remove
	if err := Remove(tmpDir); err != nil {
		t.Fatalf("failed to remove config: %v", err)
	}

	// Should not exist
	if Exists(tmpDir) {
		t.Error("config should not exist after remove")
	}
}
