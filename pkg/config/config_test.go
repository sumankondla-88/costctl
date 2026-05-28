package config

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if len(cfg.Clouds) != 3 {
		t.Errorf("Expected 3 default clouds, got %d", len(cfg.Clouds))
	}

	if cfg.Thresholds.CPUUtilizationPercent != 5.0 {
		t.Errorf("Expected default CPU utilization threshold to be 5.0, got %f", cfg.Thresholds.CPUUtilizationPercent)
	}

	if _, ok := cfg.TagMappings["owner"]; !ok {
		t.Error("Expected default tag mappings to contain 'owner'")
	}
}

func TestNormalizeTags(t *testing.T) {
	cfg := DefaultConfig()

	rawTags := map[string]string{
		"Owner-Contact": "devops@company.com",
		"env":           "production",
		"App":           "billing-system",
		"custom-tag":    "custom-value",
	}

	normalized := cfg.NormalizeTags(rawTags)

	// Check mapping from 'Owner-Contact' to canonical 'owner'
	if normalized["owner"] != "devops@company.com" {
		t.Errorf("Expected owner to be devops@company.com, got %s", normalized["owner"])
	}

	// Check mapping from 'env' to canonical 'environment'
	if normalized["environment"] != "production" {
		t.Errorf("Expected environment to be production, got %s", normalized["environment"])
	}

	// Check mapping from 'App' to canonical 'project'
	if normalized["project"] != "billing-system" {
		t.Errorf("Expected project to be billing-system, got %s", normalized["project"])
	}

	// Check that non-mapped tags are preserved in lowercase
	if normalized["custom-tag"] != "custom-value" {
		t.Errorf("Expected custom-tag to be preserved as custom-value, got %s", normalized["custom-tag"])
	}
}
