package node_test

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/blockAPI"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInternalMiniBlock_NotFoundInStorerShouldFail(t *testing.T) {
	t.Parallel()

	uint64Converter := mock.NewNonceHashConverterMock()
	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.UInt64ByteSliceConv = uint64Converter
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.HistoryRepositoryInternal = &dblookupext.HistoryRepositoryStub{}
	processComponentsMock.ShardCoord = &testscommon.ShardsCoordinatorMock{}
	dataComponentsMock := getDefaultDataComponents()

	expectedErr := errors.New("failed to get from storage")
	storerMock := &testscommon.StorerStub{
		GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
			return nil, expectedErr
		},
	}
	dataComponentsMock.Store = &mock.ChainStorerMock{
		GetStorerCalled: func(_ dataRetriever.UnitType) storage.Storer {
			return storerMock
		},
	}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
	)

	miniBlock, err := n.GetInternalMiniBlock(common.ApiOutputFormatProto, hex.EncodeToString([]byte("dummyhash")))
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, miniBlock)
}

func TestInternalMiniBlock_NotInStorerWhenHistoryRepoDisabledShouldFail(t *testing.T) {
	t.Parallel()

	historyProc := &dblookupext.HistoryRepositoryStub{
		IsEnabledCalled: func() bool {
			return false
		},
	}

	expectedErr := errors.New("failed to get from storage")
	storerMock := &testscommon.StorerStub{}
	coreComponentsMock := getDefaultCoreComponents()
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.HistoryRepositoryInternal = historyProc
	processComponentsMock.ShardCoord = &testscommon.ShardsCoordinatorMock{}
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.Store = &mock.ChainStorerMock{
		GetStorerCalled: func(_ dataRetriever.UnitType) storage.Storer {
			return storerMock
		},
		GetCalled: func(_ dataRetriever.UnitType, key []byte) ([]byte, error) {
			return nil, expectedErr
		},
	}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
	)

	miniBlock, err := n.GetInternalMiniBlock(common.ApiOutputFormatProto, hex.EncodeToString([]byte("dummyhash")))
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, miniBlock)
}

func TestInternalMiniBlock_GetEpochWhenHistoryRepoEnabledShouldFail(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("failed to get epoch")
	historyProc := &dblookupext.HistoryRepositoryStub{
		IsEnabledCalled: func() bool {
			return true
		},
		GetEpochByHashCalled: func(_ []byte) (uint32, error) {
			return 0, expectedErr
		},
	}

	storerMock := &testscommon.StorerStub{}
	coreComponentsMock := getDefaultCoreComponents()
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.HistoryRepositoryInternal = historyProc
	processComponentsMock.ShardCoord = &testscommon.ShardsCoordinatorMock{}
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.Store = &mock.ChainStorerMock{
		GetStorerCalled: func(_ dataRetriever.UnitType) storage.Storer {
			return storerMock
		},
	}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
	)

	miniBlock, err := n.GetInternalMiniBlock(common.ApiOutputFormatProto, hex.EncodeToString([]byte("dummyhash")))
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, miniBlock)
}

func TestGetInternalMiniBlock_WrongEncodedHashShoudFail(t *testing.T) {
	t.Parallel()

	historyProc := &dblookupext.HistoryRepositoryStub{}
	uint64Converter := mock.NewNonceHashConverterMock()
	coreComponentsMock := getDefaultCoreComponents()
	coreComponentsMock.UInt64ByteSliceConv = uint64Converter
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.HistoryRepositoryInternal = historyProc
	processComponentsMock.ShardCoord = &testscommon.ShardsCoordinatorMock{
		NoShards:     1,
		CurrentShard: 1,
	}
	dataComponentsMock := getDefaultDataComponents()
	dataComponentsMock.Store = &mock.ChainStorerMock{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponentsMock),
		node.WithProcessComponents(processComponentsMock),
		node.WithDataComponents(dataComponentsMock),
	)

	blk, err := n.GetInternalMiniBlock(common.ApiOutputFormatProto, "wronghashformat")
	assert.Error(t, err)
	assert.Nil(t, blk)
}

func TestGetInternalMiniBlock_InvalidOutputFormatShouldFail(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	n, _ := prepareInternalMiniBlockNode(headerHash)

	blk, err := n.GetInternalMiniBlock(2, hex.EncodeToString(headerHash))
	assert.Equal(t, blockAPI.ErrInvalidOutputFormat, err)
	assert.Nil(t, blk)
}

func TestGetInternalMiniBlock_RawDataFormatShouldWork(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	n, miniBlockBytes := prepareInternalMiniBlockNode(headerHash)

	blk, err := n.GetInternalMiniBlock(common.ApiOutputFormatProto, hex.EncodeToString(headerHash))
	assert.Nil(t, err)
	assert.Equal(t, miniBlockBytes, blk)
}

func TestGetInternalMiniBlock_InternalFormatShouldWork(t *testing.T) {
	t.Parallel()

	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	n, miniBlockBytes := prepareInternalMiniBlockNode(headerHash)
	miniBlock := &block.MiniBlock{}
	err := json.Unmarshal(miniBlockBytes, miniBlock)
	require.Nil(t, err)

	blk, err := n.GetInternalMiniBlock(common.ApiOutputFormatInternal, hex.EncodeToString(headerHash))
	assert.Nil(t, err)
	assert.Equal(t, miniBlock, blk)
}

func prepareInternalMiniBlockNode(headerHash []byte) (*node.Node, []byte) {
	nonce := uint64(1)

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
		CurrentShard: 1,
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

	miniBlock := &block.MiniBlock{
		ReceiverShardID: 1,
		SenderShardID:   1,
	}

	blockBytes, _ := json.Marshal(miniBlock)
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
