package dataValidators

import (
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/prometheus/common/log"
)

// nilTxValidator represents a tx handler validator that doesn't check the validity of provided txHandler
type storageTxValidator struct {
	txStorer    storage.Storer
	marshalizer marshal.Marshalizer
	hasher      hashing.Hasher
}

// NewNilTxValidator creates a new nil tx handler validator instance
func NewStorageTxValidator(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	txStorer storage.Storer,
) (*storageTxValidator, error) {
	return &storageTxValidator{txStorer: txStorer,
		marshalizer: marshalizer,
		hasher:      hasher}, nil
}

// IsTxValidForProcessing is a nil implementation that will return true
func (stv *storageTxValidator) IsTxValidForProcessing(txHandler process.TxValidatorHandler) bool {
	tx, ok := txHandler.(*transaction.InterceptedTransaction)
	if !ok {
		return false
	}

	err := stv.txStorer.Has(tx.Hash())
	isTxInStorage := err == nil
	if isTxInStorage {
		log.Debug("intercepted tx already processed")
		return false
	}
	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (stv *storageTxValidator) IsInterfaceNil() bool {
	if stv == nil {
		return true
	}
	return false
}

func (stv *storageTxValidator) NumRejectedTxs() uint64 {
	return 0
}
