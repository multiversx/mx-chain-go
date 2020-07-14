package databasereader

import "errors"

// ErrEmptyDbFilePath signals that an empty database file path has been provided
var ErrEmptyDbFilePath = errors.New("empty db file path")

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNilPersisterFactory signals that a nil persister factory has been provided
var ErrNilPersisterFactory = errors.New("nil persister factory")
