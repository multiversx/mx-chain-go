package transaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// TxResolver is a wrapper over Resolver that is specialized in resolving transaction requests
type TxResolver struct {
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
) (*TxResolver, error) {

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

	txResolver := &TxResolver{
		Resolver:    resolver,
		txPool:      txPool,
		txStorage:   txStorage,
		marshalizer: marshalizer,
	}
	txResolver.SetResolverHandler(txResolver.resolveTxRequest)

	return txResolver, nil
}

func (txRes *TxResolver) resolveTxRequest(rd process.RequestData) ([]byte, error) {
	if rd.Type != process.HashType {
		return nil, process.ErrResolveNotHashType
	}

	if rd.Value == nil {
		return nil, process.ErrNilValue
	}

	value, ok := txRes.txPool.SearchFirstData(rd.Value)
	if !ok {
		return txRes.txStorage.Get(rd.Value)
	}

	buff, err := txRes.marshalizer.Marshal(value)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

// RequestTransactionFromHash requests a transaction from other peers having input the tx hash
func (txRes *TxResolver) RequestTransactionFromHash(hash []byte) error {
	return txRes.RequestData(process.RequestData{
		Type:  process.HashType,
		Value: hash,
	})
}
