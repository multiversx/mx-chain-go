package metachain

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	vics "github.com/multiversx/mx-chain-go/testscommon/validatorInfoCacher"
	"github.com/stretchr/testify/assert"
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
		ShardCoordinator:     shardCoordinator,
		ValidatorInfoStorage: createMemUnit(),
		MiniBlockStorage:     createMemUnit(),
		Hasher:               &hashingMocks.HasherMock{},
		Marshalizer:          &mock.MarshalizerMock{},
		DataPool: &dataRetrieverMock.PoolsHolderStub{
			MiniBlocksCalled: func() storage.Cacher {
				return &testscommon.CacherStub{
					RemoveCalled: func(key []byte) {},
				}
			},
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vics.ValidatorInfoCacherStub{}
			},
		},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.RefactorPeersMiniBlocksFlag
			},
		},
	}
	return argsNewEpochEconomics
}

func verifyMiniBlocks(bl *block.MiniBlock, infos []*state.ValidatorInfo, marshalledShardValidatorsInfo [][]byte, marshalizer marshal.Marshalizer) bool {
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

	for i, marshalledShardValidatorInfo := range marshalledShardValidatorsInfo {
		vi := createShardValidatorInfo(infos[i])
		unmarshaledVi := &state.ShardValidatorInfo{}
		_ = marshalizer.Unmarshal(unmarshaledVi, marshalledShardValidatorInfo)
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

func TestEpochValidatorInfoCreator_NewValidatorInfoCreatorNilEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochValidatorInfoCreatorsArguments()
	arguments.EnableEpochsHandler = nil
	vic, err := NewValidatorInfoCreator(arguments)

	require.Nil(t, vic)
	require.Equal(t, epochStart.ErrNilEnableEpochsHandler, err)
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

	shardValidatorInfo := make([]*state.ShardValidatorInfo, len(validatorInfo[0]))
	marshalledShardValidatorInfo := make([][]byte, len(validatorInfo[0]))
	for i := 0; i < len(validatorInfo[0]); i++ {
		shardValidatorInfo[i] = createShardValidatorInfo(validatorInfo[0][i])
		marshalledShardValidatorInfo[i], _ = arguments.Marshalizer.Marshal(shardValidatorInfo[i])
	}
	correctMB0 := verifyMiniBlocks(mbs[0], validatorInfo[0], marshalledShardValidatorInfo, arguments.Marshalizer)
	require.True(t, correctMB0)

	shardValidatorInfo = make([]*state.ShardValidatorInfo, len(validatorInfo[core.MetachainShardId]))
	marshalledShardValidatorInfo = make([][]byte, len(validatorInfo[core.MetachainShardId]))
	for i := 0; i < len(validatorInfo[core.MetachainShardId]); i++ {
		shardValidatorInfo[i] = createShardValidatorInfo(validatorInfo[core.MetachainShardId][i])
		marshalledShardValidatorInfo[i], _ = arguments.Marshalizer.Marshal(shardValidatorInfo[i])
	}
	correctMbMeta := verifyMiniBlocks(mbs[1], validatorInfo[core.MetachainShardId], marshalledShardValidatorInfo, arguments.Marshalizer)
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

func TestEpochValidatorInfoCreator_VerifyValidatorInfoMiniBlocksTxHashDoNotMatch(t *testing.T) {
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
			shardValidatorInfo := createShardValidatorInfo(validator)
			shardValidatorInfoHash, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, shardValidatorInfo)
			miniBlock.TxHashes[index] = shardValidatorInfoHash
		}

		miniblocks = append(miniblocks, miniBlock)
	}
	return miniblocks
}

func TestEpochValidatorInfoCreator_SaveValidatorInfoBlockDataToStorage(t *testing.T) {
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
	vic.SaveBlockDataToStorage(meta, body)

	for i, mbHeader := range meta.MiniBlockHeaders {
		mb, err := miniBlockStorage.Get(mbHeader.Hash)
		require.Nil(t, err)

		unmarshaledMiniblock := &block.MiniBlock{}
		_ = marshalizer.Unmarshal(unmarshaledMiniblock, mb)

		require.True(t, reflect.DeepEqual(miniblocks[i], unmarshaledMiniblock))
	}
}

func TestEpochValidatorInfoCreator_DeleteValidatorInfoBlockDataFromStorage(t *testing.T) {
	testDeleteValidatorInfoBlockData(t, block.PeerBlock, false)
}

func TestEpochValidatorInfoCreator_DeleteValidatorInfoBlockDataFromStorageDoesDeleteOnlyPeerBlocks(t *testing.T) {
	testDeleteValidatorInfoBlockData(t, block.TxBlock, true)
}

func testDeleteValidatorInfoBlockData(t *testing.T, blockType block.Type, shouldExist bool) {
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

	body := &block.Body{}
	vic.DeleteBlockDataFromStorage(meta, body)

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

func TestEpochValidatorInfoCreator_GetShardValidatorInfoData(t *testing.T) {
	t.Parallel()

	t.Run("get shard validator info data before refactor peers mini block activation flag is set", func(t *testing.T) {
		t.Parallel()

		arguments := createMockEpochValidatorInfoCreatorsArguments()
		arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return false
			},
		}
		vic, _ := NewValidatorInfoCreator(arguments)

		shardValidatorInfo := &state.ShardValidatorInfo{
			PublicKey: []byte("x"),
		}
		marshalledShardValidatorInfo, _ := arguments.Marshalizer.Marshal(shardValidatorInfo)
		shardValidatorInfoData, _ := vic.getShardValidatorInfoData(shardValidatorInfo)
		assert.Equal(t, marshalledShardValidatorInfo, shardValidatorInfoData)
	})

	t.Run("get shard validator info data after refactor peers mini block activation flag is set", func(t *testing.T) {
		t.Parallel()

		arguments := createMockEpochValidatorInfoCreatorsArguments()
		arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.RefactorPeersMiniBlocksFlag
			},
		}
		vic, _ := NewValidatorInfoCreator(arguments)

		shardValidatorInfo := &state.ShardValidatorInfo{
			PublicKey: []byte("x"),
		}
		shardValidatorInfoHash, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, shardValidatorInfo)
		shardValidatorInfoData, _ := vic.getShardValidatorInfoData(shardValidatorInfo)
		assert.Equal(t, shardValidatorInfoHash, shardValidatorInfoData)
	})
}

func TestEpochValidatorInfoCreator_CreateMarshalledData(t *testing.T) {
	t.Parallel()

	t.Run("CreateMarshalledData should return nil before refactor peers mini block activation flag is set", func(t *testing.T) {
		t.Parallel()

		arguments := createMockEpochValidatorInfoCreatorsArguments()
		arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return false
			},
		}
		vic, _ := NewValidatorInfoCreator(arguments)

		body := createMockBlockBody(0, 1, block.TxBlock)
		marshalledData := vic.CreateMarshalledData(body)
		assert.Nil(t, marshalledData)
	})

	t.Run("CreateMarshalledData should return nil body is nil", func(t *testing.T) {
		t.Parallel()

		arguments := createMockEpochValidatorInfoCreatorsArguments()
		arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.RefactorPeersMiniBlocksFlag
			},
		}
		vic, _ := NewValidatorInfoCreator(arguments)

		marshalledData := vic.CreateMarshalledData(nil)
		assert.Nil(t, marshalledData)
	})

	t.Run("CreateMarshalledData should return empty slice when there is no peer mini block in body", func(t *testing.T) {
		t.Parallel()

		arguments := createMockEpochValidatorInfoCreatorsArguments()
		arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.RefactorPeersMiniBlocksFlag
			},
		}
		vic, _ := NewValidatorInfoCreator(arguments)

		body := createMockBlockBody(0, 1, block.TxBlock)
		marshalledData := vic.CreateMarshalledData(body)
		assert.Equal(t, make(map[string][][]byte), marshalledData)
	})

	t.Run("CreateMarshalledData should return empty slice when sender or receiver do not match", func(t *testing.T) {
		t.Parallel()

		arguments := createMockEpochValidatorInfoCreatorsArguments()
		arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.RefactorPeersMiniBlocksFlag
			},
		}
		vic, _ := NewValidatorInfoCreator(arguments)

		body := createMockBlockBody(0, 1, block.PeerBlock)
		marshalledData := vic.CreateMarshalledData(body)
		assert.Equal(t, make(map[string][][]byte), marshalledData)
	})

	t.Run("CreateMarshalledData should return empty slice when tx hash does not exist in validator info cacher", func(t *testing.T) {
		t.Parallel()

		arguments := createMockEpochValidatorInfoCreatorsArguments()
		arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.RefactorPeersMiniBlocksFlag
			},
		}
		arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vics.ValidatorInfoCacherStub{
					GetValidatorInfoCalled: func(validatorInfoHash []byte) (*state.ShardValidatorInfo, error) {
						return nil, errors.New("error")
					},
				}
			},
		}
		vic, _ := NewValidatorInfoCreator(arguments)

		body := createMockBlockBody(core.MetachainShardId, 0, block.PeerBlock)
		marshalledData := vic.CreateMarshalledData(body)
		assert.Equal(t, make(map[string][][]byte), marshalledData)
	})

	t.Run("CreateMarshalledData should work", func(t *testing.T) {
		t.Parallel()

		arguments := createMockEpochValidatorInfoCreatorsArguments()

		svi1 := &state.ShardValidatorInfo{PublicKey: []byte("x")}
		marshalledSVI1, _ := arguments.Marshalizer.Marshal(svi1)

		svi2 := &state.ShardValidatorInfo{PublicKey: []byte("y")}
		marshalledSVI2, _ := arguments.Marshalizer.Marshal(svi2)

		svi3 := &state.ShardValidatorInfo{PublicKey: []byte("z")}
		marshalledSVI3, _ := arguments.Marshalizer.Marshal(svi3)

		arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.RefactorPeersMiniBlocksFlag
			},
		}
		arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vics.ValidatorInfoCacherStub{
					GetValidatorInfoCalled: func(validatorInfoHash []byte) (*state.ShardValidatorInfo, error) {
						if bytes.Equal(validatorInfoHash, []byte("a")) {
							return svi1, nil
						}
						if bytes.Equal(validatorInfoHash, []byte("b")) {
							return svi2, nil
						}
						if bytes.Equal(validatorInfoHash, []byte("c")) {
							return svi3, nil
						}
						return nil, errors.New("error")
					},
				}
			},
		}
		vic, _ := NewValidatorInfoCreator(arguments)

		body := createMockBlockBody(core.MetachainShardId, 0, block.PeerBlock)
		marshalledData := vic.CreateMarshalledData(body)
		require.Equal(t, 1, len(marshalledData))
		require.Equal(t, 3, len(marshalledData[common.ValidatorInfoTopic]))
		assert.Equal(t, marshalledSVI1, marshalledData[common.ValidatorInfoTopic][0])
		assert.Equal(t, marshalledSVI2, marshalledData[common.ValidatorInfoTopic][1])
		assert.Equal(t, marshalledSVI3, marshalledData[common.ValidatorInfoTopic][2])
	})
}

func TestEpochValidatorInfoCreator_SetMarshalledValidatorInfoTxsShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochValidatorInfoCreatorsArguments()

	svi1 := &state.ShardValidatorInfo{PublicKey: []byte("x")}
	marshalledSVI1, _ := arguments.Marshalizer.Marshal(svi1)

	svi2 := &state.ShardValidatorInfo{PublicKey: []byte("y")}
	marshalledSVI2, _ := arguments.Marshalizer.Marshal(svi2)

	arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.RefactorPeersMiniBlocksFlag
		},
	}
	arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
		CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
			return &vics.ValidatorInfoCacherStub{
				GetValidatorInfoCalled: func(validatorInfoHash []byte) (*state.ShardValidatorInfo, error) {
					if bytes.Equal(validatorInfoHash, []byte("a")) {
						return svi1, nil
					}
					if bytes.Equal(validatorInfoHash, []byte("c")) {
						return svi2, nil
					}
					return nil, errors.New("error")
				},
			}
		},
	}
	vic, _ := NewValidatorInfoCreator(arguments)

	miniBlock := createMockMiniBlock(core.MetachainShardId, 0, block.PeerBlock)
	marshalledValidatorInfoTxs := vic.getMarshalledValidatorInfoTxs(miniBlock)

	require.Equal(t, 2, len(marshalledValidatorInfoTxs))
	assert.Equal(t, marshalledSVI1, marshalledValidatorInfoTxs[0])
	assert.Equal(t, marshalledSVI2, marshalledValidatorInfoTxs[1])
}

func TestEpochValidatorInfoCreator_GetValidatorInfoTxsShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochValidatorInfoCreatorsArguments()

	svi1 := &state.ShardValidatorInfo{PublicKey: []byte("x")}
	svi2 := &state.ShardValidatorInfo{PublicKey: []byte("y")}
	svi3 := &state.ShardValidatorInfo{PublicKey: []byte("z")}

	arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.RefactorPeersMiniBlocksFlag
		},
	}
	arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
		CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
			return &vics.ValidatorInfoCacherStub{
				GetValidatorInfoCalled: func(validatorInfoHash []byte) (*state.ShardValidatorInfo, error) {
					if bytes.Equal(validatorInfoHash, []byte("a")) {
						return svi1, nil
					}
					if bytes.Equal(validatorInfoHash, []byte("b")) {
						return svi2, nil
					}
					if bytes.Equal(validatorInfoHash, []byte("c")) {
						return svi3, nil
					}
					return nil, errors.New("error")
				},
			}
		},
	}
	vic, _ := NewValidatorInfoCreator(arguments)

	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, createMockMiniBlock(core.MetachainShardId, 0, block.TxBlock))
	body.MiniBlocks = append(body.MiniBlocks, createMockMiniBlock(core.MetachainShardId, 0, block.PeerBlock))
	mapValidatorInfoTxs := vic.GetValidatorInfoTxs(body)

	require.Equal(t, 3, len(mapValidatorInfoTxs))
	require.Equal(t, svi1, mapValidatorInfoTxs["a"])
	require.Equal(t, svi2, mapValidatorInfoTxs["b"])
	require.Equal(t, svi3, mapValidatorInfoTxs["c"])
}

func TestEpochValidatorInfoCreator_SetMapShardValidatorInfoShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochValidatorInfoCreatorsArguments()

	svi1 := &state.ShardValidatorInfo{PublicKey: []byte("x")}
	svi2 := &state.ShardValidatorInfo{PublicKey: []byte("y")}

	arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.RefactorPeersMiniBlocksFlag
		},
	}
	arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
		CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
			return &vics.ValidatorInfoCacherStub{
				GetValidatorInfoCalled: func(validatorInfoHash []byte) (*state.ShardValidatorInfo, error) {
					if bytes.Equal(validatorInfoHash, []byte("a")) {
						return svi1, nil
					}
					if bytes.Equal(validatorInfoHash, []byte("b")) {
						return svi2, nil
					}
					return nil, errors.New("error")
				},
			}
		},
	}
	vic, _ := NewValidatorInfoCreator(arguments)

	miniBlock := createMockMiniBlock(core.MetachainShardId, 0, block.TxBlock)
	mapShardValidatorInfo := make(map[string]*state.ShardValidatorInfo)
	vic.setMapShardValidatorInfo(miniBlock, mapShardValidatorInfo)

	require.Equal(t, 2, len(mapShardValidatorInfo))
	require.Equal(t, svi1, mapShardValidatorInfo["a"])
	require.Equal(t, svi2, mapShardValidatorInfo["b"])
}

func TestEpochValidatorInfoCreator_GetShardValidatorInfoShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("get shard validator info before refactor peers mini block activation flag is set", func(t *testing.T) {
		t.Parallel()

		arguments := createMockEpochValidatorInfoCreatorsArguments()

		svi := &state.ShardValidatorInfo{PublicKey: []byte("x")}
		marshalledSVI, _ := arguments.Marshalizer.Marshal(svi)

		arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return false
			},
		}
		arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vics.ValidatorInfoCacherStub{
					GetValidatorInfoCalled: func(validatorInfoHash []byte) (*state.ShardValidatorInfo, error) {
						if bytes.Equal(validatorInfoHash, []byte("a")) {
							return svi, nil
						}
						return nil, errors.New("error")
					},
				}
			},
		}
		vic, _ := NewValidatorInfoCreator(arguments)

		shardValidatorInfo, _ := vic.getShardValidatorInfo(marshalledSVI)
		require.Equal(t, svi, shardValidatorInfo)
	})

	t.Run("get shard validator info after refactor peers mini block activation flag is set", func(t *testing.T) {
		t.Parallel()

		arguments := createMockEpochValidatorInfoCreatorsArguments()

		svi := &state.ShardValidatorInfo{PublicKey: []byte("x")}

		arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.RefactorPeersMiniBlocksFlag
			},
		}
		arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vics.ValidatorInfoCacherStub{
					GetValidatorInfoCalled: func(validatorInfoHash []byte) (*state.ShardValidatorInfo, error) {
						if bytes.Equal(validatorInfoHash, []byte("a")) {
							return svi, nil
						}
						return nil, errors.New("error")
					},
				}
			},
		}
		vic, _ := NewValidatorInfoCreator(arguments)

		shardValidatorInfo, _ := vic.getShardValidatorInfo([]byte("a"))
		require.Equal(t, svi, shardValidatorInfo)
	})
}

func TestEpochValidatorInfoCreator_SaveValidatorInfoShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochValidatorInfoCreatorsArguments()

	svi1 := &state.ShardValidatorInfo{PublicKey: []byte("x")}
	marshalledSVI1, _ := arguments.Marshalizer.Marshal(svi1)

	svi2 := &state.ShardValidatorInfo{PublicKey: []byte("y")}
	marshalledSVI2, _ := arguments.Marshalizer.Marshal(svi2)

	storer := createMemUnit()
	arguments.ValidatorInfoStorage = storer
	arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
		CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
			return &vics.ValidatorInfoCacherStub{
				GetValidatorInfoCalled: func(validatorInfoHash []byte) (*state.ShardValidatorInfo, error) {
					if bytes.Equal(validatorInfoHash, []byte("a")) {
						return svi1, nil
					}
					if bytes.Equal(validatorInfoHash, []byte("b")) {
						return svi2, nil
					}
					return nil, errors.New("error")
				},
			}
		},
	}
	vic, _ := NewValidatorInfoCreator(arguments)

	miniBlock := createMockMiniBlock(core.MetachainShardId, 0, block.TxBlock)
	vic.saveValidatorInfo(miniBlock)

	msvi1, err := storer.Get([]byte("a"))
	assert.Nil(t, err)
	assert.Equal(t, marshalledSVI1, msvi1)

	msvi2, err := storer.Get([]byte("b"))
	assert.Nil(t, err)
	assert.Equal(t, marshalledSVI2, msvi2)

	msvi3, err := storer.Get([]byte("c"))
	assert.NotNil(t, err)
	assert.Nil(t, msvi3)
}

func TestEpochValidatorInfoCreator_RemoveValidatorInfoShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochValidatorInfoCreatorsArguments()

	storer := createMemUnit()
	arguments.ValidatorInfoStorage = storer
	vic, _ := NewValidatorInfoCreator(arguments)

	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, createMockMiniBlock(core.MetachainShardId, 0, block.TxBlock))
	body.MiniBlocks = append(body.MiniBlocks, createMockMiniBlock(core.MetachainShardId, 0, block.PeerBlock))

	_ = storer.Put([]byte("a"), []byte("aa"))
	_ = storer.Put([]byte("b"), []byte("bb"))
	_ = storer.Put([]byte("c"), []byte("cc"))
	_ = storer.Put([]byte("d"), []byte("dd"))

	vic.removeValidatorInfo(body)

	msvi, err := storer.Get([]byte("a"))
	assert.NotNil(t, err)
	assert.Nil(t, msvi)

	msvi, err = storer.Get([]byte("b"))
	assert.NotNil(t, err)
	assert.Nil(t, msvi)

	msvi, err = storer.Get([]byte("c"))
	assert.NotNil(t, err)
	assert.Nil(t, msvi)

	msvi, err = storer.Get([]byte("d"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("dd"), msvi)
}

func TestEpochValidatorInfoCreator_RemoveValidatorInfoFromPoolShouldWork(t *testing.T) {
	t.Parallel()

	shardedDataCacheNotifierMock := testscommon.NewShardedDataCacheNotifierMock()
	arguments := createMockEpochValidatorInfoCreatorsArguments()
	arguments.DataPool = &dataRetrieverMock.PoolsHolderStub{
		CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
			return &vics.ValidatorInfoCacherStub{}
		},
		ValidatorsInfoCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return shardedDataCacheNotifierMock
		},
	}

	vic, _ := NewValidatorInfoCreator(arguments)

	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, createMockMiniBlock(core.MetachainShardId, 0, block.TxBlock))
	body.MiniBlocks = append(body.MiniBlocks, createMockMiniBlock(core.MetachainShardId, 0, block.PeerBlock))

	svi1 := &state.ShardValidatorInfo{PublicKey: []byte("aa")}
	svi2 := &state.ShardValidatorInfo{PublicKey: []byte("bb")}
	svi3 := &state.ShardValidatorInfo{PublicKey: []byte("cc")}
	svi4 := &state.ShardValidatorInfo{PublicKey: []byte("dd")}

	shardedDataCacheNotifierMock.AddData([]byte("a"), svi1, svi1.Size(), "x")
	shardedDataCacheNotifierMock.AddData([]byte("b"), svi2, svi2.Size(), "x")
	shardedDataCacheNotifierMock.AddData([]byte("c"), svi3, svi3.Size(), "x")
	shardedDataCacheNotifierMock.AddData([]byte("d"), svi4, svi4.Size(), "x")

	vic.removeValidatorInfoFromPool(body)

	svi, found := shardedDataCacheNotifierMock.SearchFirstData([]byte("a"))
	assert.False(t, found)
	assert.Nil(t, svi)

	svi, found = shardedDataCacheNotifierMock.SearchFirstData([]byte("b"))
	assert.False(t, found)
	assert.Nil(t, svi)

	svi, found = shardedDataCacheNotifierMock.SearchFirstData([]byte("c"))
	assert.False(t, found)
	assert.Nil(t, svi)

	svi, found = shardedDataCacheNotifierMock.SearchFirstData([]byte("d"))
	assert.True(t, found)
	assert.Equal(t, svi4, svi)
}

func createMockBlockBody(senderShardID, receiverShardID uint32, blockType block.Type) *block.Body {
	return &block.Body{
		MiniBlocks: []*block.MiniBlock{createMockMiniBlock(senderShardID, receiverShardID, blockType)},
	}
}

func createMockMiniBlock(senderShardID, receiverShardID uint32, blockType block.Type) *block.MiniBlock {
	return &block.MiniBlock{
		SenderShardID:   senderShardID,
		ReceiverShardID: receiverShardID,
		Type:            blockType,
		TxHashes: [][]byte{
			[]byte("a"),
			[]byte("b"),
			[]byte("c"),
		},
	}
}

// TestValidatorInfoCreator_CreateMiniblockBackwardsCompatibility will test the sorting call for the backwards compatibility issues
func TestValidatorInfoCreator_CreateMiniblockBackwardsCompatibility(t *testing.T) {
	t.Parallel()

	t.Run("legacy mode", func(t *testing.T) {
		testCreateMiniblockBackwardsCompatibility(t, false, "./testdata/expected-legacy.data")
	})
	t.Run("full deterministic mode", func(t *testing.T) {
		// this will prevent changes to the deterministic algorithm and ensure the backward compatibility
		testCreateMiniblockBackwardsCompatibility(t, true, "./testdata/expected-deterministic.data")
	})
}

func testCreateMiniblockBackwardsCompatibility(t *testing.T, deterministFixEnabled bool, expectedDataFilename string) {
	inputRAW, err := os.ReadFile("./testdata/input.data")
	require.Nil(t, err)

	expectedRAW, err := os.ReadFile(expectedDataFilename)
	require.Nil(t, err)

	filterCutSet := " \r\n\t"
	input := strings.Split(strings.Trim(string(inputRAW), filterCutSet), "\n")
	expected := strings.Split(strings.Trim(string(expectedRAW), filterCutSet), "\n")

	require.Equal(t, len(input), len(expected))

	validators := make([]*state.ValidatorInfo, 0, len(input))
	marshaller := &marshal.GogoProtoMarshalizer{}
	for _, marshalledData := range input {
		vinfo := &state.ValidatorInfo{}
		buffMarshalledData, errDecode := hex.DecodeString(marshalledData)
		require.Nil(t, errDecode)

		err = marshaller.Unmarshal(vinfo, buffMarshalledData)
		require.Nil(t, err)

		validators = append(validators, vinfo)
	}

	arguments := createMockEpochValidatorInfoCreatorsArguments()
	arguments.Marshalizer = &marshal.GogoProtoMarshalizer{} // we need the real marshaller that generated the test set
	arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			if flag == common.DeterministicSortOnValidatorsInfoFixFlag {
				return deterministFixEnabled
			}
			return false
		},
	}

	storer := createMemUnit()
	arguments.ValidatorInfoStorage = storer
	vic, _ := NewValidatorInfoCreator(arguments)

	mb, err := vic.createMiniBlock(validators)
	require.Nil(t, err)

	// test all generated miniblock's "txhashes" are the same with the expected ones
	require.Equal(t, len(expected), len(mb.TxHashes))
	for i, hash := range mb.TxHashes {
		assert.Equal(t, expected[i], hex.EncodeToString(hash), "not matching for index %d", i)
	}
}

func TestValidatorInfoCreator_printAllMiniBlocksShouldNotPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
		}
	}()

	testSlice := []*block.MiniBlock{
		{
			TxHashes:        nil,
			ReceiverShardID: 0,
			SenderShardID:   0,
			Type:            -1,
			Reserved:        nil,
		},
		{
			TxHashes:        make([][]byte, 0),
			ReceiverShardID: core.MetachainShardId,
			SenderShardID:   core.AllShardId,
			Type:            0,
			Reserved:        nil,
		},
		{
			TxHashes:        [][]byte{[]byte("tx hash 1"), []byte("tx hash 2")},
			ReceiverShardID: 0,
			SenderShardID:   0,
			Type:            1,
			Reserved:        nil,
		},
		{
			TxHashes:        [][]byte{[]byte("tx hash 3"), nil, []byte("tx hash 4")}, // a nil tx hash should not cause panic
			ReceiverShardID: core.MetachainShardId,
			SenderShardID:   core.AllShardId,
			Type:            block.PeerBlock,
			Reserved:        nil,
		},
	}

	testSliceWithNilMiniblock := []*block.MiniBlock{
		{
			TxHashes:        [][]byte{[]byte("tx hash 3"), []byte("tx hash 4")},
			ReceiverShardID: core.MetachainShardId,
			SenderShardID:   core.AllShardId,
			Type:            block.PeerBlock,
			Reserved:        nil,
		},
		nil,
		{
			TxHashes:        make([][]byte, 0),
			ReceiverShardID: core.MetachainShardId,
			SenderShardID:   core.AllShardId,
			Type:            0,
			Reserved:        nil,
		},
	}

	arguments := createMockEpochValidatorInfoCreatorsArguments()
	vic, _ := NewValidatorInfoCreator(arguments)

	// do not run these tests in parallel so the panic will be caught by the main defer function
	t.Run("nil and empty slices, should not panic", func(t *testing.T) {
		vic.printAllMiniBlocks(nil, make([]*block.MiniBlock, 0))
		vic.printAllMiniBlocks(make([]*block.MiniBlock, 0), nil)
		vic.printAllMiniBlocks(make([]*block.MiniBlock, 0), testSlice)
		vic.printAllMiniBlocks(nil, testSlice)
	})
	t.Run("slice contains a nil miniblock, should not panic", func(t *testing.T) {
		vic.printAllMiniBlocks(testSlice, testSliceWithNilMiniblock)
		vic.printAllMiniBlocks(testSliceWithNilMiniblock, testSlice)
	})
	t.Run("marshal outputs error, should not panic", func(t *testing.T) {
		localArguments := arguments
		localArguments.Marshalizer = &testscommon.MarshallerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, fmt.Errorf("marshal error")
			},
		}
		instance, _ := NewValidatorInfoCreator(localArguments)
		instance.printAllMiniBlocks(testSlice, testSlice)
	})
}

func TestValidatorInfoCreator_sortValidators(t *testing.T) {
	t.Parallel()

	firstValidator := createTestValidatorInfo()
	firstValidator.List = "a"

	secondValidator := createTestValidatorInfo()
	secondValidator.List = "b"

	thirdValidator := createTestValidatorInfo()
	thirdValidator.List = "b"
	thirdValidator.PublicKey = []byte("xxxx")

	t.Run("legacy sort should not change order", func(t *testing.T) {
		t.Parallel()

		arguments := createMockEpochValidatorInfoCreatorsArguments()
		arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return false
			},
		}
		vic, _ := NewValidatorInfoCreator(arguments)

		list := []*state.ValidatorInfo{thirdValidator, secondValidator, firstValidator}
		vic.sortValidators(list)

		assert.Equal(t, list[0], secondValidator) // order not changed for the ones with same public key
		assert.Equal(t, list[1], firstValidator)
		assert.Equal(t, list[2], thirdValidator)
	})
	t.Run("deterministic sort should change order taking into consideration all fields", func(t *testing.T) {
		t.Parallel()

		arguments := createMockEpochValidatorInfoCreatorsArguments()
		arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.DeterministicSortOnValidatorsInfoFixFlag
			},
		}
		vic, _ := NewValidatorInfoCreator(arguments)

		list := []*state.ValidatorInfo{thirdValidator, secondValidator, firstValidator}
		vic.sortValidators(list)

		assert.Equal(t, list[0], firstValidator) // proper sorting
		assert.Equal(t, list[1], secondValidator)
		assert.Equal(t, list[2], thirdValidator)
	})
}

func createTestValidatorInfo() *state.ValidatorInfo {
	return &state.ValidatorInfo{
		PublicKey:                       []byte("pubkey"),
		ShardId:                         1,
		List:                            "new",
		Index:                           2,
		TempRating:                      3,
		Rating:                          4,
		RatingModifier:                  5,
		RewardAddress:                   []byte("reward address"),
		LeaderSuccess:                   6,
		LeaderFailure:                   7,
		ValidatorSuccess:                8,
		ValidatorFailure:                9,
		ValidatorIgnoredSignatures:      10,
		NumSelectedInSuccessBlocks:      11,
		AccumulatedFees:                 big.NewInt(12),
		TotalLeaderSuccess:              13,
		TotalLeaderFailure:              14,
		TotalValidatorSuccess:           15,
		TotalValidatorFailure:           16,
		TotalValidatorIgnoredSignatures: 17,
	}
}
