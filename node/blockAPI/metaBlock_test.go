package blockAPI

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestMetaAPIBlockProcessor_GetBlockByHash_InvalidHashShouldErr(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := mock.NewStorerMock()
	uint64Converter := mock.NewNonceHashConverterMock()
	metaAPIBlockProcessor := NewMetaApiBlockProcessor(
		&APIBlockProcessorArg{
			SelfShardID: core.MetachainShardId,
			Marshalizer: &mock.MarshalizerFake{},
			Store: &mock.ChainStorerMock{
				GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
					return storerMock
				},
				GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
					return headerHash, nil
				},
			},
			Uint64ByteSliceConverter: uint64Converter,
			HistoryRepo: &testscommon.HistoryRepositoryStub{
				IsEnabledCalled: func() bool {
					return true
				},
				GetEpochByHashCalled: func(hash []byte) (uint32, error) {
					return 1, nil
				},
			},
		},
	)

	blk, err := metaAPIBlockProcessor.GetBlockByHash([]byte("invalidHash"), false)
	assert.Nil(t, blk)
	assert.Error(t, err)
}

func TestMetaAPIBlockProcessor_GetBlockByNonce_InvalidNonceShouldErr(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := mock.NewStorerMock()
	uint64Converter := mock.NewNonceHashConverterMock()
	metaAPIBlockProcessor := NewMetaApiBlockProcessor(
		&APIBlockProcessorArg{
			SelfShardID: core.MetachainShardId,
			Marshalizer: &mock.MarshalizerFake{},
			Store: &mock.ChainStorerMock{
				GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
					return storerMock
				},
				GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
					return headerHash, nil
				},
			},
			Uint64ByteSliceConverter: uint64Converter,
			HistoryRepo: &testscommon.HistoryRepositoryStub{
				IsEnabledCalled: func() bool {
					return true
				},
				GetEpochByHashCalled: func(hash []byte) (uint32, error) {
					return 1, nil
				},
			},
		},
	)

	blk, err := metaAPIBlockProcessor.GetBlockByNonce(100, false)
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

	storerMock := mock.NewStorerMock()
	uint64Converter := mock.NewNonceHashConverterMock()
	metaAPIBlockProcessor := NewMetaApiBlockProcessor(
		&APIBlockProcessorArg{
			SelfShardID: core.MetachainShardId,
			Marshalizer: &mock.MarshalizerFake{},
			Store: &mock.ChainStorerMock{
				GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
					return storerMock
				},
				GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
					return headerHash, nil
				},
			},
			Uint64ByteSliceConverter: uint64Converter,
			HistoryRepo: &testscommon.HistoryRepositoryStub{
				IsEnabledCalled: func() bool {
					return true
				},
				GetEpochByHashCalled: func(hash []byte) (uint32, error) {
					return 1, nil
				},
			},
		},
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
	_ = storerMock.Put(nonceBytes, headerHash)

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

	blk, err := metaAPIBlockProcessor.GetBlockByHash(headerHash, false)
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}

func TestMetaAPIBlockProcessor_GetBlockByNonceFromHistoryNode(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(2)
	epoch := uint32(1)
	miniblockHeader := []byte("miniBlockHash")
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := mock.NewStorerMock()
	uint64Converter := mock.NewNonceHashConverterMock()
	metaAPIBlockProcessor := NewMetaApiBlockProcessor(
		&APIBlockProcessorArg{
			SelfShardID: core.MetachainShardId,
			Marshalizer: &mock.MarshalizerFake{},
			Store: &mock.ChainStorerMock{
				GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
					return storerMock
				},
				GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
					return headerHash, nil
				},
			},
			Uint64ByteSliceConverter: uint64Converter,
			HistoryRepo: &testscommon.HistoryRepositoryStub{
				IsEnabledCalled: func() bool {
					return true
				},
				GetEpochByHashCalled: func(hash []byte) (uint32, error) {
					return 1, nil
				},
			},
		},
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

	blk, err := metaAPIBlockProcessor.GetBlockByNonce(1, true)
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}

func TestMetaAPIBlockProcessor_GetBlockByHashFromHistoryNode_StatusReverted(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(2)
	epoch := uint32(1)
	miniblockHeader := []byte("miniBlockHash")
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := mock.NewStorerMock()
	uint64Converter := mock.NewNonceHashConverterMock()
	metaAPIBlockProcessor := NewMetaApiBlockProcessor(
		&APIBlockProcessorArg{
			SelfShardID: core.MetachainShardId,
			Marshalizer: &mock.MarshalizerFake{},
			Store: &mock.ChainStorerMock{
				GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
					return storerMock
				},
				GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
					return storerMock.Get(key)
				},
			},
			Uint64ByteSliceConverter: uint64Converter,
			HistoryRepo: &testscommon.HistoryRepositoryStub{
				IsEnabledCalled: func() bool {
					return true
				},
				GetEpochByHashCalled: func(hash []byte) (uint32, error) {
					return 1, nil
				},
			},
		},
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

	blk, err := metaAPIBlockProcessor.GetBlockByHash(headerHash, false)
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}
