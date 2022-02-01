package bootstrap

import (
	"os"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/nodeTypeProviderMock"
	"github.com/stretchr/testify/assert"
)

func TestNewMetaStorageHandler_InvalidConfigErr(t *testing.T) {
	gCfg := config.Config{}
	prefsConfig := config.PreferencesConfig{}
	coordinator := &mock.ShardCoordinatorStub{}
	pathManager := &testscommon.PathManagerStub{}
	marshalizer := &mock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	uit64Cvt := &mock.Uint64ByteSliceConverterMock{}
	nodeTypeProvider := &nodeTypeProviderMock.NodeTypeProviderStub{}

	mtStrHandler, err := NewMetaStorageHandler(gCfg, prefsConfig, coordinator, pathManager, marshalizer, hasher, 1, uit64Cvt, nodeTypeProvider)
	assert.True(t, check.IfNil(mtStrHandler))
	assert.NotNil(t, err)
}

func TestNewMetaStorageHandler_CreateForMetaErr(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	gCfg := testscommon.GetGeneralConfig()
	prefsConfig := config.PreferencesConfig{}
	coordinator := &mock.ShardCoordinatorStub{}
	pathManager := &testscommon.PathManagerStub{}
	marshalizer := &mock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	uit64Cvt := &mock.Uint64ByteSliceConverterMock{}
	nodeTypeProvider := &nodeTypeProviderMock.NodeTypeProviderStub{}

	mtStrHandler, err := NewMetaStorageHandler(gCfg, prefsConfig, coordinator, pathManager, marshalizer, hasher, 1, uit64Cvt, nodeTypeProvider)
	assert.False(t, check.IfNil(mtStrHandler))
	assert.Nil(t, err)
}

func TestMetaStorageHandler_saveLastHeader(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	gCfg := testscommon.GetGeneralConfig()
	prefsConfig := config.PreferencesConfig{}
	coordinator := &mock.ShardCoordinatorStub{}
	pathManager := &testscommon.PathManagerStub{}
	marshalizer := &mock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	uit64Cvt := &mock.Uint64ByteSliceConverterMock{}
	nodeTypeProvider := &nodeTypeProviderMock.NodeTypeProviderStub{}

	mtStrHandler, _ := NewMetaStorageHandler(gCfg, prefsConfig, coordinator, pathManager, marshalizer, hasher, 1, uit64Cvt, nodeTypeProvider)

	header := &block.MetaBlock{Nonce: 0}

	headerHash, _ := core.CalculateHash(marshalizer, hasher, header)
	expectedBootInfo := bootstrapStorage.BootstrapHeaderInfo{
		ShardId: core.MetachainShardId, Hash: headerHash,
	}

	bootHeaderInfo, err := mtStrHandler.saveLastHeader(header)
	assert.Nil(t, err)
	assert.Equal(t, expectedBootInfo, bootHeaderInfo)
}

func TestMetaStorageHandler_saveLastCrossNotarizedHeaders(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	gCfg := testscommon.GetGeneralConfig()
	prefsConfig := config.PreferencesConfig{}
	coordinator := &mock.ShardCoordinatorStub{}
	pathManager := &testscommon.PathManagerStub{}
	marshalizer := &mock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	uit64Cvt := &mock.Uint64ByteSliceConverterMock{}
	nodeTypeProvider := &nodeTypeProviderMock.NodeTypeProviderStub{}

	mtStrHandler, _ := NewMetaStorageHandler(gCfg, prefsConfig, coordinator, pathManager, marshalizer, hasher, 1, uit64Cvt, nodeTypeProvider)

	hdr1 := &block.Header{Nonce: 1}
	hdr2 := &block.Header{Nonce: 2}
	hdrHash1, _ := core.CalculateHash(marshalizer, hasher, hdr1)
	hdrHash2, _ := core.CalculateHash(marshalizer, hasher, hdr2)

	hdr3 := &block.MetaBlock{
		Nonce: 3,
		EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{
			{HeaderHash: hdrHash1}, {HeaderHash: hdrHash2},
		}},
	}

	hdrs := map[string]data.HeaderHandler{string(hdrHash1): hdr1, string(hdrHash2): hdr2}
	crossNotarizedHdrs, err := mtStrHandler.saveLastCrossNotarizedHeaders(hdr3, hdrs)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(crossNotarizedHdrs))
}

func TestMetaStorageHandler_saveTriggerRegistry(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	gCfg := testscommon.GetGeneralConfig()
	prefsConfig := config.PreferencesConfig{}
	coordinator := &mock.ShardCoordinatorStub{}
	pathManager := &testscommon.PathManagerStub{}
	marshalizer := &mock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	uit64Cvt := &mock.Uint64ByteSliceConverterMock{}
	nodeTypeProvider := &nodeTypeProviderMock.NodeTypeProviderStub{}

	mtStrHandler, _ := NewMetaStorageHandler(gCfg, prefsConfig, coordinator, pathManager, marshalizer, hasher, 1, uit64Cvt, nodeTypeProvider)

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: &block.MetaBlock{Nonce: 3},
		PreviousEpochStart:  &block.MetaBlock{Nonce: 2},
	}

	_, err := mtStrHandler.saveTriggerRegistry(components)
	assert.Nil(t, err)
}

func TestMetaStorageHandler_saveDataToStorage(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	gCfg := testscommon.GetGeneralConfig()
	prefsConfig := config.PreferencesConfig{}
	coordinator := &mock.ShardCoordinatorStub{}
	pathManager := &testscommon.PathManagerStub{}
	marshalizer := &mock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	uit64Cvt := &mock.Uint64ByteSliceConverterMock{}
	nodeTypeProvider := &nodeTypeProviderMock.NodeTypeProviderStub{}

	mtStrHandler, _ := NewMetaStorageHandler(gCfg, prefsConfig, coordinator, pathManager, marshalizer, hasher, 1, uit64Cvt, nodeTypeProvider)

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: &block.MetaBlock{Nonce: 3},
		PreviousEpochStart:  &block.MetaBlock{Nonce: 2},
	}

	err := mtStrHandler.SaveDataToStorage(components)
	assert.Nil(t, err)
}
