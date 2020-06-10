package core

import (
	"errors"
)

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer provided")

// ErrNilHasher signals that a nil hasher has been provided
var ErrNilHasher = errors.New("nil hasher provided")

// ErrNilNodesCoordinator signals a nil nodes coordinator has been provided
var ErrNilNodesCoordinator = errors.New("nil nodes coordinator")

// ErrInvalidValue signals that a nil value has been provided
var ErrInvalidValue = errors.New("invalid value provided")

// ErrNilInputData signals that a nil data has been provided
var ErrNilInputData = errors.New("nil input data")

//ErrNilUrl signals that the provided url is empty
var ErrNilUrl = errors.New("url is empty")

// ErrPemFileIsInvalid signals that a pem file is invalid
var ErrPemFileIsInvalid = errors.New("pem file is invalid")

// ErrNilPemBLock signals that the pem block is nil
var ErrNilPemBLock = errors.New("nil pem block")

// ErrNilFile signals that a nil file has been provided
var ErrNilFile = errors.New("nil file provided")

// ErrEmptyFile signals that a empty file has been provided
var ErrEmptyFile = errors.New("empty file provided")

// ErrInvalidIndex signals that an invalid private key index has been provided
var ErrInvalidIndex = errors.New("invalid private key index")

// ErrNotPositiveValue signals that a 0 or negative value has been provided
var ErrNotPositiveValue = errors.New("the provided value is not positive")

// ErrNilAppStatusHandler signals that a nil status handler has been provided
var ErrNilAppStatusHandler = errors.New("appStatusHandler is nil")

// ErrNilStatusTagProvider signals that a nil status tag provider has been given as parameter
var ErrNilStatusTagProvider = errors.New("nil status tag provider")

// ErrInvalidPollingInterval signals that an invalid polling interval has been provided
var ErrInvalidPollingInterval = errors.New("invalid polling interval ")

// ErrInvalidIdentifierForEpochStartBlockRequest signals that an invalid identifier for epoch start block request
// has been provided
var ErrInvalidIdentifierForEpochStartBlockRequest = errors.New("invalid identifier for epoch start block request")

// ErrNilEpochStartNotifier signals that nil epoch start notifier has been provided
var ErrNilEpochStartNotifier = errors.New("nil epoch start notifier")

// ErrVersionNumComponents signals that a wrong number of components was provided
var ErrVersionNumComponents = errors.New("invalid version while checking number of components")

// ErrMajorVersionMismatch signals that the major version mismatch
var ErrMajorVersionMismatch = errors.New("major version mismatch")

// ErrMinorVersionMismatch signals that the minor version mismatch
var ErrMinorVersionMismatch = errors.New("minor version mismatch")

// ErrReleaseVersionMismatch signals that the release version mismatch
var ErrReleaseVersionMismatch = errors.New("release version mismatch")
