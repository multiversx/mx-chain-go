package metachain

import (
	"bytes"
	"errors"
	"math/big"
	"reflect"
	"sort"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func createMockValidatorInfo() map[uint32][]*state.ValidatorInfo {
	validatorInfo := map[uint32][]*state.ValidatorInfo{
		0: {
			&state.ValidatorInfo{
				PublicKey:                  []byte("a1"),
				ShardId:                    0,
				List:                       "eligible",
				Index:                      1,
				TempRating:                 100,
				Rating:                     1000,
				RewardAddress:              []byte("rewardA1"),
				LeaderSuccess:              1,
				LeaderFailure:              2,
				ValidatorSuccess:           3,
				ValidatorFailure:           4,
				TotalLeaderSuccess:         10,
				TotalLeaderFailure:         20,
				TotalValidatorSuccess:      30,
				TotalValidatorFailure:      40,
				NumSelectedInSuccessBlocks: 5,
				AccumulatedFees:            big.NewInt(100),
			},
			&state.ValidatorInfo{
				PublicKey:                  []byte("a2"),
				ShardId:                    0,
				List:                       "waiting",
				Index:                      2,
				TempRating:                 101,
				Rating:                     1001,
				RewardAddress:              []byte("rewardA2"),
				LeaderSuccess:              6,
				LeaderFailure:              7,
				ValidatorSuccess:           8,
				ValidatorFailure:           9,
				TotalLeaderSuccess:         60,
				TotalLeaderFailure:         70,
				TotalValidatorSuccess:      80,
				TotalValidatorFailure:      90,
				NumSelectedInSuccessBlocks: 10,
				AccumulatedFees:            big.NewInt(101),
			},
		},
		core.MetachainShardId: {
			&state.ValidatorInfo{
				PublicKey:                  []byte("m1"),
				ShardId:                    core.MetachainShardId,
				List:                       "eligible",
				Index:                      1,
				TempRating:                 100,
				Rating:                     1000,
				RewardAddress:              []byte("rewardM1"),
				LeaderSuccess:              1,
				LeaderFailure:              2,
				ValidatorSuccess:           3,
				ValidatorFailure:           4,
				TotalLeaderSuccess:         10,
				TotalLeaderFailure:         20,
				TotalValidatorSuccess:      30,
				TotalValidatorFailure:      40,
				NumSelectedInSuccessBlocks: 5,
				AccumulatedFees:            big.NewInt(100),
			},
			&state.ValidatorInfo{
				PublicKey:                  []byte("m0"),
				ShardId:                    core.MetachainShardId,
				List:                       "waiting",
				Index:                      2,
				TempRating:                 101,
				Rating:                     1001,
				RewardAddress:              []byte("rewardM2"),
				LeaderSuccess:              6,
				LeaderFailure:              7,
				ValidatorSuccess:           8,
				ValidatorFailure:           9,
				TotalLeaderSuccess:         60,
				TotalLeaderFailure:         70,
				TotalValidatorSuccess:      80,
				TotalValidatorFailure:      90,
				NumSelectedInSuccessBlocks: 10,
				AccumulatedFees:            big.NewInt(101),
			},
		},
	}
	return validatorInfo
}

func createMockEpochValidatorInfoCreatorsArguments() ArgsNewValidatorInfoCreator {
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	_ = shardCoordinator.SetSelfId(core.MetachainShardId)

	argsNewEpochEconomics := ArgsNewValidatorInfoCreator{
		ShardCoordinator: shardCoordinator,
		MiniBlockStorage: createMemUnit(),
		Hasher:           &mock.HasherMock{},
		Marshalizer:      &mock.MarshalizerMock{},
		DataPool: &testscommon.PoolsHolderStub{
			MiniBlocksCalled: func() storage.Cacher {
				return &testscommon.CacherStub{
					RemoveCalled: func(key []byte) {},
				}
			},
		},
	}
	return argsNewEpochEconomics
}

func verifyMiniBlocks(bl *block.MiniBlock, infos []*state.ValidatorInfo, marshalizer marshal.Marshalizer) bool {
	if bl.SenderShardID != core.MetachainShardId ||
		bl.ReceiverShardID != core.AllShardId ||
		len(bl.TxHashes) == 0 ||
		bl.Type != block.PeerBlock {
		return false
	}

	validatorCopy := make([]*state.ValidatorInfo, len(infos))
	copy(validatorCopy, infos)
	sort.Slice(validatorCopy, func(a, b int) bool {
		return bytes.Compare(validatorCopy[a].PublicKey, validatorCopy[b].PublicKey) < 0
	})

	for i, txHash := range bl.TxHashes {
		vi := createShardValidatorInfo(validatorCopy[i])
		unmarshaledVi := &state.ShardValidatorInfo{}
		_ = marshalizer.Unmarshal(unmarshaledVi, txHash)
		if !reflect.DeepEqual(unmarshaledVi, vi) {
			return false
		}
	}

	return true
}

func TestEpochValidatorInfoCreator_NewValidatorInfoCreatorNilMarshalizer(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochValidatorInfoCreatorsArguments()
	arguments.Marshalizer = nil
	vic, err := NewValidatorInfoCreator(arguments)

	require.Nil(t, vic)
	require.Equal(t, epochStart.ErrNilMarshalizer, err)
}

func TestEpochValidatorInfoCreator_NewValidatorInfoCreatorNilHasher(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochValidatorInfoCreatorsArguments()
	arguments.Hasher = nil
	vic, err := NewValidatorInfoCreator(arguments)

	require.Nil(t, vic)
	require.Equal(t, epochStart.ErrNilHasher, err)
}

func TestEpochValidatorInfoCreator_NewValidatorInfoCreatorNilStore(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochValidatorInfoCreatorsArguments()
	arguments.MiniBlockStorage = nil
	vic, err := NewValidatorInfoCreator(arguments)

	require.Nil(t, vic)
	require.Equal(t, epochStart.ErrNilStorage, err)
}

func TestEpochValidatorInfoCreator_NewValidatorInfoCreatorNilShardCoordinator(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochValidatorInfoCreatorsArguments()
	arguments.ShardCoordinator = nil
	vic, err := NewValidatorInfoCreator(arguments)

	require.Nil(t, vic)
	require.Equal(t, epochStart.ErrNilShardCoordinator, err)
}

func TestEpochValidatorInfoCreator_NewValidatorInfoCreatorNilDataPool(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochValidatorInfoCreatorsArguments()
	arguments.DataPool = nil
	vic, err := NewValidatorInfoCreator(arguments)

	require.Nil(t, vic)
	require.Equal(t, epochStart.ErrNilDataPoolsHolder, err)
}

func TestEpochValidatorInfoCreator_NewValidatorInfoCreatorShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochValidatorInfoCreatorsArguments()
	vic, err := NewValidatorInfoCreator(arguments)

	require.NotNil(t, vic)
	require.Nil(t, err)
}

func TestEpochValidatorInfoCreator_CreateValidatorInfoMiniBlocksNilValidatorInfo(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochValidatorInfoCreatorsArguments()
	vic, _ := NewValidatorInfoCreator(arguments)
	mbs, err := vic.CreateValidatorInfoMiniBlocks(nil)

	require.Equal(t, epochStart.ErrNilValidatorInfo, err)
	require.Nil(t, mbs)
}

func TestEpochValidatorInfoCreator_CreateValidatorInfoMiniBlocksErrMarshal(t *testing.T) {
	t.Parallel()

	validatorInfo := createMockValidatorInfo()
	arguments := createMockEpochValidatorInfoCreatorsArguments()
	arguments.Marshalizer = &mock.MarshalizerMock{Fail: true}
	vic, _ := NewValidatorInfoCreator(arguments)
	mbs, err := vic.CreateValidatorInfoMiniBlocks(validatorInfo)

	errExpected := errors.New("MarshalizerMock generic error")
	require.Equal(t, errExpected, err)
	require.Nil(t, mbs)
}

func TestEpochValidatorInfoCreator_CreateValidatorInfoMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()

	validatorInfo := createMockValidatorInfo()
	arguments := createMockEpochValidatorInfoCreatorsArguments()
	vic, _ := NewValidatorInfoCreator(arguments)
	mbs, err := vic.CreateValidatorInfoMiniBlocks(validatorInfo)

	require.Nil(t, err)
	require.NotNil(t, mbs)
	require.Equal(t, 2, len(mbs))
}

func TestEpochValidatorInfoCreator_CreateValidatorInfoMiniBlocksShouldBeCorrect(t *testing.T) {
	t.Parallel()

	validatorInfo := createMockValidatorInfo()
	arguments := createMockEpochValidatorInfoCreatorsArguments()
	vic, _ := NewValidatorInfoCreator(arguments)
	mbs, _ := vic.CreateValidatorInfoMiniBlocks(validatorInfo)

	correctMB0 := verifyMiniBlocks(mbs[0], validatorInfo[0], arguments.Marshalizer)
	require.True(t, correctMB0)
	correctMbMeta := verifyMiniBlocks(mbs[1], validatorInfo[core.MetachainShardId], arguments.Marshalizer)
	require.True(t, correctMbMeta)
}

func TestEpochValidatorInfoCreator_VerifyValidatorInfoMiniBlocksShouldBeCorrect(t *testing.T) {
	t.Parallel()

	validatorInfo := createMockValidatorInfo()
	arguments := createMockEpochValidatorInfoCreatorsArguments()
	vic, _ := NewValidatorInfoCreator(arguments)
	miniblocks := createValidatorInfoMiniBlocks(validatorInfo, arguments)

	err := vic.VerifyValidatorInfoMiniBlocks(miniblocks, validatorInfo)
	require.Nil(t, err)
}

func TestEpochValidatorInfoCreator_VerifyValidatorInfoMiniBlocksNilValidatorInfo(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochValidatorInfoCreatorsArguments()
	vic, _ := NewValidatorInfoCreator(arguments)
	err := vic.VerifyValidatorInfoMiniBlocks(nil, nil)
	require.NotNil(t, err)
}

func TestEpochValidatorInfoCreator_VerifyValidatorInfoMiniBlocksNumberNoMatch(t *testing.T) {
	t.Parallel()

	validatorInfo := createMockValidatorInfo()
	arguments := createMockEpochValidatorInfoCreatorsArguments()
	vic, _ := NewValidatorInfoCreator(arguments)

	miniblocks := createValidatorInfoMiniBlocks(validatorInfo, arguments)

	err := vic.VerifyValidatorInfoMiniBlocks(miniblocks[0:1], validatorInfo)
	require.NotNil(t, err)
	require.Equal(t, epochStart.ErrValidatorInfoMiniBlocksNumDoesNotMatch, err)
}

func TestEpochValidatorInfoCreator_VerifyValidatorInfoMiniBlocksTxHashNoMatchT(t *testing.T) {
	t.Parallel()

	validatorInfo := createMockValidatorInfo()
	arguments := createMockEpochValidatorInfoCreatorsArguments()
	vic, _ := NewValidatorInfoCreator(arguments)

	miniblocks := createValidatorInfoMiniBlocks(validatorInfo, arguments)
	miniblocks[0].TxHashes[1] = []byte("testHash")

	err := vic.VerifyValidatorInfoMiniBlocks(miniblocks, validatorInfo)
	require.NotNil(t, err)
	require.Equal(t, epochStart.ErrValidatorMiniBlockHashDoesNotMatch, err)
}

func TestEpochValidatorInfoCreator_VerifyValidatorInfoMiniBlocksNilMiniblocks(t *testing.T) {
	t.Parallel()

	validatorInfo := createMockValidatorInfo()
	arguments := createMockEpochValidatorInfoCreatorsArguments()
	vic, _ := NewValidatorInfoCreator(arguments)
	err := vic.VerifyValidatorInfoMiniBlocks(nil, validatorInfo)
	require.NotNil(t, err)
	require.Equal(t, epochStart.ErrNilMiniblocks, err)
}

func TestEpochValidatorInfoCreator_VerifyValidatorInfoMiniBlocksNilOneMiniblock(t *testing.T) {
	t.Parallel()

	validatorInfo := createMockValidatorInfo()
	arguments := createMockEpochValidatorInfoCreatorsArguments()
	vic, _ := NewValidatorInfoCreator(arguments)
	miniblocks := createValidatorInfoMiniBlocks(validatorInfo, arguments)
	miniblocks[1] = nil
	err := vic.VerifyValidatorInfoMiniBlocks(miniblocks, validatorInfo)
	require.NotNil(t, err)
	require.Equal(t, epochStart.ErrNilMiniblock, err)
}

func createValidatorInfoMiniBlocks(
	validatorInfo map[uint32][]*state.ValidatorInfo,
	arguments ArgsNewValidatorInfoCreator,
) []*block.MiniBlock {
	miniblocks := make([]*block.MiniBlock, 0)
	for _, validators := range validatorInfo {
		if len(validators) == 0 {
			continue
		}

		miniBlock := &block.MiniBlock{}
		miniBlock.SenderShardID = arguments.ShardCoordinator.SelfId()
		miniBlock.ReceiverShardID = core.AllShardId
		miniBlock.TxHashes = make([][]byte, len(validators))
		miniBlock.Type = block.PeerBlock

		validatorCopy := make([]*state.ValidatorInfo, len(validators))
		copy(validatorCopy, validators)
		sort.Slice(validatorCopy, func(a, b int) bool {
			return bytes.Compare(validatorCopy[a].PublicKey, validatorCopy[b].PublicKey) < 0
		})

		for index, validator := range validatorCopy {
			shardValidator := createShardValidatorInfo(validator)
			marshalizedValidator, _ := arguments.Marshalizer.Marshal(shardValidator)
			miniBlock.TxHashes[index] = marshalizedValidator
		}

		miniblocks = append(miniblocks, miniBlock)
	}
	return miniblocks
}

func TestEpochValidatorInfoCreator_SaveValidatorInfoBlocksToStorage(t *testing.T) {
	validatorInfo := createMockValidatorInfo()
	arguments := createMockEpochValidatorInfoCreatorsArguments()
	arguments.MiniBlockStorage = mock.NewStorerMock()

	vic, _ := NewValidatorInfoCreator(arguments)
	miniblocks := createValidatorInfoMiniBlocks(validatorInfo, arguments)
	miniblockHeaders := make([]block.MiniBlockHeader, 0)

	hasher := arguments.Hasher
	marshalizer := arguments.Marshalizer
	miniBlockStorage := arguments.MiniBlockStorage

	for _, mb := range miniblocks {
		mMb, _ := marshalizer.Marshal(mb)
		hash := hasher.Compute(string(mMb))
		mbHeader := block.MiniBlockHeader{
			Hash:            hash,
			SenderShardID:   mb.SenderShardID,
			ReceiverShardID: mb.ReceiverShardID,
			TxCount:         uint32(len(mb.TxHashes)),
			Type:            block.PeerBlock,
		}
		miniblockHeaders = append(miniblockHeaders, mbHeader)
	}

	meta := &block.MetaBlock{
		Nonce:                  0,
		Round:                  0,
		TimeStamp:              0,
		ShardInfo:              nil,
		Signature:              nil,
		LeaderSignature:        nil,
		PubKeysBitmap:          nil,
		PrevHash:               nil,
		PrevRandSeed:           nil,
		RandSeed:               nil,
		RootHash:               nil,
		ValidatorStatsRootHash: nil,
		MiniBlockHeaders:       miniblockHeaders,
		ReceiptsHash:           nil,
		EpochStart:             block.EpochStart{},
		ChainID:                nil,
		Epoch:                  0,
		TxCount:                0,
		AccumulatedFees:        nil,
		AccumulatedFeesInEpoch: nil,
	}

	body := &block.Body{MiniBlocks: miniblocks}
	vic.SaveValidatorInfoBlocksToStorage(meta, body)

	for i, mbHeader := range meta.MiniBlockHeaders {
		mb, err := miniBlockStorage.Get(mbHeader.Hash)
		require.Nil(t, err)

		unmarshaledMiniblock := &block.MiniBlock{}
		_ = marshalizer.Unmarshal(unmarshaledMiniblock, mb)

		require.True(t, reflect.DeepEqual(miniblocks[i], unmarshaledMiniblock))
	}
}

func TestEpochValidatorInfoCreator_DeleteValidatorInfoBlocksFromStorage(t *testing.T) {
	testDeleteValidatorInfoBlock(t, block.PeerBlock, false)
}

func TestEpochValidatorInfoCreator_DeleteValidatorInfoBlocksFromStorageDoesDeleteOnlyPeerBlocks(t *testing.T) {
	testDeleteValidatorInfoBlock(t, block.TxBlock, true)
}

func testDeleteValidatorInfoBlock(t *testing.T, blockType block.Type, shouldExist bool) {
	validatorInfo := createMockValidatorInfo()
	arguments := createMockEpochValidatorInfoCreatorsArguments()
	arguments.MiniBlockStorage = mock.NewStorerMock()

	vic, _ := NewValidatorInfoCreator(arguments)
	miniblocks := createValidatorInfoMiniBlocks(validatorInfo, arguments)
	miniblockHeaders := make([]block.MiniBlockHeader, 0)

	hasher := arguments.Hasher
	marshalizer := arguments.Marshalizer
	mbStorage := arguments.MiniBlockStorage

	for _, mb := range miniblocks {
		mMb, _ := marshalizer.Marshal(mb)
		hash := hasher.Compute(string(mMb))
		mbHeader := block.MiniBlockHeader{
			Hash:            hash,
			SenderShardID:   mb.SenderShardID,
			ReceiverShardID: mb.ReceiverShardID,
			TxCount:         uint32(len(mb.TxHashes)),
			Type:            blockType,
		}
		miniblockHeaders = append(miniblockHeaders, mbHeader)
		_ = mbStorage.Put(hash, mMb)
	}

	meta := &block.MetaBlock{
		Nonce:                  0,
		Round:                  0,
		TimeStamp:              0,
		ShardInfo:              nil,
		Signature:              nil,
		LeaderSignature:        nil,
		PubKeysBitmap:          nil,
		PrevHash:               nil,
		PrevRandSeed:           nil,
		RandSeed:               nil,
		RootHash:               nil,
		ValidatorStatsRootHash: nil,
		MiniBlockHeaders:       miniblockHeaders,
		ReceiptsHash:           nil,
		EpochStart:             block.EpochStart{},
		ChainID:                nil,
		Epoch:                  0,
		TxCount:                0,
		AccumulatedFees:        nil,
		AccumulatedFeesInEpoch: nil,
	}

	for _, mbHeader := range meta.MiniBlockHeaders {
		mb, err := mbStorage.Get(mbHeader.Hash)
		require.NotNil(t, mb)
		require.Nil(t, err)
	}

	vic.DeleteValidatorInfoBlocksFromStorage(meta)

	for _, mbHeader := range meta.MiniBlockHeaders {
		mb, err := mbStorage.Get(mbHeader.Hash)
		if shouldExist {
			require.NotNil(t, mb)
			require.Nil(t, err)
		} else {
			require.Nil(t, mb)
			require.NotNil(t, err)
		}
	}
}

func TestEpochValidatorInfoCreator_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochValidatorInfoCreatorsArguments()
	vic, _ := NewValidatorInfoCreator(arguments)
	require.False(t, vic.IsInterfaceNil())
}
