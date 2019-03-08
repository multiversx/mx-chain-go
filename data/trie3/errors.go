package trie3

import (
	"github.com/pkg/errors"
)

// ErrNotAtLeaf is raised when the iterator it's not positioned at a leaf
var ErrNotAtLeaf = errors.New("iterator not positioned at leaf")

// ErrIterationEnd is raised when the trie iteration has reached an end
var ErrIterationEnd = errors.New("end of iteration")

// ErrProve is raised when a proof can't be provided
var ErrProve = errors.New("can't provide proof")

// ErrInvalidNode is raised when we reach an invalid node
var ErrInvalidNode = errors.New("invalid node")

// ErrNilHasher is raised when the NewTrie() function is called, but a hasher isn't provided
var ErrNilHasher = errors.New("no hasher provided")

// ErrNilMarshalizer is raised when the NewTrie() function is called, but a marshalizer isn't provided
var ErrNilMarshalizer = errors.New("no marshalizer provided")

// ErrNilEncoder is raised when the NewTrie() function is called, but an encoder isn't provided
var ErrNilEncoder = errors.New("no encoder provided")

// ErrNilDatabase is raised when a database operation is called, but no database is provided
var ErrNilDatabase = errors.New("no database provided")

// ErrInvalidEncoding is raised when it is tried to decode an invalid encoding
var ErrInvalidEncoding = errors.New("cannot decode this invalid encoding")

var ErrValueTooShort = errors.New("cannot remove bytes from value because value is too short")
