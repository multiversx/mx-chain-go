package transaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// txResolver is a wrapper over Resolver that is specialized in resolving transaction requests
type txResolver struct {
	process.Resolver
	txPool      data.ShardedDataCacherNotifier
	txStorage   storage.Storer
	marshalizer marshal.Marshalizer
}

// NewTxResolver creates a new transaction resolver
func NewTxResolver(
	resolver process.Resolver,
	txPool data.ShardedDataCacherNotifier,
	txStorage storage.Storer,
	marshalizer marshal.Marshalizer,
) (*txResolver, error) {

	if resolver == nil {
		return nil, process.ErrNilResolver
	}

	if txPool == nil {
		return nil, process.ErrNilTxDataPool
	}

	if txStorage == nil {
		return nil, process.ErrNilTxStorage
	}

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}

	txResolver := &txResolver{
		Resolver:    resolver,
		txPool:      txPool,
		txStorage:   txStorage,
		marshalizer: marshalizer,
	}
	txResolver.SetResolverHandler(txResolver.resolveTxRequest)

	return txResolver, nil
}

func (txRes *txResolver) resolveTxRequest(rd process.RequestData) ([]byte, error) {
	if rd.Type != process.HashType {
		return nil, process.ErrResolveNotHashType
	}

	if rd.Value == nil {
		return nil, process.ErrNilValue
	}

	dataMap := txRes.txPool.SearchData(rd.Value)
	if len(dataMap) > 0 {
		for _, v := range dataMap {
			//since there might be multiple entries, it shall return the first one that it finds
			buff, err := txRes.marshalizer.Marshal(v)
			if err != nil {
				return nil, err
			}

			return buff, nil
		}
	}

	return txRes.txStorage.Get(rd.Value)
}

// RequestTransactionFromHash requests a transaction from other peers having input the tx hash
func (txRes *txResolver) RequestTransactionFromHash(hash []byte) error {
	return txRes.RequestData(process.RequestData{
		Type:  process.HashType,
		Value: hash,
	})
}
