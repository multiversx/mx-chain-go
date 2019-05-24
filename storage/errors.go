package storage

import (
	"errors"
)

var errNilPersister = errors.New("expected not nil persister")

var errNilCacher = errors.New("expected not nil cacher")

var errNilBloomFilter = errors.New("expected not nil bloom filter")

var errNotSupportedCacheType = errors.New("not supported cache type")

var errMissingOrInvalidParameter = errors.New("parameter is missing or invalid")

var errNotSupportedDBType = errors.New("nit supported db type")

var errNotSupportedHashType = errors.New("hash type not supported")
