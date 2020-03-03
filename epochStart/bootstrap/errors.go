package bootstrap

import "errors"

// ErrNilMessenger signals that a nil messenger has been provider
var ErrNilMessenger = errors.New("nil messenger")

// ErrNilMarshalizer signals that a nil marshalizer has been provider
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNilHasher signals that a nil hasher has been provider
var ErrNilHasher = errors.New("nil hasher")

// ErrNilNodesConfigProvider signals that a nil nodes config provider has been given
var ErrNilNodesConfigProvider = errors.New("nil nodes config provider")

// ErrNilMetaBlockInterceptor signals that a metablock interceptor has been provided
var ErrNilMetaBlockInterceptor = errors.New("nil metablock interceptor")

// ErrNilShardHeaderInterceptor signals that a nil shard header interceptor has been provided
var ErrNilShardHeaderInterceptor = errors.New("nil shard header interceptor")

// ErrNilMetaBlockResolver signals that a nil metablock resolver has been provided
var ErrNilMetaBlockResolver = errors.New("nil metablock resolver")

// ErrNumTriesExceeded signals that there were too many tries for fetching a metablock
var ErrNumTriesExceeded = errors.New("num of tries exceeded. try re-request")
