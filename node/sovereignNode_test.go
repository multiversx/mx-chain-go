package node_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/node"
)

var nativeESDT = "WEGLD-bd4d79"

func TestNewSovereignNode(t *testing.T) {
	t.Parallel()

	t.Run("valid node should work", func(t *testing.T) {
		t.Parallel()

		n, err := node.NewNode()
		require.Nil(t, err)
		require.NotNil(t, n)

		sn, err := node.NewSovereignNode(n, nativeESDT)
		require.Nil(t, err)
		require.False(t, sn.IsInterfaceNil())
	})
	t.Run("nil node should error", func(t *testing.T) {
		t.Parallel()

		sn, err := node.NewSovereignNode(nil, nativeESDT)
		require.NotNil(t, err)
		require.Equal(t, errors.ErrNilNode, err)
		require.Nil(t, sn)
	})
	t.Run("empty native esdt should error", func(t *testing.T) {
		t.Parallel()

		n, err := node.NewNode()
		require.Nil(t, err)
		require.NotNil(t, n)

		sn, err := node.NewSovereignNode(n, "")
		require.NotNil(t, err)
		require.Equal(t, node.ErrEmptyNativeEsdt, err)
		require.Nil(t, sn)
	})
}

func TestSovereignNode_GetAllESDTTokens(t *testing.T) {
	t.Parallel()

	testNodeGetAllIssuedESDTs(t, node.NewSovereignNodeFactory(nativeESDT), core.SovereignChainShardId)
	testNodeGetAllIssuedESDTs(t, node.NewSovereignNodeFactory(nativeESDT), core.MainChainShardId)
}

func TestSovereignNode_GetNFTTokenIDsRegisteredByAddress(t *testing.T) {
	t.Parallel()

	testNodeGetNFTTokenIDsRegisteredByAddress(t, node.NewSovereignNodeFactory(nativeESDT), core.SovereignChainShardId)
	testNodeGetNFTTokenIDsRegisteredByAddress(t, node.NewSovereignNodeFactory(nativeESDT), core.MainChainShardId)
}

func TestSovereignNode_GetESDTsWithRole(t *testing.T) {
	t.Parallel()

	testNodeGetESDTsWithRole(t, node.NewSovereignNodeFactory(nativeESDT), core.SovereignChainShardId)
	testNodeGetESDTsWithRole(t, node.NewSovereignNodeFactory(nativeESDT), core.MainChainShardId)
}

func TestSovereignNode_GetESDTsRoles(t *testing.T) {
	t.Parallel()

	testNodeGetESDTsRoles(t, node.NewSovereignNodeFactory(nativeESDT), core.SovereignChainShardId)
	testNodeGetESDTsRoles(t, node.NewSovereignNodeFactory(nativeESDT), core.MainChainShardId)
}
