package metachain

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRewardsCreator_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	args.ShardCoordinator = nil

	rwd, err := NewRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilShardCoordinator, err)
}

func TestNewRewardsCreator_NilPubkeyConverter(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	args.PubkeyConverter = nil

	rwd, err := NewRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilPubkeyConverter, err)
}

func TestNewRewardsCreator_NilRewardsStorage(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	args.RewardsStorage = nil

	rwd, err := NewRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilStorage, err)
}

func TestNewRewardsCreator_NilMiniBlockStorage(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	args.MiniBlockStorage = nil

	rwd, err := NewRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilStorage, err)
}

func TestNewRewardsCreator_NilHasher(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	args.Hasher = nil

	rwd, err := NewRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilHasher, err)
}

func TestNewRewardsCreator_NilMarshalizer(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	args.Marshalizer = nil

	rwd, err := NewRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilMarshalizer, err)
}

func TestNewRewardsCreator_EmptyProtocolSustainabilityAddress(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	args.ProtocolSustainabilityAddress = ""

	rwd, err := NewRewardsCreator(args)
	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilProtocolSustainabilityAddress, err)
}

func TestNewRewardsCreator_InvalidProtocolSustainabilityAddress(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	args.ProtocolSustainabilityAddress = "xyz" // not a hex string

	rwd, err := NewRewardsCreator(args)
	assert.True(t, check.IfNil(rwd))
	assert.NotNil(t, err)
}

func TestNewRewardsCreator_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	rwd, err := NewRewardsCreator(args)

	assert.False(t, check.IfNil(rwd))
	assert.Nil(t, err)
}

func TestRewardsCreator_CreateRewardsMiniBlocks(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	rwd, err := NewRewardsCreator(args)
	require.Nil(t, err)

	mb := &block.MetaBlock{
		EpochStart:     getDefaultEpochStart(),
		DevFeesInEpoch: big.NewInt(0),
	}
	valInfo := make(map[uint32][]*state.ValidatorInfo)
	valInfo[0] = []*state.ValidatorInfo{
		{
			PublicKey:       []byte("pubkey"),
			ShardId:         0,
			AccumulatedFees: big.NewInt(100),
		},
	}
	bdy, err := rwd.CreateRewardsMiniBlocks(mb, valInfo, &mb.EpochStart.Economics)
	assert.Nil(t, err)
	assert.NotNil(t, bdy)
}

func TestRewardsCreator_VerifyRewardsMiniBlocksHashDoesNotMatch(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	rwd, _ := NewRewardsCreator(args)

	bdy := block.MiniBlock{
		TxHashes:        [][]byte{},
		ReceiverShardID: 0,
		SenderShardID:   core.MetachainShardId,
		Type:            block.RewardsBlock,
	}
	mbh := block.MiniBlockHeader{
		Hash:            nil,
		SenderShardID:   core.MetachainShardId,
		ReceiverShardID: 0,
		TxCount:         1,
		Type:            block.RewardsBlock,
	}
	mbHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &mock.HasherMock{}, bdy)
	mbh.Hash = mbHash

	mb := &block.MetaBlock{
		EpochStart: getDefaultEpochStart(),
		MiniBlockHeaders: []block.MiniBlockHeader{
			mbh,
		},
		DevFeesInEpoch: big.NewInt(0),
	}
	valInfo := make(map[uint32][]*state.ValidatorInfo)
	valInfo[0] = []*state.ValidatorInfo{
		{
			PublicKey:       []byte("pubkey"),
			ShardId:         0,
			AccumulatedFees: big.NewInt(100),
		},
	}

	err := rwd.VerifyRewardsMiniBlocks(mb, valInfo, &mb.EpochStart.Economics)
	assert.Equal(t, epochStart.ErrRewardMiniBlockHashDoesNotMatch, err)
}

func TestRewardsCreator_VerifyRewardsMiniBlocksRewardsMbNumDoesNotMatch(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	rwd, _ := NewRewardsCreator(args)
	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Value:   big.NewInt(100),
		RcvAddr: []byte{},
		Epoch:   0,
	}
	rwdTxHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &mock.HasherMock{}, rwdTx)

	mb := &block.MetaBlock{
		EpochStart:     getDefaultEpochStart(),
		DevFeesInEpoch: big.NewInt(0),
	}
	protocolSustainabilityRewardTx := rewardTx.RewardTx{
		Round:   0,
		Value:   big.NewInt(50),
		RcvAddr: []byte{17},
		Epoch:   0,
	}
	mb.EpochStart.Economics.RewardsForProtocolSustainability.Set(protocolSustainabilityRewardTx.Value)
	mb.EpochStart.Economics.TotalToDistribute.Set(big.NewInt(0).Add(rwdTx.Value, protocolSustainabilityRewardTx.Value))
	commRwdTxHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &mock.HasherMock{}, protocolSustainabilityRewardTx)

	bdy := block.MiniBlock{
		TxHashes:        [][]byte{commRwdTxHash, rwdTxHash},
		ReceiverShardID: 0,
		SenderShardID:   core.MetachainShardId,
		Type:            block.RewardsBlock,
	}

	mbh := block.MiniBlockHeader{
		Hash:            nil,
		SenderShardID:   core.MetachainShardId,
		ReceiverShardID: 0,
		TxCount:         2,
		Type:            block.RewardsBlock,
	}
	mbHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &mock.HasherMock{}, bdy)
	mbh.Hash = mbHash

	mb.MiniBlockHeaders = []block.MiniBlockHeader{mbh, mbh}
	valInfo := make(map[uint32][]*state.ValidatorInfo)
	valInfo[0] = []*state.ValidatorInfo{
		{
			PublicKey:       []byte("pubkey"),
			ShardId:         0,
			AccumulatedFees: big.NewInt(100),
			LeaderSuccess:   1,
		},
	}

	err := rwd.VerifyRewardsMiniBlocks(mb, valInfo, &mb.EpochStart.Economics)
	assert.Equal(t, epochStart.ErrRewardMiniBlocksNumDoesNotMatch, err)
}

func TestRewardsCreator_adjustProtocolSustainabilityRewardsPositiveValue(t *testing.T) {
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
	mbSlice := createDefaultMiniBlocksSlice()
	_ = rwd.addProtocolRewardToMiniBlocks(protRwTx, mbSlice, protRwShard)

	dust := big.NewInt(1000)
	rwd1 := rewardsCreator{
		baseRewardsCreator: rwd,
	}
	rwd1.adjustProtocolSustainabilityRewards(protRwTx, dust)
	require.Zero(t, protRwTx.Value.Cmp(big.NewInt(0).Add(dust, initialProtRewardValue)))
	setProtValue := rwd.GetProtocolSustainabilityRewards()
	require.Zero(t, protRwTx.Value.Cmp(setProtValue))
}

func TestRewardsCreator_adjustProtocolSustainabilityRewardsNegValueShouldWork(t *testing.T) {
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
	mbSlice := createDefaultMiniBlocksSlice()
	_ = rwd.addProtocolRewardToMiniBlocks(protRwTx, mbSlice, protRwShard)

	rwd1 := rewardsCreator{
		baseRewardsCreator: rwd,
	}

	dust := big.NewInt(-10)
	rwd1.adjustProtocolSustainabilityRewards(protRwTx, dust)
	expected := big.NewInt(0).Add(dust, initialProtRewardValue).String()
	assert.Equal(t, expected, protRwTx.Value.String())
	setProtValue := rwd.GetProtocolSustainabilityRewards()
	require.Zero(t, protRwTx.Value.Cmp(setProtValue))
}

func TestRewardsCreator_adjustProtocolSustainabilityRewardsInitialNegativeValue(t *testing.T) {
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
	mbSlice := createDefaultMiniBlocksSlice()
	_ = rwd.addProtocolRewardToMiniBlocks(protRwTx, mbSlice, protRwShard)

	rwd1 := rewardsCreator{
		baseRewardsCreator: rwd,
	}

	dust := big.NewInt(0)
	rwd1.adjustProtocolSustainabilityRewards(protRwTx, dust)
	require.Zero(t, protRwTx.Value.Cmp(big.NewInt(0)))
	setProtValue := rwd.GetProtocolSustainabilityRewards()
	require.Zero(t, protRwTx.Value.Cmp(setProtValue))
}

func TestRewardsCreator_VerifyRewardsMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	rwd, _ := NewRewardsCreator(args)
	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Value:   big.NewInt(100),
		RcvAddr: []byte{},
		Epoch:   0,
	}
	rwdTxHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &mock.HasherMock{}, rwdTx)

	protocolSustainabilityRewardTx := rewardTx.RewardTx{
		Round:   0,
		Value:   big.NewInt(50),
		RcvAddr: []byte{17},
		Epoch:   0,
	}
	commRwdTxHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &mock.HasherMock{}, protocolSustainabilityRewardTx)

	bdy := block.MiniBlock{
		TxHashes:        [][]byte{commRwdTxHash, rwdTxHash},
		ReceiverShardID: 0,
		SenderShardID:   core.MetachainShardId,
		Type:            block.RewardsBlock,
	}
	mbh := block.MiniBlockHeader{
		Hash:            nil,
		SenderShardID:   core.MetachainShardId,
		ReceiverShardID: 0,
		TxCount:         2,
		Type:            block.RewardsBlock,
	}
	mbHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &mock.HasherMock{}, bdy)
	mbh.Hash = mbHash

	mb := &block.MetaBlock{
		EpochStart: getDefaultEpochStart(),
		MiniBlockHeaders: []block.MiniBlockHeader{
			mbh,
		},
		DevFeesInEpoch: big.NewInt(0),
	}
	mb.EpochStart.Economics.RewardsForProtocolSustainability.Set(protocolSustainabilityRewardTx.Value)
	mb.EpochStart.Economics.TotalToDistribute.Set(big.NewInt(0).Add(rwdTx.Value, protocolSustainabilityRewardTx.Value))

	valInfo := make(map[uint32][]*state.ValidatorInfo)
	valInfo[0] = []*state.ValidatorInfo{
		{
			PublicKey:       []byte("pubkey"),
			ShardId:         0,
			AccumulatedFees: big.NewInt(100),
			LeaderSuccess:   1,
		},
	}

	err := rwd.VerifyRewardsMiniBlocks(mb, valInfo, &mb.EpochStart.Economics)
	assert.Nil(t, err)
}

func TestRewardsCreator_VerifyRewardsMiniBlocksShouldWorkEvenIfNotAllShardsHaveRewards(t *testing.T) {
	t.Parallel()

	receivedShardID := uint32(5)
	shardCoordinator := &mock.ShardCoordinatorStub{
		ComputeIdCalled: func(address []byte) uint32 {
			return receivedShardID
		},
		NumberOfShardsCalled: func() uint32 {
			return receivedShardID + 1
		}}
	args := getRewardsArguments()
	args.ShardCoordinator = shardCoordinator
	rwd, _ := NewRewardsCreator(args)
	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Value:   big.NewInt(100),
		RcvAddr: []byte{},
		Epoch:   0,
	}
	rwdTxHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &mock.HasherMock{}, rwdTx)

	protocolSustainabilityRewardTx := rewardTx.RewardTx{
		Round:   0,
		Value:   big.NewInt(50),
		RcvAddr: []byte{17},
		Epoch:   0,
	}
	commRwdTxHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &mock.HasherMock{}, protocolSustainabilityRewardTx)

	bdy := block.MiniBlock{
		TxHashes:        [][]byte{commRwdTxHash, rwdTxHash},
		ReceiverShardID: receivedShardID,
		SenderShardID:   core.MetachainShardId,
		Type:            block.RewardsBlock,
	}
	mbh := block.MiniBlockHeader{
		Hash:            nil,
		SenderShardID:   core.MetachainShardId,
		ReceiverShardID: receivedShardID,
		TxCount:         2,
		Type:            block.RewardsBlock,
	}
	mbHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &mock.HasherMock{}, bdy)
	mbh.Hash = mbHash

	mb := &block.MetaBlock{
		EpochStart: getDefaultEpochStart(),
		MiniBlockHeaders: []block.MiniBlockHeader{
			mbh,
		},
		DevFeesInEpoch: big.NewInt(0),
	}
	mb.EpochStart.Economics.RewardsForProtocolSustainability.Set(protocolSustainabilityRewardTx.Value)
	mb.EpochStart.Economics.TotalToDistribute.Set(big.NewInt(0).Add(rwdTx.Value, protocolSustainabilityRewardTx.Value))

	valInfo := make(map[uint32][]*state.ValidatorInfo)
	valInfo[0] = []*state.ValidatorInfo{
		{
			PublicKey:       []byte("pubkey"),
			ShardId:         receivedShardID,
			AccumulatedFees: big.NewInt(100),
			LeaderSuccess:   1,
		},
	}

	err := rwd.VerifyRewardsMiniBlocks(mb, valInfo, &mb.EpochStart.Economics)
	assert.Nil(t, err)
}

func TestRewardsCreator_CreateMarshalizedData(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	rwd, _ := NewRewardsCreator(args)

	mb := &block.MetaBlock{
		EpochStart:     getDefaultEpochStart(),
		DevFeesInEpoch: big.NewInt(0),
	}
	valInfo := make(map[uint32][]*state.ValidatorInfo)
	valInfo[0] = []*state.ValidatorInfo{
		{
			PublicKey:       []byte("pubkey"),
			ShardId:         0,
			AccumulatedFees: big.NewInt(100),
		},
	}
	_, _ = rwd.CreateRewardsMiniBlocks(mb, valInfo, &mb.EpochStart.Economics)

	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Value:   big.NewInt(100),
		RcvAddr: []byte{},
		Epoch:   0,
	}
	rwdTxHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &mock.HasherMock{}, rwdTx)

	bdy := block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				ReceiverShardID: 0,
				Type:            block.RewardsBlock,
				TxHashes:        [][]byte{rwdTxHash},
			},
		},
	}
	res := rwd.CreateMarshalizedData(&bdy)

	assert.NotNil(t, res)
}

func TestRewardsCreator_SaveTxBlockToStorage(t *testing.T) {
	t.Parallel()

	putRwdTxWasCalled := false
	putMbWasCalled := false

	args := getRewardsArguments()
	args.RewardsStorage = &testscommon.StorerStub{
		PutCalled: func(_, _ []byte) error {
			putRwdTxWasCalled = true
			return nil
		},
	}
	args.MiniBlockStorage = &testscommon.StorerStub{
		PutCalled: func(_, _ []byte) error {
			putMbWasCalled = true
			return nil
		},
	}
	rwd, _ := NewRewardsCreator(args)

	mb := &block.MetaBlock{
		EpochStart:     getDefaultEpochStart(),
		DevFeesInEpoch: big.NewInt(0),
	}
	valInfo := make(map[uint32][]*state.ValidatorInfo)
	valInfo[0] = []*state.ValidatorInfo{
		{
			PublicKey:       []byte("pubkey"),
			ShardId:         0,
			AccumulatedFees: big.NewInt(100),
			LeaderSuccess:   1,
		},
	}
	_, _ = rwd.CreateRewardsMiniBlocks(mb, valInfo, &mb.EpochStart.Economics)

	mb2 := block.MetaBlock{
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Type: block.RewardsBlock,
			},
		},
	}
	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Value:   big.NewInt(100),
		RcvAddr: []byte{},
		Epoch:   0,
	}
	rwdTxHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &mock.HasherMock{}, rwdTx)
	bdy := block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				ReceiverShardID: 0,
				SenderShardID:   core.MetachainShardId,
				Type:            block.RewardsBlock,
				TxHashes:        [][]byte{rwdTxHash},
			},
		},
	}
	rwd.SaveTxBlockToStorage(&mb2, &bdy)

	assert.True(t, putRwdTxWasCalled)
	assert.True(t, putMbWasCalled)
}

func TestRewardsCreator_addValidatorRewardsToMiniBlocks(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	rwdc, _ := NewRewardsCreator(args)

	mb := &block.MetaBlock{
		EpochStart: getDefaultEpochStart(),
	}

	miniBlocks := make(block.MiniBlockSlice, rwdc.shardCoordinator.NumberOfShards())
	miniBlocks[0] = &block.MiniBlock{}
	miniBlocks[0].SenderShardID = core.MetachainShardId
	miniBlocks[0].ReceiverShardID = 0
	miniBlocks[0].Type = block.RewardsBlock
	miniBlocks[0].TxHashes = make([][]byte, 0)

	cloneMb := &(*miniBlocks[0]) //nolint
	cloneMb.TxHashes = make([][]byte, 0)
	expectedRwdTx := &rewardTx.RewardTx{
		Round:   0,
		Value:   big.NewInt(100),
		RcvAddr: []byte("pubkey"),
		Epoch:   0,
	}
	expectedRwdTxHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &mock.HasherMock{}, expectedRwdTx)
	cloneMb.TxHashes = append(cloneMb.TxHashes, expectedRwdTxHash)

	valInfo := make(map[uint32][]*state.ValidatorInfo)
	valInfo[0] = []*state.ValidatorInfo{
		{
			PublicKey:       []byte("pubkey"),
			ShardId:         0,
			AccumulatedFees: big.NewInt(100),
			LeaderSuccess:   1,
		},
	}

	rwdc.fillBaseRewardsPerBlockPerNode(mb.EpochStart.Economics.RewardsPerBlock)
	err := rwdc.addValidatorRewardsToMiniBlocks(valInfo, mb, miniBlocks, &rewardTx.RewardTx{})
	assert.Nil(t, err)
	assert.Equal(t, cloneMb, miniBlocks[0])
}

func TestRewardsCreator_ProtocolRewardsForValidatorFromMultipleShards(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	args.NodesConfigProvider = &mock.NodesCoordinatorStub{
		ConsensusGroupSizeCalled: func(shardID uint32) int {
			if shardID == core.MetachainShardId {
				return 400
			}
			return 63
		},
	}
	rwdc, _ := NewRewardsCreator(args)

	mb := &block.MetaBlock{
		EpochStart: getDefaultEpochStart(),
	}

	pubkey := "pubkey"
	valInfo := make(map[uint32][]*state.ValidatorInfo)
	valInfo[0] = []*state.ValidatorInfo{
		{
			RewardAddress:              []byte(pubkey),
			ShardId:                    0,
			AccumulatedFees:            big.NewInt(100),
			NumSelectedInSuccessBlocks: 100,
			LeaderSuccess:              1,
		},
	}
	valInfo[core.MetachainShardId] = []*state.ValidatorInfo{
		{
			RewardAddress:              []byte(pubkey),
			ShardId:                    core.MetachainShardId,
			AccumulatedFees:            big.NewInt(100),
			NumSelectedInSuccessBlocks: 200,
			LeaderSuccess:              1,
		},
	}

	rwdc.fillBaseRewardsPerBlockPerNode(mb.EpochStart.Economics.RewardsPerBlock)
	rwdInfoData := rwdc.computeValidatorInfoPerRewardAddress(valInfo, &rewardTx.RewardTx{}, 0)
	assert.Equal(t, 1, len(rwdInfoData))
	rwdInfo := rwdInfoData[pubkey]
	assert.Equal(t, rwdInfo.address, pubkey)

	assert.Equal(t, rwdInfo.accumulatedFees.Cmp(big.NewInt(200)), 0)
	protocolRewards := uint64(valInfo[0][0].NumSelectedInSuccessBlocks) * (mb.EpochStart.Economics.RewardsPerBlock.Uint64() / uint64(args.NodesConfigProvider.ConsensusGroupSize(0)))
	protocolRewards += uint64(valInfo[core.MetachainShardId][0].NumSelectedInSuccessBlocks) * (mb.EpochStart.Economics.RewardsPerBlock.Uint64() / uint64(args.NodesConfigProvider.ConsensusGroupSize(core.MetachainShardId)))
	assert.Equal(t, rwdInfo.rewardsFromProtocol.Uint64(), protocolRewards)
}

func TestRewardsCreator_CreateProtocolSustainabilityRewardTransaction(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	rwdc, _ := NewRewardsCreator(args)
	mb := &block.MetaBlock{
		EpochStart: getDefaultEpochStart(),
	}
	expectedRewardTx := &rewardTx.RewardTx{
		Round:   0,
		Value:   big.NewInt(50),
		RcvAddr: []byte{17},
		Epoch:   0,
	}

	rwdTx, _, err := rwdc.createProtocolSustainabilityRewardTransaction(mb, &mb.EpochStart.Economics)
	assert.Equal(t, expectedRewardTx, rwdTx)
	assert.Nil(t, err)
}

func TestRewardsCreator_AddProtocolSustainabilityRewardToMiniBlocks(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	rwdc, _ := NewRewardsCreator(args)
	metaBlk := &block.MetaBlock{
		EpochStart:     getDefaultEpochStart(),
		DevFeesInEpoch: big.NewInt(0),
	}

	miniBlocks := make(block.MiniBlockSlice, rwdc.shardCoordinator.NumberOfShards())
	miniBlocks[0] = &block.MiniBlock{}
	miniBlocks[0].SenderShardID = core.MetachainShardId
	miniBlocks[0].ReceiverShardID = 0
	miniBlocks[0].Type = block.RewardsBlock
	miniBlocks[0].TxHashes = make([][]byte, 0)

	cloneMb := &(*miniBlocks[0]) //nolint
	cloneMb.TxHashes = make([][]byte, 0)
	expectedRewardTx := &rewardTx.RewardTx{
		Round:   0,
		Value:   big.NewInt(50),
		RcvAddr: []byte{17},
		Epoch:   0,
	}
	expectedRwdTxHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &mock.HasherMock{}, expectedRewardTx)
	cloneMb.TxHashes = append(cloneMb.TxHashes, expectedRwdTxHash)
	metaBlk.EpochStart.Economics.RewardsForProtocolSustainability.Set(expectedRewardTx.Value)
	metaBlk.EpochStart.Economics.TotalToDistribute.Set(expectedRewardTx.Value)

	miniBlocks, err := rwdc.CreateRewardsMiniBlocks(metaBlk, make(map[uint32][]*state.ValidatorInfo), &metaBlk.EpochStart.Economics)
	assert.Nil(t, err)
	assert.Equal(t, cloneMb, miniBlocks[0])
}

func TestRewardsCreator_ValidatorInfoWithMetaAddressAddedToProtocolSustainabilityReward(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	args.NodesConfigProvider = &mock.NodesCoordinatorStub{}
	args.ShardCoordinator, _ = sharding.NewMultiShardCoordinator(1, core.MetachainShardId)
	rwdc, _ := NewRewardsCreator(args)
	metaBlk := &block.MetaBlock{
		EpochStart:     getDefaultEpochStart(),
		DevFeesInEpoch: big.NewInt(0),
	}
	metaBlk.EpochStart.Economics.TotalToDistribute = big.NewInt(20250)
	valInfo := make(map[uint32][]*state.ValidatorInfo)
	valInfo[0] = []*state.ValidatorInfo{
		{
			RewardAddress:              vm.StakingSCAddress,
			ShardId:                    0,
			AccumulatedFees:            big.NewInt(100),
			NumSelectedInSuccessBlocks: 1,
			LeaderSuccess:              1,
		},
		{
			RewardAddress:              vm.FirstDelegationSCAddress,
			ShardId:                    0,
			AccumulatedFees:            big.NewInt(100),
			NumSelectedInSuccessBlocks: 1,
			LeaderSuccess:              1,
		},
	}

	acc, _ := args.UserAccountsDB.LoadAccount(vm.FirstDelegationSCAddress)
	userAcc, _ := acc.(state.UserAccountHandler)
	_ = userAcc.DataTrieTracker().SaveKeyValue([]byte(core.DelegationSystemSCKey), []byte(core.DelegationSystemSCKey))
	_ = args.UserAccountsDB.SaveAccount(userAcc)

	miniBlocks, err := rwdc.CreateRewardsMiniBlocks(metaBlk, valInfo, &metaBlk.EpochStart.Economics)
	assert.Nil(t, err)
	assert.Equal(t, len(miniBlocks), 2)
	assert.Equal(t, len(miniBlocks[0].TxHashes), 1)
	assert.Equal(t, len(miniBlocks[1].TxHashes), 1)

	expectedProtocolSustainabilityValue := big.NewInt(0).Add(metaBlk.EpochStart.Economics.RewardsForProtocolSustainability, metaBlk.EpochStart.Economics.RewardsPerBlock)
	expectedProtocolSustainabilityValue.Add(expectedProtocolSustainabilityValue, big.NewInt(100))
	protocolSustainabilityReward, err := rwdc.currTxs.GetTx(miniBlocks[0].TxHashes[0])
	assert.Nil(t, err)
	assert.True(t, expectedProtocolSustainabilityValue.Cmp(protocolSustainabilityReward.GetValue()) == 0)
}

func getDefaultEpochStart() block.EpochStart {
	return block.EpochStart{
		Economics: block.Economics{
			TotalSupply:                      big.NewInt(10000),
			TotalToDistribute:                big.NewInt(10000),
			TotalNewlyMinted:                 big.NewInt(10000),
			RewardsPerBlock:                  big.NewInt(10000),
			NodePrice:                        big.NewInt(10000),
			RewardsForProtocolSustainability: big.NewInt(50),
		},
	}
}

func getRewardsArguments() ArgsNewRewardsCreator {
	return ArgsNewRewardsCreator{
		BaseRewardsCreatorArgs: getBaseRewardsArguments(),
	}
}
