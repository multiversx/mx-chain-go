package enablers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
)

func TestSovereignEnableEpochsFactory_CreateEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	factory := NewSovereignEnableEpochsFactory(config.SovereignEpochConfig{})
	require.False(t, factory.IsInterfaceNil())

	eeh, err := factory.CreateEnableEpochsHandler(config.EnableEpochs{}, &epochNotifier.EpochNotifierStub{})
	require.Nil(t, err)
	require.NotNil(t, eeh)
	require.IsType(t, &sovereignEnableEpochsHandler{}, eeh)
}
