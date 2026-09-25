package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateEpConsumerCommand_Structure(t *testing.T) {
	cmd := NewGenerateEpConsumerCommand()

	if cmd.Use != "consumer [name]" {
		t.Errorf("expected Use to be 'consumer [name]', got '%s'", cmd.Use)
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

func TestGenerateEpConsumerCommand_Flags(t *testing.T) {
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
			name:     "usecases flag is registered",
			flagName: "usecases",
			wantType: "stringSlice",
		},
		{
			name:     "queue flag is registered",
			flagName: "queue",
			wantType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateEpConsumerCommand()
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

func TestGenerateEpConsumerCommand_Args(t *testing.T) {
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
			args:      []string{"UserEventConsumer"},
			wantError: false,
		},
		{
			name:      "two args returns error",
			args:      []string{"UserEventConsumer", "Extra"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateEpConsumerCommand()
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

func TestGenerateEpConsumerCommand_ValidationError(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewGenerateEpConsumerCommand()
	cmd.SetArgs([]string{"UserEventConsumer", "--broker", "rabbitmq", "--pattern", "event"})

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for uninitialized project, got none")
	}
}

func TestGenerateEpConsumerCommand_KafkaPlaceholder(t *testing.T) {
	tmpDir := newInitializedTempProject(t)

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	cmd := NewGenerateEpConsumerCommand()
	cmd.SetArgs([]string{"OrderConsumer", "--broker", "kafka", "--pattern", "event"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected kafka to be rejected as not implemented")
	}
}

func TestGenerateEpConsumerCommand_Success_RabbitMQ(t *testing.T) {
	tmpDir := newInitializedTempProject(t)

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	cmd := NewGenerateEpConsumerCommand()
	cmd.SetArgs([]string{
		"UserEventConsumer",
		"--broker", "rabbitmq",
		"--pattern", "event",
		"--usecases", "createUser:CreateUserUseCase",
		"--queue", "user.created",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	consumerPath := filepath.Join(tmpDir, "internal", "entrypoints", "messaging", "usereventconsumer_consumer.go")
	if _, err := os.Stat(consumerPath); os.IsNotExist(err) {
		t.Error("consumer file was not created")
	}

	moduleFile := filepath.Join(tmpDir, "internal", "entrypoints", "messaging", "usereventconsumer_module.go")
	if _, err := os.Stat(moduleFile); os.IsNotExist(err) {
		t.Error("per-component module file was not created")
	}

	foundationFile := filepath.Join(tmpDir, "internal", "infrastructure", "messaging", "rabbitmq", "module.go")
	if _, err := os.Stat(foundationFile); os.IsNotExist(err) {
		t.Error("foundation module file was not created")
	}
}

func TestGenerateEpConsumerCommand_Success_IBMMQ(t *testing.T) {
	tmpDir := newInitializedTempProject(t)

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	cmd := NewGenerateEpConsumerCommand()
	cmd.SetArgs([]string{
		"InvoiceListener",
		"--broker", "ibmmq",
		"--pattern", "listener",
		"--queue", "DEV.QUEUE.INVOICES",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundationFile := filepath.Join(tmpDir, "internal", "infrastructure", "messaging", "ibmmq", "module.go")
	if _, err := os.Stat(foundationFile); os.IsNotExist(err) {
		t.Error("ibmmq foundation module file was not created")
	}
}

func TestGenerateEpConsumerCommand_Example(t *testing.T) {
	cmd := NewGenerateEpConsumerCommand()

	if cmd.Example == "" {
		t.Error("expected Example to be set")
	}

	if !contains(cmd.Example, "scaffold-go generate ep consumer") {
		t.Error("example should contain 'scaffold-go generate ep consumer'")
	}

	if !contains(cmd.Example, "--broker") {
		t.Error("example should contain '--broker' flag")
	}

	if !contains(cmd.Example, "--pattern") {
		t.Error("example should contain '--pattern' flag")
	}
}
