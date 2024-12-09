package broadcast

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/alarm"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func createDefaultDelayedBroadcasterArgs() *ArgsDelayedBlockBroadcaster {
	return &ArgsDelayedBlockBroadcaster{
		ShardCoordinator:      &mock.ShardCoordinatorMock{},
		InterceptorsContainer: &testscommon.InterceptorsContainerStub{},
		HeadersSubscriber:     &testscommon.HeadersCacherStub{},
		LeaderCacheSize:       2,
		ValidatorCacheSize:    2,
		AlarmScheduler:        alarm.NewAlarmScheduler(),
	}
}

func TestNewSovereignDelayedBlockBroadcaster(t *testing.T) {
	t.Parallel()

	hdrKey := fmt.Sprintf("%s_%d", factory.ShardBlocksTopic, core.SovereignChainShardId)
	mbKey := fmt.Sprintf("%s_%d", factory.MiniBlocksTopic, core.SovereignChainShardId)

	wasHdrInterceptorSet := false
	hdrInterceptor := &testscommon.InterceptorStub{
		RegisterHandlerCalled: func(handler func(topic string, hash []byte, data interface{})) {
			require.Contains(t, getFunctionName(handler), "(*delayedBlockBroadcaster).interceptedHeader")
			wasHdrInterceptorSet = true
		},
	}

	wasMBInterceptorSet := false
	mbInterceptor := &testscommon.InterceptorStub{
		RegisterHandlerCalled: func(handler func(topic string, hash []byte, data interface{})) {
			require.Contains(t, getFunctionName(handler), "(*delayedBlockBroadcaster).interceptedMiniBlockData")
			wasMBInterceptorSet = true
		},
	}

	container := &testscommon.InterceptorsContainerStub{
		GetCalled: func(key string) (process.Interceptor, error) {
			switch key {
			case hdrKey:
				return hdrInterceptor, nil
			case mbKey:
				return mbInterceptor, nil
			default:
				require.Fail(t, "should not have get other container key")
			}

			return nil, nil
		},
	}

	args := createDefaultDelayedBroadcasterArgs()
	args.InterceptorsContainer = container
	sovBroadcaster, err := NewSovereignDelayedBlockBroadcaster(args)
	require.Nil(t, err)
	require.False(t, sovBroadcaster.IsInterfaceNil())

	require.True(t, wasHdrInterceptorSet)
	require.True(t, wasMBInterceptorSet)
}

func TestSovereignDelayedBroadcastData_SetValidatorData(t *testing.T) {
	t.Parallel()

	args := createDefaultDelayedBroadcasterArgs()
	sovBroadcaster, _ := NewSovereignDelayedBlockBroadcaster(args)

	err := sovBroadcaster.SetValidatorData(&delayedBroadcastData{
		header: &block.SovereignChainHeader{},
	})
	require.Nil(t, err)
	require.Empty(t, sovBroadcaster.valBroadcastData[0].miniBlockHashes)
}
