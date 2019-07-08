package unsigned

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/feeTx"
	"github.com/ElrondNetwork/elrond-go/process"
	"sync"
)

type feeTxHandler struct {
	address process.SpecialAddressHandler
	mutTxs  sync.Mutex
	feeTxs  []*feeTx.FeeTx
}

func NewFeeTxHandler(address process.SpecialAddressHandler) (*feeTxHandler, error) {
	ftxh := &feeTxHandler{
		address: address,
	}
	ftxh.feeTxs = make([]*feeTx.FeeTx, 0)

	return ftxh, nil
}

// CleanProcessedUTxs deletes the cached data
func (ftxh *feeTxHandler) CleanProcessedUTxs() {
	ftxh.mutTxs.Lock()
	ftxh.feeTxs = make([]*feeTx.FeeTx, 0)
	ftxh.mutTxs.Unlock()
}

// AddProcessedUTx adds a new feeTx to the cache
func (ftxh *feeTxHandler) AddProcessedUTx(tx data.TransactionHandler) {
	currFeeTx, ok := tx.(*feeTx.FeeTx)
	if !ok {
		log.Debug(process.ErrWrongTypeAssertion.Error())
	}

	ftxh.mutTxs.Lock()
	ftxh.feeTxs = append(ftxh.feeTxs, currFeeTx)
	ftxh.mutTxs.Unlock()
}

// CreateAllUtxs creates all the needed fee transactions
// According to economic paper 50% burn, 40% to the leader, 10% to Elrond community fund
func (ftxh *feeTxHandler) CreateAllUTxs() []data.TransactionHandler {
	panic("implement me")
}

// VerifyCreatedUTxs
func (ftxh *feeTxHandler) VerifyCreatedUTxs(block block.Body) {
	panic("implement me")
}
