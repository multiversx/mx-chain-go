package nodesCoordinator

import (
	"errors"
)

// ErrNilInputNodesMap signals that a nil nodes map was provided
var ErrNilInputNodesMap = errors.New("nil input nodes map")

// ErrNilInputNodesList signals that a nil nodes list was provided
var ErrNilInputNodesList = errors.New("nil input nodes list")

// ErrSmallEligibleListSize signals that the eligible validators list's size is less than the consensus size
var ErrSmallEligibleListSize = errors.New("small eligible list size")

// ErrInvalidConsensusGroupSize signals that the consensus size is invalid (e.g. value is negative)
var ErrInvalidConsensusGroupSize = errors.New("invalid consensus group size")

// ErrEligibleSelectionMismatch signals a mismatch between the eligible list and the group selection bitmap
var ErrEligibleSelectionMismatch = errors.New("invalid eligible validator selection")

// ErrLeaderNotSelectedInBitmap signals an invalid validators selection from a consensus group as leader is not marked
var ErrLeaderNotSelectedInBitmap = errors.New("bitmap invalid as leader is not selected")

// ErrEligibleTooManySelections signals an invalid selection for consensus group
var ErrEligibleTooManySelections = errors.New("too many selections for consensus group")

// ErrEligibleTooFewSelections signals an invalid selection for consensus group
var ErrEligibleTooFewSelections = errors.New("too few selections for consensus group")

// ErrNilRandomness signals that a nil randomness source has been provided
var ErrNilRandomness = errors.New("nil randomness source")

// ErrNilHasher signals that a nil hasher has been provided
var ErrNilHasher = errors.New("nil hasher")
