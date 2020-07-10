package databasereader

import "errors"

// ErrEmptyDbFilePath signals that an empty database file path has been provided
var ErrEmptyDbFilePath = errors.New("empty db file path")
