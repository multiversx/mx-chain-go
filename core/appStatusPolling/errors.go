package appStatusPolling

import "errors"

// ErrNilAppStatusHandler will be returned when the AppStatusHandler is nil
var ErrNilAppStatusHandler = errors.New("appStatusHandler is nil")

// ErrPollingDurationToSmall will be returned when the polling duration is to small
var ErrPollingDurationToSmall = errors.New("polling duration it's to small")

// ErrNilHandlerFunc will be returned when the handler function is nil
var ErrNilHandlerFunc = errors.New("handler function is nil")
