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

//ErrEmptyUserName signals that user name for elastic search is empty
var ErrEmptyUserName = errors.New("user name is empty")

//ErrEmptyPassword signals that password for elastic search is empty
var ErrEmptyPassword = errors.New("password is empty")

// ErrNilPubkeyConverter signals that an operation has been attempted to or with a nil public key converter implementation
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")
