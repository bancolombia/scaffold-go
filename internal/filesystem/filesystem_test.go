package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCreateDirectory tests directory creation
func TestCreateDirectory(t *testing.T) {
	fs := New()
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "create single directory",
			path:    filepath.Join(tmpDir, "testdir"),
			wantErr: false,
		},
		{
			name:    "create nested directories",
			path:    filepath.Join(tmpDir, "nested", "deep", "directory"),
			wantErr: false,
		},
		{
			name:    "create existing directory",
			path:    tmpDir,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fs.CreateDirectory(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify directory was created
				if _, err := os.Stat(tt.path); os.IsNotExist(err) {
					t.Errorf("CreateDirectory() did not create directory at %s", tt.path)
				}
			}
		})
	}
}

// TestWriteFile tests file writing with automatic directory creation
func TestWriteFile(t *testing.T) {
	fs := New()
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		content []byte
		wantErr bool
	}{
		{
			name:    "write file in existing directory",
			path:    filepath.Join(tmpDir, "test.txt"),
			content: []byte("test content"),
			wantErr: false,
		},
		{
			name:    "write file with nested directory creation",
			path:    filepath.Join(tmpDir, "nested", "dir", "file.txt"),
			content: []byte("nested content"),
			wantErr: false,
		},
		{
			name:    "write empty file",
			path:    filepath.Join(tmpDir, "empty.txt"),
			content: []byte(""),
			wantErr: false,
		},
		{
			name:    "overwrite existing file",
			path:    filepath.Join(tmpDir, "overwrite.txt"),
			content: []byte("new content"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For overwrite test, create file first
			if tt.name == "overwrite existing file" {
				if err := os.WriteFile(tt.path, []byte("old content"), 0644); err != nil {
					t.Fatalf("Failed to setup test: %v", err)
				}
			}

			err := fs.WriteFile(tt.path, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file was created
				data, err := os.ReadFile(tt.path)
				if err != nil {
					t.Errorf("WriteFile() created file but cannot read it: %v", err)
					return
				}

				// Verify content
				if string(data) != string(tt.content) {
					t.Errorf("WriteFile() content = %s, want %s", string(data), string(tt.content))
				}
			}
		})
	}
}

// TestExists tests path existence checking
func TestExists(t *testing.T) {
	fs := New()
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to setup test: %v", err)
	}

	// Create a test directory
	testDir := filepath.Join(tmpDir, "existsdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to setup test: %v", err)
	}

	tests := []struct {
		name   string
		path   string
		exists bool
	}{
		{
			name:   "existing file",
			path:   testFile,
			exists: true,
		},
		{
			name:   "existing directory",
			path:   testDir,
			exists: true,
		},
		{
			name:   "non-existing file",
			path:   filepath.Join(tmpDir, "nonexistent.txt"),
			exists: false,
		},
		{
			name:   "non-existing directory",
			path:   filepath.Join(tmpDir, "nonexistent"),
			exists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fs.Exists(tt.path)
			if result != tt.exists {
				t.Errorf("Exists() = %v, want %v", result, tt.exists)
			}
		})
	}
}

// TestFormatGoCode tests gofmt integration
func TestFormatGoCode(t *testing.T) {
	fs := New()

	tests := []struct {
		name    string
		input   []byte
		wantErr bool
		check   func([]byte) bool
	}{
		{
			name: "format valid Go code",
			input: []byte(`package main
func main(){
fmt.Println("hello")
}`),
			wantErr: false,
			check: func(output []byte) bool {
				// Should contain properly formatted code
				return len(output) > 0
			},
		},
		{
			name: "format already formatted code",
			input: []byte(`package main

func main() {
	fmt.Println("hello")
}
`),
			wantErr: false,
			check: func(output []byte) bool {
				return len(output) > 0
			},
		},
		{
			name: "invalid Go code",
			input: []byte(`package main
func main() {
	this is not valid go code
}`),
			wantErr: true,
			check:   nil,
		},
		{
			name:    "empty input",
			input:   []byte(""),
			wantErr: false,
			check: func(output []byte) bool {
				return len(output) == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := fs.FormatGoCode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatGoCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				if !tt.check(output) {
					t.Errorf("FormatGoCode() output validation failed")
				}
			}
		})
	}
}

// TestWriteFiles tests atomic file operations
func TestWriteFiles(t *testing.T) {
	fs := New()
	tmpDir := t.TempDir()

	t.Run("write multiple files successfully", func(t *testing.T) {
		ops := []FileOperation{
			{
				Path:    filepath.Join(tmpDir, "file1.txt"),
				Content: []byte("content1"),
			},
			{
				Path:    filepath.Join(tmpDir, "file2.txt"),
				Content: []byte("content2"),
			},
			{
				Path:    filepath.Join(tmpDir, "nested", "file3.txt"),
				Content: []byte("content3"),
			},
		}

		err := fs.WriteFiles(ops)
		if err != nil {
			t.Errorf("WriteFiles() error = %v, want nil", err)
			return
		}

		// Verify all files were created
		for _, op := range ops {
			data, err := os.ReadFile(op.Path)
			if err != nil {
				t.Errorf("WriteFiles() created file but cannot read %s: %v", op.Path, err)
				continue
			}
			if string(data) != string(op.Content) {
				t.Errorf("WriteFiles() content at %s = %s, want %s", op.Path, string(data), string(op.Content))
			}
		}
	})

	t.Run("rollback on failure", func(t *testing.T) {
		rollbackDir := filepath.Join(tmpDir, "rollback")
		if err := os.MkdirAll(rollbackDir, 0755); err != nil {
			t.Fatalf("Failed to setup test: %v", err)
		}

		// Create operations where the last one will fail
		ops := []FileOperation{
			{
				Path:    filepath.Join(rollbackDir, "success1.txt"),
				Content: []byte("content1"),
			},
			{
				Path:    filepath.Join(rollbackDir, "success2.txt"),
				Content: []byte("content2"),
			},
			{
				// This will fail because we'll make the directory read-only
				Path:    filepath.Join(rollbackDir, "readonly", "fail.txt"),
				Content: []byte("content3"),
			},
		}

		// Make the readonly directory exist but be read-only
		readonlyDir := filepath.Join(rollbackDir, "readonly")
		if err := os.MkdirAll(readonlyDir, 0755); err != nil {
			t.Fatalf("Failed to setup test: %v", err)
		}
		if err := os.Chmod(readonlyDir, 0444); err != nil {
			t.Fatalf("Failed to setup test: %v", err)
		}
		defer os.Chmod(readonlyDir, 0755) // Cleanup

		err := fs.WriteFiles(ops)
		if err == nil {
			t.Error("WriteFiles() expected error, got nil")
			return
		}

		// Verify rollback: first two files should be deleted
		if fs.Exists(ops[0].Path) {
			t.Errorf("WriteFiles() rollback failed: %s still exists", ops[0].Path)
		}
		if fs.Exists(ops[1].Path) {
			t.Errorf("WriteFiles() rollback failed: %s still exists", ops[1].Path)
		}
	})

	t.Run("empty operations list", func(t *testing.T) {
		err := fs.WriteFiles([]FileOperation{})
		if err != nil {
			t.Errorf("WriteFiles() with empty list error = %v, want nil", err)
		}
	})
}

// TestRollback tests the rollback mechanism
func TestRollback(t *testing.T) {
	fs := New()
	tmpDir := t.TempDir()

	// Create test files
	files := []string{
		filepath.Join(tmpDir, "file1.txt"),
		filepath.Join(tmpDir, "file2.txt"),
		filepath.Join(tmpDir, "file3.txt"),
	}

	for _, file := range files {
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to setup test: %v", err)
		}
	}

	// Verify files exist
	for _, file := range files {
		if !fs.Exists(file) {
			t.Fatalf("Setup failed: %s does not exist", file)
		}
	}

	// Create a transaction and execute operations
	tx := NewTransaction(fs)
	for _, file := range files {
		tx.AddWriteFile(file, []byte("updated"))
	}

	// Execute to populate executed operations
	if err := tx.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Backup original content
	originalContent := make(map[string][]byte)
	for _, file := range files {
		data, _ := os.ReadFile(file)
		originalContent[file] = data
	}

	// Call rollback
	if err := tx.Rollback(); err != nil {
		t.Errorf("Rollback() error = %v, want nil", err)
	}

	// Verify files were restored to original content
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Errorf("Failed to read file after rollback: %v", err)
			continue
		}
		if string(data) != "test" {
			t.Errorf("Rollback() did not restore original content for %s, got %s, want 'test'", file, string(data))
		}
	}
}

// TestTransaction tests the transaction mechanism
func TestTransaction(t *testing.T) {
	fs := New()
	tmpDir := t.TempDir()

	t.Run("successful transaction with multiple operations", func(t *testing.T) {
		tx := NewTransaction(fs)

		// Add multiple operations
		tx.AddCreateDirectory(filepath.Join(tmpDir, "txdir1"))
		tx.AddWriteFile(filepath.Join(tmpDir, "txdir1", "file1.txt"), []byte("content1"))
		tx.AddWriteFile(filepath.Join(tmpDir, "txdir1", "file2.txt"), []byte("content2"))

		err := tx.Execute()
		if err != nil {
			t.Errorf("Transaction.Execute() error = %v, want nil", err)
			return
		}

		// Verify all operations completed
		if !fs.Exists(filepath.Join(tmpDir, "txdir1")) {
			t.Error("Transaction did not create directory")
		}
		if !fs.Exists(filepath.Join(tmpDir, "txdir1", "file1.txt")) {
			t.Error("Transaction did not create file1.txt")
		}
		if !fs.Exists(filepath.Join(tmpDir, "txdir1", "file2.txt")) {
			t.Error("Transaction did not create file2.txt")
		}
	})

	t.Run("transaction rollback on failure", func(t *testing.T) {
		rollbackDir := filepath.Join(tmpDir, "rollback_test")
		if err := os.MkdirAll(rollbackDir, 0755); err != nil {
			t.Fatalf("Failed to setup test: %v", err)
		}

		tx := NewTransaction(fs)

		// Add operations that will succeed
		tx.AddWriteFile(filepath.Join(rollbackDir, "success1.txt"), []byte("content1"))
		tx.AddWriteFile(filepath.Join(rollbackDir, "success2.txt"), []byte("content2"))

		// Add operation that will fail
		readonlyDir := filepath.Join(rollbackDir, "readonly")
		if err := os.MkdirAll(readonlyDir, 0755); err != nil {
			t.Fatalf("Failed to setup test: %v", err)
		}
		if err := os.Chmod(readonlyDir, 0444); err != nil {
			t.Fatalf("Failed to setup test: %v", err)
		}
		defer os.Chmod(readonlyDir, 0755) // Cleanup

		tx.AddWriteFile(filepath.Join(readonlyDir, "fail.txt"), []byte("content3"))

		err := tx.Execute()
		if err == nil {
			t.Error("Transaction.Execute() expected error, got nil")
			return
		}

		// Verify rollback: first two files should not exist
		if fs.Exists(filepath.Join(rollbackDir, "success1.txt")) {
			t.Error("Transaction rollback failed: success1.txt still exists")
		}
		if fs.Exists(filepath.Join(rollbackDir, "success2.txt")) {
			t.Error("Transaction rollback failed: success2.txt still exists")
		}
	})

	t.Run("transaction with existing file overwrite", func(t *testing.T) {
		existingFile := filepath.Join(tmpDir, "existing.txt")
		originalContent := []byte("original content")
		if err := os.WriteFile(existingFile, originalContent, 0644); err != nil {
			t.Fatalf("Failed to setup test: %v", err)
		}

		tx := NewTransaction(fs)
		tx.AddWriteFile(existingFile, []byte("new content"))

		err := tx.Execute()
		if err != nil {
			t.Errorf("Transaction.Execute() error = %v, want nil", err)
			return
		}

		// Verify file was updated
		data, _ := os.ReadFile(existingFile)
		if string(data) != "new content" {
			t.Errorf("Transaction did not update file, got %s, want 'new content'", string(data))
		}

		// Test rollback
		err = tx.Rollback()
		if err != nil {
			t.Errorf("Transaction.Rollback() error = %v, want nil", err)
			return
		}

		// Verify file was restored
		data, _ = os.ReadFile(existingFile)
		if string(data) != string(originalContent) {
			t.Errorf("Transaction rollback did not restore file, got %s, want %s", string(data), string(originalContent))
		}
	})

	t.Run("empty transaction", func(t *testing.T) {
		tx := NewTransaction(fs)
		err := tx.Execute()
		if err != nil {
			t.Errorf("Empty transaction error = %v, want nil", err)
		}
	})

	t.Run("transaction with directory creation", func(t *testing.T) {
		newDir := filepath.Join(tmpDir, "newdir", "nested")
		tx := NewTransaction(fs)
		tx.AddCreateDirectory(newDir)

		err := tx.Execute()
		if err != nil {
			t.Errorf("Transaction.Execute() error = %v, want nil", err)
			return
		}

		if !fs.Exists(newDir) {
			t.Error("Transaction did not create directory")
		}

		// Test rollback
		err = tx.Rollback()
		if err != nil {
			t.Errorf("Transaction.Rollback() error = %v, want nil", err)
			return
		}

		// Verify directory was removed
		if fs.Exists(newDir) {
			t.Error("Transaction rollback did not remove directory")
		}
	})
}

// TestWriteFileOperation tests the WriteFileOperation
func TestWriteFileOperation(t *testing.T) {
	fs := New()
	tmpDir := t.TempDir()

	t.Run("execute new file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "newfile.txt")
		content := []byte("test content")

		op := &WriteFileOperation{
			fs:      fs,
			path:    path,
			content: content,
		}

		err := op.Execute()
		if err != nil {
			t.Errorf("WriteFileOperation.Execute() error = %v, want nil", err)
			return
		}

		// Verify file was created
		data, _ := os.ReadFile(path)
		if string(data) != string(content) {
			t.Errorf("WriteFileOperation created wrong content, got %s, want %s", string(data), string(content))
		}

		// Test rollback
		err = op.Rollback()
		if err != nil {
			t.Errorf("WriteFileOperation.Rollback() error = %v, want nil", err)
			return
		}

		// Verify file was removed
		if fs.Exists(path) {
			t.Error("WriteFileOperation rollback did not remove file")
		}
	})

	t.Run("execute existing file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "existing.txt")
		originalContent := []byte("original")
		newContent := []byte("new")

		// Create existing file
		if err := os.WriteFile(path, originalContent, 0644); err != nil {
			t.Fatalf("Failed to setup test: %v", err)
		}

		op := &WriteFileOperation{
			fs:      fs,
			path:    path,
			content: newContent,
		}

		err := op.Execute()
		if err != nil {
			t.Errorf("WriteFileOperation.Execute() error = %v, want nil", err)
			return
		}

		// Verify file was updated
		data, _ := os.ReadFile(path)
		if string(data) != string(newContent) {
			t.Errorf("WriteFileOperation updated wrong content, got %s, want %s", string(data), string(newContent))
		}

		// Test rollback
		err = op.Rollback()
		if err != nil {
			t.Errorf("WriteFileOperation.Rollback() error = %v, want nil", err)
			return
		}

		// Verify file was restored
		data, _ = os.ReadFile(path)
		if string(data) != string(originalContent) {
			t.Errorf("WriteFileOperation rollback did not restore content, got %s, want %s", string(data), string(originalContent))
		}
	})
}

// TestCreateDirectoryOperation tests the CreateDirectoryOperation
func TestCreateDirectoryOperation(t *testing.T) {
	fs := New()
	tmpDir := t.TempDir()

	t.Run("execute new directory", func(t *testing.T) {
		path := filepath.Join(tmpDir, "newdir")

		op := &CreateDirectoryOperation{
			fs:   fs,
			path: path,
		}

		err := op.Execute()
		if err != nil {
			t.Errorf("CreateDirectoryOperation.Execute() error = %v, want nil", err)
			return
		}

		// Verify directory was created
		if !fs.Exists(path) {
			t.Error("CreateDirectoryOperation did not create directory")
		}

		// Test rollback
		err = op.Rollback()
		if err != nil {
			t.Errorf("CreateDirectoryOperation.Rollback() error = %v, want nil", err)
			return
		}

		// Verify directory was removed
		if fs.Exists(path) {
			t.Error("CreateDirectoryOperation rollback did not remove directory")
		}
	})

	t.Run("execute existing directory", func(t *testing.T) {
		path := filepath.Join(tmpDir, "existingdir")

		// Create existing directory
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("Failed to setup test: %v", err)
		}

		op := &CreateDirectoryOperation{
			fs:   fs,
			path: path,
		}

		err := op.Execute()
		if err != nil {
			t.Errorf("CreateDirectoryOperation.Execute() error = %v, want nil", err)
			return
		}

		// Verify directory still exists
		if !fs.Exists(path) {
			t.Error("CreateDirectoryOperation removed existing directory")
		}

		// Test rollback (should not remove existing directory)
		err = op.Rollback()
		if err != nil {
			t.Errorf("CreateDirectoryOperation.Rollback() error = %v, want nil", err)
			return
		}

		// Verify directory still exists
		if !fs.Exists(path) {
			t.Error("CreateDirectoryOperation rollback removed existing directory")
		}
	})
}
