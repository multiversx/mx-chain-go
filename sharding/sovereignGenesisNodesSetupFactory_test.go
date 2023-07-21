package sharding

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/sharding/mock"
	"github.com/stretchr/testify/require"
)

func TestSovereignGenesisNodesSetupFactory_CreateNodesSetup(t *testing.T) {
	t.Parallel()

	factory := NewSovereignGenesisNodesSetupFactory()
	require.False(t, factory.IsInterfaceNil())

	nodesHandler, err := factory.CreateNodesSetup(&NodesSetupArgs{
		NodesFilePath:            "mock/testdata/sovereignNodesSetupMock.json",
		AddressPubKeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubKeyConverter: mock.NewPubkeyConverterMock(96),
		GenesisMaxNumShards:      100,
	})
	require.Nil(t, err)
	require.Equal(t, "*sharding.SovereignNodesSetup", fmt.Sprintf("%T", nodesHandler))
}
