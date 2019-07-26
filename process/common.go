package process

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// EmptyChannel empties the given channel
func EmptyChannel(ch chan bool) {
	for len(ch) > 0 {
		<-ch
	}
}

// GetShardHeader gets the header, which is associated with the given hash, from pool or storage
func GetShardHeader(
	hash []byte,
	cacher storage.Cacher,
	uint64SyncMapCacher dataRetriever.Uint64SyncMapCacher,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) (*block.Header, error) {

	if cacher == nil {
		return nil, ErrNilCacher
	}
	if uint64SyncMapCacher == nil {
		return nil, ErrNilUint64SyncMapCacher
	}
	if marshalizer == nil {
		return nil, ErrNilMarshalizer
	}
	if storageService == nil {
		return nil, ErrNilStorage
	}

	hdr, err := GetShardHeaderFromPool(hash, cacher)
	if err != nil {
		hdr, err = GetShardHeaderFromStorage(hash, marshalizer, storageService)
		if err != nil {
			return nil, err
		}

		cacher.Put(hash, hdr)
		syncMap := &dataPool.ShardIdHashSyncMap{}
		syncMap.Store(hdr.GetShardID(), hash)
		uint64SyncMapCacher.Merge(hdr.GetNonce(), syncMap)
	}

	return hdr, nil
}

// GetMetaHeader gets the header, which is associated with the given hash, from pool or storage
func GetMetaHeader(
	hash []byte,
	cacher storage.Cacher,
	uint64SyncMapCacher dataRetriever.Uint64SyncMapCacher,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) (*block.MetaBlock, error) {

	if cacher == nil {
		return nil, ErrNilCacher
	}
	if uint64SyncMapCacher == nil {
		return nil, ErrNilUint64SyncMapCacher
	}
	if marshalizer == nil {
		return nil, ErrNilMarshalizer
	}
	if storageService == nil {
		return nil, ErrNilStorage
	}

	hdr, err := GetMetaHeaderFromPool(hash, cacher)
	if err != nil {
		hdr, err = GetMetaHeaderFromStorage(hash, marshalizer, storageService)
		if err != nil {
			return nil, err
		}

		cacher.Put(hash, hdr)
		syncMap := &dataPool.ShardIdHashSyncMap{}
		syncMap.Store(hdr.GetShardID(), hash)
		uint64SyncMapCacher.Merge(hdr.GetNonce(), syncMap)
	}

	return hdr, nil
}

// GetShardHeaderFromPool gets the header, which is associated with the given hash, from pool
func GetShardHeaderFromPool(
	hash []byte,
	cacher storage.Cacher,
) (*block.Header, error) {

	if cacher == nil {
		return nil, ErrNilCacher
	}

	obj, ok := cacher.Peek(hash)
	if !ok {
		return nil, ErrMissingHeader
	}

	hdr, ok := obj.(*block.Header)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	return hdr, nil
}

// GetMetaHeaderFromPool gets the header, which is associated with the given hash, from pool
func GetMetaHeaderFromPool(
	hash []byte,
	cacher storage.Cacher,
) (*block.MetaBlock, error) {

	if cacher == nil {
		return nil, ErrNilCacher
	}

	obj, ok := cacher.Peek(hash)
	if !ok {
		return nil, ErrMissingHeader
	}

	hdr, ok := obj.(*block.MetaBlock)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	return hdr, nil
}

// GetShardHeaderFromStorage gets the header, which is associated with the given hash, from storage
func GetShardHeaderFromStorage(
	hash []byte,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) (*block.Header, error) {

	if marshalizer == nil {
		return nil, ErrNilMarshalizer
	}
	if storageService == nil {
		return nil, ErrNilStorage
	}

	buffHdr, err := GetMarshalizedHeaderFromStorage(dataRetriever.BlockHeaderUnit, hash, marshalizer, storageService)
	if err != nil {
		return nil, err
	}

	hdr := &block.Header{}
	err = marshalizer.Unmarshal(hdr, buffHdr)
	if err != nil {
		return nil, ErrUnmarshalWithoutSuccess
	}

	return hdr, nil
}

// GetMetaHeaderFromStorage gets the header, which is associated with the given hash, from storage
func GetMetaHeaderFromStorage(
	hash []byte,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) (*block.MetaBlock, error) {

	if marshalizer == nil {
		return nil, ErrNilMarshalizer
	}
	if storageService == nil {
		return nil, ErrNilStorage
	}

	buffHdr, err := GetMarshalizedHeaderFromStorage(dataRetriever.MetaBlockUnit, hash, marshalizer, storageService)
	if err != nil {
		return nil, err
	}

	hdr := &block.MetaBlock{}
	err = marshalizer.Unmarshal(hdr, buffHdr)
	if err != nil {
		return nil, ErrUnmarshalWithoutSuccess
	}

	return hdr, nil
}

// GetMarshalizedHeaderFromStorage gets the marshalized header, which is associated with the given hash, from storage
func GetMarshalizedHeaderFromStorage(
	blockUnit dataRetriever.UnitType,
	hash []byte,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) ([]byte, error) {

	if marshalizer == nil {
		return nil, ErrNilMarshalizer
	}
	if storageService == nil {
		return nil, ErrNilStorage
	}

	hdrStore := storageService.GetStorer(blockUnit)
	if hdrStore == nil {
		return nil, ErrNilHeadersStorage
	}

	buffHdr, err := hdrStore.Get(hash)
	if err != nil {
		return nil, ErrMissingHeader
	}

	return buffHdr, nil
}

// GetMetaHeaderWithNonce method returns a meta block header with a given nonce
func GetMetaHeaderWithNonce(
	nonce uint64,
	cacher storage.Cacher,
	uint64SyncMapCacher dataRetriever.Uint64SyncMapCacher,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
) (*block.MetaBlock, []byte, error) {

	if cacher == nil {
		return nil, nil, ErrNilCacher
	}
	if uint64SyncMapCacher == nil {
		return nil, nil, ErrNilUint64SyncMapCacher
	}
	if marshalizer == nil {
		return nil, nil, ErrNilMarshalizer
	}
	if storageService == nil {
		return nil, nil, ErrNilStorage
	}
	if uint64Converter == nil {
		return nil, nil, ErrNilUint64Converter
	}

	hdr, hash, err := GetMetaHeaderFromPoolWithNonce(nonce, cacher, uint64SyncMapCacher)
	if err != nil {
		hdr, hash, err = GetMetaHeaderFromStorageWithNonce(nonce, storageService, uint64Converter, marshalizer)
		if err != nil {
			return nil, nil, err
		}

		cacher.Put(hash, hdr)
		syncMap := &dataPool.ShardIdHashSyncMap{}
		syncMap.Store(hdr.GetShardID(), hash)
		uint64SyncMapCacher.Merge(hdr.GetNonce(), syncMap)
	}

	return hdr, hash, nil
}

// GetMetaHeaderFromPoolWithNonce method returns a meta block header from pool with a given nonce
func GetMetaHeaderFromPoolWithNonce(
	nonce uint64,
	cacher storage.Cacher,
	uint64SyncMapCacher dataRetriever.Uint64SyncMapCacher,
) (*block.MetaBlock, []byte, error) {

	if cacher == nil {
		return nil, nil, ErrNilCacher
	}
	if uint64SyncMapCacher == nil {
		return nil, nil, ErrNilUint64SyncMapCacher
	}

	syncMap, ok := uint64SyncMapCacher.Get(nonce)
	if !ok {
		return nil, nil, ErrMissingHashForHeaderNonce
	}

	hash, ok := syncMap.Load(sharding.MetachainShardId)
	if hash == nil || !ok {
		return nil, nil, ErrMissingHashForHeaderNonce
	}

	obj, ok := cacher.Peek(hash)
	if !ok {
		return nil, nil, ErrMissingHeader
	}

	hdr, ok := obj.(*block.MetaBlock)
	if !ok {
		return nil, nil, ErrWrongTypeAssertion
	}

	return hdr, hash, nil
}

// GetMetaHeaderFromStorageWithNonce method returns a meta block header from storage with a given nonce
func GetMetaHeaderFromStorageWithNonce(
	nonce uint64,
	storageService dataRetriever.StorageService,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
	marshalizer marshal.Marshalizer,
) (*block.MetaBlock, []byte, error) {

	if storageService == nil {
		return nil, nil, ErrNilStorage
	}
	if uint64Converter == nil {
		return nil, nil, ErrNilUint64Converter
	}
	if marshalizer == nil {
		return nil, nil, ErrNilMarshalizer
	}

	headerStore := storageService.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	if headerStore == nil {
		return nil, nil, ErrNilHeadersStorage
	}

	nonceToByteSlice := uint64Converter.ToByteSlice(nonce)
	hash, err := headerStore.Get(nonceToByteSlice)
	if err != nil {
		return nil, nil, ErrMissingHashForHeaderNonce
	}

	hdr, err := GetMetaHeaderFromStorage(hash, marshalizer, storageService)
	if err != nil {
		return nil, nil, err
	}

	return hdr, hash, nil
}

// GetShardHeaderWithNonce method returns a shard block header with a given nonce and shardId
func GetShardHeaderWithNonce(
	nonce uint64,
	shardId uint32,
	cacher storage.Cacher,
	uint64SyncMapCacher dataRetriever.Uint64SyncMapCacher,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
) (*block.Header, []byte, error) {

	if cacher == nil {
		return nil, nil, ErrNilCacher
	}
	if uint64SyncMapCacher == nil {
		return nil, nil, ErrNilUint64SyncMapCacher
	}
	if marshalizer == nil {
		return nil, nil, ErrNilMarshalizer
	}
	if storageService == nil {
		return nil, nil, ErrNilStorage
	}
	if uint64Converter == nil {
		return nil, nil, ErrNilUint64Converter
	}

	hdr, hash, err := GetShardHeaderFromPoolWithNonce(nonce, shardId, cacher, uint64SyncMapCacher)
	if err != nil {
		hdr, hash, err = GetShardHeaderFromStorageWithNonce(nonce, shardId, storageService, uint64Converter, marshalizer)
		if err != nil {
			return nil, nil, err
		}

		cacher.Put(hash, hdr)
		syncMap := &dataPool.ShardIdHashSyncMap{}
		syncMap.Store(hdr.GetShardID(), hash)
		uint64SyncMapCacher.Merge(hdr.GetNonce(), syncMap)
	}

	return hdr, hash, nil
}

// GetShardHeaderFromPoolWithNonce method returns a shard block header from pool with a given nonce and shardId
func GetShardHeaderFromPoolWithNonce(
	nonce uint64,
	shardId uint32,
	cacher storage.Cacher,
	uint64SyncMapCacher dataRetriever.Uint64SyncMapCacher,
) (*block.Header, []byte, error) {

	if cacher == nil {
		return nil, nil, ErrNilCacher
	}
	if uint64SyncMapCacher == nil {
		return nil, nil, ErrNilUint64SyncMapCacher
	}

	syncMap, ok := uint64SyncMapCacher.Get(nonce)
	if !ok {
		return nil, nil, ErrMissingHashForHeaderNonce
	}

	hash, ok := syncMap.Load(shardId)
	if hash == nil || !ok {
		return nil, nil, ErrMissingHashForHeaderNonce
	}

	obj, ok := cacher.Peek(hash)
	if !ok {
		return nil, nil, ErrMissingHeader
	}

	hdr, ok := obj.(*block.Header)
	if !ok {
		return nil, nil, ErrWrongTypeAssertion
	}

	return hdr, hash, nil
}

// GetShardHeaderFromStorageWithNonce method returns a shard block header from storage with a given nonce and shardId
func GetShardHeaderFromStorageWithNonce(
	nonce uint64,
	shardId uint32,
	storageService dataRetriever.StorageService,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
	marshalizer marshal.Marshalizer,
) (*block.Header, []byte, error) {

	if storageService == nil {
		return nil, nil, ErrNilStorage
	}
	if uint64Converter == nil {
		return nil, nil, ErrNilUint64Converter
	}
	if marshalizer == nil {
		return nil, nil, ErrNilMarshalizer
	}

	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardId)
	headerStore := storageService.GetStorer(hdrNonceHashDataUnit)
	if headerStore == nil {
		return nil, nil, ErrNilHeadersStorage
	}

	nonceToByteSlice := uint64Converter.ToByteSlice(nonce)
	hash, err := headerStore.Get(nonceToByteSlice)
	if err != nil {
		return nil, nil, ErrMissingHashForHeaderNonce
	}

	hdr, err := GetShardHeaderFromStorage(hash, marshalizer, storageService)
	if err != nil {
		return nil, nil, err
	}

	return hdr, hash, nil
}
