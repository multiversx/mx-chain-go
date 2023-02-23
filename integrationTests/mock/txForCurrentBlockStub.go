package mock

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
)

// TxForCurrentBlockStub -
type TxForCurrentBlockStub struct {
	CleanCalled func()
	GetTxCalled func(txHash []byte) (data.TransactionHandler, error)
	AddTxCalled func(txHash []byte, tx data.TransactionHandler)
}

// Clean -
func (t *TxForCurrentBlockStub) Clean() {
	if t.CleanCalled != nil {
		t.CleanCalled()
	}
}

// GetTx -
func (t *TxForCurrentBlockStub) GetTx(txHash []byte) (data.TransactionHandler, error) {
	if t.GetTxCalled != nil {
		return t.GetTxCalled(txHash)
	}
	return &rewardTx.RewardTx{Value: big.NewInt(0)}, nil
}

// AddTx -
func (t *TxForCurrentBlockStub) AddTx(txHash []byte, tx data.TransactionHandler) {
	if t.AddTxCalled != nil {
		t.AddTxCalled(txHash, tx)
	}
}

// IsInterfaceNil -
func (t *TxForCurrentBlockStub) IsInterfaceNil() bool {
	return t == nil
}
