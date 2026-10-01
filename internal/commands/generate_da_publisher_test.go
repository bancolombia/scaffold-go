package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateDaPublisherCommand_Structure(t *testing.T) {
	cmd := NewGenerateDaPublisherCommand()

	if cmd.Use != "publisher [name]" {
		t.Errorf("expected Use to be 'publisher [name]', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	if cmd.RunE == nil {
		t.Error("expected RunE function to be set")
	}
}

func TestGenerateDaPublisherCommand_Flags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		wantType string
	}{
		{
			name:     "broker flag is registered",
			flagName: "broker",
			wantType: "string",
		},
		{
			name:     "pattern flag is registered",
			flagName: "pattern",
			wantType: "string",
		},
		{
			name:     "target flag is registered",
			flagName: "target",
			wantType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateDaPublisherCommand()
			flag := cmd.Flags().Lookup(tt.flagName)

			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}

			if flag.Value.Type() != tt.wantType {
				t.Errorf("flag %q type = %q, want %q", tt.flagName, flag.Value.Type(), tt.wantType)
			}
		})
	}
}

func TestGenerateDaPublisherCommand_Args(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError bool
	}{
		{
			name:      "no args returns error",
			args:      []string{},
			wantError: true,
		},
		{
			name:      "one arg is valid",
			args:      []string{"UserEventPublisher"},
			wantError: false,
		},
		{
			name:      "two args returns error",
			args:      []string{"UserEventPublisher", "Extra"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateDaPublisherCommand()
			cmd.SetArgs(tt.args)

			err := cmd.Args(cmd, tt.args)

			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGenerateDaPublisherCommand_ValidationError(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewGenerateDaPublisherCommand()
	cmd.SetArgs([]string{"UserEventPublisher", "--broker", "rabbitmq", "--pattern", "event"})

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for uninitialized project, got none")
	}
}

func TestGenerateDaPublisherCommand_KafkaPlaceholder(t *testing.T) {
	tmpDir := newInitializedTempProject(t)

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	cmd := NewGenerateDaPublisherCommand()
	cmd.SetArgs([]string{"OrderEventPublisher", "--broker", "kafka", "--pattern", "event"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected kafka to be rejected as not implemented")
	}
}

func TestGenerateDaPublisherCommand_Success_RabbitMQ(t *testing.T) {
	tmpDir := newInitializedTempProject(t)

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	cmd := NewGenerateDaPublisherCommand()
	cmd.SetArgs([]string{
		"UserEventPublisher",
		"--broker", "rabbitmq",
		"--pattern", "event",
		"--target", "user.created",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	publisherPath := filepath.Join(tmpDir, "internal", "drivenadapters", "messaging", "rabbitmq", "usereventpublisher_publisher.go")
	if _, err := os.Stat(publisherPath); os.IsNotExist(err) {
		t.Error("publisher file was not created")
	}

	moduleFile := filepath.Join(tmpDir, "internal", "drivenadapters", "messaging", "rabbitmq", "module.go")
	if _, err := os.Stat(moduleFile); os.IsNotExist(err) {
		t.Error("per-component module file was not created")
	}

	foundationFile := filepath.Join(tmpDir, "internal", "infrastructure", "messaging", "rabbitmq", "module.go")
	if _, err := os.Stat(foundationFile); os.IsNotExist(err) {
		t.Error("foundation module file was not created")
	}
}

func TestGenerateDaPublisherCommand_Success_IBMMQ(t *testing.T) {
	tmpDir := newInitializedTempProject(t)

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	cmd := NewGenerateDaPublisherCommand()
	cmd.SetArgs([]string{
		"InvoiceSender",
		"--broker", "ibmmq",
		"--pattern", "sender",
		"--target", "DEV.QUEUE.INVOICES",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	publisherPath := filepath.Join(tmpDir, "internal", "drivenadapters", "messaging", "ibmmq", "invoicesender_publisher.go")
	if _, err := os.Stat(publisherPath); os.IsNotExist(err) {
		t.Error("publisher file was not created")
	}

	foundationFile := filepath.Join(tmpDir, "internal", "infrastructure", "messaging", "ibmmq", "module.go")
	if _, err := os.Stat(foundationFile); os.IsNotExist(err) {
		t.Error("ibmmq foundation module file was not created")
	}
}

func TestGenerateDaPublisherCommand_Example(t *testing.T) {
	cmd := NewGenerateDaPublisherCommand()

	if cmd.Example == "" {
		t.Error("expected Example to be set")
	}

	if !contains(cmd.Example, "scaffold-go generate da publisher") {
		t.Error("example should contain 'scaffold-go generate da publisher'")
	}

	if !contains(cmd.Example, "--broker") {
		t.Error("example should contain '--broker' flag")
	}

	if !contains(cmd.Example, "--pattern") {
		t.Error("example should contain '--pattern' flag")
	}
}
