package block

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// HeaderResolver is a wrapper over Resolver that is specialized in resolving headers requests
type HeaderResolver struct {
	process.Resolver
	hdrPool     data.ShardedDataCacherNotifier
	hdrNonces   data.Uint64Cacher
	hdrStorage  storage.Storer
	marshalizer marshal.Marshalizer
}

// NewHeaderResolver creates a new header resolver
func NewHeaderResolver(
	resolver process.Resolver,
	hdrPool data.ShardedDataCacherNotifier,
	hdrNonces data.Uint64Cacher,
	hdrStorage storage.Storer,
	marshalizer marshal.Marshalizer,
) (*HeaderResolver, error) {

	if resolver == nil {
		return nil, process.ErrNilResolver
	}

	if hdrPool == nil {
		return nil, process.ErrNilHeadersDataPool
	}

	if hdrNonces == nil {
		return nil, process.ErrNilHeadersNoncesDataPool
	}

	if hdrStorage == nil {
		return nil, process.ErrNilHeadersStorage
	}

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}

	hdrResolver := &HeaderResolver{
		Resolver:    resolver,
		hdrPool:     hdrPool,
		hdrNonces:   hdrNonces,
		hdrStorage:  hdrStorage,
		marshalizer: marshalizer,
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

// RequestTransactionFromHash requests a transaction from other peers having input the tx hash
func (txRes *TxResolver) RequestTransactionFromHash(hash []byte) error {
	return txRes.RequestData(process.RequestData{
		Type:  process.HashType,
		Value: hash,
	})
}
