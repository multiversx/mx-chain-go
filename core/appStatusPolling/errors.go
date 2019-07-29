package appStatusPolling

import "github.com/pkg/errors"

// NilAppStatusHandler will be returned when the AppStatusHandler is nil
var NilAppStatusHandler = errors.New("AppStatusHandler is nil")

// PollingDurationNegative will be returned when the polling duration is not a positive number
var PollingDurationNegative = errors.New("Polling duration must be a positive number")

// NilConnectedAddressesHandler will be returned when ConnectedAddressesHandler is nil
var NilConnectedAddressesHandler = errors.New("ConnectedAddressesHandler must not be nil")
