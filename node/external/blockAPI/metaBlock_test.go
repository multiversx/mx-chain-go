package blockAPI

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/outport/process/alteredaccounts/shared"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/state"
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
		ReceiptsRepository:      &testscommon.ReceiptsRepositoryStub{},
		AddressPubkeyConverter:  &testscommon.PubkeyConverterMock{},
		AlteredAccountsProvider: &testscommon.AlteredAccountsProviderStub{},
		AccountsRepository:      &state.AccountsRepositoryStub{},
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

func TestMetaAPIBlockProcessor_CheckAllFieldsAreAccountedFor(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(2)
	epoch := uint32(1)
	miniblockHeader := []byte("miniBlockHash")
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")
	prevHash := []byte("prevHash")
	prevRandSeed := []byte("prevRandSeed")
	randSeed := []byte("randSeed")
	pubkeysBitmap := []byte("pubkeysBitmap")
	timestamp := uint64(12345678)
	stateRootHash := []byte("stateRootHash")
	accumulatedFees := big.NewInt(11)
	developerFees := big.NewInt(12)
	accumulatedFeesInEpoch := big.NewInt(13)
	developerFeesInEpoch := big.NewInt(14)
	numTxs := uint32(2)

	storerMock := genericMocks.NewStorerMockWithEpoch(epoch)
	metaAPIBlockProcessor := createMockMetaAPIProcessor(
		headerHash,
		storerMock,
		true,
		false,
	)

	header := &block.MetaBlock{
		Nonce:                  nonce,
		Epoch:                  epoch,
		Round:                  round,
		TimeStamp:              timestamp,
		ShardInfo:              nil,
		PeerInfo:               nil,
		Signature:              nil,
		LeaderSignature:        nil,
		PubKeysBitmap:          pubkeysBitmap,
		PrevHash:               prevHash,
		PrevRandSeed:           prevRandSeed,
		RandSeed:               randSeed,
		RootHash:               stateRootHash,
		ValidatorStatsRootHash: nil,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Hash:    miniblockHeader,
				Type:    block.TxBlock,
				TxCount: numTxs,
			},
		},
		ReceiptsHash:           nil,
		EpochStart:             block.EpochStart{},
		ChainID:                nil,
		SoftwareVersion:        nil,
		AccumulatedFees:        accumulatedFees,
		AccumulatedFeesInEpoch: accumulatedFeesInEpoch,
		DeveloperFees:          developerFees,
		DevFeesInEpoch:         developerFeesInEpoch,
		TxCount:                numTxs,
		Reserved:               nil,
	}
	headerBytes, _ := json.Marshal(header)
	_ = storerMock.Put(headerHash, headerBytes)

	expectedBlock := &api.Block{
		Nonce:                  nonce,
		Round:                  round,
		Epoch:                  epoch,
		Shard:                  core.MetachainShardId,
		NumTxs:                 numTxs,
		Hash:                   hex.EncodeToString(headerHash),
		PrevBlockHash:          hex.EncodeToString(prevHash),
		StateRootHash:          hex.EncodeToString(stateRootHash),
		AccumulatedFees:        accumulatedFees.String(),
		DeveloperFees:          developerFees.String(),
		AccumulatedFeesInEpoch: accumulatedFeesInEpoch.String(),
		DeveloperFeesInEpoch:   developerFeesInEpoch.String(),
		Status:                 BlockStatusOnChain,
		RandSeed:               hex.EncodeToString(randSeed),
		PrevRandSeed:           hex.EncodeToString(prevRandSeed),
		Timestamp:              time.Duration(timestamp),
		NotarizedBlocks:        []*api.NotarizedBlock{},
		MiniBlocks: []*api.MiniBlock{
			{
				Hash: hex.EncodeToString(miniblockHeader),
				Type: block.TxBlock.String(),
			},
		},
		EpochStartInfo:       nil,
		EpochStartShardsData: nil,
		ScheduledData:        nil,
	}

	blk, err := metaAPIBlockProcessor.GetBlockByHash(headerHash, api.BlockQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)

	expectedBlockValue := reflect.ValueOf(expectedBlock)
	expectedBlockIndirect := reflect.Indirect(expectedBlockValue)
	expectedBlockType := expectedBlockIndirect.Type()

	numFieldsOnBlockOnGoCore1125 := 21
	assert.Equal(t, numFieldsOnBlockOnGoCore1125, expectedBlockType.NumField(), "please add the new fields to this test and update the numFields")
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

func TestMetaAPIBlockProcessor_GetAlteredAccountsForBlock(t *testing.T) {
	t.Parallel()

	t.Run("header not found in storage - should err", func(t *testing.T) {
		t.Parallel()

		headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

		storerMock := genericMocks.NewStorerMockWithEpoch(1)
		metaAPIBlockProc := createMockMetaAPIProcessor(
			headerHash,
			storerMock,
			true,
			true,
		)

		res, err := metaAPIBlockProc.GetAlteredAccountsForBlock(api.GetAlteredAccountsForBlockOptions{
			GetBlockParameters: api.GetBlockParameters{
				RequestType: api.BlockFetchTypeByHash,
				Hash:        headerHash,
			},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
		require.Nil(t, res)
	})

	t.Run("get altered account by block hash - should work", func(t *testing.T) {
		t.Parallel()

		marshaller := &testscommon.MarshalizerMock{}
		headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")
		mbHash := []byte("mb-hash")
		txHash0, txHash1 := []byte("tx-hash-0"), []byte("tx-hash-1")

		mbhReserved := block.MiniBlockHeaderReserved{}

		mbhReserved.IndexOfLastTxProcessed = 1
		reserved, _ := mbhReserved.Marshal()

		metaBlock := &block.MetaBlock{
			Nonce: 37,
			Epoch: 1,
			MiniBlockHeaders: []block.MiniBlockHeader{
				{
					Hash:     mbHash,
					Reserved: reserved,
				},
			},
		}
		miniBlock := &block.MiniBlock{
			TxHashes: [][]byte{txHash0, txHash1},
		}
		tx0 := &transaction.Transaction{
			SndAddr: []byte("addr0"),
			RcvAddr: []byte("addr1"),
		}
		tx1 := &transaction.Transaction{
			SndAddr: []byte("addr2"),
			RcvAddr: []byte("addr3"),
		}
		miniBlockBytes, _ := marshaller.Marshal(miniBlock)
		metaBlockBytes, _ := marshaller.Marshal(metaBlock)
		tx0Bytes, _ := marshaller.Marshal(tx0)
		tx1Bytes, _ := marshaller.Marshal(tx1)

		storerMock := genericMocks.NewStorerMockWithEpoch(1)
		_ = storerMock.Put(headerHash, metaBlockBytes)
		_ = storerMock.Put(mbHash, miniBlockBytes)
		_ = storerMock.Put(txHash0, tx0Bytes)
		_ = storerMock.Put(txHash1, tx1Bytes)

		metaAPIBlockProc := createMockMetaAPIProcessor(
			headerHash,
			storerMock,
			true,
			true,
		)

		metaAPIBlockProc.apiTransactionHandler = &mock.TransactionAPIHandlerStub{
			UnmarshalTransactionCalled: func(txBytes []byte, _ transaction.TxType) (*transaction.ApiTransactionResult, error) {
				var tx transaction.Transaction
				_ = marshaller.Unmarshal(&tx, txBytes)

				return &transaction.ApiTransactionResult{
					Type:     "normal",
					Sender:   hex.EncodeToString(tx.SndAddr),
					Receiver: hex.EncodeToString(tx.RcvAddr),
				}, nil
			},
		}
		metaAPIBlockProc.txStatusComputer = &mock.StatusComputerStub{}

		metaAPIBlockProc.logsFacade = &testscommon.LogsFacadeStub{}
		metaAPIBlockProc.alteredAccountsProvider = &testscommon.AlteredAccountsProviderStub{
			ExtractAlteredAccountsFromPoolCalled: func(outportPool *outportcore.Pool, options shared.AlteredAccountsOptions) (map[string]*outportcore.AlteredAccount, error) {
				retMap := map[string]*outportcore.AlteredAccount{}
				for _, tx := range outportPool.Txs {
					retMap[string(tx.GetSndAddr())] = &outportcore.AlteredAccount{
						Address: string(tx.GetSndAddr()),
						Balance: "10",
					}
				}

				return retMap, nil
			},
		}

		res, err := metaAPIBlockProc.GetAlteredAccountsForBlock(api.GetAlteredAccountsForBlockOptions{
			GetBlockParameters: api.GetBlockParameters{
				RequestType: api.BlockFetchTypeByHash,
				Hash:        headerHash,
			},
		})
		require.NoError(t, err)
		require.True(t, areAlteredAccountsResponsesTheSame([]*outportcore.AlteredAccount{
			{
				Address: "addr0",
				Balance: "10",
			},
			{
				Address: "addr2",
				Balance: "10",
			},
		},
			res))
	})
}
