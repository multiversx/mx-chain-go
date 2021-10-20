package block

import "errors"

// ErrInvalidSoftwareVersion signals that an invalid software version was provided
var ErrInvalidSoftwareVersion = errors.New("invalid software version")

// ErrInvalidVersionOnEpochValues signals that the version element is not accepted because the epoch values are invalid
var ErrInvalidVersionOnEpochValues = errors.New("invalid version provided on epoch values")

// ErrEmptyVersionsByEpochsList signals that an empty versions by epochs list was provided
var ErrEmptyVersionsByEpochsList = errors.New("empty versions by epochs list")

// ErrInvalidVersionStringTooLong signals that the version element is not accepted because it contains too large strings
var ErrInvalidVersionStringTooLong = errors.New("invalid version provided: string too large")

// ErrSoftwareVersionMismatch signals a software version mismatch
var ErrSoftwareVersionMismatch = errors.New("software versions mismatch")

// ErrNilCacher signals that a nil cacher has been provided
var ErrNilCacher = errors.New("nil cacher")

// ErrNilHeaderVersionHandler signals that a nil header version handler was provided
var ErrNilHeaderVersionHandler = errors.New("nil error version handler")
