package peer_test

import (
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewValidatorStatisticsProcessorFactory(t *testing.T) {
	vspf, err := peer.NewValidatorStatisticsProcessorFactory()
	require.Nil(t, err)
	require.NotNil(t, vspf)
}

func TestValidatorStatisticsProcessorFactory_CreateValidatorStatisticsProcessor(t *testing.T) {
	vspf, _ := peer.NewValidatorStatisticsProcessorFactory()

	vsp, err := vspf.CreateValidatorStatisticsProcessor(peer.ArgValidatorStatisticsProcessor{})
	require.NotNil(t, err)
	require.Nil(t, vsp)

	vsp, err = vspf.CreateValidatorStatisticsProcessor(createMockArguments())
	require.Nil(t, err)
	require.NotNil(t, vsp)
}

func TestValidatorStatisticsProcessorFactory_IsInterfaceNil(t *testing.T) {
	vspf, _ := peer.NewValidatorStatisticsProcessorFactory()
	require.False(t, vspf.IsInterfaceNil())
}
