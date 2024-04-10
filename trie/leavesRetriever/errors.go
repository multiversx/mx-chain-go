package leavesRetriever

import "errors"

// ErrNilDB is returned when the given db is nil
var ErrNilDB = errors.New("nil db")

// ErrNilMarshaller is returned when the given marshaller is nil
var ErrNilMarshaller = errors.New("nil marshaller")

// ErrNilHasher is returned when the given hasher is nil
var ErrNilHasher = errors.New("nil hasher")

// ErrIteratorNotFound is returned when the iterator is not found
var ErrIteratorNotFound = errors.New("iterator not found")
