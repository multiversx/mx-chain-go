package metachain

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseRewardsCreator_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	args.ShardCoordinator = nil

	rwd, err := NewBaseRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilShardCoordinator, err)
}

func TestBaseRewardsCreator_NilPubkeyConverter(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	args.PubkeyConverter = nil

	rwd, err := NewBaseRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilPubkeyConverter, err)
}

func TestBaseRewardsCreator_NilRewardsStorage(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	args.RewardsStorage = nil

	rwd, err := NewBaseRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilStorage, err)
}

func TestBaseRewardsCreator_NilMiniBlockStorage(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	args.MiniBlockStorage = nil

	rwd, err := NewBaseRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilStorage, err)
}

func TestBaseRewardsCreator_NilHasher(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	args.Hasher = nil

	rwd, err := NewBaseRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilHasher, err)
}

func TestBaseRewardsCreator_NilMarshalizer(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	args.Marshalizer = nil

	rwd, err := NewBaseRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilMarshalizer, err)
}

func TestBaseRewardsCreator_EmptyProtocolSustainabilityAddress(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	args.ProtocolSustainabilityAddress = ""

	rwd, err := NewBaseRewardsCreator(args)
	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilProtocolSustainabilityAddress, err)
}

func TestBaseRewardsCreator_InvalidProtocolSustainabilityAddress(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	args.ProtocolSustainabilityAddress = "xyz" // not a hex string

	rwd, err := NewBaseRewardsCreator(args)
	assert.True(t, check.IfNil(rwd))
	assert.NotNil(t, err)
}

func TestBaseRewardsCreator_ProtocolSustainabilityAddressInMetachainShouldErr(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	var err error
	args.ShardCoordinator, err = sharding.NewMultiShardCoordinator(2, 0)
	// wrong configuration of staking system SC address (in metachain) as protocol sustainability address
	args.ProtocolSustainabilityAddress = hex.EncodeToString([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255})

	rwd, err := NewBaseRewardsCreator(args)
	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrProtocolSustainabilityAddressInMetachain, err)
}

func TestBaseRewardsCreator_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)

	assert.False(t, check.IfNil(rwd))
	assert.Nil(t, err)
}

func TestBaseRewardsCreator_GetLocalTxCache(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)
	require.Nil(t, err)

	txCache := rwd.GetLocalTxCache()
	require.False(t, check.IfNil(txCache))
}

func TestBaseRewardsCreator_GetProtocolSustainabilityRewards(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	// should return 0 as just initialized
	rewards := rwd.GetProtocolSustainabilityRewards()
	require.Zero(t, big.NewInt(0).Cmp(rewards))
}

func TestBaseRewardsCreator_adjustProtocolSustainabilityRewardsPositiveValue(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	initialProtRewardValue := big.NewInt(1000000)
	protRwAddr, _ := args.PubkeyConverter.Decode(args.ProtocolSustainabilityAddress)
	protRwTx := &rewardTx.RewardTx{
		Round:   100,
		Value:   big.NewInt(0).Set(initialProtRewardValue),
		RcvAddr: protRwAddr,
		Epoch:   1,
	}

	protRwShard := args.ShardCoordinator.ComputeId(protRwAddr)
	mbSlice := createDefaultMiniblocksSlice()
	err = rwd.addProtocolRewardToMiniBlocks(protRwTx, mbSlice, protRwShard)

	dust := big.NewInt(1000)
	rwd.adjustProtocolSustainabilityRewards(protRwTx, dust)
	require.Zero(t, protRwTx.Value.Cmp(big.NewInt(0).Add(dust, initialProtRewardValue)))
	setProtValue := rwd.GetProtocolSustainabilityRewards()
	require.Zero(t, protRwTx.Value.Cmp(setProtValue))
}

func TestBaseRewardsCreator_adjustProtocolSustainabilityRewardsNegValueNotAccepted(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	initialProtRewardValue := big.NewInt(10)
	protRwAddr, _ := args.PubkeyConverter.Decode(args.ProtocolSustainabilityAddress)
	protRwTx := &rewardTx.RewardTx{
		Round:   100,
		Value:   big.NewInt(0).Set(initialProtRewardValue),
		RcvAddr: protRwAddr,
		Epoch:   1,
	}

	protRwShard := args.ShardCoordinator.ComputeId(protRwAddr)
	mbSlice := createDefaultMiniblocksSlice()
	err = rwd.addProtocolRewardToMiniBlocks(protRwTx, mbSlice, protRwShard)

	dust := big.NewInt(-10)
	rwd.adjustProtocolSustainabilityRewards(protRwTx, dust)
	require.Zero(t, protRwTx.Value.Cmp(initialProtRewardValue))
	setProtValue := rwd.GetProtocolSustainabilityRewards()
	require.Zero(t, protRwTx.Value.Cmp(setProtValue))
}

func TestBaseRewardsCreator_adjustProtocolSustainabilityRewardsInitialNegativeValue(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	initialProtRewardValue := big.NewInt(-100)
	protRwAddr, _ := args.PubkeyConverter.Decode(args.ProtocolSustainabilityAddress)
	protRwTx := &rewardTx.RewardTx{
		Round:   100,
		Value:   big.NewInt(0).Set(initialProtRewardValue),
		RcvAddr: protRwAddr,
		Epoch:   1,
	}

	protRwShard := args.ShardCoordinator.ComputeId(protRwAddr)
	mbSlice := createDefaultMiniblocksSlice()
	err = rwd.addProtocolRewardToMiniBlocks(protRwTx, mbSlice, protRwShard)

	dust := big.NewInt(0)
	rwd.adjustProtocolSustainabilityRewards(protRwTx, dust)
	require.Zero(t, protRwTx.Value.Cmp(big.NewInt(0)))
	setProtValue := rwd.GetProtocolSustainabilityRewards()
	require.Zero(t, protRwTx.Value.Cmp(setProtValue))
}

func TestBaseRewardsCreator_addProtocolRewardToMiniblocks(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	initialProtRewardValue := big.NewInt(-100)
	protRwAddr, _ := args.PubkeyConverter.Decode(args.ProtocolSustainabilityAddress)
	protRwTx := &rewardTx.RewardTx{
		Round:   100,
		Value:   big.NewInt(0).Set(initialProtRewardValue),
		RcvAddr: protRwAddr,
		Epoch:   1,
	}

	marshalled, err := args.Marshalizer.Marshal(protRwTx)
	require.Nil(t, err)

	protRwTxHash := args.Hasher.Compute(string(marshalled))

	protRwShard := args.ShardCoordinator.ComputeId(protRwAddr)
	mbSlice := createDefaultMiniblocksSlice()
	err = rwd.addProtocolRewardToMiniBlocks(protRwTx, mbSlice, protRwShard)

	found := false
	for _, mb := range mbSlice {
		for _, txHash := range mb.TxHashes {
			if bytes.Compare(txHash, protRwTxHash) == 0 {
				found = true
			}
		}
	}
	require.True(t, found)
}

func TestBaseRewardsCreator_CreateMarshalizedDataNilMiniblocksEmptyMap(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	result := rwd.CreateMarshalizedData(nil)
	require.Equal(t, 0, len(result))
}

func TestBaseRewardsCreator_CreateMarshalizedDataEmptyMiniblocksEmptyMap(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	result := rwd.CreateMarshalizedData(&block.Body{})
	require.Equal(t, 0, len(result))
}

func TestBaseRewardsCreator_CreateMarshalizedDataOnlyRewardsMiniblocksGetMarshalized(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	dummyMiniBlock := createDummyRewardTxMiniblock(rwd)

	miniBlockTypes := []block.Type{
		block.TxBlock,
		block.StateBlock,
		block.PeerBlock,
		block.SmartContractResultBlock,
		block.InvalidBlock,
		block.ReceiptBlock,
	}

	for _, mbType := range miniBlockTypes {
		dummyMiniBlock.Type = mbType
		result := rwd.CreateMarshalizedData(&block.Body{
			MiniBlocks: block.MiniBlockSlice{
				dummyMiniBlock,
			},
		})
		require.Equal(t, 0, len(result))
	}

	dummyMiniBlock.Type = block.RewardsBlock
	result := rwd.CreateMarshalizedData(&block.Body{
		MiniBlocks: block.MiniBlockSlice{
			dummyMiniBlock,
		},
	})
	require.Greater(t, len(result), 0)

	readRwTx := &rewardTx.RewardTx{}
	for _, data := range result {
		for _, tx := range data {
			err = args.Marshalizer.Unmarshal(readRwTx, tx)
			require.Nil(t, err)
			expectedTx, err := rwd.currTxs.GetTx(dummyMiniBlock.TxHashes[0])
			require.Nil(t, err)
			require.Equal(t, expectedTx, readRwTx)
		}
	}
}

func TestBaseRewardsCreator_CreateMarshalizedDataWrongSenderNotIncluded(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	dummyMiniBlock := createDummyRewardTxMiniblock(rwd)
	dummyMiniBlock.Type = block.RewardsBlock
	dummyMiniBlock.SenderShardID = args.ShardCoordinator.SelfId() + 1
	result := rwd.CreateMarshalizedData(&block.Body{
		MiniBlocks: block.MiniBlockSlice{
			dummyMiniBlock,
		},
	})
	require.Equal(t, 0, len(result))
}

func TestBaseRewardsCreator_CreateMarshalizedDataNotFoundTxHashIgnored(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	dummyMiniBlock := createDummyRewardTxMiniblock(rwd)
	dummyMiniBlock.Type = block.RewardsBlock
	dummyMiniBlock.TxHashes = [][]byte{[]byte("not found txHash")}
	result := rwd.CreateMarshalizedData(&block.Body{
		MiniBlocks: block.MiniBlockSlice{
			dummyMiniBlock,
		},
	})
	require.Equal(t, 0, len(result))
}

func TestBaseRewardsCreator_GetRewardsTxsNonRewardsMiniBlocksGetIgnored(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	dummyMb := createDummyRewardTxMiniblock(rwd)
	dummyMb.Type = block.StateBlock
	result := rwd.GetRewardsTxs(&block.Body{MiniBlocks: block.MiniBlockSlice{dummyMb}})
	require.Equal(t, 0, len(result))
}

func TestBaseRewardsCreator_GetRewardsTxsRewardMiniBlockNotFoundTxIgnored(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)
	rwTxHash := []byte("not found")
	dummyMb := createDummyRewardTxMiniblock(rwd)
	dummyMb.TxHashes = [][]byte{rwTxHash}

	result := rwd.GetRewardsTxs(&block.Body{MiniBlocks: block.MiniBlockSlice{dummyMb}})
	require.Equal(t, 0, len(result))
}

func TestBaseRewardsCreator_GetRewardsTxsRewardMiniBlockOK(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)
	rwTx := &rewardTx.RewardTx{
		Value: big.NewInt(1000),
	}
	rwTxHash := []byte("rwTxHash")
	rwd.currTxs.AddTx(rwTxHash, rwTx)

	dummyMb := createDummyRewardTxMiniblock(rwd)
	dummyMb.TxHashes = [][]byte{rwTxHash}

	result := rwd.GetRewardsTxs(&block.Body{MiniBlocks: block.MiniBlockSlice{dummyMb}})
	require.Equal(t, 1, len(result))
	for _, tx := range result {
		require.Equal(t, rwTx, tx)
	}
}

func getBaseRewardsArguments() BaseRewardsCreatorArgs {
	hasher := sha256.Sha256{}
	marshalizer := &marshal.GogoProtoMarshalizer{}
	trieFactoryManager, _ := trie.NewTrieStorageManagerWithoutPruning(createMemUnit())
	userAccountsDB := createAccountsDB(hasher, marshalizer, factory.NewAccountCreator(), trieFactoryManager)
	return BaseRewardsCreatorArgs{
		ShardCoordinator:              mock.NewMultiShardsCoordinatorMock(2),
		PubkeyConverter:               mock.NewPubkeyConverterMock(32),
		RewardsStorage:                mock.NewStorerMock(),
		MiniBlockStorage:              mock.NewStorerMock(),
		Hasher:                        &mock.HasherMock{},
		Marshalizer:                   &mock.MarshalizerMock{},
		DataPool:                      testscommon.NewPoolsHolderStub(),
		ProtocolSustainabilityAddress: "11", // string hex => 17 decimal
		NodesConfigProvider:           &mock.NodesCoordinatorStub{},
		UserAccountsDB:                userAccountsDB,
	}
}

func createDummyRewardTxMiniblock(rwd *baseRewardsCreator) *block.MiniBlock {
	dummyTx := &rewardTx.RewardTx{}
	dummyTxHash := []byte("rwdTxHash")
	rwd.currTxs.AddTx(dummyTxHash, dummyTx)

	return &block.MiniBlock{
		Type:            block.RewardsBlock,
		SenderShardID:   rwd.shardCoordinator.SelfId(),
		ReceiverShardID: rwd.shardCoordinator.SelfId() + 1,
		TxHashes:        [][]byte{dummyTxHash},
	}
}

func createDefaultMiniblocksSlice() block.MiniBlockSlice {
	return block.MiniBlockSlice{
		&block.MiniBlock{
			TxHashes:        nil,
			ReceiverShardID: 0,
			SenderShardID:   core.MetachainShardId,
			Type:            block.RewardsBlock,
			Reserved:        nil,
		},
		&block.MiniBlock{
			TxHashes:        nil,
			ReceiverShardID: 1,
			SenderShardID:   core.MetachainShardId,
			Type:            block.RewardsBlock,
			Reserved:        nil,
		},
	}
}
