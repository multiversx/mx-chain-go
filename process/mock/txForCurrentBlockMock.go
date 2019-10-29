package mock

import "github.com/ElrondNetwork/elrond-go/data"

type TxForCurrentBlockStub struct {
	CleanCalled func()
	GetTxCalled func(txHash []byte) (data.TransactionHandler, error)
	AddTxCalled func(txHash []byte, tx data.TransactionHandler)
}

func (t *TxForCurrentBlockStub) Clean() {
	if t.CleanCalled != nil {
		t.CleanCalled()
	}
}

func (t *TxForCurrentBlockStub) GetTx(txHash []byte) (data.TransactionHandler, error) {
	if t.GetTxCalled != nil {
		return t.GetTxCalled(txHash)
	}
	return nil, nil
}

func (t *TxForCurrentBlockStub) AddTx(txHash []byte, tx data.TransactionHandler) {
	if t.AddTxCalled != nil {
		t.AddTxCalled(txHash, tx)
	}
}

func (t *TxForCurrentBlockStub) IsInterfaceNil() bool {
	if t == nil {
		return true
	}
	return false
}
