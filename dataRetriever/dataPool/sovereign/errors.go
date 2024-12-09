package sovereign

import "errors"

var errHashOfHashesNotFound = errors.New("hash of hashes in bridge operations pool not found")

var errHashOfBridgeOpNotFound = errors.New("hash of bridge operation not found in pool")
