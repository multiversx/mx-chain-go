package shard

import "errors"

// ErrInvalidVersionOnEpochValues signals that the version element is not accepted because the epoch values are invalid
var ErrInvalidVersionOnEpochValues = errors.New("invalid version provided on epoch values")

// ErrEmptyVersionsByEpochsList signals that an empty versions by epochs list was provided
var ErrEmptyVersionsByEpochsList = errors.New("empty versions by epochs list")

// ErrInvalidVersionStringTooLong signals that the version element is not accepted because it contains too large strings
var ErrInvalidVersionStringTooLong = errors.New("invalid version provided: string too large")
