package process

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// GetShardHeader gets the header, which is associated with the given hash, from pool or storage
func GetShardHeader(hash []byte,
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
func GetMetaHeader(hash []byte,
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
func GetShardHeaderFromPool(hash []byte,
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

// GetMetaHeaderFromPool gets the header, which is associated with the given hash, from pool
func GetMetaHeaderFromPool(hash []byte,
	cacher storage.Cacher,
) (*block.MetaBlock, error) {
	if cacher == nil {
		return nil, ErrNilCacher
	}

	hdr, ok := cacher.Peek(hash)
	if !ok {
		return nil, ErrMissingHeader
	}

	header, ok := hdr.(*block.MetaBlock)
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

	buffHeader, err := GetMarshalizedHeaderFromStorage(dataRetriever.BlockHeaderUnit, hash, marshalizer, storageService)
	if err != nil {
		return nil, err
	}

	header := &block.Header{}
	err = marshalizer.Unmarshal(header, buffHeader)
	if err != nil {
		return nil, ErrUnmarshalWithoutSuccess
	}

	return header, nil
}

// GetMetaHeaderFromStorage gets the header, which is associated with the given hash, from storage
func GetMetaHeaderFromStorage(
	hash []byte,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) (*block.MetaBlock, error) {

	buffHeader, err := GetMarshalizedHeaderFromStorage(dataRetriever.MetaBlockUnit, hash, marshalizer, storageService)
	if err != nil {
		return nil, err
	}

	header := &block.MetaBlock{}
	err = marshalizer.Unmarshal(header, buffHeader)
	if err != nil {
		return nil, ErrUnmarshalWithoutSuccess
	}

	return header, nil
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

	headerStore := storageService.GetStorer(blockUnit)
	if headerStore == nil {
		return nil, ErrNilHeadersStorage
	}

	buffHeader, err := headerStore.Get(hash)
	if err != nil {
		return nil, ErrMissingHeader
	}

	return buffHeader, nil
}
