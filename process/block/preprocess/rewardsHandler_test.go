package preprocess

import (
    "github.com/ElrondNetwork/elrond-go/data"
    "math/big"
    "testing"

    "github.com/ElrondNetwork/elrond-go/data/rewardTx"
    "github.com/ElrondNetwork/elrond-go/process"
    "github.com/ElrondNetwork/elrond-go/process/mock"
    "github.com/stretchr/testify/assert"
)

func TestNewRewardTxHandler_NilSpecialAddressShouldErr(t *testing.T) {
    t.Parallel()

    tdp := initDataPool()
    th, err := NewRewardTxHandler(
        nil,
        &mock.HasherMock{},
        &mock.MarshalizerMock{},
        mock.NewMultiShardsCoordinatorMock(3),
        &mock.AddressConverterMock{},
        &mock.ChainStorerMock{},
        tdp.RewardTransactions(),
    )

    assert.Nil(t, th)
    assert.Equal(t, process.ErrNilSpecialAddressHandler, err)
}

func TestNewRewardTxHandler_NilHasher(t *testing.T) {
    t.Parallel()

    tdp := initDataPool()
    th, err := NewRewardTxHandler(
        &mock.SpecialAddressHandlerMock{},
        nil,
        &mock.MarshalizerMock{},
        mock.NewMultiShardsCoordinatorMock(3),
        &mock.AddressConverterMock{},
        &mock.ChainStorerMock{},
        tdp.RewardTransactions(),
    )

    assert.Nil(t, th)
    assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewRewardTxHandler_NilMarshalizer(t *testing.T) {
    t.Parallel()

    tdp := initDataPool()
    th, err := NewRewardTxHandler(
        &mock.SpecialAddressHandlerMock{},
        &mock.HasherMock{},
        nil,
        mock.NewMultiShardsCoordinatorMock(3),
        &mock.AddressConverterMock{},
        &mock.ChainStorerMock{},
        tdp.RewardTransactions(),
    )

    assert.Nil(t, th)
    assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewRewardTxHandler_NilShardCoordinator(t *testing.T) {
    t.Parallel()

    tdp := initDataPool()
    th, err := NewRewardTxHandler(
        &mock.SpecialAddressHandlerMock{},
        &mock.HasherMock{},
        &mock.MarshalizerMock{},
        nil,
        &mock.AddressConverterMock{},
        &mock.ChainStorerMock{},
        tdp.RewardTransactions(),
    )

    assert.Nil(t, th)
    assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewRewardTxHandler_NilAddressConverter(t *testing.T) {
    t.Parallel()

    tdp := initDataPool()
    th, err := NewRewardTxHandler(
        &mock.SpecialAddressHandlerMock{},
        &mock.HasherMock{},
        &mock.MarshalizerMock{},
        mock.NewMultiShardsCoordinatorMock(3),
        nil,
        &mock.ChainStorerMock{},
        tdp.RewardTransactions(),
    )

    assert.Nil(t, th)
    assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewRewardTxHandler_NilChainStorer(t *testing.T) {
    t.Parallel()

    tdp := initDataPool()
    th, err := NewRewardTxHandler(
        &mock.SpecialAddressHandlerMock{},
        &mock.HasherMock{},
        &mock.MarshalizerMock{},
        mock.NewMultiShardsCoordinatorMock(3),
        &mock.AddressConverterMock{},
        nil,
        tdp.RewardTransactions(),
    )

    assert.Nil(t, th)
    assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewRewardTxHandler_NilRewardsPool(t *testing.T) {
    t.Parallel()

    th, err := NewRewardTxHandler(
        &mock.SpecialAddressHandlerMock{},
        &mock.HasherMock{},
        &mock.MarshalizerMock{},
        mock.NewMultiShardsCoordinatorMock(3),
        &mock.AddressConverterMock{},
        &mock.ChainStorerMock{},
        nil,
    )

    assert.Nil(t, th)
    assert.NotNil(t, process.ErrNilRewardTxDataPool, err)
}

func TestNewRewardTxHandler_ValsOk(t *testing.T) {
    t.Parallel()

    tdp := initDataPool()
    th, err := NewRewardTxHandler(
        &mock.SpecialAddressHandlerMock{},
        &mock.HasherMock{},
        &mock.MarshalizerMock{},
        mock.NewMultiShardsCoordinatorMock(3),
        &mock.AddressConverterMock{},
        &mock.ChainStorerMock{},
        tdp.RewardTransactions(),
    )

    assert.Nil(t, err)
    assert.NotNil(t, th)
}

func TestRewardTxHandlerAddIntermediateTransactions(t *testing.T) {
    t.Parallel()

    tdp := initDataPool()
    th, err := NewRewardTxHandler(
        &mock.SpecialAddressHandlerMock{},
        &mock.HasherMock{},
        &mock.MarshalizerMock{},
        mock.NewMultiShardsCoordinatorMock(3),
        &mock.AddressConverterMock{},
        &mock.ChainStorerMock{},
        tdp.RewardTransactions(),
    )

    assert.Nil(t, err)
    assert.NotNil(t, th)

    err = th.AddIntermediateTransactions(nil)
    assert.Nil(t, err)
}

func TestRewardTxHandlerProcessTransactionFee(t *testing.T) {
    t.Parallel()

    tdp := initDataPool()
    th, err := NewRewardTxHandler(
        &mock.SpecialAddressHandlerMock{},
        &mock.HasherMock{},
        &mock.MarshalizerMock{},
        mock.NewMultiShardsCoordinatorMock(3),
        &mock.AddressConverterMock{},
        &mock.ChainStorerMock{},
        tdp.RewardTransactions(),
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

func TestRewardTxHandlerCleanProcessedUTxs(t *testing.T) {
    t.Parallel()

    tdp := initDataPool()
    th, err := NewRewardTxHandler(
        &mock.SpecialAddressHandlerMock{},
        &mock.HasherMock{},
        &mock.MarshalizerMock{},
        mock.NewMultiShardsCoordinatorMock(3),
        &mock.AddressConverterMock{},
        &mock.ChainStorerMock{},
        tdp.RewardTransactions(),
    )

    assert.Nil(t, err)
    assert.NotNil(t, th)

    th.ProcessTransactionFee(big.NewInt(10))
    _ = th.AddIntermediateTransactions([]data.TransactionHandler{&rewardTx.RewardTx{}})
    assert.Equal(t, big.NewInt(10), th.accumulatedFees)
    assert.Equal(t, 1, len(th.rewardTxsForBlock))

    th.cleanCachedData()
    assert.Equal(t, big.NewInt(0), th.accumulatedFees)
    assert.Equal(t, 0, len(th.rewardTxsForBlock))
}

func TestRewardTxHandlerCreateAllUTxs(t *testing.T) {
    t.Parallel()

    tdp := initDataPool()
    th, err := NewRewardTxHandler(
        &mock.SpecialAddressHandlerMock{},
        &mock.HasherMock{},
        &mock.MarshalizerMock{},
        mock.NewMultiShardsCoordinatorMock(3),
        &mock.AddressConverterMock{},
        &mock.ChainStorerMock{},
        tdp.RewardTransactions(),
    )

    assert.Nil(t, err)
    assert.NotNil(t, th)

    txs := th.createRewardFromFees()
    assert.Equal(t, 0, len(txs))

    currTxFee := big.NewInt(50)
    th.ProcessTransactionFee(currTxFee)

    txs = th.createRewardFromFees()
    assert.Equal(t, 3, len(txs))

    totalSum := txs[0].GetValue().Uint64()
    totalSum += txs[1].GetValue().Uint64()
    totalSum += txs[2].GetValue().Uint64()

    assert.Equal(t, currTxFee.Uint64(), totalSum)
}

func TestRewardTxHandler_VerifyCreatedRewardsTxsRewardTxNotFound(t *testing.T) {
    t.Parallel()

    tdp := initDataPool()
    adrConv := &mock.AddressConverterMock{}
    shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
    addr := mock.NewSpecialAddressHandlerMock(adrConv, shardCoordinator)
    th, err := NewRewardTxHandler(
        addr,
        &mock.HasherMock{},
        &mock.MarshalizerMock{},
        shardCoordinator,
        adrConv,
        &mock.ChainStorerMock{},
        tdp.RewardTransactions(),
    )

    assert.Nil(t, err)
    assert.NotNil(t, th)

    err = th.verifyCreatedRewardsTxs()
    assert.Nil(t, err)

    currTxFee := big.NewInt(50)
    th.ProcessTransactionFee(currTxFee)
    _ = th.CreateAllInterMiniBlocks()
    err = th.verifyCreatedRewardsTxs()
    assert.Equal(t, process.ErrRewardTxNotFound, err)
}

func TestRewardTxHandler_VerifyCreatedRewardsTxsTotalTxsFeesDoNotMatch(t *testing.T) {
    t.Parallel()

    tdp := initDataPool()
    adrConv := &mock.AddressConverterMock{}
    shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
    addr := mock.NewSpecialAddressHandlerMock(adrConv, shardCoordinator)
    th, err := NewRewardTxHandler(
        addr,
        &mock.HasherMock{},
        &mock.MarshalizerMock{},
        shardCoordinator,
        adrConv,
        &mock.ChainStorerMock{},
        tdp.RewardTransactions(),
    )

    assert.Nil(t, err)
    assert.NotNil(t, th)

    err = th.verifyCreatedRewardsTxs()
    assert.Nil(t, err)

    currTxFee := big.NewInt(50)
    th.ProcessTransactionFee(currTxFee)
    extraVal := big.NewInt(100)
    _ = th.AddIntermediateTransactions([]data.TransactionHandler{&rewardTx.RewardTx{Value: big.NewInt(5), RcvAddr: addr.ElrondCommunityAddress()}})
    _ = th.AddIntermediateTransactions([]data.TransactionHandler{&rewardTx.RewardTx{Value: big.NewInt(25), RcvAddr: addr.LeaderAddress()}})
    _ = th.AddIntermediateTransactions([]data.TransactionHandler{&rewardTx.RewardTx{Value: big.NewInt(20), RcvAddr: addr.BurnAddress()}})
    _ = th.AddIntermediateTransactions([]data.TransactionHandler{&rewardTx.RewardTx{Value: extraVal, RcvAddr: addr.BurnAddress()}})
    _ = th.CreateAllInterMiniBlocks()
    err = th.verifyCreatedRewardsTxs()
    assert.Equal(t, process.ErrTotalTxsFeesDoNotMatch, err)
}

func TestRewardTxHandlerVerifyCreatedRewardsTxsOK(t *testing.T) {
    t.Parallel()

    tdp := initDataPool()
    adrConv := &mock.AddressConverterMock{}
    shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
    addr := mock.NewSpecialAddressHandlerMock(adrConv, shardCoordinator)
    th, err := NewRewardTxHandler(
        addr,
        &mock.HasherMock{},
        &mock.MarshalizerMock{},
        shardCoordinator,
        adrConv,
        &mock.ChainStorerMock{},
        tdp.RewardTransactions(),
    )

    assert.Nil(t, err)
    assert.NotNil(t, th)

    currTxFee := big.NewInt(50)
    th.ProcessTransactionFee(currTxFee)
    _ = th.AddIntermediateTransactions([]data.TransactionHandler{&rewardTx.RewardTx{Value: big.NewInt(5), RcvAddr: addr.ElrondCommunityAddress()}})
    _ = th.AddIntermediateTransactions([]data.TransactionHandler{&rewardTx.RewardTx{Value: big.NewInt(25), RcvAddr: addr.LeaderAddress()}})
    _ = th.AddIntermediateTransactions([]data.TransactionHandler{&rewardTx.RewardTx{Value: big.NewInt(20), RcvAddr: addr.BurnAddress()}})
    _ = th.CreateAllInterMiniBlocks()
    err = th.verifyCreatedRewardsTxs()
    assert.Nil(t, err)
}

func TestRewardTxHandlerCreateAllInterMiniBlocksOK(t *testing.T) {
    t.Parallel()

    shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
    tdp := initDataPool()
    th, err := NewRewardTxHandler(
        &mock.SpecialAddressHandlerMock{
            AdrConv:          &mock.AddressConverterMock{},
            ShardCoordinator: shardCoordinator},
        &mock.HasherMock{},
        &mock.MarshalizerMock{},
        shardCoordinator,
        &mock.AddressConverterMock{},
        &mock.ChainStorerMock{},
        tdp.RewardTransactions(),
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
