package configs

import "errors"

// ErrEmptyProcessConfigsByEpoch signals that an empty process configs by epoch has been provided
var ErrEmptyProcessConfigsByEpoch = errors.New("empty process configs by epoch")

// ErrEmptyProcessConfigsByRound signals that an empty process configs by round has been provided
var ErrEmptyProcessConfigsByRound = errors.New("empty process configs by round")

// ErrEmptyConsensusConfigsByRound signals that an empty consensus configs by round has been provided
var ErrEmptyConsensusConfigsByRound = errors.New("empty consensus configs by round")

// ErrDuplicatedEpochConfig signals that a duplicated config section has been provided
var ErrDuplicatedEpochConfig = errors.New("duplicated epoch config")

// ErrDuplicatedRoundConfig signals that a duplicated config section has been provided
var ErrDuplicatedRoundConfig = errors.New("duplicated round config")

// ErrMissingEpochZeroConfig signals that epoch zero configuration is missing
var ErrMissingEpochZeroConfig = errors.New("missing configuration for epoch 0")

// ErrMissingRoundZeroConfig signals that base round configuration is missing
var ErrMissingRoundZeroConfig = errors.New("missing base configuration for round")

// ErrNegativeSubroundTiming signals that a negative subround timing value was provided
var ErrNegativeSubroundTiming = errors.New("negative subround timing value")

// ErrInvalidSubroundTimingRange signals that a subround start time is not less than its end time
var ErrInvalidSubroundTimingRange = errors.New("subround start time must be less than end time")

// ErrOverlappingSubroundTiming signals that subround timing ranges overlap
var ErrOverlappingSubroundTiming = errors.New("subround timing ranges overlap")

// ErrSubroundTimingExceedsRound signals that a subround timing exceeds round
var ErrSubroundTimingExceedsRound = errors.New("subround timing exceeds round")

// ErrInvalidProcessingThreshold signals that an invalid processing threshold percent has been provided
var ErrInvalidProcessingThreshold = errors.New("invalid processing threshold percent")

// ErrInvalidSubroundsTimingCount signals that the number of subround timing entries does not equal the expected count
var ErrInvalidSubroundsTimingCount = errors.New("invalid subrounds timing count")
