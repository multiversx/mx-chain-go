package databasereader

import "errors"

// ErrEmptyDbFilePath signals that an empty database file path has been provided
var ErrEmptyDbFilePath = errors.New("empty db file path")

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNilPersisterFactory signals that a nil persister factory has been provided
var ErrNilPersisterFactory = errors.New("nil persister factory")

// ErrNilDirectoryReader signals that a nil directory reader has been provided
var ErrNilDirectoryReader = errors.New("nil directory reader")

// ErrNoHeader signals that no header has been found during database read
var ErrNoHeader = errors.New("no header")

// ErrNoDatabaseFound signals that no database has been found in the provided path
var ErrNoDatabaseFound = errors.New("no database found")
