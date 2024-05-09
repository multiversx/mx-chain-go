package node_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/errors"
	runType "github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/mock"
	"github.com/multiversx/mx-chain-go/node"
	"github.com/multiversx/mx-chain-go/testscommon"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/consensus/factoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"

	"github.com/stretchr/testify/require"
)

func createRunTypeComponents() runType.RunTypeComponentsHolder {
	coreArgs := componentsMock.GetCoreArgs(testscommon.GetGeneralConfig())
	coreArgs.NodesFilename = "../factory/mock/testdata/nodesSetupMock.json"
	coreComp := componentsMock.GetNewCoreComponents(coreArgs)
	cryptoComp := componentsMock.GetCryptoComponents(coreComp)
	return componentsMock.GetRunTypeComponents(coreComp, cryptoComp)
}

func TestCreateNode(t *testing.T) {
	t.Parallel()

	t.Run("nil node factory should not work", func(t *testing.T) {
		t.Parallel()

		nodeHandler, err := node.CreateNode(
			&config.Config{},
			createRunTypeComponents(),
			&factory.StatusCoreComponentsStub{},
			getDefaultBootstrapComponents(),
			getDefaultCoreComponents(),
			getDefaultCryptoComponents(),
			getDefaultDataComponents(),
			getDefaultNetworkComponents(),
			getDefaultProcessComponents(),
			getDefaultStateComponents(),
			&mainFactoryMocks.StatusComponentsStub{},
			&mock.HeartbeatV2ComponentsStub{},
			&factoryMocks.ConsensusComponentsStub{
				GroupSize: 1,
			},
			0,
			false,
			nil)

		require.NotNil(t, err)
		require.Equal(t, errors.ErrNilNode, err)
		require.Nil(t, nodeHandler)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		nodeHandler, err := node.CreateNode(
			&config.Config{},
			createRunTypeComponents(),
			&factory.StatusCoreComponentsStub{},
			getDefaultBootstrapComponents(),
			getDefaultCoreComponents(),
			getDefaultCryptoComponents(),
			getDefaultDataComponents(),
			getDefaultNetworkComponents(),
			getDefaultProcessComponents(),
			getDefaultStateComponents(),
			&mainFactoryMocks.StatusComponentsStub{},
			&mock.HeartbeatV2ComponentsStub{},
			&factoryMocks.ConsensusComponentsStub{
				GroupSize: 1,
			},
			0,
			false,
			node.NewSovereignNodeFactory())

		require.Nil(t, err)
		require.NotNil(t, nodeHandler)
	})
}
