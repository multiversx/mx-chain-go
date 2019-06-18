package indexer

import (
	"errors"
)

// ErrCannotCreateIndex signals that we could not create an elasticsearch index
var ErrCannotCreateIndex = errors.New("cannot create elasitc index")

// ErrBodyTypeAssertion signals that we could not create an elasticsearch index
var ErrBodyTypeAssertion = errors.New("elasticsearch - body type assertion failed")

// ErrNoHeader signals that we could not create an elasticsearch index
var ErrNoHeader = errors.New("elasticsearch - no header")

// ErrNoMiniblocks signals that we could not create an elasticsearch index
var ErrNoMiniblocks = errors.New("elasticsearch - no miniblocks")
