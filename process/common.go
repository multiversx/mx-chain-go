package process

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

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

// GetShardHeaderFromPool gets the header, which is associated with the given hash, from pool
func GetShardHeaderFromPool(
	hash []byte,
	cacher storage.Cacher,
) (*block.Header, error) {
	if cacher == nil {
		return nil, ErrNilCacher
	}

	hdr, ok := cacher.Peek(hash)
	if !ok {
		return nil, ErrMissingHeader
	}

	header, ok := hdr.(*block.Header)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	return header, nil
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

	headerStore := storageService.GetStorer(dataRetriever.BlockHeaderUnit)
	if headerStore == nil {
		return nil, ErrNilHeadersStorage
	}

	buffHeader, err := headerStore.Get(hash)
	if err != nil {
		return nil, ErrMissingHeader
	}

	header := &block.Header{}
	err = marshalizer.Unmarshal(header, buffHeader)
	if err != nil {
		return nil, ErrUnmarshalWithoutSuccess
	}

	return header, nil
}

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

	headerStore := storageService.GetStorer(dataRetriever.MetaBlockUnit)
	if headerStore == nil {
		return nil, ErrNilHeadersStorage
	}

	buffHeader, err := headerStore.Get(hash)
	if err != nil {
		return nil, ErrMissingHeader
	}

	header := &block.MetaBlock{}
	err = marshalizer.Unmarshal(header, buffHeader)
	if err != nil {
		return nil, ErrUnmarshalWithoutSuccess
	}

	return header, nil
}
