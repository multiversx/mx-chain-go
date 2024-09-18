package bootstrap

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestBootStrapShardProcessor_applyCurrentShardIDOnMiniblocksCopy(t *testing.T) {
	t.Parallel()

	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockStorageEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.GeneralConfig = testscommon.GetGeneralConfig()

	expectedShardId := uint32(3)
	args.ImportDbConfig = config.ImportDbConfig{
		ImportDBTargetShardID: expectedShardId,
	}
	sesb, _ := NewStorageEpochStartBootstrap(args)
	bsr := &bootStrapShardProcessor{
		sesb.epochStartBootstrap,
	}

	metaBlock := &block.MetaBlock{
		Epoch: 2,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Hash:          []byte("hdrHash1"),
				SenderShardID: 1,
			},
			{
				Hash:          []byte("hdrHash2"),
				SenderShardID: 2,
			},
		},
	}
	err := bsr.applyCurrentShardIDOnMiniblocksCopy(metaBlock, sesb.importDbConfig.ImportDBTargetShardID)

	assert.Nil(t, err)
	for _, miniBlock := range metaBlock.GetMiniBlockHeaderHandlers() {
		assert.Equal(t, expectedShardId, miniBlock.GetSenderShardID())
	}
}
