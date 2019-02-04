package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// HeaderResolver is a wrapper over Resolver that is specialized in resolving headers requests
type HeaderResolver struct {
	process.Resolver
	hdrPool        data.ShardedDataCacherNotifier
	hdrNonces      data.Uint64Cacher
	hdrStorage     storage.Storer
	marshalizer    marshal.Marshalizer
	nonceConverter typeConverters.Uint64ByteSliceConverter
}

// GenericBlockBodyResolver is a wrapper over Resolver that is specialized in resolving block body requests
type GenericBlockBodyResolver struct {
	process.Resolver
	blockBodyPool storage.Cacher
	blockStorage  storage.Storer
	marshalizer   marshal.Marshalizer
}

//------- headerResolver

// NewHeaderResolver creates a new header resolver
func NewHeaderResolver(
	resolver process.Resolver,
	transient data.TransientDataHolder,
	hdrStorage storage.Storer,
	marshalizer marshal.Marshalizer,
	nonceConverter typeConverters.Uint64ByteSliceConverter,
) (*HeaderResolver, error) {

	if resolver == nil {
		return nil, process.ErrNilResolver
	}

	if transient == nil {
		return nil, process.ErrNilTransientPool
	}

	headers := transient.Headers()
	if headers == nil {
		return nil, process.ErrNilHeadersDataPool
	}

	headersNonces := transient.HeadersNonces()
	if headersNonces == nil {
		return nil, process.ErrNilHeadersNoncesDataPool
	}

	if hdrStorage == nil {
		return nil, process.ErrNilHeadersStorage
	}

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}

	if nonceConverter == nil {
		return nil, process.ErrNilNonceConverter
	}

	hdrResolver := &HeaderResolver{
		Resolver:       resolver,
		hdrPool:        transient.Headers(),
		hdrNonces:      transient.HeadersNonces(),
		hdrStorage:     hdrStorage,
		marshalizer:    marshalizer,
		nonceConverter: nonceConverter,
	}
	hdrResolver.SetResolverHandler(hdrResolver.resolveHdrRequest)

	return hdrResolver, nil
}

func (hdrRes *HeaderResolver) resolveHdrRequest(rd process.RequestData) ([]byte, error) {
	if rd.Value == nil {
		return nil, process.ErrNilValue
	}

	var buff []byte
	var err error

	switch rd.Type {
	case process.HashType:
		buff, err = hdrRes.resolveHeaderFromHash(rd.Value)
	case process.NonceType:
		buff, err = hdrRes.resolveHeaderFromNonce(rd.Value)
	default:
		return nil, process.ErrResolveTypeUnknown
	}

	return buff, err
}

func (hdrRes *HeaderResolver) resolveHeaderFromHash(key []byte) ([]byte, error) {
	value, ok := hdrRes.hdrPool.SearchFirstData(key)

	if !ok {
		return hdrRes.hdrStorage.Get(key)
	}

	buff, err := hdrRes.marshalizer.Marshal(value)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

func (hdrRes *HeaderResolver) resolveHeaderFromNonce(key []byte) ([]byte, error) {
	//key is now an encoded nonce (uint64)

	//Step 1. decode the nonce from the key
	nonce, err := hdrRes.nonceConverter.ToUint64(key)
	if err != nil {
		return nil, process.ErrInvalidNonceByteSlice
	}

	//Step 2. search the nonce-key pair
	hash, _ := hdrRes.hdrNonces.Get(nonce)
	if hash == nil {
		return nil, nil
	}

	//Step 3. search header by key (hash)
	value, ok := hdrRes.hdrPool.SearchFirstData(hash)
	if !ok {
		return hdrRes.hdrStorage.Get(hash)
	}

	//since there might be multiple entries, it shall return the first one that it finds
	buff, err := hdrRes.marshalizer.Marshal(value)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

// RequestHeaderFromHash requests a header from other peers having input the hdr hash
func (hdrRes *HeaderResolver) RequestHeaderFromHash(hash []byte) error {
	return hdrRes.RequestData(process.RequestData{
		Type:  process.HashType,
		Value: hash,
	})
}

// RequestHeaderFromNonce requests a header from other peers having input the hdr nonce
func (hdrRes *HeaderResolver) RequestHeaderFromNonce(nonce uint64) error {
	return hdrRes.RequestData(process.RequestData{
		Type:  process.NonceType,
		Value: hdrRes.nonceConverter.ToByteSlice(nonce),
	})
}

//------- genericBlockBodyResolver

// NewGenericBlockBodyResolver creates a new block body resolver
func NewGenericBlockBodyResolver(
	resolver process.Resolver,
	blockBodyPool storage.Cacher,
	blockBodyStorage storage.Storer,
	marshalizer marshal.Marshalizer) (*GenericBlockBodyResolver, error) {

	if resolver == nil {
		return nil, process.ErrNilResolver
	}

	if blockBodyPool == nil {
		return nil, process.ErrNilBlockBodyPool
	}

	if blockBodyStorage == nil {
		return nil, process.ErrNilBlockBodyStorage
	}

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}

	bbResolver := &GenericBlockBodyResolver{
		Resolver:      resolver,
		blockBodyPool: blockBodyPool,
		blockStorage:  blockBodyStorage,
		marshalizer:   marshalizer,
	}
	bbResolver.SetResolverHandler(bbResolver.resolveBlockBodyRequest)

	return bbResolver, nil
}

func (gbbRes *GenericBlockBodyResolver) resolveBlockBodyRequest(rd process.RequestData) ([]byte, error) {
	if rd.Type != process.HashType {
		return nil, process.ErrResolveNotHashType
	}

	if rd.Value == nil {
		return nil, process.ErrNilValue
	}

	blockBody, _ := gbbRes.blockBodyPool.Get(rd.Value)
	if blockBody != nil {
		buff, err := gbbRes.marshalizer.Marshal(blockBody)
		if err != nil {
			return nil, err
		}

		return buff, nil
	}

	return gbbRes.blockStorage.Get(rd.Value)
}

// RequestBlockBodyFromHash requests a block body from other peers having input the block body hash
func (gbbRes *GenericBlockBodyResolver) RequestBlockBodyFromHash(hash []byte) error {
	return gbbRes.RequestData(process.RequestData{
		Type:  process.HashType,
		Value: hash,
	})
}
