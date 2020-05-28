package headerCheck

import "errors"

// ErrNotEnoughSignatures signals that a block is not signed by at least the minimum number of validators from
// the consensus group
var ErrNotEnoughSignatures = errors.New("not enough signatures in block")

// ErrWrongSizeBitmap signals that the provided bitmap's length is bigger than the one that was required
var ErrWrongSizeBitmap = errors.New("wrong size bitmap has been provided")

// ErrInvalidReferenceChainID signals that the provided reference chain ID is not valid
var ErrInvalidReferenceChainID = errors.New("invalid reference Chain ID provided")

// ErrInvalidChainID signals that an invalid chain ID has been provided
var ErrInvalidChainID = errors.New("invalid chain ID")

// ErrInvalidSoftwareVersion signals that invalid software version was provided
var ErrInvalidSoftwareVersion = errors.New("invalid software version")
