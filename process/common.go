package process

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
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
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) (*block.Header, error) {
	if cacher == nil {
		return nil, ErrNilCacher
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
	}

	return hdr, nil
}

// GetMetaHeader gets the header, which is associated with the given hash, from pool or storage
func GetMetaHeader(
	hash []byte,
	cacher storage.Cacher,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) (*block.MetaBlock, error) {
	if cacher == nil {
		return nil, ErrNilCacher
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
