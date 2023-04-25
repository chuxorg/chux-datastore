package errors

// ChuxDataStoreError is a custom error type
// that wraps an error and adds a message
// to the error.
// This is the error that is returned by
// all functions in chux-datastore that return
// an error.
type ChuxDataStoreError struct {
	// Message is the message that is
	// given by chux-parser when an error
	// occurs.
	// This message is used to provide
	// more context to the error.
	// The Err field contains the actual
	// error that occurred.
	Message string
	Err     error
	code    int
}

// NewChuxParserError returns a new ChuxDataStoreError
func NewChuxDataStoreError(message string, code int, err error) *ChuxDataStoreError {
	return &ChuxDataStoreError{
		Message: message,
		Err:     err,
		code:    code,
	}
}

func (e *ChuxDataStoreError) Error() string {
	return e.Message
}

// Unwrap returns the underlying error without
// the message added by chux-parser.
func (e *ChuxDataStoreError) Unwrap() error {
	return e.Err
}
