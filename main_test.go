package main

import (
	"testing"
)

func TestIPFilterGeolocation(t *testing.T) {
	// Basic test to ensure the plugin can be instantiated and configured

	matcher := &IPFilterGeolocation{
		AllowCountries: []string{"AU"},
		BlockByDefault: true,
	}

	// Test configuration validation
	err := matcher.Validate()
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	// Test that the matcher has the right configuration
	if len(matcher.AllowCountries) != 1 {
		t.Errorf("Expected 1 allowed country, got %d", len(matcher.AllowCountries))
	}

	if matcher.AllowCountries[0] != "AU" {
		t.Errorf("Expected AU as allowed country, got %s", matcher.AllowCountries[0])
	}

	if !matcher.BlockByDefault {
		t.Error("Expected BlockByDefault to be true")
	}
}

func TestIPFilterGeolocationConflicts(t *testing.T) {
	// Test that conflicting allow/deny configurations are rejected

	matcher := &IPFilterGeolocation{
		AllowCountries: []string{"AU"},
		DenyCountries:  []string{"US"},
	}

	err := matcher.Validate()
	if err == nil {
		t.Error("Expected validation to fail with conflicting allow/deny countries")
	}
}

func TestIPFilterGeolocationEmpty(t *testing.T) {
	// Test that empty configuration is valid

	matcher := &IPFilterGeolocation{}

	err := matcher.Validate()
	if err != nil {
		t.Errorf("Expected empty configuration to be valid, got error: %v", err)
	}
}
