package commands

import (
	"testing"
)

func TestGenerateCommand_Structure(t *testing.T) {
	cmd := NewGenerateCommand()

	if cmd.Use != "generate" {
		t.Errorf("expected Use to be 'generate', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	if len(cmd.Aliases) == 0 {
		t.Error("expected aliases to be set")
	}

	hasGenAlias := false
	for _, alias := range cmd.Aliases {
		if alias == "gen" {
			hasGenAlias = true
			break
		}
	}
	if !hasGenAlias {
		t.Error("expected 'gen' alias to be present")
	}
}

func TestGenerateCommand_Subcommands(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
	}{
		{
			name:        "entity subcommand is registered",
			commandName: "entity",
		},
		{
			name:        "usecase subcommand is registered",
			commandName: "usecase",
		},
		{
			name:        "ep subcommand is registered",
			commandName: "ep",
		},
		{
			name:        "da subcommand is registered",
			commandName: "da",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateCommand()
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

func TestGenerateCommand_LegacySubcommandsAreRemoved(t *testing.T) {
	removed := []string{"service", "handler", "consumer", "adapter"}

	cmd := NewGenerateCommand()

	for _, name := range removed {
		t.Run(name+" is no longer a top-level subcommand", func(t *testing.T) {
			subCmd, _, err := cmd.Find([]string{name})
			if err == nil && subCmd != nil && subCmd.Name() == name {
				t.Errorf("legacy subcommand %q should not be registered on `generate` anymore", name)
			}
		})
	}
}

func TestGenerateEntryPointCommand_Structure(t *testing.T) {
	cmd := NewGenerateEntryPointCommand()

	if cmd.Use != "ep" {
		t.Errorf("expected Use to be 'ep', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}
}

func TestGenerateEntryPointCommand_Subcommands(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
	}{
		{
			name:        "http subcommand is registered",
			commandName: "http",
		},
		{
			name:        "consumer subcommand is registered",
			commandName: "consumer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateEntryPointCommand()
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

func TestGenerateDrivenAdapterCommand_Structure(t *testing.T) {
	cmd := NewGenerateDrivenAdapterCommand()

	if cmd.Use != "da" {
		t.Errorf("expected Use to be 'da', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}
}

func TestGenerateDrivenAdapterCommand_Subcommands(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
	}{
		{
			name:        "database subcommand is registered",
			commandName: "database",
		},
		{
			name:        "api subcommand is registered",
			commandName: "api",
		},
		{
			name:        "publisher subcommand is registered",
			commandName: "publisher",
		},
		{
			name:        "cache subcommand is registered",
			commandName: "cache",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateDrivenAdapterCommand()
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

func TestGenerateCommand_HelpOutput(t *testing.T) {
	cmd := NewGenerateCommand()

	if cmd.Long == "" {
		t.Error("expected Long description to provide help information")
	}

	if !cmd.HasSubCommands() {
		t.Error("generate command should have subcommands")
	}
}

func TestGenerateDrivenAdapterCommand_HelpOutput(t *testing.T) {
	cmd := NewGenerateDrivenAdapterCommand()

	if cmd.Long == "" {
		t.Error("expected Long description to provide help information")
	}

	if !cmd.HasSubCommands() {
		t.Error("da command should have subcommands")
	}
}

func TestGenerateEntryPointCommand_HelpOutput(t *testing.T) {
	cmd := NewGenerateEntryPointCommand()

	if cmd.Long == "" {
		t.Error("expected Long description to provide help information")
	}

	if !cmd.HasSubCommands() {
		t.Error("ep command should have subcommands")
	}
}
