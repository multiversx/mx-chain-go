package transactionsfee

import (
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type txGetter struct {
	storer      storage.Storer
	marshalizer marshal.Marshalizer
}

func newTxGetter(storer storage.Storer, marshalizer marshal.Marshalizer) *txGetter {
	return &txGetter{
		storer:      storer,
		marshalizer: marshalizer,
	}
}

func (tg *txGetter) getTxByHash(txHash []byte) (*transaction.Transaction, error) {
	txBytes, err := tg.storer.Get(txHash)
	if err != nil {
		return nil, err
	}

	tx := &transaction.Transaction{}
	err = tg.marshalizer.Unmarshal(tx, txBytes)
	return tx, err
}
