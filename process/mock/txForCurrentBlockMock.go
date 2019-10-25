package mock

import "github.com/ElrondNetwork/elrond-go/data"

type TxForCurrentBlockMock struct {
	CleanCalled func()
	GetTxCalled func(txHash []byte) (data.TransactionHandler, error)
	AddTxCalled func(txHash []byte, tx data.TransactionHandler)
}

func (t *TxForCurrentBlockMock) Clean() {
	if t.CleanCalled != nil {
		t.CleanCalled()
	}
}

func (t *TxForCurrentBlockMock) GetTx(txHash []byte) (data.TransactionHandler, error) {
	if t.GetTxCalled != nil {
		return t.GetTxCalled(txHash)
	}
	return nil, nil
}

func (t *TxForCurrentBlockMock) AddTx(txHash []byte, tx data.TransactionHandler) {
	if t.AddTxCalled != nil {
		t.AddTxCalled(txHash, tx)
	}
}

func (t *TxForCurrentBlockMock) IsInterfaceNil() bool {
	if t == nil {
		return true
	}
	return false
}
