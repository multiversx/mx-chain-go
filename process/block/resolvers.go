package block

import (
	"fmt"

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

//------- HeaderResolver

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

func (hdrRes *HeaderResolver) resolveHdrRequest(rd process.RequestData) []byte {
	if rd.Value == nil {
		return nil
	}

	var buff []byte
	var err error

	switch rd.Type {
	case process.HashType:
		buff, err = hdrRes.resolveHeaderFromHash(rd.Value)
	case process.NonceType:
		buff, err = hdrRes.resolveHeaderFromNonce(rd.Value)
	default:
		log.Debug(fmt.Sprintf("unknown request type %d", rd.Type))
		return nil
	}

	if err != nil {
		log.Debug(err.Error())
	}
	return buff
}

func (hdrRes *HeaderResolver) resolveHeaderFromHash(key []byte) ([]byte, error) {
	dataMap := hdrRes.hdrPool.SearchData(key)
	if len(dataMap) > 0 {
		for _, v := range dataMap {
			//since there might be multiple entries, it shall return the first one that it finds
			buff, err := hdrRes.marshalizer.Marshal(v)
			if err != nil {
				return nil, err
			}

			return buff, nil
		}
	}

	return hdrRes.hdrStorage.Get(key)
}

func (hdrRes *HeaderResolver) resolveHeaderFromNonce(key []byte) ([]byte, error) {
	//key is now an encoded nonce (uint64)

	//Step 1. decode the nonce from the key
	nonce := hdrRes.nonceConverter.ToUint64(key)
	if nonce == nil {
		return nil, process.ErrInvalidNonceByteSlice
	}

	//Step 2. search the nonce-key pair
	hash, _ := hdrRes.hdrNonces.Get(*nonce)
	if hash == nil {
		//not found
		return nil, nil
	}

	//Step 3. search header by key (hash)
	dataMap := hdrRes.hdrPool.SearchData(hash)
	if len(dataMap) > 0 {
		for _, v := range dataMap {
			//since there might be multiple entries, it shall return the first one that it finds
			buff, err := hdrRes.marshalizer.Marshal(v)
			if err != nil {
				return nil, err
			}

			return buff, nil
		}
	}

	return hdrRes.hdrStorage.Get(hash)
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
