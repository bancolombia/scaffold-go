package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func TestNewRootCommand(t *testing.T) {
	tests := []struct {
		name string
		want struct {
			use   string
			short string
			long  string
		}
	}{
		{
			name: "root command has correct metadata",
			want: struct {
				use   string
				short string
				long  string
			}{
				use:   "scaffold-go",
				short: "A Go CLI tool for scaffolding microservice projects",
				long:  "scaffold-go is a code generation tool that scaffolds Go microservice projects \nfollowing clean architecture principles.\n\nThe tool generates projects with three distinct layers:\n  • Domain Layer: Pure business logic with no external dependencies\n  • EntryPoints Layer: External interfaces (HTTP, messaging)\n  • DrivenAdapters Layer: Infrastructure implementations (databases, APIs, caching)\n\nThe dependency rule is strictly enforced: dependencies point inward toward the domain.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCommand()

			if cmd.Use != tt.want.use {
				t.Errorf("Use = %q, want %q", cmd.Use, tt.want.use)
			}
			if cmd.Short != tt.want.short {
				t.Errorf("Short = %q, want %q", cmd.Short, tt.want.short)
			}
			if cmd.Long != tt.want.long {
				t.Errorf("Long = %q, want %q", cmd.Long, tt.want.long)
			}
		})
	}
}

func TestRootCommand_GlobalFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		wantType string
	}{
		{
			name:     "config flag is registered",
			flagName: "config",
			wantType: "string",
		},
		{
			name:     "verbose flag is registered",
			flagName: "verbose",
			wantType: "bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCommand()
			flag := cmd.PersistentFlags().Lookup(tt.flagName)

			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}

			if flag.Value.Type() != tt.wantType {
				t.Errorf("flag %q type = %q, want %q", tt.flagName, flag.Value.Type(), tt.wantType)
			}
		})
	}
}

func TestRootCommand_VerboseFlagDefault(t *testing.T) {
	cmd := NewRootCommand()
	flag := cmd.PersistentFlags().Lookup("verbose")

	if flag == nil {
		t.Fatal("verbose flag not found")
	}

	if flag.DefValue != "false" {
		t.Errorf("verbose flag default = %q, want %q", flag.DefValue, "false")
	}
}

func TestRootCommand_Subcommands(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
	}{
		{
			name:        "init subcommand is registered",
			commandName: "init",
		},
		{
			name:        "generate subcommand is registered",
			commandName: "generate",
		},
		{
			name:        "version subcommand is registered",
			commandName: "version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCommand()
			subCmd, _, err := cmd.Find([]string{tt.commandName})

			if err != nil {
				t.Fatalf("error finding subcommand %q: %v", tt.commandName, err)
			}

			if subCmd.Name() != tt.commandName {
				t.Errorf("subcommand name = %q, want %q", subCmd.Name(), tt.commandName)
			}
		})
	}
}

func TestRootCommand_FlagParsing(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError bool
		checkFunc func(*testing.T, *bytes.Buffer)
	}{
		{
			name:      "parse config flag",
			args:      []string{"--config", "/path/to/config.yaml"},
			wantError: false,
			checkFunc: func(t *testing.T, buf *bytes.Buffer) {
				if cfgFile != "/path/to/config.yaml" {
					t.Errorf("cfgFile = %q, want %q", cfgFile, "/path/to/config.yaml")
				}
			},
		},
		{
			name:      "parse verbose flag long form",
			args:      []string{"--verbose"},
			wantError: false,
			checkFunc: func(t *testing.T, buf *bytes.Buffer) {
				if !verbose {
					t.Error("verbose flag not set")
				}
			},
		},
		{
			name:      "parse verbose flag short form",
			args:      []string{"-v"},
			wantError: false,
			checkFunc: func(t *testing.T, buf *bytes.Buffer) {
				if !verbose {
					t.Error("verbose flag not set")
				}
			},
		},
		{
			name:      "parse both flags",
			args:      []string{"--config", "test.yaml", "--verbose"},
			wantError: false,
			checkFunc: func(t *testing.T, buf *bytes.Buffer) {
				if cfgFile != "test.yaml" {
					t.Errorf("cfgFile = %q, want %q", cfgFile, "test.yaml")
				}
				if !verbose {
					t.Error("verbose flag not set")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global variables
			cfgFile = ""
			verbose = false
			logger = nil

			cmd := NewRootCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, buf)
			}
		})
	}
}

func TestInitConfig_WithCustomConfigFile(t *testing.T) {
	// Create temporary config file
	tempDir, err := os.MkdirTemp("", "root-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "test-config.yaml")
	configContent := []byte("test_key: test_value\n")
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Reset viper
	viper.Reset()

	// Set config file
	cfgFile = configPath

	// Call initConfig
	initConfig()

	// Verify config was loaded
	if viper.ConfigFileUsed() != configPath {
		t.Errorf("ConfigFileUsed = %q, want %q", viper.ConfigFileUsed(), configPath)
	}

	// Verify config value
	if viper.GetString("test_key") != "test_value" {
		t.Errorf("test_key = %q, want %q", viper.GetString("test_key"), "test_value")
	}
}

func TestInitConfig_WithoutCustomConfigFile(t *testing.T) {
	// Reset viper
	viper.Reset()

	// Clear config file
	cfgFile = ""

	// Call initConfig
	initConfig()

	// Verify default config paths are set
	// Note: We can't easily test the actual config loading without a real config file,
	// but we can verify the function doesn't panic
	if viper.ConfigFileUsed() != "" {
		// Config file was found and loaded (this is fine)
		t.Logf("Config file loaded: %s", viper.ConfigFileUsed())
	}
}

func TestInitLogger_ProductionMode(t *testing.T) {
	// Reset logger
	logger = nil
	verbose = false

	// Call initLogger
	initLogger()

	// Verify logger is initialized
	if logger == nil {
		t.Fatal("logger not initialized")
	}

	// Verify logger is production logger (we can't directly check the type,
	// but we can verify it was created without panic)
	defer logger.Sync()
}

func TestInitLogger_DevelopmentMode(t *testing.T) {
	// Reset logger
	logger = nil
	verbose = true

	// Call initLogger
	initLogger()

	// Verify logger is initialized
	if logger == nil {
		t.Fatal("logger not initialized")
	}

	// Verify logger is development logger
	defer logger.Sync()
}

func TestGetLogger(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func()
		wantNonNil  bool
		verboseMode bool
	}{
		{
			name: "get logger when already initialized",
			setupFunc: func() {
				logger, _ = zap.NewProduction()
			},
			wantNonNil:  true,
			verboseMode: false,
		},
		{
			name: "get logger when not initialized",
			setupFunc: func() {
				logger = nil
				verbose = false
			},
			wantNonNil:  true,
			verboseMode: false,
		},
		{
			name: "get logger in verbose mode",
			setupFunc: func() {
				logger = nil
				verbose = true
			},
			wantNonNil:  true,
			verboseMode: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			got := GetLogger()

			if tt.wantNonNil && got == nil {
				t.Error("GetLogger() returned nil, want non-nil")
			}

			if got != nil {
				defer got.Sync()
			}
		})
	}
}

func TestGetValidator(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func()
		wantNonNil bool
	}{
		{
			name: "get validator when not initialized",
			setupFunc: func() {
				validatorInstance = nil
			},
			wantNonNil: true,
		},
		{
			name: "get validator when already initialized",
			setupFunc: func() {
				// Validator is already initialized from previous test
			},
			wantNonNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			got := GetValidator()

			if tt.wantNonNil && got == nil {
				t.Error("GetValidator() returned nil, want non-nil")
			}

			// Verify it's the same instance on subsequent calls
			got2 := GetValidator()
			if got != got2 {
				t.Error("GetValidator() returned different instances")
			}
		})
	}
}

func TestRootCommand_HelpOutput(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check that help output contains key information
	requiredStrings := []string{
		"scaffold-go",
		"Usage:",
		"Available Commands:",
		"Flags:",
		"--config",
		"--verbose",
		"-v",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(output, required) {
			t.Errorf("help output missing %q", required)
		}
	}
}

func TestRootCommand_PersistentPreRun(t *testing.T) {
	tests := []struct {
		name        string
		verboseFlag bool
		wantLogger  bool
	}{
		{
			name:        "logger initialized in production mode",
			verboseFlag: false,
			wantLogger:  true,
		},
		{
			name:        "logger initialized in development mode",
			verboseFlag: true,
			wantLogger:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset logger
			logger = nil
			verbose = tt.verboseFlag

			cmd := NewRootCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			args := []string{"version"} // Use version subcommand to trigger PersistentPreRun
			if tt.verboseFlag {
				args = append([]string{"--verbose"}, args...)
			}
			cmd.SetArgs(args)

			// Execute command (this will trigger PersistentPreRun)
			_ = cmd.Execute()

			// Check logger was initialized
			if tt.wantLogger && logger == nil {
				t.Error("logger not initialized after PersistentPreRun")
			}

			if logger != nil {
				defer logger.Sync()
			}
		})
	}
}
