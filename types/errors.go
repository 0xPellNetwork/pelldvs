package types

import "fmt"

type (
	// ErrInvalidCommitHeight is returned when we encounter a commit with an
	// unexpected height.
	ErrInvalidCommitHeight struct {
		Expected int64 // The height we expected to see
		Actual   int64 // The height we actually saw
	}

	// ErrInvalidCommitSignatures is returned when we encounter a commit where
	// the number of signatures doesn't match the number of validators.
	ErrInvalidCommitSignatures struct {
		Expected int // The number of signatures we expected
		Actual   int // The number of signatures we actually got
	}
)

// NewErrInvalidCommitHeight creates a new error for cases when the commit height
// doesn't match the expected height in the consensus process
func NewErrInvalidCommitHeight(expected, actual int64) ErrInvalidCommitHeight {
	return ErrInvalidCommitHeight{
		Expected: expected,
		Actual:   actual,
	}
}

// Error implements the error interface for ErrInvalidCommitHeight
// providing a descriptive error message about the height mismatch
func (e ErrInvalidCommitHeight) Error() string {
	return fmt.Sprintf("Invalid commit -- wrong height: %v vs %v", e.Expected, e.Actual)
}

// NewErrInvalidCommitSignatures creates a new error for cases when the number
// of signatures in a commit doesn't match the expected validator set size
func NewErrInvalidCommitSignatures(expected, actual int) ErrInvalidCommitSignatures {
	return ErrInvalidCommitSignatures{
		Expected: expected,
		Actual:   actual,
	}
}

// Error implements the error interface for ErrInvalidCommitSignatures
// providing a descriptive error message about the signature count mismatch
func (e ErrInvalidCommitSignatures) Error() string {
	return fmt.Sprintf("Invalid commit -- wrong set size: %v vs %v", e.Expected, e.Actual)
}
