package indexer

import (
	"errors"
)

// ErrCannotCreateIndex signals that we could not creeate an elasticsearch index
var ErrCannotCreateIndex = errors.New("cannot create elasitc index")

// ErrHeaderTypeAssertion signals that we could not convert to a header
var ErrHeaderTypeAssertion = errors.New("elasticsearch - header type assertion failed")

// ErrCannotCreateIndex signals that we could not creeate an elasticsearch index
var ErrBodyTypeAssertion = errors.New("elasticsearch - body type assertion failed")

// ErrCannotCreateIndex signals that we could not creeate an elasticsearch index
var ErrNoHeader = errors.New("elasticsearch - no header")

// ErrCannotCreateIndex signals that we could not creeate an elasticsearch index
var ErrNoMiniblocks = errors.New("elasticsearch - no miniblocks")
