package node_test

import (
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/node"
	"github.com/stretchr/testify/require"
)

func TestCreateNode(t *testing.T) {
	t.Parallel()

	t.Run("nil node factory should not work", func(t *testing.T) {
		t.Parallel()

		nodeHandler, err := node.CreateNode(
			&config.Config{},
			getDefaultStatusCoreComponents(),
			getDefaultBootstrapComponents(),
			getDefaultCoreComponents(),
			getDefaultCryptoComponents(),
			getDefaultDataComponents(),
			getDefaultNetworkComponents(),
			getDefaultProcessComponents(),
			getDefaultStateComponents(),
			getDefaultStatusComponents(),
			getDefaultHeartbeatV2Components(),
			getDefaultConsensusComponents(),
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
			getDefaultStatusCoreComponents(),
			getDefaultBootstrapComponents(),
			getDefaultCoreComponents(),
			getDefaultCryptoComponents(),
			getDefaultDataComponents(),
			getDefaultNetworkComponents(),
			getDefaultProcessComponents(),
			getDefaultStateComponents(),
			&mainFactoryMocks.StatusComponentsStub{},
			getDefaultHeartbeatV2Components(),
			getDefaultConsensusComponents(),
			0,
			false,
			node.NewSovereignNodeFactory())

		require.Nil(t, err)
		require.NotNil(t, nodeHandler)
	})
}
