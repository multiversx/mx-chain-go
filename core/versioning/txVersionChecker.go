package versioning

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

const (
	maskSignedWithHash = uint32(16777216)
)

// TxVersionChecker represents transaction option decoder
type TxVersionChecker struct {
	txOptions    uint32
	minTxVersion uint32
	txVersion    uint32
}

// NewTxVersionChecker will create a new instance of TxOptionsChecker
func NewTxVersionChecker(minTxVersion uint32, tx *transaction.Transaction) *TxVersionChecker {
	return &TxVersionChecker{
		txVersion:    tx.Version,
		txOptions:    tx.Options,
		minTxVersion: minTxVersion,
	}
}

// IsSignedWithHash will return true is transaction is signed with hash
func (toc *TxVersionChecker) IsSignedWithHash() bool {
	if toc.txVersion > toc.minTxVersion {
		// transaction is signed with hash if first byte from first octet from options is set with 1
		return toc.txOptions&maskSignedWithHash > 0
	}

	return false
}

// CheckTxVersion will check transaction version
func (toc *TxVersionChecker) CheckTxVersion() error {
	if (toc.txVersion == 1 && toc.txOptions != 0) || toc.txVersion < toc.minTxVersion {
		return core.ErrInvalidTransactionVersion
	}

	return nil
}
