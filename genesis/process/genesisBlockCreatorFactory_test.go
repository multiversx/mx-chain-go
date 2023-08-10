package process

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/genesis/mock"
	"github.com/stretchr/testify/require"
)

func TestNewGenesisBlockCreatorFactory(t *testing.T) {
	factory := NewGenesisBlockCreatorFactory()
	require.False(t, factory.IsInterfaceNil())
	require.Implements(t, new(GenesisBlockCreatorFactory), factory)
}

func TestNewGenesisBlockCreatorFactory_CreateGenesisBlockCreator(t *testing.T) {
	factory := NewGenesisBlockCreatorFactory()

	args := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	blockCreator, err := factory.CreateGenesisBlockCreator(args)
	require.Nil(t, err)
	require.IsType(t, &genesisBlockCreator{}, blockCreator)
}
