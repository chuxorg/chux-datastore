package logging

// Logger is the interface that all loggers must implement.
type ILogger interface {
	// Debug logs a message at level Debug on the standard logger.
	Debug(msg string, args ...interface{})
	// Info logs a message at level Info on the standard logger.
	Info(msg string, args ...interface{})
	// Warn logs a message at level Warn on the standard logger.
	Warn(msg string, args ...interface{})
	// Error logs a message at level Error on the standard logger.
	Error(msg string, args ...interface{})
}
