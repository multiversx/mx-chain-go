package peer_test

import (
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewSovereignValidatorStatisticsProcessorFactory(t *testing.T) {
	vspf, err := peer.NewSovereignValidatorStatisticsProcessorFactory(nil)
	require.NotNil(t, err)
	require.Nil(t, vspf)

	vs, _ := peer.NewValidatorStatisticsProcessorFactory()
	vspf, err = peer.NewSovereignValidatorStatisticsProcessorFactory(vs)
	require.Nil(t, err)
	require.NotNil(t, vspf)
}

func TestSovereignValidatorStatisticsProcessorFactory_CreateValidatorStatisticsProcessor(t *testing.T) {
	vs, _ := peer.NewValidatorStatisticsProcessorFactory()
	vspf, _ := peer.NewSovereignValidatorStatisticsProcessorFactory(vs)

	vsp, err := vspf.CreateValidatorStatisticsProcessor(peer.ArgValidatorStatisticsProcessor{})
	require.NotNil(t, err)
	require.Nil(t, vsp)

	vsp, err = vspf.CreateValidatorStatisticsProcessor(createMockArguments())
	require.Nil(t, err)
	require.NotNil(t, vsp)
}

func TestSovereignValidatorStatisticsProcessorFactory_IsInterfaceNil(t *testing.T) {
	vs, _ := peer.NewValidatorStatisticsProcessorFactory()
	vspf, _ := peer.NewSovereignValidatorStatisticsProcessorFactory(vs)
	require.False(t, vspf.IsInterfaceNil())
}
