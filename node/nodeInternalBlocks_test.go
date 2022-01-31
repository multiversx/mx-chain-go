package node_test

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateInternalBlockProcessor_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	uint64Converter := mock.NewNonceHashConverterMock()
	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.UInt64ByteSliceConv = uint64Converter
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.HistoryRepositoryInternal = &dblookupext.HistoryRepositoryStub{}
	processComponentsMock.ShardCoord = &testscommon.ShardsCoordinatorMock{}
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.Store = nil

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
	)

	err := n.CreateInternalBlockProcessor()
	assert.Error(t, err)
}

func TestCreateInternalBlockProcessor_NilUint64ByteSliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.UInt64ByteSliceConv = nil
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.HistoryRepositoryInternal = &dblookupext.HistoryRepositoryStub{}
	processComponentsMock.ShardCoord = &testscommon.ShardsCoordinatorMock{}
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.Store = &mock.ChainStorerMock{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
	)

	err := n.CreateInternalBlockProcessor()
	assert.Error(t, err)
}

// ---- GetInternalMetaBlockByHash

func TestGetInternalMetaBlockByHash_WrongEncodedHashShoudFail(t *testing.T) {
	t.Parallel()

	historyProc := &dblookupext.HistoryRepositoryStub{}
	uint64Converter := mock.NewNonceHashConverterMock()
	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.UInt64ByteSliceConv = uint64Converter
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.HistoryRepositoryInternal = historyProc
	processComponentsMock.ShardCoord = &testscommon.ShardsCoordinatorMock{
		NoShards:     1,
		CurrentShard: core.MetachainShardId,
	}
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.Store = &mock.ChainStorerMock{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
	)

	err := n.CreateInternalBlockProcessor()
	require.Nil(t, err)

	blk, err := n.GetInternalMetaBlockByHash(common.ApiOutputFormatInternal, "wronghashformat")
	assert.Error(t, err)
	assert.Nil(t, blk)
}

func TestGetInternalMetaBlockByHash_MetaChainOnlyEndpointShoudFail(t *testing.T) {
	t.Parallel()

	historyProc := &dblookupext.HistoryRepositoryStub{}
	uint64Converter := mock.NewNonceHashConverterMock()
	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.UInt64ByteSliceConv = uint64Converter
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.HistoryRepositoryInternal = historyProc
	processComponentsMock.ShardCoord = &testscommon.ShardsCoordinatorMock{
		NoShards:     1,
		CurrentShard: 0,
	}
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.Store = &mock.ChainStorerMock{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
	)

	blk, err := n.GetInternalMetaBlockByHash(common.ApiOutputFormatInternal, "dummyhash")
	assert.Equal(t, node.ErrMetachainOnlyEndpoint, err)
	assert.Nil(t, blk)
}

func TestGetInternalMetaBlockByHash_ShouldWork(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(2)
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	n, headerBytes := prepareInternalBlockNode(nonce, round, headerHash, core.MetachainShardId)
	err := n.CreateInternalBlockProcessor()
	require.Nil(t, err)

	blk, err := n.GetInternalMetaBlockByHash(common.ApiOutputFormatProto, hex.EncodeToString(headerHash))
	assert.Nil(t, err)
	assert.Equal(t, headerBytes, blk)
}

// ---- GetInternalMetaBlockByNonce

func TestGetInternalMetaBlockByNonce_MetaChainOnlyEndpointShoudFail(t *testing.T) {
	t.Parallel()

	historyProc := &dblookupext.HistoryRepositoryStub{}
	uint64Converter := mock.NewNonceHashConverterMock()
	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.UInt64ByteSliceConv = uint64Converter
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.HistoryRepositoryInternal = historyProc
	processComponentsMock.ShardCoord = &testscommon.ShardsCoordinatorMock{
		NoShards:     1,
		CurrentShard: 0,
	}
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.Store = &mock.ChainStorerMock{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
	)

	blk, err := n.GetInternalMetaBlockByNonce(common.ApiOutputFormatInternal, 1)
	assert.Equal(t, node.ErrMetachainOnlyEndpoint, err)
	assert.Nil(t, blk)
}

func TestGetInternalMetaBlockByNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	nonce := uint64(1)
	round := uint64(2)

	n, headerBytes := prepareInternalBlockNode(nonce, round, headerHash, core.MetachainShardId)
	err := n.CreateInternalBlockProcessor()
	require.Nil(t, err)

	blk, err := n.GetInternalMetaBlockByNonce(common.ApiOutputFormatProto, nonce)
	assert.Nil(t, err)
	assert.Equal(t, headerBytes, blk)
}

// ---- GetInternalMetaBlockByRound

func TestGetInternalMetaBlockByRound_MetaChainOnlyEndpointShoudFail(t *testing.T) {
	t.Parallel()

	historyProc := &dblookupext.HistoryRepositoryStub{}
	uint64Converter := mock.NewNonceHashConverterMock()
	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.UInt64ByteSliceConv = uint64Converter
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.HistoryRepositoryInternal = historyProc
	processComponentsMock.ShardCoord = &testscommon.ShardsCoordinatorMock{
		NoShards:     1,
		CurrentShard: 0,
	}
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.Store = &mock.ChainStorerMock{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
	)

	blk, err := n.GetInternalMetaBlockByRound(common.ApiOutputFormatInternal, 1)
	assert.Equal(t, node.ErrMetachainOnlyEndpoint, err)
	assert.Nil(t, blk)
}

func TestGetInternalMetaBlockByRound_ShouldWork(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(2)

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	n, headerBytes := prepareInternalBlockNode(nonce, round, headerHash, core.MetachainShardId)
	err := n.CreateInternalBlockProcessor()
	require.Nil(t, err)

	blk, err := n.GetInternalMetaBlockByRound(common.ApiOutputFormatProto, round)
	assert.Nil(t, err)
	assert.Equal(t, headerBytes, blk)
}

func prepareInternalBlockNode(nonce uint64, round uint64, headerHash []byte, shardId uint32) (*node.Node, []byte) {
	historyProc := &dblookupext.HistoryRepositoryStub{
		IsEnabledCalled: func() bool {
			return true
		},
		GetEpochByHashCalled: func(_ []byte) (uint32, error) {
			return 1, nil
		},
	}
	uint64Converter := mock.NewNonceHashConverterMock()
	coreComponentsMock := getDefaultCoreComponents()
	storerMock := mock.NewStorerMock()
	coreComponentsMock.UInt64ByteSliceConv = uint64Converter
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.HistoryRepositoryInternal = historyProc
	processComponentsMock.ShardCoord = &testscommon.ShardsCoordinatorMock{
		NoShards:     1,
		CurrentShard: shardId,
	}
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.Store = &mock.ChainStorerMock{
		GetStorerCalled: func(_ dataRetriever.UnitType) storage.Storer {
			return storerMock
		},
		GetCalled: func(_ dataRetriever.UnitType, _ []byte) ([]byte, error) {
			return headerHash, nil
		},
	}

	var blockBytes []byte
	if shardId == core.MetachainShardId {
		header := &block.MetaBlock{
			Nonce: nonce,
			Round: round,
		}
		blockBytes, _ = json.Marshal(header)
	} else {
		header := &block.Header{
			Nonce: nonce,
			Round: round,
		}
		blockBytes, _ = json.Marshal(header)
	}
	_ = storerMock.Put(headerHash, blockBytes)

	nonceBytes := uint64Converter.ToByteSlice(nonce)
	_ = storerMock.Put(nonceBytes, headerHash)

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
	)

	return n, blockBytes
}

// ---- GetInternalShardBlockByHash

func TestGetInternalShardBlockByHash_WrongEncodedHashShoudFail(t *testing.T) {
	t.Parallel()

	historyProc := &dblookupext.HistoryRepositoryStub{}
	uint64Converter := mock.NewNonceHashConverterMock()
	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.UInt64ByteSliceConv = uint64Converter
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.HistoryRepositoryInternal = historyProc
	processComponentsMock.ShardCoord = &testscommon.ShardsCoordinatorMock{
		NoShards:     1,
		CurrentShard: 0,
	}
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.Store = &mock.ChainStorerMock{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
	)

	err := n.CreateInternalBlockProcessor()
	require.Nil(t, err)

	blk, err := n.GetInternalShardBlockByHash(common.ApiOutputFormatInternal, "wronghashformat")
	assert.Error(t, err)
	assert.Nil(t, blk)
}

func TestGetInternalShardBlockByHash_ShardOnlyEndpointShoudFail(t *testing.T) {
	t.Parallel()

	historyProc := &dblookupext.HistoryRepositoryStub{}
	uint64Converter := mock.NewNonceHashConverterMock()
	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.UInt64ByteSliceConv = uint64Converter
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.HistoryRepositoryInternal = historyProc
	processComponentsMock.ShardCoord = &testscommon.ShardsCoordinatorMock{
		NoShards:     1,
		CurrentShard: core.MetachainShardId,
	}
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.Store = &mock.ChainStorerMock{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
	)

	blk, err := n.GetInternalShardBlockByHash(common.ApiOutputFormatInternal, "dummyhash")
	assert.Equal(t, node.ErrShardOnlyEndpoint, err)
	assert.Nil(t, blk)
}

func TestGetInternalShardBlockByHash_ShouldWork(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(2)
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	n, headerBytes := prepareInternalBlockNode(nonce, round, headerHash, 0)
	err := n.CreateInternalBlockProcessor()
	require.Nil(t, err)

	blk, err := n.GetInternalShardBlockByHash(common.ApiOutputFormatProto, hex.EncodeToString(headerHash))
	assert.Nil(t, err)
	assert.Equal(t, headerBytes, blk)
}

// ---- GetInternalShardBlockByNonce

func TestGetInternalShardBlockByNonce_ShardOnlyEndpointShoudFail(t *testing.T) {
	t.Parallel()

	historyProc := &dblookupext.HistoryRepositoryStub{}
	uint64Converter := mock.NewNonceHashConverterMock()
	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.UInt64ByteSliceConv = uint64Converter
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.HistoryRepositoryInternal = historyProc
	processComponentsMock.ShardCoord = &testscommon.ShardsCoordinatorMock{
		NoShards:     1,
		CurrentShard: core.MetachainShardId,
	}
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.Store = &mock.ChainStorerMock{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
	)

	blk, err := n.GetInternalShardBlockByNonce(common.ApiOutputFormatInternal, 1)
	assert.Equal(t, node.ErrShardOnlyEndpoint, err)
	assert.Nil(t, blk)
}

func TestGetInternalShardBlockByNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	nonce := uint64(1)
	round := uint64(2)

	n, headerBytes := prepareInternalBlockNode(nonce, round, headerHash, 0)
	err := n.CreateInternalBlockProcessor()
	require.Nil(t, err)

	blk, err := n.GetInternalShardBlockByNonce(common.ApiOutputFormatProto, nonce)
	assert.Nil(t, err)
	assert.Equal(t, headerBytes, blk)
}

// ---- GetInternalShardBlockByRound

func TestGetInternalShardBlockByRound_ShardOnlyEndpointShoudFail(t *testing.T) {
	t.Parallel()

	historyProc := &dblookupext.HistoryRepositoryStub{}
	uint64Converter := mock.NewNonceHashConverterMock()
	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.UInt64ByteSliceConv = uint64Converter
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.HistoryRepositoryInternal = historyProc
	processComponentsMock.ShardCoord = &testscommon.ShardsCoordinatorMock{
		NoShards:     1,
		CurrentShard: core.MetachainShardId,
	}
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.Store = &mock.ChainStorerMock{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
	)

	blk, err := n.GetInternalShardBlockByRound(common.ApiOutputFormatInternal, 1)
	assert.Equal(t, node.ErrShardOnlyEndpoint, err)
	assert.Nil(t, blk)
}

func TestGetInternalShardBlockByRound_ShouldWork(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(2)

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	n, headerBytes := prepareInternalBlockNode(nonce, round, headerHash, 0)
	err := n.CreateInternalBlockProcessor()
	require.Nil(t, err)

	blk, err := n.GetInternalShardBlockByRound(common.ApiOutputFormatProto, round)
	assert.Nil(t, err)
	assert.Equal(t, headerBytes, blk)
}
