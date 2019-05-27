package storage

import (
	"errors"
)

var ErrNilPersister = errors.New("expected not nil persister")

var ErrNilCacher = errors.New("expected not nil cacher")

var ErrNilBloomFilter = errors.New("expected not nil bloom filter")

var ErrNotSupportedCacheType = errors.New("not supported cache type")

var ErrMissingOrInvalidParameter = errors.New("parameter is missing or invalid")

var ErrNotSupportedDBType = errors.New("nit supported db type")

var ErrNotSupportedHashType = errors.New("hash type not supported")

var ErrKeyNotFound = errors.New("key not found")

var ErrInvalidBatch = errors.New("batch is invalid")
