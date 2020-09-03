package indexer

import (
	"errors"
)

// ErrBodyTypeAssertion signals that we could not create an elasticsearch index
var ErrBodyTypeAssertion = errors.New("elasticsearch - body type assertion failed")

// ErrNoHeader signals that we could not create an elasticsearch index
var ErrNoHeader = errors.New("elasticsearch - no header")

// ErrBackOff -
var ErrBackOff = errors.New("back off")

// ErrNoElasticUrlProvided -
var ErrNoElasticUrlProvided = errors.New("no elastic url provided")

// ErrCouldNotCreatePolicy -
var ErrCouldNotCreatePolicy = errors.New("could not create policy")

// ErrNilPubkeyConverter signals that an operation has been attempted to or with a nil public key converter implementation
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")

// ErrInvalidWorkItemData signas when receiving unexpected data in a work item
var ErrInvalidWorkItemData = errors.New("invalid work item data")
