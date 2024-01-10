package node_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node"
	"github.com/stretchr/testify/require"
)

func TestCreateNode(t *testing.T) {
	t.Parallel()

	nodeHandler, err := node.CreateNode(
		&config.Config{},
		nil,
		getDefaultBootstrapComponents(),
		getDefaultCoreComponents(),
		getDefaultCryptoComponents(),
		getDefaultDataComponents(),
		getDefaultNetworkComponents(),
		getDefaultProcessComponents(),
		getDefaultStateComponents(),
		nil,
		nil,
		nil,
		0,
		false,
		nil)

	require.NotNil(t, err)
	require.Equal(t, node.ErrNilNodeFactory, err)
	require.Nil(t, nodeHandler)
}
