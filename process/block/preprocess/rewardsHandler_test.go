package preprocess

import (
	"bytes"
	"math/big"
	"reflect"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
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

func TestRewardsHandler_AddIntermediateTransactions(t *testing.T) {
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

func TestRewardsHandler_ProcessTransactionFee(t *testing.T) {
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

func TestRewardsHandler_cleanCachedData(t *testing.T) {
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

func TestRewardsHandler_CreateRewardsFromFees(t *testing.T) {
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

func TestRewardsHandler_VerifyCreatedRewardsTxsRewardTxNotFound(t *testing.T) {
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
	_ = th.AddIntermediateTransactions([]data.TransactionHandler{&rewardTx.RewardTx{Value: big.NewInt(5), RcvAddr: addr.LeaderAddress()}})
	_ = th.AddIntermediateTransactions([]data.TransactionHandler{&rewardTx.RewardTx{Value: big.NewInt(5), RcvAddr: addr.ElrondCommunityAddress()}})
	_ = th.AddIntermediateTransactions([]data.TransactionHandler{&rewardTx.RewardTx{Value: big.NewInt(5), RcvAddr: addr.BurnAddress()}})
	err = th.verifyCreatedRewardsTxs()
	assert.Equal(t, process.ErrRewardTxNotFound, err)
}

func TestRewardsHandler_VerifyCreatedRewardsTxsTotalTxsFeesDoNotMatch(t *testing.T) {
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
	assert.Equal(t, process.ErrRewardTxsMismatchCreatedReceived, err)
}

func TestRewardsHandler_VerifyCreatedRewardsTxsOK(t *testing.T) {
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

func TestRewardsHandler_CreateAllInterMiniBlocksOK(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
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

func TestRewardsHandler_GetAllCurrentFinishedTxs(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	tdp := initDataPool()
	specialAddress := &mock.SpecialAddressHandlerMock{
		AdrConv:          &mock.AddressConverterMock{},
		ShardCoordinator: shardCoordinator,
	}

	consensusAddresses := []string{
		"1000000000000000000000000000000000000000000000000000000000000000",
		"2000000000000000000000000000000000000000000000000000000000000000",
	}

	specialAddress.SetConsensusData(consensusAddresses, 0, 0)

	th, err := NewRewardTxHandler(
		specialAddress,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		&mock.AddressConverterMock{},
		&mock.ChainStorerMock{},
		tdp.RewardTransactions(),
	)

	assert.Nil(t, err)
	assert.NotNil(t, th)

	txs := []data.TransactionHandler{
		&rewardTx.RewardTx{
			Round:   0,
			Epoch:   0,
			Value:   big.NewInt(1),
			RcvAddr: []byte(consensusAddresses[0]),
			ShardId: 0,
		},
		&rewardTx.RewardTx{
			Round:   0,
			Epoch:   0,
			Value:   big.NewInt(1),
			RcvAddr: []byte(consensusAddresses[1]),
			ShardId: 0,
		},
	}

	err = th.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	finishedTxs := th.GetAllCurrentFinishedTxs()
	assert.Equal(t, 2, len(txs))

	for _, ftx := range finishedTxs {
		found := false
		for _, tx := range txs {
			if reflect.DeepEqual(tx, ftx) {
				found = true
				break
			}
		}

		assert.True(t, found)
	}
}

func TestRewardsHandler_CreateMarshalizedDataShouldWork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	th, _ := NewRewardTxHandler(
		&mock.SpecialAddressHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AddressConverterMock{},
		&mock.ChainStorerMock{},
		tdp.RewardTransactions(),
	)

	txs := []data.TransactionHandler{
		&rewardTx.RewardTx{
			Round:   0,
			Epoch:   0,
			Value:   big.NewInt(1),
			RcvAddr: []byte("rcvr1"),
			ShardId: 0,
		},
		&rewardTx.RewardTx{
			Round:   0,
			Epoch:   0,
			Value:   big.NewInt(1),
			RcvAddr: []byte("rcvr2"),
			ShardId: 0,
		},
	}

	err := th.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	var expectedMarshalizedTxs [][]byte
	marshTx1, _ := th.marshalizer.Marshal(txs[0])
	marshTx2, _ := th.marshalizer.Marshal(txs[1])
	expectedMarshalizedTxs = append(expectedMarshalizedTxs, marshTx1, marshTx2)

	var txsHashes [][]byte
	tx1Hash, _ := core.CalculateHash(th.marshalizer, th.hasher, txs[0])
	tx2Hash, _ := core.CalculateHash(th.marshalizer, th.hasher, txs[1])
	txsHashes = append(txsHashes, tx1Hash, tx2Hash)

	res, err := th.CreateMarshalizedData(txsHashes)
	assert.Nil(t, err)
	assert.Equal(t, len(txs), len(res))
	assert.True(t, bytes.Equal(expectedMarshalizedTxs[0], res[0]))
	assert.True(t, bytes.Equal(expectedMarshalizedTxs[1], res[1]))
}

func TestRewardsHandler_CreateBlockStartedShouldCreateProtocolReward(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	th, _ := NewRewardTxHandler(
		&mock.SpecialAddressHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AddressConverterMock{},
		&mock.ChainStorerMock{},
		tdp.RewardTransactions(),
	)

	assert.Nil(t, th.protocolRewards)

	th.CreateBlockStarted()
	assert.NotNil(t, th.protocolRewards)
}

func TestRewardsHandler_SaveCurrentIntermediateTxToStorageShouldWork(t *testing.T) {
	t.Parallel()

	putWasCalled := false
	tdp := initDataPool()
	th, _ := NewRewardTxHandler(
		&mock.SpecialAddressHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AddressConverterMock{},
		&mock.ChainStorerMock{
			PutCalled: func(unitType dataRetriever.UnitType, key []byte, value []byte) error {
				putWasCalled = true
				return nil
			},
		},
		tdp.RewardTransactions(),
	)

	txs := []data.TransactionHandler{
		&rewardTx.RewardTx{
			Round:   0,
			Epoch:   0,
			Value:   big.NewInt(1),
			RcvAddr: []byte("rcvr1"),
			ShardId: 0,
		},
		&rewardTx.RewardTx{
			Round:   0,
			Epoch:   0,
			Value:   big.NewInt(1),
			RcvAddr: []byte("rcvr2"),
			ShardId: 0,
		},
	}

	err := th.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	err = th.SaveCurrentIntermediateTxToStorage()
	assert.Nil(t, err)
	assert.True(t, putWasCalled)
}
