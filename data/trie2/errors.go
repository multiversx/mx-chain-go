package trie2

import (
	"errors"
)

// ErrNotAtLeaf is raised when the iterator is not positioned at a leaf
var ErrNotAtLeaf = errors.New("iterator not positioned at leaf")

// ErrIterationEnd is raised when the trie iteration has reached an end
var ErrIterationEnd = errors.New("end of iteration")

// ErrInvalidNode is raised when we reach an invalid node
var ErrInvalidNode = errors.New("invalid node")

// ErrNilHasher is raised when the NewTrie() function is called, but a hasher isn't provided
var ErrNilHasher = errors.New("no hasher provided")

// ErrNilMarshalizer is raised when the NewTrie() function is called, but a marshalizer isn't provided
var ErrNilMarshalizer = errors.New("no marshalizer provided")

// ErrNilDatabase is raised when a database operation is called, but no database is provided
var ErrNilDatabase = errors.New("no database provided")

// ErrInvalidEncoding is raised when the encoded information cannot be decoded
var ErrInvalidEncoding = errors.New("cannot decode this invalid encoding")

// ErrValueTooShort is raised when we try to remove something from a value, and the value is too short
var ErrValueTooShort = errors.New("cannot remove bytes from value because value is too short")

// ErrChildPosOutOfRange is raised when the position of a child in a branch node is less than 0 or greater than 16
var ErrChildPosOutOfRange = errors.New("the position of the child is out of range")

// ErrNodeNotFound is raised when we try to get a node that is not present in the trie
var ErrNodeNotFound = errors.New("the node is not present in the trie")

// ErrEmptyNode is raised when we reach an empty node (a node with no children or no value)
var ErrEmptyNode = errors.New("the node is empty")

// ErrNilNode is raised when we reach a nil node
var ErrNilNode = errors.New("the node is nil")
