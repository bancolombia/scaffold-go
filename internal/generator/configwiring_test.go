package generator

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfigYAML(t *testing.T, projectDir, contents string) string {
	t.Helper()
	cfgDir := filepath.Join(projectDir, "config")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	path := filepath.Join(cfgDir, "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}
	return path
}

func TestEnsureConfigBlock_AppendsWhenKeyAbsent(t *testing.T) {
	projectDir := t.TempDir()
	path := writeConfigYAML(t, projectDir, "server:\n  host: localhost\n  port: 8080\n")

	warn, err := EnsureConfigBlock(projectDir, "mongodb", "mongodb:\n  uri: mongodb://localhost:27017\n  database: demo\n")
	if err != nil {
		t.Fatalf("EnsureConfigBlock: %v", err)
	}
	if warn != nil {
		t.Fatalf("unexpected warning: %s", warn)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config.yaml: %v", err)
	}
	body := string(got)
	if !strings.Contains(body, "mongodb:\n  uri: mongodb://localhost:27017\n  database: demo") {
		t.Fatalf("appended block missing from config.yaml:\n%s", body)
	}
	if !strings.Contains(body, "server:\n  host: localhost") {
		t.Fatalf("existing content clobbered:\n%s", body)
	}
	if !strings.HasSuffix(body, "\n") {
		t.Fatalf("expected trailing newline")
	}
}

func TestEnsureConfigBlock_Idempotent(t *testing.T) {
	projectDir := t.TempDir()
	initial := "server:\n  host: localhost\n\nmongodb:\n  uri: existing-uri\n  database: demo\n"
	path := writeConfigYAML(t, projectDir, initial)

	warn, err := EnsureConfigBlock(projectDir, "mongodb", "mongodb:\n  uri: mongodb://localhost:27017\n  database: replaced\n")
	if err != nil {
		t.Fatalf("EnsureConfigBlock: %v", err)
	}
	if warn != nil {
		t.Fatalf("unexpected warning: %s", warn)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config.yaml: %v", err)
	}
	if !bytes.Equal(got, []byte(initial)) {
		t.Fatalf("file was modified on idempotent re-run:\nwant:\n%s\ngot:\n%s", initial, got)
	}
}

func TestEnsureConfigBlock_MissingFileWarns(t *testing.T) {
	projectDir := t.TempDir()
	warn, err := EnsureConfigBlock(projectDir, "mongodb", "mongodb:\n  uri: whatever\n")
	if err != nil {
		t.Fatalf("EnsureConfigBlock: %v", err)
	}
	if warn == nil {
		t.Fatalf("expected a warning when config.yaml is absent")
	}
	if !strings.Contains(warn.Error(), "file not found") {
		t.Fatalf("unexpected warning message: %s", warn.Error())
	}
}

func TestEnsureConfigBlock_MultipleKeysAppendSequentially(t *testing.T) {
	projectDir := t.TempDir()
	path := writeConfigYAML(t, projectDir, "server:\n  host: localhost\n")

	if _, err := EnsureConfigBlock(projectDir, "mongodb", "mongodb:\n  uri: m\n"); err != nil {
		t.Fatalf("first EnsureConfigBlock: %v", err)
	}
	if _, err := EnsureConfigBlock(projectDir, "dynamodb", "dynamodb:\n  region: us-east-1\n"); err != nil {
		t.Fatalf("second EnsureConfigBlock: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	body := string(got)
	mIdx := strings.Index(body, "\nmongodb:")
	dIdx := strings.Index(body, "\ndynamodb:")
	if mIdx < 0 || dIdx < 0 {
		t.Fatalf("both blocks should be present:\n%s", body)
	}
	if mIdx >= dIdx {
		t.Fatalf("mongodb block should come before dynamodb block:\n%s", body)
	}
}

func TestHasTopLevelKey_IgnoresNestedAndSubstrings(t *testing.T) {
	tests := []struct {
		name string
		body string
		key  string
		want bool
	}{
		{
			name: "top-level key with value",
			body: "mongodb:\n  uri: x\n",
			key:  "mongodb",
			want: true,
		},
		{
			name: "top-level key alone on line",
			body: "mongodb:\n",
			key:  "mongodb",
			want: true,
		},
		{
			name: "nested key is not top level",
			body: "database:\n  mongodb: legacy\n",
			key:  "mongodb",
			want: false,
		},
		{
			name: "substring inside a string value doesn't count",
			body: "notes: contains mongodb: reference\n",
			key:  "mongodb",
			want: false,
		},
		{
			name: "similar prefix doesn't count",
			body: "mongodbfoo:\n  x: y\n",
			key:  "mongodb",
			want: false,
		},
		{
			name: "key followed by comment",
			body: "mongodb: # a comment\n",
			key:  "mongodb",
			want: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := hasTopLevelKey([]byte(tc.body), tc.key)
			if got != tc.want {
				t.Fatalf("hasTopLevelKey(%q, %q) = %v, want %v", tc.body, tc.key, got, tc.want)
			}
		})
	}
}
