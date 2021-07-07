package bootstrap

import (
	"os"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/nodeTypeProviderMock"
	"github.com/stretchr/testify/assert"
)

func TestNewShardStorageHandler_ShouldWork(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	gCfg := testscommon.GetGeneralConfig()
	prefsConfig := config.PreferencesConfig{}
	coordinator := &mock.ShardCoordinatorStub{}
	pathManager := &testscommon.PathManagerStub{}
	marshalizer := &mock.MarshalizerMock{}
	hasher := &mock.HasherMock{}
	uit64Cvt := &mock.Uint64ByteSliceConverterMock{}
	nodeTypeProvider := &nodeTypeProviderMock.NodeTypeProviderStub{}

	shardStrHandler, err := NewShardStorageHandler(gCfg, prefsConfig, coordinator, pathManager, marshalizer, hasher, 1, uit64Cvt, nodeTypeProvider)
	assert.False(t, check.IfNil(shardStrHandler))
	assert.Nil(t, err)
}

func TestShardStorageHandler_SaveDataToStorageShardDataNotFound(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	gCfg := testscommon.GetGeneralConfig()
	prefsConfig := config.PreferencesConfig{}
	coordinator := &mock.ShardCoordinatorStub{}
	pathManager := &testscommon.PathManagerStub{}
	marshalizer := &mock.MarshalizerMock{}
	hasher := &mock.HasherMock{}
	uit64Cvt := &mock.Uint64ByteSliceConverterMock{}
	nodeTypeProvider := &nodeTypeProviderMock.NodeTypeProviderStub{}

	shardStrHandler, _ := NewShardStorageHandler(gCfg, prefsConfig, coordinator, pathManager, marshalizer, hasher, 1, uit64Cvt, nodeTypeProvider)

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: &block.MetaBlock{Epoch: 1},
		PreviousEpochStart:  &block.MetaBlock{Epoch: 1},
		ShardHeader:         &block.Header{Nonce: 1},
	}

	err := shardStrHandler.SaveDataToStorage(components)
	assert.Equal(t, epochStart.ErrEpochStartDataForShardNotFound, err)
}

func TestShardStorageHandler_SaveDataToStorageMissingHeader(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	gCfg := testscommon.GetGeneralConfig()
	prefsConfig := config.PreferencesConfig{}
	coordinator := &mock.ShardCoordinatorStub{}
	pathManager := &testscommon.PathManagerStub{}
	marshalizer := &mock.MarshalizerMock{}
	hasher := &mock.HasherMock{}
	uit64Cvt := &mock.Uint64ByteSliceConverterMock{}
	nodeTypeProvider := &nodeTypeProviderMock.NodeTypeProviderStub{}

	shardStrHandler, _ := NewShardStorageHandler(gCfg, prefsConfig, coordinator, pathManager, marshalizer, hasher, 1, uit64Cvt, nodeTypeProvider)

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: &block.MetaBlock{
			Epoch: 1,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{ShardID: 0, Nonce: 1},
				},
			},
		},
		PreviousEpochStart: &block.MetaBlock{Epoch: 1},
		ShardHeader:        &block.Header{Nonce: 1},
	}

	err := shardStrHandler.SaveDataToStorage(components)
	assert.Equal(t, epochStart.ErrMissingHeader, err)
}

func TestShardStorageHandler_SaveDataToStorage(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	gCfg := testscommon.GetGeneralConfig()
	prefsConfig := config.PreferencesConfig{}
	coordinator := &mock.ShardCoordinatorStub{}
	pathManager := &testscommon.PathManagerStub{}
	marshalizer := &mock.MarshalizerMock{}
	hasher := &mock.HasherMock{}
	uit64Cvt := &mock.Uint64ByteSliceConverterMock{}
	nodeTypeProvider := &nodeTypeProviderMock.NodeTypeProviderStub{}

	shardStrHandler, _ := NewShardStorageHandler(gCfg, prefsConfig, coordinator, pathManager, marshalizer, hasher, 1, uit64Cvt, nodeTypeProvider)

	hash1 := []byte("hash1")
	hdr1 := block.MetaBlock{
		Nonce: 1,
	}
	headers := map[string]data.HeaderHandler{
		string(hash1): &hdr1,
	}

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: &block.MetaBlock{
			Epoch: 1,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{ShardID: 0, Nonce: 1, FirstPendingMetaBlock: hash1, LastFinishedMetaBlock: hash1},
				},
			},
		},
		PreviousEpochStart: &block.MetaBlock{Epoch: 1},
		ShardHeader:        &block.Header{Nonce: 1},
		Headers:            headers,
		NodesConfig:        &sharding.NodesCoordinatorRegistry{},
	}

	err := shardStrHandler.SaveDataToStorage(components)
	assert.Nil(t, err)
}

func TestGetAllMiniBlocksWithDst(t *testing.T) {
	t.Parallel()

	hash1 := []byte("hash1")
	hash2 := []byte("hash2")
	shardMiniBlockHeader := block.MiniBlockHeader{SenderShardID: 1, Hash: hash1}
	miniBlockHeader := block.MiniBlockHeader{SenderShardID: 1, Hash: hash2}
	metablock := &block.MetaBlock{
		ShardInfo: []block.ShardData{
			{
				ShardID: 1,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					shardMiniBlockHeader,
					{SenderShardID: 0},
				},
			},
			{ShardID: 0},
		},
		MiniBlockHeaders: []block.MiniBlockHeader{
			{SenderShardID: 0},
			miniBlockHeader,
		},
	}

	shardMbHeaders := getAllMiniBlocksWithDst(metablock, 0)
	assert.Equal(t, shardMbHeaders[string(hash1)], shardMiniBlockHeader)
	assert.NotNil(t, shardMbHeaders[string(hash2)])
}
