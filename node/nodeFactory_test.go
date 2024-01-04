package node_test

import (
	"fmt"
	"github.com/multiversx/mx-chain-go/node"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewNodeFactory(t *testing.T) {
	t.Parallel()

	nodeFactory := node.NewNodeFactory()
	require.False(t, nodeFactory.IsInterfaceNil())
}

func TestNodeFactory_CreateNewNode(t *testing.T) {
	t.Parallel()

	nodeFactory := node.NewNodeFactory()

	n, err := nodeFactory.CreateNewNode()
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, "*node.Node", fmt.Sprintf("%T", n))
}
