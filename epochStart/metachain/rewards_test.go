package metachain

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/stretchr/testify/assert"
)

func TestNewEpochStartRewardsCreator_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	args.ShardCoordinator = nil

	rwd, err := NewEpochStartRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilShardCoordinator, err)
}

func TestNewEpochStartRewardsCreator_NilAddressConverter(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	args.AddrConverter = nil

	rwd, err := NewEpochStartRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilAddressConverter, err)
}

func TestNewEpochStartRewardsCreator_NilRewardsStorage(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	args.RewardsStorage = nil

	rwd, err := NewEpochStartRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilStorage, err)
}

func TestNewEpochStartRewardsCreator_NilMiniBlockStorage(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	args.MiniBlockStorage = nil

	rwd, err := NewEpochStartRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilStorage, err)
}

func TestNewEpochStartRewardsCreator_NilHasher(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	args.Hasher = nil

	rwd, err := NewEpochStartRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilHasher, err)
}

func TestNewEpochStartRewardsCreator_NilMarshalizer(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	args.Marshalizer = nil

	rwd, err := NewEpochStartRewardsCreator(args)

	assert.True(t, check.IfNil(rwd))
	assert.Equal(t, epochStart.ErrNilMarshalizer, err)
}

func TestNewEpochStartRewardsCreator_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	rwd, err := NewEpochStartRewardsCreator(args)

	assert.False(t, check.IfNil(rwd))
	assert.Nil(t, err)
}

func TestRewardsCreator_CreateRewardsMiniBlocks(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	rwd, _ := NewEpochStartRewardsCreator(args)

	mb := &block.MetaBlock{
		EpochStart: block.EpochStart{
			Economics: block.Economics{
				TotalSupply:            big.NewInt(10000),
				TotalToDistribute:      big.NewInt(10000),
				TotalNewlyMinted:       big.NewInt(10000),
				RewardsPerBlockPerNode: big.NewInt(10000),
				NodePrice:              big.NewInt(10000),
			},
		},
	}
	valInfo := make(map[uint32][]*state.ValidatorInfoData)
	valInfo[0] = []*state.ValidatorInfoData{
		{
			PublicKey:       []byte("pubkey"),
			ShardId:         0,
			AccumulatedFees: big.NewInt(100),
		},
	}
	bdy, err := rwd.CreateRewardsMiniBlocks(mb, valInfo)
	assert.Nil(t, err)
	assert.NotNil(t, bdy)
}

func TestRewardsCreator_VerifyRewardsMiniBlocksHashDoesNotMatch(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	rwd, _ := NewEpochStartRewardsCreator(args)

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
		EpochStart: block.EpochStart{
			Economics: block.Economics{
				TotalSupply:            big.NewInt(10000),
				TotalToDistribute:      big.NewInt(10000),
				TotalNewlyMinted:       big.NewInt(10000),
				RewardsPerBlockPerNode: big.NewInt(10000),
				NodePrice:              big.NewInt(10000),
			},
		},
		MiniBlockHeaders: []block.MiniBlockHeader{
			mbh,
		},
	}
	valInfo := make(map[uint32][]*state.ValidatorInfoData)
	valInfo[0] = []*state.ValidatorInfoData{
		{
			PublicKey:       []byte("pubkey"),
			ShardId:         0,
			AccumulatedFees: big.NewInt(100),
		},
	}

	err := rwd.VerifyRewardsMiniBlocks(mb, valInfo)
	assert.Equal(t, epochStart.ErrRewardMiniBlockHashDoesNotMatch, err)
}

func TestRewardsCreator_VerifyRewardsMiniBlocksRewardsMbNumDoesNotMatch(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	rwd, _ := NewEpochStartRewardsCreator(args)
	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Value:   big.NewInt(100),
		RcvAddr: []byte{},
		Epoch:   0,
	}
	rwdTxHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &mock.HasherMock{}, rwdTx)
	bdy := block.MiniBlock{
		TxHashes:        [][]byte{rwdTxHash},
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
		EpochStart: block.EpochStart{
			Economics: block.Economics{
				TotalSupply:            big.NewInt(10000),
				TotalToDistribute:      big.NewInt(10000),
				TotalNewlyMinted:       big.NewInt(10000),
				RewardsPerBlockPerNode: big.NewInt(10000),
				NodePrice:              big.NewInt(10000),
			},
		},
		MiniBlockHeaders: []block.MiniBlockHeader{
			mbh,
			mbh,
		},
	}
	valInfo := make(map[uint32][]*state.ValidatorInfoData)
	valInfo[0] = []*state.ValidatorInfoData{
		{
			PublicKey:       []byte("pubkey"),
			ShardId:         0,
			AccumulatedFees: big.NewInt(100),
		},
	}

	err := rwd.VerifyRewardsMiniBlocks(mb, valInfo)
	assert.Equal(t, epochStart.ErrRewardMiniBlocksNumDoesNotMatch, err)
}

func TestRewardsCreator_VerifyRewardsMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	rwd, _ := NewEpochStartRewardsCreator(args)
	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Value:   big.NewInt(100),
		RcvAddr: []byte{},
		Epoch:   0,
	}
	rwdTxHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &mock.HasherMock{}, rwdTx)
	bdy := block.MiniBlock{
		TxHashes:        [][]byte{rwdTxHash},
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
		EpochStart: block.EpochStart{
			Economics: block.Economics{
				TotalSupply:            big.NewInt(10000),
				TotalToDistribute:      big.NewInt(10000),
				TotalNewlyMinted:       big.NewInt(10000),
				RewardsPerBlockPerNode: big.NewInt(10000),
				NodePrice:              big.NewInt(10000),
			},
		},
		MiniBlockHeaders: []block.MiniBlockHeader{
			mbh,
		},
	}
	valInfo := make(map[uint32][]*state.ValidatorInfoData)
	valInfo[0] = []*state.ValidatorInfoData{
		{
			PublicKey:       []byte("pubkey"),
			ShardId:         0,
			AccumulatedFees: big.NewInt(100),
		},
	}

	err := rwd.VerifyRewardsMiniBlocks(mb, valInfo)
	assert.Nil(t, err)
}

func TestRewardsCreator_CreateMarshalizedData(t *testing.T) {
	t.Parallel()

	args := getRewardsArguments()
	rwd, _ := NewEpochStartRewardsCreator(args)

	mb := &block.MetaBlock{
		EpochStart: block.EpochStart{
			Economics: block.Economics{
				TotalSupply:            big.NewInt(10000),
				TotalToDistribute:      big.NewInt(10000),
				TotalNewlyMinted:       big.NewInt(10000),
				RewardsPerBlockPerNode: big.NewInt(10000),
				NodePrice:              big.NewInt(10000),
			},
		},
	}
	valInfo := make(map[uint32][]*state.ValidatorInfoData)
	valInfo[0] = []*state.ValidatorInfoData{
		{
			PublicKey:       []byte("pubkey"),
			ShardId:         0,
			AccumulatedFees: big.NewInt(100),
		},
	}
	_, _ = rwd.CreateRewardsMiniBlocks(mb, valInfo)

	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Value:   big.NewInt(100),
		RcvAddr: []byte{},
		Epoch:   0,
	}
	rwdTxHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &mock.HasherMock{}, rwdTx)

	bdy := block.Body{
		{
			ReceiverShardID: 0,
			Type:            block.RewardsBlock,
			TxHashes:        [][]byte{rwdTxHash},
		},
	}
	res := rwd.CreateMarshalizedData(bdy)

	assert.NotNil(t, res)
}

func TestRewardsCreator_SaveTxBlockToStorage(t *testing.T) {
	t.Parallel()

	putRwdTxWasCalled := false
	putMbWasCalled := false

	args := getRewardsArguments()
	args.RewardsStorage = &mock.StorerStub{
		PutCalled: func(_, _ []byte) error {
			putRwdTxWasCalled = true
			return nil
		},
	}
	args.MiniBlockStorage = &mock.StorerStub{
		PutCalled: func(_, _ []byte) error {
			putMbWasCalled = true
			return nil
		},
	}
	rwd, _ := NewEpochStartRewardsCreator(args)

	mb := &block.MetaBlock{
		EpochStart: block.EpochStart{
			Economics: block.Economics{
				TotalSupply:            big.NewInt(10000),
				TotalToDistribute:      big.NewInt(10000),
				TotalNewlyMinted:       big.NewInt(10000),
				RewardsPerBlockPerNode: big.NewInt(10000),
				NodePrice:              big.NewInt(10000),
			},
		},
	}
	valInfo := make(map[uint32][]*state.ValidatorInfoData)
	valInfo[0] = []*state.ValidatorInfoData{
		{
			PublicKey:       []byte("pubkey"),
			ShardId:         0,
			AccumulatedFees: big.NewInt(100),
		},
	}
	_, _ = rwd.CreateRewardsMiniBlocks(mb, valInfo)

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
		{
			ReceiverShardID: 0,
			SenderShardID:   core.MetachainShardId,
			Type:            block.RewardsBlock,
			TxHashes:        [][]byte{rwdTxHash},
		},
	}
	rwd.SaveTxBlockToStorage(&mb2, bdy)

	assert.True(t, putRwdTxWasCalled)
	assert.True(t, putMbWasCalled)
}

func getRewardsArguments() ArgsNewRewardsCreator {
	return ArgsNewRewardsCreator{
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(2),
		AddrConverter:    &mock.AddressConverterMock{},
		RewardsStorage:   &mock.StorerStub{},
		MiniBlockStorage: &mock.StorerStub{},
		Hasher:           &mock.HasherMock{},
		Marshalizer:      &mock.MarshalizerMock{},
	}
}
