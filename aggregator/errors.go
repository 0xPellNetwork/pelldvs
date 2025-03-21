package aggregator

import "fmt"

// Error codes
const (
	// ErrorCodeNone represents no error
	ErrorCodeNone = 0
	// ErrorCodeStakeThresholdsNotMet represents stake thresholds not met error
	ErrorCodeStakeThresholdsNotMet = 1001
	// ErrorCodeNoSignatures represents no signatures to aggregate error
	ErrorCodeNoSignatures = 1002
	// ErrorCodeOperatorInfoNotFound represents operator info not found error
	ErrorCodeOperatorInfoNotFound = 1003
	// ErrorCodeInvalidIndices represents invalid indices error
	ErrorCodeInvalidIndices = 1004
)

// AggregatorError is an error type with an error code
type AggregatorError struct {
	Code    int
	Message string
}

// Error returns the error message
func (e *AggregatorError) Error() string {
	return fmt.Sprintf("[Code: %d] %s", e.Code, e.Message)
}

// NewStakeThresholdsNotMetError creates a new error for stake thresholds not met
func NewStakeThresholdsNotMetError(digest interface{}) *AggregatorError {
	return &AggregatorError{
		Code:    ErrorCodeStakeThresholdsNotMet,
		Message: fmt.Sprintf("stake thresholds not met for digest: %v", digest),
	}
}

// NewNoSignaturesError creates a new error for no signatures to aggregate
func NewNoSignaturesError() *AggregatorError {
	return &AggregatorError{
		Code:    ErrorCodeNoSignatures,
		Message: "no signatures to aggregate",
	}
}

// NewOperatorInfoNotFoundError creates a new error for operator info not found
func NewOperatorInfoNotFoundError(id interface{}) *AggregatorError {
	return &AggregatorError{
		Code:    ErrorCodeOperatorInfoNotFound,
		Message: fmt.Sprintf("failed to get operator info by ID: %v", id),
	}
}

// NewInvalidIndicesError creates a new error for invalid indices
func NewInvalidIndicesError(err error) *AggregatorError {
	return &AggregatorError{
		Code:    ErrorCodeInvalidIndices,
		Message: fmt.Sprintf("failed to get check signatures indices: %v", err),
	}
}
