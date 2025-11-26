package types

import "fmt"

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// ExternalAPIError represents an error from external API/service
type ExternalAPIError struct {
	Service string
	URL     string
	Err     error
}

func (e *ExternalAPIError) Error() string {
	return fmt.Sprintf("external API error [%s] %s: %v", e.Service, e.URL, e.Err)
}

func (e *ExternalAPIError) Unwrap() error {
	return e.Err
}

// DatabaseError represents a database operation error
type DatabaseError struct {
	Operation string
	Err       error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database error during %s: %v", e.Operation, e.Err)
}

func (e *DatabaseError) Unwrap() error {
	return e.Err
}
