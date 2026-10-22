package filesystem

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"

	"github.com/bancolombia/scaffold-go/internal/domain"
)

type FileSystem struct{}

func New() *FileSystem {
	return &FileSystem{}
}

// Operation represents a file system operation that can be rolled back
type Operation interface {
	Execute() error
	Rollback() error
	Description() string
}

// WriteFileOperation represents a file write operation
type WriteFileOperation struct {
	fs      *FileSystem
	path    string
	content []byte
	existed bool
	oldData []byte
}

func (op *WriteFileOperation) Execute() error {
	// Check if file exists and backup if it does
	if _, err := os.Stat(op.path); err == nil {
		op.existed = true
		data, err := os.ReadFile(op.path)
		if err != nil {
			return &domain.FileSystemError{
				Operation: "backup",
				Path:      op.path,
				Err:       err,
			}
		}
		op.oldData = data
	}

	// Write the file
	if err := op.fs.WriteFile(op.path, op.content); err != nil {
		return err
	}

	return nil
}

func (op *WriteFileOperation) Rollback() error {
	if op.existed {
		// Restore old content
		return os.WriteFile(op.path, op.oldData, 0644)
	}
	// Remove newly created file
	return os.Remove(op.path)
}

func (op *WriteFileOperation) Description() string {
	return fmt.Sprintf("write file: %s", op.path)
}

// CreateDirectoryOperation represents a directory creation operation
type CreateDirectoryOperation struct {
	fs      *FileSystem
	path    string
	existed bool
}

func (op *CreateDirectoryOperation) Execute() error {
	// Check if directory exists
	if _, err := os.Stat(op.path); err == nil {
		op.existed = true
		return nil
	}

	// Create directory
	if err := op.fs.CreateDirectory(op.path); err != nil {
		return err
	}

	return nil
}

func (op *CreateDirectoryOperation) Rollback() error {
	if !op.existed {
		// Remove newly created directory
		return os.RemoveAll(op.path)
	}
	return nil
}

func (op *CreateDirectoryOperation) Description() string {
	return fmt.Sprintf("create directory: %s", op.path)
}

// Transaction manages a set of file system operations with rollback capability
type Transaction struct {
	fs         *FileSystem
	operations []Operation
	executed   []Operation
}

// NewTransaction creates a new transaction
func NewTransaction(fs *FileSystem) *Transaction {
	return &Transaction{
		fs:         fs,
		operations: make([]Operation, 0),
		executed:   make([]Operation, 0),
	}
}

// AddWriteFile adds a file write operation to the transaction
func (t *Transaction) AddWriteFile(path string, content []byte) {
	op := &WriteFileOperation{
		fs:      t.fs,
		path:    path,
		content: content,
	}
	t.operations = append(t.operations, op)
}

// AddCreateDirectory adds a directory creation operation to the transaction
func (t *Transaction) AddCreateDirectory(path string) {
	op := &CreateDirectoryOperation{
		fs:   t.fs,
		path: path,
	}
	t.operations = append(t.operations, op)
}

// Execute executes all operations in the transaction
// If any operation fails, all completed operations are rolled back
func (t *Transaction) Execute() error {
	for _, op := range t.operations {
		if err := op.Execute(); err != nil {
			// Rollback all executed operations
			rollbackErr := t.Rollback()
			if rollbackErr != nil {
				return &domain.FileSystemError{
					Operation: "transaction",
					Path:      "multiple",
					Err:       fmt.Errorf("operation failed: %w, rollback also failed: %v", err, rollbackErr),
				}
			}
			return &domain.FileSystemError{
				Operation: "transaction",
				Path:      "multiple",
				Err:       fmt.Errorf("operation failed and rolled back: %w", err),
			}
		}
		t.executed = append(t.executed, op)
	}
	return nil
}

// Rollback rolls back all executed operations in reverse order
func (t *Transaction) Rollback() error {
	var errors []error

	// Rollback in reverse order
	for i := len(t.executed) - 1; i >= 0; i-- {
		op := t.executed[i]
		if err := op.Rollback(); err != nil {
			errors = append(errors, fmt.Errorf("failed to rollback %s: %w", op.Description(), err))
		}
	}

	if len(errors) > 0 {
		return &domain.FileSystemError{
			Operation: "rollback",
			Path:      "multiple",
			Err:       fmt.Errorf("rollback errors: %v", errors),
		}
	}

	// Clear executed operations
	t.executed = make([]Operation, 0)
	return nil
}

func (fs *FileSystem) CreateDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

func (fs *FileSystem) WriteFile(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := fs.CreateDirectory(dir); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0644)
}

func (fs *FileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (fs *FileSystem) FormatGoCode(content []byte) ([]byte, error) {
	return format.Source(content)
}

type FileOperation struct {
	Path    string
	Content []byte
}

func (fs *FileSystem) WriteFiles(operations []FileOperation) error {
	// Use transaction for atomic operations
	tx := NewTransaction(fs)

	for _, op := range operations {
		tx.AddWriteFile(op.Path, op.Content)
	}

	return tx.Execute()
}
