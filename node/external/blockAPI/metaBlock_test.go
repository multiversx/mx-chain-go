package blockAPI

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	storageMocks "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockMetaAPIProcessor(
	blockHeaderHash []byte,
	storerMock *genericMocks.StorerMock,
	withHistory bool,
	withKey bool,
) *metaAPIBlockProcessor {
	return newMetaApiBlockProcessor(&ArgAPIBlockProcessor{
		APITransactionHandler: &mock.TransactionAPIHandlerStub{},
		SelfShardID:           core.MetachainShardId,
		Marshalizer:           &mock.MarshalizerFake{},
		Store: &storageMocks.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return storerMock, nil
			},
			GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
				if withKey {
					return storerMock.Get(key)
				}
				return blockHeaderHash, nil
			},
		},
		Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
		HistoryRepo: &dblookupext.HistoryRepositoryStub{
			GetEpochByHashCalled: func(hash []byte) (uint32, error) {
				return 1, nil
			},
			IsEnabledCalled: func() bool {
				return withHistory
			},
		},
		ReceiptsRepository: &testscommon.ReceiptsRepositoryStub{},
	}, nil)
}

func TestMetaAPIBlockProcessor_GetBlockByHashInvalidHashShouldErr(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := genericMocks.NewStorerMock()

	metaAPIBlockProcessor := createMockMetaAPIProcessor(
		headerHash,
		storerMock,
		true,
		false,
	)

	blk, err := metaAPIBlockProcessor.GetBlockByHash([]byte("invalidHash"), api.BlockQueryOptions{})
	assert.Nil(t, blk)
	assert.Error(t, err)
}

func TestMetaAPIBlockProcessor_GetBlockByNonceInvalidNonceShouldErr(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := genericMocks.NewStorerMock()

	metaAPIBlockProcessor := createMockMetaAPIProcessor(
		headerHash,
		storerMock,
		true,
		false,
	)

	blk, err := metaAPIBlockProcessor.GetBlockByNonce(100, api.BlockQueryOptions{})
	assert.Nil(t, blk)
	assert.Error(t, err)
}

func TestMetaAPIBlockProcessor_GetBlockByRoundInvalidRoundShouldErr(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := genericMocks.NewStorerMock()

	metaAPIBlockProcessor := createMockMetaAPIProcessor(
		headerHash,
		storerMock,
		true,
		true,
	)

	blk, err := metaAPIBlockProcessor.GetBlockByRound(100, api.BlockQueryOptions{})
	assert.Nil(t, blk)
	assert.Error(t, err)
}

func TestMetaAPIBlockProcessor_GetBlockByHashFromHistoryNode(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(2)
	epoch := uint32(1)
	miniblockHeader := []byte("miniBlockHash")
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := genericMocks.NewStorerMockWithEpoch(epoch)
	metaAPIBlockProcessor := createMockMetaAPIProcessor(
		headerHash,
		storerMock,
		true,
		false,
	)

	header := &block.MetaBlock{
		Nonce: nonce,
		Round: round,
		Epoch: epoch,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Hash: miniblockHeader,
				Type: block.TxBlock,
			},
		},
		AccumulatedFees:        big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(10),
		DevFeesInEpoch:         big.NewInt(5),
	}
	headerBytes, _ := json.Marshal(header)
	_ = storerMock.Put(headerHash, headerBytes)

	expectedBlock := &api.Block{
		Nonce:           nonce,
		Round:           round,
		Shard:           core.MetachainShardId,
		Epoch:           epoch,
		Hash:            hex.EncodeToString(headerHash),
		NotarizedBlocks: []*api.NotarizedBlock{},
		MiniBlocks: []*api.MiniBlock{
			{
				Hash: hex.EncodeToString(miniblockHeader),
				Type: block.TxBlock.String(),
			},
		},
		AccumulatedFees:        "0",
		DeveloperFees:          "0",
		AccumulatedFeesInEpoch: "10",
		DeveloperFeesInEpoch:   "5",
		Status:                 BlockStatusOnChain,
	}

	blk, err := metaAPIBlockProcessor.GetBlockByHash(headerHash, api.BlockQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}

func TestMetaAPIBlockProcessor_GetBlockByHashFromGenesis(t *testing.T) {
	t.Parallel()

	nonce := uint64(0)
	round := uint64(0)
	epoch := uint32(0)
	miniblockHeader := []byte("miniBlockHash")
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := genericMocks.NewStorerMockWithEpoch(epoch)

	metaAPIBlockProcessor := createMockMetaAPIProcessor(
		headerHash,
		storerMock,
		true,
		true,
	)
	historyRepository := &dblookupext.HistoryRepositoryStub{
		GetEpochByHashCalled: func(hash []byte) (uint32, error) {
			return epoch, nil
		},
	}
	metaAPIBlockProcessor.historyRepo = historyRepository

	header := &block.MetaBlock{
		Nonce: nonce,
		Round: round,
		Epoch: epoch,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Hash: miniblockHeader,
				Type: block.TxBlock,
			},
		},
		AccumulatedFees:        big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(10),
		DevFeesInEpoch:         big.NewInt(5),
	}
	headerBytes, _ := json.Marshal(header)
	_ = storerMock.Put(headerHash, headerBytes)

	nonceConverterMock := mock.NewNonceHashConverterMock()
	nonceBytes := nonceConverterMock.ToByteSlice(nonce)
	_ = storerMock.Put(nonceBytes, headerHash)

	alteredHeader := &block.MetaBlock{
		Nonce: nonce,
		Round: round,
		Epoch: epoch,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Hash: miniblockHeader,
				Type: block.TxBlock,
			},
		},
		AccumulatedFees:        big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(10),
		DevFeesInEpoch:         big.NewInt(5),
	}
	alteredHeaderHash := make([]byte, 0)
	alteredHeaderHash = append(alteredHeaderHash, headerHash...)
	alteredHeaderHash = append(alteredHeaderHash, []byte(common.GenesisStorageSuffix)...)
	alteredHeaderBytes, _ := json.Marshal(alteredHeader)
	_ = storerMock.Put(alteredHeaderHash, alteredHeaderBytes)
	nonceBytes = append(nonceBytes, []byte(common.GenesisStorageSuffix)...)
	_ = storerMock.Put(nonceBytes, alteredHeaderHash)

	expectedBlock := &api.Block{
		Nonce:           nonce,
		Round:           round,
		Shard:           core.MetachainShardId,
		Epoch:           epoch,
		Hash:            hex.EncodeToString(headerHash),
		NotarizedBlocks: []*api.NotarizedBlock{},
		MiniBlocks: []*api.MiniBlock{
			{
				Hash: hex.EncodeToString(miniblockHeader),
				Type: block.TxBlock.String(),
			},
		},
		AccumulatedFees:        "0",
		DeveloperFees:          "0",
		AccumulatedFeesInEpoch: "10",
		DeveloperFeesInEpoch:   "5",
		Status:                 BlockStatusOnChain,
	}

	blk, err := metaAPIBlockProcessor.GetBlockByHash(headerHash, api.BlockQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}

func TestMetaAPIBlockProcessor_GetBlockByNonceFromHistoryNode(t *testing.T) {
	t.Parallel()

	testEpoch := uint32(7)
	testNonce := uint64(42)
	testRound := uint64(42)

	marshalizer := &marshal.GogoProtoMarshalizer{}
	storageService := genericMocks.NewChainStorerMock(testEpoch)
	historyRepository := &dblookupext.HistoryRepositoryStub{
		GetEpochByHashCalled: func(hash []byte) (uint32, error) {
			return testEpoch, nil
		},
	}

	processor := createMockMetaAPIProcessor(nil, nil, true, false)
	processor.store = storageService
	processor.marshalizer = marshalizer
	processor.historyRepo = historyRepository

	// Set up a miniblock
	miniblockHash := []byte{0xff}
	miniblock := &block.MiniBlock{
		Type:     block.TxBlock,
		TxHashes: [][]byte{},
	}

	// Set up a block containing the miniblock
	metablockHash := []byte{0xaa, 0xbb}
	metablock := &block.MetaBlock{
		Nonce: testNonce,
		Epoch: testEpoch,
		Round: testRound,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Hash: miniblockHash,
				Type: block.TxBlock,
			},
		},
		AccumulatedFees:        big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(10),
		DevFeesInEpoch:         big.NewInt(5),
	}

	// Store the miniblock and the metablock
	miniblockBytes, _ := processor.marshalizer.Marshal(miniblock)
	metablockBytes, _ := processor.marshalizer.Marshal(metablock)
	metablockNonceBytes := mock.NewNonceHashConverterMock().ToByteSlice(testNonce)
	_ = storageService.Miniblocks.PutInEpoch(miniblockHash, miniblockBytes, testEpoch)
	_ = storageService.Metablocks.PutInEpoch(metablockHash, metablockBytes, testEpoch)
	_ = storageService.MetaHdrNonce.PutInEpoch(metablockNonceBytes, metablockHash, testEpoch)

	expectedApiBlock := &api.Block{
		Epoch:           testEpoch,
		Nonce:           testNonce,
		Round:           testRound,
		Shard:           core.MetachainShardId,
		Hash:            hex.EncodeToString(metablockHash),
		NotarizedBlocks: []*api.NotarizedBlock{},
		MiniBlocks: []*api.MiniBlock{
			{
				Hash: hex.EncodeToString(miniblockHash),
				Type: block.TxBlock.String(),
			},
		},
		AccumulatedFees:        "0",
		DeveloperFees:          "0",
		AccumulatedFeesInEpoch: "10",
		DeveloperFeesInEpoch:   "5",
		Status:                 BlockStatusOnChain,
	}

	fetchedApiBlock, err := processor.GetBlockByHash(metablockHash, api.BlockQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, expectedApiBlock, fetchedApiBlock)
}

func TestMetaAPIBlockProcessor_GetBlockByNonceFromGenesis(t *testing.T) {
	t.Parallel()

	nonce := uint64(0)
	round := uint64(0)
	epoch := uint32(0)
	miniblockHeader := []byte("miniBlockHash")
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := genericMocks.NewStorerMockWithEpoch(epoch)

	metaAPIBlockProcessor := createMockMetaAPIProcessor(
		headerHash,
		storerMock,
		true,
		true,
	)
	historyRepository := &dblookupext.HistoryRepositoryStub{
		GetEpochByHashCalled: func(hash []byte) (uint32, error) {
			return epoch, nil
		},
	}
	metaAPIBlockProcessor.historyRepo = historyRepository

	header := &block.MetaBlock{
		Nonce: nonce,
		Round: round,
		Epoch: epoch,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Hash: miniblockHeader,
				Type: block.TxBlock,
			},
		},
		AccumulatedFees:        big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(10),
		DevFeesInEpoch:         big.NewInt(5),
	}
	headerBytes, _ := json.Marshal(header)
	_ = storerMock.Put(headerHash, headerBytes)

	nonceConverterMock := mock.NewNonceHashConverterMock()
	nonceBytes := nonceConverterMock.ToByteSlice(nonce)
	_ = storerMock.Put(nonceBytes, headerHash)

	alteredHeader := &block.MetaBlock{
		Nonce: nonce,
		Round: round,
		Epoch: epoch,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Hash: miniblockHeader,
				Type: block.TxBlock,
			},
		},
		AccumulatedFees:        big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(10),
		DevFeesInEpoch:         big.NewInt(5),
	}
	alteredHeaderHash := make([]byte, 0)
	alteredHeaderHash = append(alteredHeaderHash, headerHash...)
	alteredHeaderHash = append(alteredHeaderHash, []byte(common.GenesisStorageSuffix)...)
	alteredHeaderBytes, _ := json.Marshal(alteredHeader)
	_ = storerMock.Put(alteredHeaderHash, alteredHeaderBytes)
	nonceBytes = append(nonceBytes, []byte(common.GenesisStorageSuffix)...)
	_ = storerMock.Put(nonceBytes, alteredHeaderHash)

	expectedBlock := &api.Block{
		Nonce:           nonce,
		Round:           round,
		Shard:           core.MetachainShardId,
		Epoch:           epoch,
		Hash:            hex.EncodeToString(headerHash),
		NotarizedBlocks: []*api.NotarizedBlock{},
		MiniBlocks: []*api.MiniBlock{
			{
				Hash: hex.EncodeToString(miniblockHeader),
				Type: block.TxBlock.String(),
			},
		},
		AccumulatedFees:        "0",
		DeveloperFees:          "0",
		AccumulatedFeesInEpoch: "10",
		DeveloperFeesInEpoch:   "5",
		Status:                 BlockStatusOnChain,
	}

	blk, err := metaAPIBlockProcessor.GetBlockByNonce(nonce, api.BlockQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}

func TestMetaAPIBlockProcessor_GetBlockByRoundFromStorer(t *testing.T) {
	t.Parallel()

	round := uint64(2)
	epoch := uint32(1)
	miniblockHeader := []byte("miniBlockHash")
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := genericMocks.NewStorerMockWithEpoch(epoch)

	metaAPIBlockProcessor := createMockMetaAPIProcessor(
		headerHash,
		storerMock,
		true,
		true,
	)

	header := &block.MetaBlock{
		Round: round,
		Epoch: epoch,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Hash: miniblockHeader,
				Type: block.TxBlock,
			},
		},
		AccumulatedFees:        big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(10),
		DevFeesInEpoch:         big.NewInt(5),
	}

	headerBytes, _ := json.Marshal(header)
	_ = storerMock.Put(headerHash, headerBytes)

	uint64Converter := metaAPIBlockProcessor.uint64ByteSliceConverter
	roundBytes := uint64Converter.ToByteSlice(round)
	_ = storerMock.Put(roundBytes, headerHash)

	expectedBlock := &api.Block{
		Round:           round,
		Shard:           core.MetachainShardId,
		Epoch:           epoch,
		Hash:            hex.EncodeToString(headerHash),
		NotarizedBlocks: []*api.NotarizedBlock{},
		MiniBlocks: []*api.MiniBlock{
			{
				Hash: hex.EncodeToString(miniblockHeader),
				Type: block.TxBlock.String(),
			},
		},
		AccumulatedFees:        "0",
		DeveloperFees:          "0",
		AccumulatedFeesInEpoch: "10",
		DeveloperFeesInEpoch:   "5",
		Status:                 BlockStatusOnChain,
	}

	blk, err := metaAPIBlockProcessor.GetBlockByRound(round+1, api.BlockQueryOptions{})
	assert.NotNil(t, err)
	assert.Nil(t, blk)

	blk, err = metaAPIBlockProcessor.GetBlockByRound(round, api.BlockQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}

func TestMetaAPIBlockProcessor_GetBlockByHashFromHistoryNodeStatusReverted(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(2)
	epoch := uint32(1)
	miniblockHeader := []byte("miniBlockHash")
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := genericMocks.NewStorerMockWithEpoch(epoch)
	uint64Converter := mock.NewNonceHashConverterMock()

	metaAPIBlockProcessor := createMockMetaAPIProcessor(
		headerHash,
		storerMock,
		true,
		true,
	)

	header := &block.MetaBlock{
		Nonce: nonce,
		Round: round,
		Epoch: epoch,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Hash: miniblockHeader,
				Type: block.TxBlock,
			},
		},
		AccumulatedFees:        big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(10),
		DevFeesInEpoch:         big.NewInt(5),
	}
	headerBytes, _ := json.Marshal(header)
	_ = storerMock.Put(headerHash, headerBytes)

	nonceBytes := uint64Converter.ToByteSlice(nonce)
	correctHash := []byte("correct-hash")
	_ = storerMock.Put(nonceBytes, correctHash)

	expectedBlock := &api.Block{
		Nonce:           nonce,
		Round:           round,
		Shard:           core.MetachainShardId,
		Epoch:           epoch,
		Hash:            hex.EncodeToString(headerHash),
		NotarizedBlocks: []*api.NotarizedBlock{},
		MiniBlocks: []*api.MiniBlock{
			{
				Hash: hex.EncodeToString(miniblockHeader),
				Type: block.TxBlock.String(),
			},
		},
		AccumulatedFees:        "0",
		DeveloperFees:          "0",
		AccumulatedFeesInEpoch: "10",
		DeveloperFeesInEpoch:   "5",
		Status:                 BlockStatusReverted,
	}

	blk, err := metaAPIBlockProcessor.GetBlockByHash(headerHash, api.BlockQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}

func TestMetaAPIBlockProcessor_GetBlockByRound_GetBlockByNonce_EpochStartBlock(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(2)
	epoch := uint32(1)
	miniblockHeader := []byte("miniBlockHash")
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := genericMocks.NewStorerMockWithEpoch(epoch)

	metaAPIBlockProc := createMockMetaAPIProcessor(
		headerHash,
		storerMock,
		true,
		true,
	)

	header := &block.MetaBlock{
		Nonce: nonce,
		Round: round,
		Epoch: epoch,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Hash: miniblockHeader,
				Type: block.TxBlock,
			},
		},
		AccumulatedFees:        big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(10),
		DevFeesInEpoch:         big.NewInt(5),
		ShardInfo: []block.ShardData{
			{
				HeaderHash: []byte("hash"),
				ShardID:    0,
				Nonce:      1,
				Round:      2,
			},
		},
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					ShardID:               1,
					Nonce:                 1234,
					Round:                 1500,
					Epoch:                 10,
					HeaderHash:            []byte("hh"),
					RootHash:              []byte("rh"),
					ScheduledRootHash:     []byte("sch"),
					FirstPendingMetaBlock: []byte("fpmb"),
					LastFinishedMetaBlock: []byte("lfmb"),
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash:            []byte("mbh1"),
							SenderShardID:   0,
							ReceiverShardID: 1,
							Type:            block.TxBlock,
							Reserved:        []byte("rrr"),
						},
						{
							Hash:            []byte("mbh2"),
							SenderShardID:   1,
							ReceiverShardID: 2,
							Type:            block.SmartContractResultBlock,
							Reserved:        []byte("rrr"),
						},
					},
				},
				{
					ShardID:               2,
					Nonce:                 2200,
					Round:                 2340,
					Epoch:                 10,
					HeaderHash:            []byte("hh2"),
					RootHash:              []byte("rh2"),
					ScheduledRootHash:     []byte("sch2"),
					FirstPendingMetaBlock: []byte("fpmb2"),
					LastFinishedMetaBlock: []byte("lfmb2"),
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash:            []byte("mmm1"),
							SenderShardID:   1,
							ReceiverShardID: 0,
							Type:            block.TxBlock,
							Reserved:        []byte("rrr"),
						},
						{
							Hash:            []byte("mmm2"),
							SenderShardID:   0,
							ReceiverShardID: 2,
							Type:            block.SmartContractResultBlock,
							Reserved:        []byte("rrr"),
						},
					},
				},
			},
			Economics: block.Economics{
				TotalSupply:                      big.NewInt(100),
				TotalToDistribute:                big.NewInt(55),
				TotalNewlyMinted:                 big.NewInt(20),
				RewardsPerBlock:                  big.NewInt(15),
				RewardsForProtocolSustainability: big.NewInt(2),
				NodePrice:                        big.NewInt(10),
				PrevEpochStartRound:              222,
				PrevEpochStartHash:               []byte("prevEpoch"),
			},
		},
	}

	headerBytes, _ := json.Marshal(header)
	_ = storerMock.Put(headerHash, headerBytes)

	uint64Converter := metaAPIBlockProc.uint64ByteSliceConverter
	roundBytes := uint64Converter.ToByteSlice(round)
	nonceBytes := uint64Converter.ToByteSlice(nonce)
	_ = storerMock.Put(roundBytes, headerHash)
	_ = storerMock.Put(nonceBytes, headerHash)

	expectedBlock := &api.Block{
		Nonce: nonce,
		Round: round,
		Shard: core.MetachainShardId,
		Epoch: epoch,
		Hash:  hex.EncodeToString(headerHash),
		NotarizedBlocks: []*api.NotarizedBlock{
			{
				Hash:  "68617368",
				Shard: 0,
				Nonce: 1,
				Round: 2,
			},
		},
		MiniBlocks: []*api.MiniBlock{
			{
				Hash: hex.EncodeToString(miniblockHeader),
				Type: block.TxBlock.String(),
			},
		},
		AccumulatedFees:        "0",
		DeveloperFees:          "0",
		AccumulatedFeesInEpoch: "10",
		DeveloperFeesInEpoch:   "5",
		Status:                 BlockStatusOnChain,
		EpochStartInfo: &api.EpochStartInfo{
			TotalSupply:                      "100",
			TotalToDistribute:                "55",
			TotalNewlyMinted:                 "20",
			RewardsPerBlock:                  "15",
			RewardsForProtocolSustainability: "2",
			NodePrice:                        "10",
			PrevEpochStartRound:              222,
			PrevEpochStartHash:               "7072657645706f6368",
		},
		EpochStartShardsData: []*api.EpochStartShardData{
			{
				ShardID:               1,
				Epoch:                 10,
				Round:                 1500,
				Nonce:                 1234,
				HeaderHash:            "6868",
				RootHash:              "7268",
				ScheduledRootHash:     "736368",
				FirstPendingMetaBlock: "66706d62",
				LastFinishedMetaBlock: "6c666d62",
				PendingMiniBlockHeaders: []*api.MiniBlock{
					{
						Hash:             "6d626831",
						SourceShard:      0,
						DestinationShard: 1,
						Type:             "TxBlock",
					},
					{
						Hash:             "6d626832",
						SourceShard:      1,
						DestinationShard: 2,
						Type:             "SmartContractResultBlock",
					},
				},
			},
			{
				ShardID:               2,
				Epoch:                 10,
				Round:                 2340,
				Nonce:                 2200,
				HeaderHash:            "686832",
				RootHash:              "726832",
				ScheduledRootHash:     "73636832",
				FirstPendingMetaBlock: "66706d6232",
				LastFinishedMetaBlock: "6c666d6232",
				PendingMiniBlockHeaders: []*api.MiniBlock{
					{
						Hash:             "6d6d6d31",
						SourceShard:      1,
						DestinationShard: 0,
						Type:             "TxBlock",
					},
					{
						Hash:             "6d6d6d32",
						SourceShard:      0,
						DestinationShard: 2,
						Type:             "SmartContractResultBlock",
					},
				},
			},
		},
	}

	blk, err := metaAPIBlockProc.GetBlockByNonce(nonce, api.BlockQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)

	blk, err = metaAPIBlockProc.GetBlockByRound(round, api.BlockQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}
