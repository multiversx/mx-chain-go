package appStatusPolling

import "errors"

// ErrNilAppStatusHandler will be returned when the AppStatusHandler is nil
var ErrNilAppStatusHandler = errors.New("appStatusHandler is nil")

// ErrPollingDurationNegative will be returned when the polling duration is not a positive number
var ErrPollingDurationNegative = errors.New("polling duration must be a positive number")

// ErrNilHandlerFunc will be returned when the handler function is nil
var ErrNilHandlerFunc = errors.New("handler function is nil")
