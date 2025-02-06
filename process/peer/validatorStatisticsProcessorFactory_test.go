package peer_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/peer"
)

func TestNewValidatorStatisticsProcessorFactory(t *testing.T) {
	t.Parallel()

	vspf := peer.NewValidatorStatisticsProcessorFactory()
	require.NotNil(t, vspf)
	require.Implements(t, new(peer.ValidatorStatisticsProcessorCreator), vspf)
}

func TestValidatorStatisticsProcessorFactory_CreateValidatorStatisticsProcessor(t *testing.T) {
	t.Parallel()

	vspf := peer.NewValidatorStatisticsProcessorFactory()

	vsp, err := vspf.CreateValidatorStatisticsProcessor(peer.ArgValidatorStatisticsProcessor{})
	require.NotNil(t, err)
	require.Nil(t, vsp)

	vsp, err = vspf.CreateValidatorStatisticsProcessor(createMockArguments())
	require.Nil(t, err)
	require.NotNil(t, vsp)
	require.Implements(t, new(process.ValidatorStatisticsProcessor), vsp)
}

func TestValidatorStatisticsProcessorFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	vspf := peer.NewValidatorStatisticsProcessorFactory()
	require.False(t, vspf.IsInterfaceNil())
}
