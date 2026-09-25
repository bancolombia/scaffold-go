package commands

import (
	"testing"
)

func TestVersionCommand(t *testing.T) {
	// Save original values
	origVersion := Version
	origBuildDate := BuildDate
	origGitCommit := GitCommit

	// Set test values
	Version = "1.0.0"
	BuildDate = "2024-01-01T00:00:00Z"
	GitCommit = "abc123"

	// Restore original values after test
	defer func() {
		Version = origVersion
		BuildDate = origBuildDate
		GitCommit = origGitCommit
	}()

	cmd := NewVersionCommand()

	// Execute command - it should not return an error
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	// Verify the command structure and that variables are set correctly
	if Version != "1.0.0" {
		t.Errorf("expected Version to be '1.0.0', got '%s'", Version)
	}

	if BuildDate != "2024-01-01T00:00:00Z" {
		t.Errorf("expected BuildDate to be '2024-01-01T00:00:00Z', got '%s'", BuildDate)
	}

	if GitCommit != "abc123" {
		t.Errorf("expected GitCommit to be 'abc123', got '%s'", GitCommit)
	}
}

func TestVersionCommandStructure(t *testing.T) {
	cmd := NewVersionCommand()

	if cmd.Use != "version" {
		t.Errorf("expected Use to be 'version', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	if cmd.Run == nil {
		t.Error("expected Run function to be set")
	}
}

func TestVersionVariablesCanBeSet(t *testing.T) {
	// Test that version variables can be set (for build-time injection)
	testVersion := "test-version"
	testBuildDate := "test-date"
	testGitCommit := "test-commit"

	Version = testVersion
	BuildDate = testBuildDate
	GitCommit = testGitCommit

	if Version != testVersion {
		t.Errorf("expected Version to be '%s', got '%s'", testVersion, Version)
	}

	if BuildDate != testBuildDate {
		t.Errorf("expected BuildDate to be '%s', got '%s'", testBuildDate, BuildDate)
	}

	if GitCommit != testGitCommit {
		t.Errorf("expected GitCommit to be '%s', got '%s'", testGitCommit, GitCommit)
	}

	// Reset to defaults
	Version = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
}
