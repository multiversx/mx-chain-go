package shard

import "errors"

// ErrInvalidSoftwareVersion signals that invalid software version was provided
var ErrInvalidSoftwareVersion = errors.New("invalid software version")

// ErrInvalidVersionOnEpochValues signals that the version element is not accepted because the epoch values are invalid
var ErrInvalidVersionOnEpochValues = errors.New("invalid version provided on epoch values")

// ErrEmptyVersionsByEpochsList signals that an empty versions by epochs list was provided
var ErrEmptyVersionsByEpochsList = errors.New("empty versions by epochs list")

// ErrInvalidVersionStringTooLong signals that the version element is not accepted because it contains too large strings
var ErrInvalidVersionStringTooLong = errors.New("invalid version provided: string too large")

// ErrSoftwareVersionMismatch signals that the software versions mismatch
var ErrSoftwareVersionMismatch = errors.New("software versions mismatch")

// ErrArwenOutOfProcessUnsupported signals that the version of Arwen does not support out-of-process instantiation
var ErrArwenOutOfProcessUnsupported = errors.New("out-of-process Arwen instance not supported")
