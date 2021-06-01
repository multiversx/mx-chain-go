package errors

import (
	"errors"
	"fmt"
	"runtime"
)

const skipStackLevels = 2
const nilPlaceholder = "<nil>"

type errorWithLocation struct {
	err      error
	location string
}

type wrappableError struct {
	errsWithLocation []errorWithLocation
}

// WrapError constructs a WrappableError from an error
func WrapError(err error) WrappableError {
	errAsWrappable, ok := err.(WrappableError)
	if ok {
		return errAsWrappable
	}
	return &wrappableError{
		errsWithLocation: []errorWithLocation{createErrorWithLocation(err, skipStackLevels)},
	}
}

// WrapWithMessage wraps the target error with a new one, created using the input message
func (werr *wrappableError) WrapWithMessage(errMessage string) WrappableError {
	return werr.wrapWithErrorWithSkipLevels(errors.New(errMessage), skipStackLevels+1)
}

// WrapWithStackTrace wraps the target error with a new one, without any message only a stack frame trace
func (werr *wrappableError) WrapWithStackTrace() WrappableError {
	return werr.wrapWithErrorWithSkipLevels(errors.New(""), skipStackLevels+1)
}

// WrapWithError wraps the target error with the provided one
func (werr *wrappableError) WrapWithError(err error) WrappableError {
	return werr.wrapWithErrorWithSkipLevels(err, skipStackLevels+1)
}

// GetBaseError gets the core error
func (werr *wrappableError) GetBaseError() error {
	errorSlice := werr.errsWithLocation
	return errorSlice[0].err
}

// GetLastError gets the last wrapped error
func (werr *wrappableError) GetLastError() error {
	errorSlice := werr.errsWithLocation
	return errorSlice[len(errorSlice)-1].err
}

func (werr *wrappableError) wrapWithErrorWithSkipLevels(err error, skipStackLevels int) *wrappableError {
	newErrs := make([]errorWithLocation, len(werr.errsWithLocation))
	copy(newErrs, werr.errsWithLocation)
	if err == nil {
		return &wrappableError{
			errsWithLocation: newErrs,
		}
	}
	return &wrappableError{
		errsWithLocation: append(newErrs, createErrorWithLocation(err, skipStackLevels)),
	}
}

func createErrorWithLocation(err error, skipStackLevels int) errorWithLocation {
	_, file, line, _ := runtime.Caller(skipStackLevels)
	locationLine := fmt.Sprintf("%s:%d", file, line)
	errWithLocation := errorWithLocation{err: err, location: locationLine}
	return errWithLocation
}

// Error is the standard error function implementation for wrappable errors
func (werr *wrappableError) Error() string {
	strErr := ""
	errorSlice := werr.errsWithLocation
	for idxErr := range errorSlice {
		errWithLocation := errorSlice[len(errorSlice)-1-idxErr]
		errMsg := nilPlaceholder
		if errWithLocation.err != nil {
			errMsg = errWithLocation.err.Error()
		}
		suffix := ""
		if errMsg != "" {
			suffix = " [" + errMsg + "]"
		}
		strErr += "\t" + errWithLocation.location + suffix
		if idxErr < len(errorSlice)-1 {
			strErr += "\n"
		}
	}
	return strErr
}

// Unwrap represents standard error function implementation for wrappable errors
func (werr *wrappableError) Unwrap() error {
	wrappingErr := werr.errsWithLocation
	if len(wrappingErr) == 1 {
		return wrappingErr[0].err
	}

	return &wrappableError{
		errsWithLocation: werr.errsWithLocation[:len(werr.errsWithLocation)-1],
	}
}

// Is represents standard error function implementation for wrappable errors
func (werr *wrappableError) Is(target error) bool {
	for _, err := range werr.errsWithLocation {
		if errors.Is(err.err, target) {
			return true
		}
	}
	return false
}
