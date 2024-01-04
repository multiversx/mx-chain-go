package node_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/node"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignNodeFactory(t *testing.T) {
	t.Parallel()

	sovereignNodeFactory := node.NewSovereignNodeFactory()
	require.False(t, sovereignNodeFactory.IsInterfaceNil())
}

func TestSovereignNodeFactory_CreateNewNode(t *testing.T) {
	t.Parallel()

	sovereignNodeFactory := node.NewSovereignNodeFactory()

	sn, err := sovereignNodeFactory.CreateNewNode()
	require.Nil(t, err)
	require.NotNil(t, sn)
	require.Equal(t, "*node.sovereignNode", fmt.Sprintf("%T", sn))
}
