package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigWiringForDatabaseAdapter validates that generating a mongodb DA
// idempotently appends the `mongodb:` block to the generated project's
// config/config.yaml, and leaves the pre-existing top-level keys intact.
func TestConfigWiringForDatabaseAdapter(t *testing.T) {
	tmpDir := t.TempDir()
	cliPath := buildCLI(t)

	projectName := "config-wiring-test"
	modulePath := "github.com/test/configwiring"
	projectDir := filepath.Join(tmpDir, projectName)

	// Init project.
	initCmd := exec.Command(cliPath, "init", projectName,
		"--module", modulePath,
		"--output", tmpDir)
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\n%s", err, out)
	}

	// Snapshot the initial config.yaml — auto-injection must not touch it.
	initialCfg, err := os.ReadFile(filepath.Join(projectDir, "config", "config.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(initialCfg), "server:")
	assert.Contains(t, string(initialCfg), "logging:")
	assert.NotContains(t, string(initialCfg), "mongodb:", "mongodb block should not be in the initial config")

	// Generate a Product entity so the DA has a valid target.
	entityCmd := exec.Command(cliPath, "generate", "entity", "Product",
		"--fields", "Name:string,Price:float64")
	entityCmd.Dir = projectDir
	if out, err := entityCmd.CombinedOutput(); err != nil {
		t.Fatalf("generate entity failed: %v\n%s", err, out)
	}

	daCmd := exec.Command(cliPath, "generate", "da", "database",
		"--type", "mongodb",
		"--entity", "Product",
		"--fields", "Name:string,Price:float64")
	daCmd.Dir = projectDir
	if out, err := daCmd.CombinedOutput(); err != nil {
		t.Fatalf("generate da database failed: %v\n%s", err, out)
	}

	// mongodb block should now be present.
	cfgAfter, err := os.ReadFile(filepath.Join(projectDir, "config", "config.yaml"))
	require.NoError(t, err)
	body := string(cfgAfter)
	assert.Contains(t, body, "mongodb:", "mongodb block should be injected after mongo DA generation")
	assert.Contains(t, body, "uri: mongodb://%s:%s@localhost:27017", "mongodb.uri default should be written")
	//assert.Contains(t, body, "database: "+projectName, "mongodb.database default should use the project name")
	// Existing content preserved.
	assert.Contains(t, body, "server:", "server block should still be present")
	assert.Contains(t, body, "logging:", "logging block should still be present")

	// Re-run — must be idempotent (exactly one top-level mongodb key).
	daCmd2 := exec.Command(cliPath, "generate", "da", "database",
		"--type", "mongodb",
		"--entity", "Product",
		"--fields", "Name:string,Price:float64")
	daCmd2.Dir = projectDir
	if out, err := daCmd2.CombinedOutput(); err != nil {
		t.Fatalf("re-run generate da database failed: %v\n%s", err, out)
	}
	cfgRerun, err := os.ReadFile(filepath.Join(projectDir, "config", "config.yaml"))
	require.NoError(t, err)
	got := strings.Count(string(cfgRerun), "\nmongodb:")
	assert.Equal(t, 1, got, "mongodb block should appear exactly once after re-run")
}
