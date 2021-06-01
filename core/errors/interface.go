package errors

// WrappableError is an interface that extends error and represents a multi-layer error
type WrappableError interface {
	error

	WrapWithMessage(errMessage string) WrappableError
	WrapWithStackTrace() WrappableError
	WrapWithError(err error) WrappableError
	GetBaseError() error
	GetLastError() error

	Unwrap() error
	Is(target error) bool
}
