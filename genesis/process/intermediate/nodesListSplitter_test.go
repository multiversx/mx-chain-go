package intermediate

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/genesis/mock"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockNodesHandler() *mock.InitialNodesHandlerStub {
	initRating := uint32(50)
	eligible := map[uint32][]nodesCoordinator.GenesisNodeInfoHandler{
		0: {
			mock.NewNodeInfo([]byte("addr1"), []byte("pubkey1"), 0, initRating),
			mock.NewNodeInfo([]byte("addr2"), []byte("pubkey2"), 0, initRating),
		},
		1: {
			mock.NewNodeInfo([]byte("addr1"), []byte("pubkey3"), 1, initRating),
			mock.NewNodeInfo([]byte("addr4"), []byte("pubkey4"), 1, initRating),
		},
		core.MetachainShardId: {
			mock.NewNodeInfo([]byte("addr5"), []byte("pubkey5"), core.MetachainShardId, initRating),
			mock.NewNodeInfo([]byte("addr6"), []byte("pubkey6"), core.MetachainShardId, initRating),
		},
	}

	waiting := map[uint32][]nodesCoordinator.GenesisNodeInfoHandler{
		0: {
			mock.NewNodeInfo([]byte("addr7"), []byte("pubkey7"), 0, initRating),
		},
		1: {
			mock.NewNodeInfo([]byte("addr8"), []byte("pubkey8"), 1, initRating),
		},
		core.MetachainShardId: {
			mock.NewNodeInfo([]byte("addr9"), []byte("pubkey9"), core.MetachainShardId, initRating),
		},
	}

	return &mock.InitialNodesHandlerStub{
		InitialNodesInfoCalled: func() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			return eligible, waiting
		},
	}
}

func createInitialAccountWithDelegation(delegationAddress string) *data.InitialAccount {
	ia := &data.InitialAccount{
		Address:      "",
		Supply:       big.NewInt(10),
		Balance:      big.NewInt(1),
		StakingValue: big.NewInt(2),
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(7),
		},
	}
	ia.Delegation.SetAddressBytes([]byte(delegationAddress))

	return ia
}

func TestNewNodesListSplitter_NilNodesHandlerShouldErr(t *testing.T) {
	t.Parallel()

	nls, err := NewNodesListSplitter(
		nil,
		&mock.AccountsParserStub{},
	)

	assert.True(t, check.IfNil(nls))
	assert.Equal(t, genesis.ErrNilNodesSetup, err)
}

func TestNewNodesListSplitter_NilAccountsParserShouldErr(t *testing.T) {
	t.Parallel()

	nls, err := NewNodesListSplitter(
		&mock.InitialNodesHandlerStub{},
		nil,
	)

	assert.True(t, check.IfNil(nls))
	assert.Equal(t, genesis.ErrNilAccountsParser, err)
}

func TestNewNodesListSplitter_GetAllNodes(t *testing.T) {
	t.Parallel()

	nls, err := NewNodesListSplitter(
		createMockNodesHandler(),
		&mock.AccountsParserStub{},
	)

	require.Nil(t, err)

	nodes := nls.GetAllNodes()
	require.Equal(t, 9, len(nodes))

	//this will also test the order of the elements
	//shard 0
	assert.Equal(t, "addr1", string(nodes[0].AddressBytes()))
	assert.Equal(t, "addr2", string(nodes[1].AddressBytes()))
	assert.Equal(t, "addr7", string(nodes[2].AddressBytes()))
	//shard 1
	assert.Equal(t, "addr1", string(nodes[3].AddressBytes())) //mind the fact that addr3 is missing
	assert.Equal(t, "addr4", string(nodes[4].AddressBytes()))
	assert.Equal(t, "addr8", string(nodes[5].AddressBytes()))
	//meta
	assert.Equal(t, "addr5", string(nodes[6].AddressBytes()))
	assert.Equal(t, "addr6", string(nodes[7].AddressBytes()))
	assert.Equal(t, "addr9", string(nodes[8].AddressBytes()))
}

func TestNodesListSplitter_GetDelegatedNodes(t *testing.T) {
	t.Parallel()

	nls, _ := NewNodesListSplitter(
		createMockNodesHandler(),
		&mock.AccountsParserStub{
			InitialAccountsCalled: func() []genesis.InitialAccountHandler {
				return []genesis.InitialAccountHandler{
					createInitialAccountWithDelegation("addr6"),
					createInitialAccountWithDelegation("addr1"),
					createInitialAccountWithDelegation("addr9"),
					createInitialAccountWithDelegation("addr3"),
				}
			},
		},
	)

	nodes := nls.GetDelegatedNodes([]byte("addr6"))
	require.Equal(t, 1, len(nodes))
	assert.Equal(t, "pubkey6", string(nodes[0].PubKeyBytes()))

	nodes = nls.GetDelegatedNodes([]byte("addr1"))
	require.Equal(t, 2, len(nodes))
	//Test the order
	assert.Equal(t, "pubkey1", string(nodes[0].PubKeyBytes()))
	assert.Equal(t, "pubkey3", string(nodes[1].PubKeyBytes()))
}
