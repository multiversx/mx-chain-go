package leavesRetriever

import "errors"

// ErrNilMarshaller is returned when the given marshaller is nil
var ErrNilMarshaller = errors.New("nil marshaller")

// ErrNilHasher is returned when the given hasher is nil
var ErrNilHasher = errors.New("nil hasher")
