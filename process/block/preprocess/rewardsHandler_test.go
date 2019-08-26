package preprocess

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewRewardTxHandler_NilSpecialAddress(t *testing.T) {
	t.Parallel()

	th, err := NewRewardTxHandler(
		nil,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, th)
	assert.Equal(t, process.ErrNilSpecialAddressHandler, err)
}

func TestNewRewardTxHandler_NilHasher(t *testing.T) {
	t.Parallel()

	th, err := NewRewardTxHandler(
		&mock.SpecialAddressHandlerMock{},
		nil,
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, th)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewRewardTxHandler_NilMarshalizer(t *testing.T) {
	t.Parallel()

	th, err := NewRewardTxHandler(
		&mock.SpecialAddressHandlerMock{},
		&mock.HasherMock{},
		nil,
	)

	assert.Nil(t, th)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewRewardTxHandler_ValsOk(t *testing.T) {
	t.Parallel()

	th, err := NewRewardTxHandler(
		&mock.SpecialAddressHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)
}

func TestRewardTxHandlerAddIntermediateTransactions(t *testing.T) {
	t.Parallel()

	th, err := NewRewardTxHandler(
		&mock.SpecialAddressHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)

	err = th.AddIntermediateTransactions(nil)
	assert.Nil(t, err)
}

func TestRewardTxHandlerProcessTransactionFee(t *testing.T) {
	t.Parallel()

	th, err := NewRewardTxHandler(
		&mock.SpecialAddressHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)

	th.ProcessTransactionFee(nil)
	assert.Equal(t, big.NewInt(0), th.accumulatedFees)

	th.ProcessTransactionFee(big.NewInt(10))
	assert.Equal(t, big.NewInt(10), th.accumulatedFees)

	th.ProcessTransactionFee(big.NewInt(100))
	assert.Equal(t, big.NewInt(110), th.accumulatedFees)
}

func TestRewardTxHandlerAddTxFeeFromBlock(t *testing.T) {
	t.Parallel()

	th, err := NewRewardTxHandler(
		&mock.SpecialAddressHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)

	th.AddRewardTxFromBlock(nil)
	assert.Equal(t, 0, len(th.rewardTxsFromBlock))

	th.AddRewardTxFromBlock(&transaction.Transaction{})
	assert.Equal(t, 0, len(th.rewardTxsFromBlock))

	th.AddRewardTxFromBlock(&rewardTx.RewardTx{})
	assert.Equal(t, 1, len(th.rewardTxsFromBlock))
}

func TestRewardTxHandlerCleanProcessedUTxs(t *testing.T) {
	t.Parallel()

	th, err := NewRewardTxHandler(
		&mock.SpecialAddressHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)

	th.ProcessTransactionFee(big.NewInt(10))
	th.AddRewardTxFromBlock(&rewardTx.RewardTx{})
	assert.Equal(t, big.NewInt(10), th.accumulatedFees)
	assert.Equal(t, 1, len(th.rewardTxsFromBlock))

	th.CleanProcessedUTxs()
	assert.Equal(t, big.NewInt(0), th.accumulatedFees)
	assert.Equal(t, 0, len(th.rewardTxsFromBlock))
}

func TestRewardTxHandlerCreateAllUTxs(t *testing.T) {
	t.Parallel()

	th, err := NewRewardTxHandler(
		&mock.SpecialAddressHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)

	txs := th.CreateAllUTxs()
	assert.Equal(t, 0, len(txs))

	currTxFee := big.NewInt(50)
	th.ProcessTransactionFee(currTxFee)

	txs = th.CreateAllUTxs()
	assert.Equal(t, 3, len(txs))

	totalSum := txs[0].GetValue().Uint64()
	totalSum += txs[1].GetValue().Uint64()
	totalSum += txs[2].GetValue().Uint64()

	assert.Equal(t, currTxFee.Uint64(), totalSum)
}

func TestRewardTxHandlerVerifyCreatedUTxs(t *testing.T) {
	t.Parallel()

	addr := &mock.SpecialAddressHandlerMock{}
	th, err := NewRewardTxHandler(
		addr,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)

	err = th.VerifyCreatedUTxs()
	assert.Nil(t, err)

	currTxFee := big.NewInt(50)
	th.ProcessTransactionFee(currTxFee)

	err = th.VerifyCreatedUTxs()
	assert.Equal(t, process.ErrTxsFeesNotFound, err)

	badValue := big.NewInt(100)
	th.AddRewardTxFromBlock(&rewardTx.RewardTx{Value: badValue})

	err = th.VerifyCreatedUTxs()
	assert.Equal(t, process.ErrTotalTxsFeesDoNotMatch, err)

	th.CleanProcessedUTxs()

	currTxFee = big.NewInt(50)
	halfCurrTxFee := big.NewInt(25)
	th.ProcessTransactionFee(currTxFee)
	th.AddRewardTxFromBlock(&rewardTx.RewardTx{Value: halfCurrTxFee})

	err = th.VerifyCreatedUTxs()
	assert.Equal(t, process.ErrTxsFeesNotFound, err)

	th.CleanProcessedUTxs()

	currTxFee = big.NewInt(50)
	th.ProcessTransactionFee(currTxFee)
	th.AddRewardTxFromBlock(&rewardTx.RewardTx{Value: big.NewInt(5), RcvAddr: addr.ElrondCommunityAddress()})
	th.AddRewardTxFromBlock(&rewardTx.RewardTx{Value: big.NewInt(20), RcvAddr: addr.LeaderAddress()})
	th.AddRewardTxFromBlock(&rewardTx.RewardTx{Value: big.NewInt(25), RcvAddr: addr.BurnAddress()})

	err = th.VerifyCreatedUTxs()
	assert.Nil(t, err)
}

func TestRewardTxHandlerCreateAllInterMiniBlocks(t *testing.T) {
	t.Parallel()

	th, err := NewRewardTxHandler(
		&mock.SpecialAddressHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)

	mbs := th.CreateAllInterMiniBlocks()
	assert.Equal(t, 0, len(mbs))

	currTxFee := big.NewInt(50)
	th.ProcessTransactionFee(currTxFee)

	mbs = th.CreateAllInterMiniBlocks()
	assert.Equal(t, 1, len(mbs))
}

func TestRewardTxHandlerVerifyInterMiniBlocks(t *testing.T) {
	t.Parallel()

	addr := &mock.SpecialAddressHandlerMock{}
	th, err := NewRewardTxHandler(
		addr,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)

	err = th.VerifyInterMiniBlocks(nil)
	assert.Nil(t, err)

	currTxFee := big.NewInt(50)
	th.ProcessTransactionFee(currTxFee)

	err = th.VerifyInterMiniBlocks(nil)
	assert.Equal(t, process.ErrTxsFeesNotFound, err)

	badValue := big.NewInt(100)
	th.AddRewardTxFromBlock(&rewardTx.RewardTx{Value: badValue})

	err = th.VerifyInterMiniBlocks(nil)
	assert.Equal(t, process.ErrTotalTxsFeesDoNotMatch, err)

	th.CleanProcessedUTxs()

	currTxFee = big.NewInt(50)
	halfCurrTxFee := big.NewInt(25)
	th.ProcessTransactionFee(currTxFee)
	th.AddRewardTxFromBlock(&rewardTx.RewardTx{Value: halfCurrTxFee})

	err = th.VerifyInterMiniBlocks(nil)
	assert.Equal(t, process.ErrTxsFeesNotFound, err)

	th.CleanProcessedUTxs()

	currTxFee = big.NewInt(50)
	th.ProcessTransactionFee(currTxFee)
	th.AddRewardTxFromBlock(&rewardTx.RewardTx{Value: big.NewInt(5), RcvAddr: addr.ElrondCommunityAddress()})
	th.AddRewardTxFromBlock(&rewardTx.RewardTx{Value: big.NewInt(20), RcvAddr: addr.LeaderAddress()})
}
