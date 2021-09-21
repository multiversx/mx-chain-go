package txstatus

import "errors"

// ErrNilUint64ByteSliceConverter signals that a nil uint64 byte slice converter has beed provided
var ErrNilUint64ByteSliceConverter = errors.New("nil uint64 byte slice converter")

// ErrNiStorageService signals that a nil storage service has been provided
var ErrNiStorageService = errors.New("nil storage service")

// ErrNilApiTransactionResult signals that a nil api transaction result has been provided
var ErrNilApiTransactionResult = errors.New("nil ApiTransactionResult")
