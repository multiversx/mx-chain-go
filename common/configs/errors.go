package configs

import "errors"

// ErrEmptyProcessConfigsByEpoch signals that an empty process configs by epoch has been provided
var ErrEmptyProcessConfigsByEpoch = errors.New("empty process configs by epoch")

// ErrEmptyProcessConfigsByRound signals that an empty process configs by round has been provided
var ErrEmptyProcessConfigsByRound = errors.New("empty process configs by round")

// ErrDuplicatedEpochConfig signals that a duplicated config section has been provided
var ErrDuplicatedEpochConfig = errors.New("duplicated epoch config")

// ErrDuplicatedRoundConfig signals that a duplicated config section has been provided
var ErrDuplicatedRoundConfig = errors.New("duplicated round config")

// ErrMissingEpochZeroConfig signals that epoch zero configuration is missing
var ErrMissingEpochZeroConfig = errors.New("missing configuration for epoch 0")

// ErrMissingRoundZeroConfig signals that base round configuration is missing
var ErrMissingRoundZeroConfig = errors.New("missing base configuration for round")
