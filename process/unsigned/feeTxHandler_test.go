package unsigned

import (
	"github.com/ElrondNetwork/elrond-go/data/feeTx"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestNewFeeTxHandler_NilSpecialAddress(t *testing.T) {
	t.Parallel()

	th, err := NewFeeTxHandler(
		nil,
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, th)
	assert.Equal(t, process.ErrNilSpecialAddressHandler, err)
}

func TestNewFeeTxHandler_NilHasher(t *testing.T) {
	t.Parallel()

	th, err := NewFeeTxHandler(
		&mock.SpecialAddressHandlerMock{},
		mock.NewMultipleShardsCoordinatorMock(),
		nil,
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, th)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewFeeTxHandler_NilMarshalizer(t *testing.T) {
	t.Parallel()

	th, err := NewFeeTxHandler(
		&mock.SpecialAddressHandlerMock{},
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.HasherMock{},
		nil,
	)

	assert.Nil(t, th)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewFeeTxHandler_ValsOk(t *testing.T) {
	t.Parallel()

	th, err := NewFeeTxHandler(
		&mock.SpecialAddressHandlerMock{},
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)
}

func TestFeeTxHandler_AddIntermediateTransactions(t *testing.T) {
	t.Parallel()

	th, err := NewFeeTxHandler(
		&mock.SpecialAddressHandlerMock{},
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)

	err = th.AddIntermediateTransactions(nil)
	assert.Nil(t, err)
}

func TestFeeTxHandler_AddProcessedUTx(t *testing.T) {
	t.Parallel()

	th, err := NewFeeTxHandler(
		&mock.SpecialAddressHandlerMock{},
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)

	th.AddProcessedUTx(nil)
	assert.Equal(t, 0, len(th.feeTxs))

	th.AddProcessedUTx(&transaction.Transaction{})
	assert.Equal(t, 0, len(th.feeTxs))

	th.AddProcessedUTx(&feeTx.FeeTx{})
	assert.Equal(t, 1, len(th.feeTxs))
}

func TestFeeTxHandler_AddTxFeeFromBlock(t *testing.T) {
	t.Parallel()

	th, err := NewFeeTxHandler(
		&mock.SpecialAddressHandlerMock{},
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)

	th.AddTxFeeFromBlock(nil)
	assert.Equal(t, 0, len(th.feeTxsFromBlock))

	th.AddTxFeeFromBlock(&transaction.Transaction{})
	assert.Equal(t, 0, len(th.feeTxsFromBlock))

	th.AddTxFeeFromBlock(&feeTx.FeeTx{})
	assert.Equal(t, 1, len(th.feeTxsFromBlock))
}

func TestFeeTxHandler_CleanProcessedUTxs(t *testing.T) {
	t.Parallel()

	th, err := NewFeeTxHandler(
		&mock.SpecialAddressHandlerMock{},
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)

	th.AddProcessedUTx(&feeTx.FeeTx{})
	th.AddTxFeeFromBlock(&feeTx.FeeTx{})
	assert.Equal(t, 1, len(th.feeTxs))
	assert.Equal(t, 1, len(th.feeTxsFromBlock))

	th.CleanProcessedUTxs()
	assert.Equal(t, 0, len(th.feeTxs))
	assert.Equal(t, 0, len(th.feeTxsFromBlock))
}

func TestFeeTxHandler_CreateAllUTxs(t *testing.T) {
	t.Parallel()

	th, err := NewFeeTxHandler(
		&mock.SpecialAddressHandlerMock{},
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)

	txs := th.CreateAllUTxs()
	assert.Equal(t, 0, len(txs))

	currTxFee := big.NewInt(50)
	th.AddProcessedUTx(&feeTx.FeeTx{Value: currTxFee})

	txs = th.CreateAllUTxs()
	assert.Equal(t, 3, len(txs))

	totalSum := txs[0].GetValue().Uint64()
	totalSum += txs[1].GetValue().Uint64()
	totalSum += txs[2].GetValue().Uint64()

	assert.Equal(t, currTxFee.Uint64(), totalSum)
}

func TestFeeTxHandler_VerifyCreatedUTxs(t *testing.T) {
	t.Parallel()

	addr := &mock.SpecialAddressHandlerMock{}
	th, err := NewFeeTxHandler(
		addr,
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)

	err = th.VerifyCreatedUTxs()
	assert.Nil(t, err)

	currTxFee := big.NewInt(50)
	th.AddProcessedUTx(&feeTx.FeeTx{Value: currTxFee})

	err = th.VerifyCreatedUTxs()
	assert.Equal(t, process.ErrTxsFeesNotFound, err)

	badValue := big.NewInt(100)
	th.AddTxFeeFromBlock(&feeTx.FeeTx{Value: badValue})

	err = th.VerifyCreatedUTxs()
	assert.Equal(t, process.ErrTotalTxsFeesDoNotMatch, err)

	th.CleanProcessedUTxs()

	currTxFee = big.NewInt(50)
	halfCurrTxFee := big.NewInt(25)
	th.AddProcessedUTx(&feeTx.FeeTx{Value: currTxFee})
	th.AddTxFeeFromBlock(&feeTx.FeeTx{Value: halfCurrTxFee})

	err = th.VerifyCreatedUTxs()
	assert.Equal(t, process.ErrTxsFeesNotFound, err)

	th.CleanProcessedUTxs()

	currTxFee = big.NewInt(50)
	th.AddProcessedUTx(&feeTx.FeeTx{Value: currTxFee})
	th.AddTxFeeFromBlock(&feeTx.FeeTx{Value: big.NewInt(5), RcvAddr: addr.ElrondCommunityAddress()})
	th.AddTxFeeFromBlock(&feeTx.FeeTx{Value: big.NewInt(20), RcvAddr: addr.LeaderAddress()})
	th.AddTxFeeFromBlock(&feeTx.FeeTx{Value: big.NewInt(25), RcvAddr: addr.BurnAddress()})

	err = th.VerifyCreatedUTxs()
	assert.Nil(t, err)
}

func TestFeeTxHandler_CreateAllInterMiniBlocks(t *testing.T) {
	t.Parallel()

	th, err := NewFeeTxHandler(
		&mock.SpecialAddressHandlerMock{},
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)

	mbs := th.CreateAllInterMiniBlocks()
	assert.Equal(t, 0, len(mbs))

	currTxFee := big.NewInt(50)
	th.AddProcessedUTx(&feeTx.FeeTx{Value: currTxFee})

	mbs = th.CreateAllInterMiniBlocks()
	assert.Equal(t, 1, len(mbs))
}

func TestFeeTxHandler_VerifyInterMiniBlocks(t *testing.T) {
	t.Parallel()

	addr := &mock.SpecialAddressHandlerMock{}
	th, err := NewFeeTxHandler(
		addr,
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)

	err = th.VerifyInterMiniBlocks(nil)
	assert.Nil(t, err)

	currTxFee := big.NewInt(50)
	th.AddProcessedUTx(&feeTx.FeeTx{Value: currTxFee})

	err = th.VerifyInterMiniBlocks(nil)
	assert.Equal(t, process.ErrTxsFeesNotFound, err)

	badValue := big.NewInt(100)
	th.AddTxFeeFromBlock(&feeTx.FeeTx{Value: badValue})

	err = th.VerifyInterMiniBlocks(nil)
	assert.Equal(t, process.ErrTotalTxsFeesDoNotMatch, err)

	th.CleanProcessedUTxs()

	currTxFee = big.NewInt(50)
	halfCurrTxFee := big.NewInt(25)
	th.AddProcessedUTx(&feeTx.FeeTx{Value: currTxFee})
	th.AddTxFeeFromBlock(&feeTx.FeeTx{Value: halfCurrTxFee})

	err = th.VerifyInterMiniBlocks(nil)
	assert.Equal(t, process.ErrTxsFeesNotFound, err)

	th.CleanProcessedUTxs()

	currTxFee = big.NewInt(50)
	th.AddProcessedUTx(&feeTx.FeeTx{Value: currTxFee})
	th.AddTxFeeFromBlock(&feeTx.FeeTx{Value: big.NewInt(5), RcvAddr: addr.ElrondCommunityAddress()})
	th.AddTxFeeFromBlock(&feeTx.FeeTx{Value: big.NewInt(20), RcvAddr: addr.LeaderAddress()})
}
