package dblookupext

import (
	"encoding/hex"
	"errors"
	"fmt"
)

// ErrNotFoundInStorage signals that an item was not found in storage
var ErrNotFoundInStorage = errors.New("not found in storage")

var errCannotCastToBlockBody = errors.New("cannot cast to block body")

var errNilESDTSuppliesHandler = errors.New("nil esdt supplies handler")

func newErrCannotSaveEpochByHash(what string, hash []byte, originalErr error) error {
	return fmt.Errorf("cannot save epoch num for [%s] hash [%s]: %w", what, hex.EncodeToString(hash), originalErr)
}

func newErrCannotSaveMiniblockMetadata(hash []byte, originalErr error) error {
	return fmt.Errorf("cannot save miniblock metadata, hash [%s]: %w", hex.EncodeToString(hash), originalErr)
}

// ErrInvalidHashSize signals that an invalid hash size has been provided
var ErrInvalidHashSize = errors.New("invalid hash size")

// ErrInvalidHash signals that an invalid hash has been provided
var ErrInvalidHash = errors.New("invalid hash")
