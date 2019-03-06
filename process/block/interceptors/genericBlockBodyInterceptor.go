package interceptors

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// GenericBlockBodyInterceptor represents an interceptor used for all types of block bodies
type GenericBlockBodyInterceptor struct {
	*messageChecker
	marshalizer      marshal.Marshalizer
	cache            storage.Cacher
	hasher           hashing.Hasher
	storer           storage.Storer
	shardCoordinator sharding.ShardCoordinator
}

// NewGenericBlockBodyInterceptor hooks a new interceptor for block bodies
// Fetched data blocks will be placed inside the cache
func NewGenericBlockBodyInterceptor(
	marshalizer marshal.Marshalizer,
	cache storage.Cacher,
	storer storage.Storer,
	hasher hashing.Hasher,
	shardCoordinator sharding.ShardCoordinator,
) (*GenericBlockBodyInterceptor, error) {

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}

	if cache == nil {
		return nil, process.ErrNilCacher
	}

	if storer == nil {
		return nil, process.ErrNilBlockBodyStorage
	}

	if hasher == nil {
		return nil, process.ErrNilHasher
	}

	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}

	bbIntercept := &GenericBlockBodyInterceptor{
		messageChecker:   &messageChecker{},
		marshalizer:      marshalizer,
		cache:            cache,
		storer:           storer,
		hasher:           hasher,
		shardCoordinator: shardCoordinator,
	}

	return bbIntercept, nil
}

func (gbbi *GenericBlockBodyInterceptor) processBlockBody(messageData []byte, body process.InterceptedBlockBody) error {
	hash := gbbi.hasher.Compute(string(messageData))
	body.SetHash(hash)

	err := body.IntegrityAndValidity(gbbi.shardCoordinator)
	if err != nil {
		return err
	}

	isBlockInStorage, _ := gbbi.storer.Has(hash)

	if isBlockInStorage {
		log.Debug("intercepted block body already processed")
		return nil
	}

	blockBody, ok := body.GetUnderlyingObject().(block.Body)
	if !ok {
		return process.ErrCouldNotDecodeUnderlyingBody
	}

	for _, miniblock := range blockBody {
		mbBytes, err := gbbi.marshalizer.Marshal(miniblock)
		if err != nil {
			return err
		}

		_ = gbbi.cache.Put(gbbi.hasher.Compute(string(mbBytes)), miniblock)
	}

	return nil
}
