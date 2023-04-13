package db

import "fmt"

// An Error that is returned from the chux-mongo library
type ChuxMongoError struct {
	Code       int
	Message    string
	InnerError error
}

// Creates a new ChuxMongoError
func NewChuxMongoError(message string, code int, innerError error) *ChuxMongoError {

	err := ChuxMongoError{
		Code:       code,
		InnerError: innerError,
		Message:    message,
	}

	return &err
}

func (e ChuxMongoError) Error() string {
	return fmt.Sprintf("ChuxMongoError: Code: %d, Message: %s, InnerError: %v", e.Code, e.Message, e.InnerError)
}
