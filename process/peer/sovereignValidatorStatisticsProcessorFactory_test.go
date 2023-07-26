package peer_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignValidatorStatisticsProcessorFactory(t *testing.T) {
	t.Parallel()

	vspf, err := peer.NewSovereignValidatorStatisticsProcessorFactory(nil)
	require.NotNil(t, err)
	require.Nil(t, vspf)

	vs, _ := peer.NewValidatorStatisticsProcessorFactory()
	vspf, err = peer.NewSovereignValidatorStatisticsProcessorFactory(vs)
	require.Nil(t, err)
	require.NotNil(t, vspf)
	require.Implements(t, new(peer.ValidatorStatisticsProcessorCreator), vspf)
}

func TestSovereignValidatorStatisticsProcessorFactory_CreateValidatorStatisticsProcessor(t *testing.T) {
	t.Parallel()

	vs, _ := peer.NewValidatorStatisticsProcessorFactory()
	vspf, _ := peer.NewSovereignValidatorStatisticsProcessorFactory(vs)

	vsp, err := vspf.CreateValidatorStatisticsProcessor(peer.ArgValidatorStatisticsProcessor{})
	require.NotNil(t, err)
	require.Nil(t, vsp)

	vsp, err = vspf.CreateValidatorStatisticsProcessor(createMockArguments())
	require.Nil(t, err)
	require.NotNil(t, vsp)
	require.Implements(t, new(process.ValidatorStatisticsProcessor), vspf)
}

func TestSovereignValidatorStatisticsProcessorFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	vs, _ := peer.NewValidatorStatisticsProcessorFactory()
	vspf, _ := peer.NewSovereignValidatorStatisticsProcessorFactory(vs)
	require.False(t, vspf.IsInterfaceNil())
}
