package unsigned

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/feeTx"
	"github.com/ElrondNetwork/elrond-go/process"
	"math/big"
	"sync"
)

const burnPercentage = 0.5      // 100 = 100%, 0 = 0%
const communityPercentage = 0.1 // 10 = 100%, 0 = 0%
const leaderPercentage = 0.4    // 10 = 100%, 0 = 0%

type feeTxHandler struct {
	address process.SpecialAddressHandler
	mutTxs  sync.Mutex
	feeTxs  []*feeTx.FeeTx

	createdTxs []*feeTx.FeeTx
}

// NewFeeTxHandler constructor for the fx tee handler
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
	ftxh.createdTxs = make([]*feeTx.FeeTx, 0)
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

func getPercentageOfValue(value *big.Int, percentage float64) *big.Int {
	x := new(big.Float).SetInt(value)
	y := big.NewFloat(percentage)

	z := new(big.Float).Mul(x, y)

	op := big.NewInt(0)
	result, _ := z.Int(op)

	return result
}

func (ftxh *feeTxHandler) createLeaderTx(totalGathered *big.Int) *feeTx.FeeTx {
	currTx := &feeTx.FeeTx{}

	currTx.Value = getPercentageOfValue(totalGathered, leaderPercentage)
	currTx.RcvAddr = ftxh.address.GetMyOwnAddress()

	return currTx
}

func (ftxh *feeTxHandler) createCommunityTx(totalGathered *big.Int) *feeTx.FeeTx {
	currTx := &feeTx.FeeTx{}

	currTx.Value = getPercentageOfValue(totalGathered, communityPercentage)
	currTx.RcvAddr = ftxh.address.GetElrondCommunityAddress()

	return currTx
}

// CreateAllUtxs creates all the needed fee transactions
// According to economic paper 50% burn, 40% to the leader, 10% to Elrond community fund
func (ftxh *feeTxHandler) CreateAllUTxs() []data.TransactionHandler {
	ftxh.mutTxs.Lock()
	defer ftxh.mutTxs.Unlock()

	totalFee := big.NewInt(0)
	for _, val := range ftxh.feeTxs {
		totalFee = totalFee.Add(totalFee, val.Value)
	}

	if totalFee.Cmp(big.NewInt(1)) < 0 {
		return nil
	}

	leaderTx := ftxh.createLeaderTx(totalFee)
	communityTx := ftxh.createCommunityTx(totalFee)

	currFeeTxs := make([]data.TransactionHandler, 0)
	currFeeTxs = append(currFeeTxs, leaderTx)
	currFeeTxs = append(currFeeTxs, communityTx)

	return currFeeTxs
}

// VerifyCreatedUTxs creates all fee txs from added values, than verifies if in block the values are the same
func (ftxh *feeTxHandler) VerifyCreatedUTxs() error {

	return nil
}
