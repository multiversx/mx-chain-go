package chainparametersnotifier

import (
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/require"
)

func TestNewChainParametersNotifier(t *testing.T) {
	t.Parallel()

	notifier := NewChainParametersNotifier()
	require.False(t, check.IfNil(notifier))
}

func TestChainParametersNotifier_UpdateCurrentChainParameters(t *testing.T) {
	t.Parallel()

	notifier := NewChainParametersNotifier()
	require.False(t, check.IfNil(notifier))

	chainParams := config.ChainParametersByEpochConfig{
		EnableEpoch: 7,
		Adaptivity:  true,
		Hysteresis:  0.7,
	}
	notifier.UpdateCurrentChainParameters(chainParams)

	resultedChainParams := notifier.CurrentChainParameters()
	require.NotNil(t, resultedChainParams)

	// update with same epoch but other params - should not change (impossible scenario in production, but easier for tests)
	chainParams.Hysteresis = 0.8
	notifier.UpdateCurrentChainParameters(chainParams)
	require.Equal(t, float32(0.7), notifier.CurrentChainParameters().Hysteresis)

	chainParams.Hysteresis = 0.8
	chainParams.EnableEpoch = 8
	notifier.UpdateCurrentChainParameters(chainParams)
	require.Equal(t, float32(0.8), notifier.CurrentChainParameters().Hysteresis)
}

func TestChainParametersNotifier_RegisterNotifyHandler(t *testing.T) {
	t.Parallel()

	notifier := NewChainParametersNotifier()
	require.False(t, check.IfNil(notifier))

	// register a nil handler - should not panic
	notifier.RegisterNotifyHandler(nil)

	testNotifee := &dummyNotifee{}
	notifier.RegisterNotifyHandler(testNotifee)

	chainParams := config.ChainParametersByEpochConfig{
		ShardMinNumNodes: 37,
	}
	notifier.UpdateCurrentChainParameters(chainParams)

	require.Equal(t, chainParams, testNotifee.receivedChainParameters)
}

func TestChainParametersNotifier_UnRegisterAll(t *testing.T) {
	t.Parallel()

	notifier := NewChainParametersNotifier()
	require.False(t, check.IfNil(notifier))

	testNotifee := &dummyNotifee{}
	notifier.RegisterNotifyHandler(testNotifee)
	notifier.UnRegisterAll()

	chainParams := config.ChainParametersByEpochConfig{
		ShardMinNumNodes: 37,
	}
	notifier.UpdateCurrentChainParameters(chainParams)

	require.Empty(t, testNotifee.receivedChainParameters)
}

func TestChainParametersNotifier_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	notifier := NewChainParametersNotifier()

	numOperations := 500
	wg := sync.WaitGroup{}
	wg.Add(numOperations)
	for i := 0; i < numOperations; i++ {
		go func(idx int) {
			switch idx {
			case 0:
				notifier.RegisterNotifyHandler(&dummyNotifee{})
			case 1:
				_ = notifier.CurrentChainParameters()
			case 2:
				notifier.UpdateCurrentChainParameters(config.ChainParametersByEpochConfig{})
			case 3:
				notifier.UnRegisterAll()
			case 4:
				_ = notifier.IsInterfaceNil()
			}

			wg.Done()
		}(i % 5)
	}

	wg.Wait()
}

type dummyNotifee struct {
	receivedChainParameters config.ChainParametersByEpochConfig
}

// ChainParametersChanged -
func (dn *dummyNotifee) ChainParametersChanged(chainParameters config.ChainParametersByEpochConfig) {
	dn.receivedChainParameters = chainParameters
}

// IsInterfaceNil -
func (dn *dummyNotifee) IsInterfaceNil() bool {
	return dn == nil
}
