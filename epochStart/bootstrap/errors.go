package bootstrap

import "errors"

// ErrNilPublicKey signals that a nil public key has been provided
var ErrNilPublicKey = errors.New("nil public key")

// ErrNilMessenger signals that a nil messenger has been provided
var ErrNilMessenger = errors.New("nil messenger")

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNilHasher signals that a nil hasher has been provider
var ErrNilHasher = errors.New("nil hasher")

// ErrNilNodesConfigProvider signals that a nil nodes config provider has been given
var ErrNilNodesConfigProvider = errors.New("nil nodes config provider")

// ErrNilEpochStartMetaBlockInterceptor signals that a epoch start metablock interceptor has been provided
var ErrNilEpochStartMetaBlockInterceptor = errors.New("nil epoch start metablock interceptor")

// ErrNilMetaBlockInterceptor signals that a metablock interceptor has been provided
var ErrNilMetaBlockInterceptor = errors.New("nil metablock interceptor")

// ErrNilShardHeaderInterceptor signals that a nil shard header interceptor has been provided
var ErrNilShardHeaderInterceptor = errors.New("nil shard header interceptor")

// ErrNilMiniBlockInterceptor signals that a nil mini block interceptor has been provided
var ErrNilMiniBlockInterceptor = errors.New("nil mini block interceptor")

// ErrNumTriesExceeded signals that there were too many tries for fetching a metablock
var ErrNumTriesExceeded = errors.New("num of tries exceeded. try re-request")

// ErrNilShardCoordinator signals that a nil shard coordinator has been provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrShardDataNotFound signals that no shard header has been found for the calculated shard
var ErrShardDataNotFound = errors.New("shard data not found")
