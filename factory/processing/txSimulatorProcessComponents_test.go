package processing_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/factory/processing"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagedProcessComponents_createAPITransactionEvaluator(t *testing.T) {
	t.Parallel()

	shardCoordinatorForShardID2 := mock.NewMultiShardsCoordinatorMock(3)
	shardCoordinatorForShardID2.CurrentShard = 2

	shardCoordinatorForMetachain := mock.NewMultiShardsCoordinatorMock(3)
	shardCoordinatorForMetachain.CurrentShard = core.MetachainShardId

	// no further t.Parallel as these tests are quite heavy (they open netMessengers and other components that start a lot of goroutines)
	t.Run("invalid VMOutputCacher config should error", func(t *testing.T) {
		processArgs := components.GetProcessComponentsFactoryArgs(shardCoordinatorForShardID2)
		processArgs.Config.VMOutputCacher.Type = "invalid"
		pcf, err := processing.NewProcessComponentsFactory(processArgs)
		require.Nil(t, err)

		apiTransactionEvaluator, vmContainerFactory, err := pcf.CreateAPITransactionEvaluator(&testscommon.EpochStartTriggerStub{})
		assert.NotNil(t, err)
		assert.True(t, check.IfNil(apiTransactionEvaluator))
		assert.True(t, check.IfNil(vmContainerFactory))
		assert.Contains(t, err.Error(), "not supported cache type")
	})
	t.Run("should work for shard", func(t *testing.T) {
		processArgs := components.GetProcessComponentsFactoryArgs(shardCoordinatorForShardID2)
		pcf, err := processing.NewProcessComponentsFactory(processArgs)
		require.Nil(t, err)

		apiTransactionEvaluator, vmContainerFactory, err := pcf.CreateAPITransactionEvaluator(&testscommon.EpochStartTriggerStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(apiTransactionEvaluator))
		assert.False(t, check.IfNil(vmContainerFactory))
		require.NoError(t, vmContainerFactory.Close())
	})
	t.Run("should work for metachain", func(t *testing.T) {
		processArgs := components.GetProcessComponentsFactoryArgs(shardCoordinatorForMetachain)
		pcf, err := processing.NewProcessComponentsFactory(processArgs)
		require.Nil(t, err)

		apiTransactionEvaluator, vmContainerFactory, err := pcf.CreateAPITransactionEvaluator(&testscommon.EpochStartTriggerStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(apiTransactionEvaluator))
		assert.False(t, check.IfNil(vmContainerFactory))
		require.NoError(t, vmContainerFactory.Close())
	})
}
