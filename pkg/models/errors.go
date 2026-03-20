// Package models provides shared error types and setup logic for AI model
// packages.
package models

import "fmt"

// FileNotFoundError indicates that a required model file is missing.
type FileNotFoundError struct {
	// Path is the expected filesystem path of the missing model file.
	Path string
	// Err is the underlying error from the filesystem operation.
	Err error
}

func (e *FileNotFoundError) Error() string {
	return "model file not found: " + e.Path
}

func (e *FileNotFoundError) Unwrap() error {
	return e.Err
}

// CorruptFileError indicates that a model file exists but has an incorrect hash.
type CorruptFileError struct {
	// Path is the filesystem path of the corrupt model file.
	Path string
	// Expected is the expected SHA-256 hash of the file.
	Expected string
	// Actual is the actual SHA-256 hash computed from the file contents.
	Actual string
}

func (e *CorruptFileError) Error() string {
	return fmt.Sprintf("model file %s: hash mismatch (expected %s, got %s)", e.Path, e.Expected, e.Actual)
}
