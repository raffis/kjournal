package log

type Logger interface {
	// Infof logs a formatted action message.
	Infof(format string, a ...interface{})
	// Generatef logs a formatted generate message.
	Generatef(format string, a ...interface{})
	// Waitingf logs a formatted waiting message.
	Waitingf(format string, a ...interface{})
	// Successf logs a formatted success message.
	Successf(format string, a ...interface{})
	// Warningf logs a formatted warning message.
	Warningf(format string, a ...interface{})
	// Failuref logs a formatted failure message.
	Failuref(format string, a ...interface{})
}
