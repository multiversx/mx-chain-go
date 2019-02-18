package trie2

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
