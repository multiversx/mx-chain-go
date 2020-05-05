package intermediate

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/data"
	"github.com/ElrondNetwork/elrond-go/genesis/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockNodesHandler() *mock.InitialNodesHandlerStub {
	eligible := map[uint32][]sharding.GenesisNodeInfoHandler{
		0: {
			mock.NewNodeInfo([]byte("addr1"), []byte("pubkey1"), 0),
			mock.NewNodeInfo([]byte("addr2"), []byte("pubkey2"), 0),
		},
		1: {
			mock.NewNodeInfo([]byte("addr1"), []byte("pubkey3"), 1),
			mock.NewNodeInfo([]byte("addr4"), []byte("pubkey4"), 1),
		},
		core.MetachainShardId: {
			mock.NewNodeInfo([]byte("addr5"), []byte("pubkey5"), core.MetachainShardId),
			mock.NewNodeInfo([]byte("addr6"), []byte("pubkey6"), core.MetachainShardId),
		},
	}

	waiting := map[uint32][]sharding.GenesisNodeInfoHandler{
		0: {
			mock.NewNodeInfo([]byte("addr7"), []byte("pubkey7"), 0),
		},
		1: {
			mock.NewNodeInfo([]byte("addr8"), []byte("pubkey8"), 1),
		},
		core.MetachainShardId: {
			mock.NewNodeInfo([]byte("addr9"), []byte("pubkey9"), core.MetachainShardId),
		},
	}

	return &mock.InitialNodesHandlerStub{
		InitialNodesInfoCalled: func() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler) {
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

func TestNewNodesHandler_NilNodesHandlerShouldErr(t *testing.T) {
	t.Parallel()

	nh, err := NewNodesHandler(
		nil,
		&mock.AccountsParserStub{},
	)

	assert.True(t, check.IfNil(nh))
	assert.Equal(t, genesis.ErrNilNodesSetup, err)
}

func TestNewNodesHandler_NilAccountsParserShouldErr(t *testing.T) {
	t.Parallel()

	nh, err := NewNodesHandler(
		&mock.InitialNodesHandlerStub{},
		nil,
	)

	assert.True(t, check.IfNil(nh))
	assert.Equal(t, genesis.ErrNilAccountsParser, err)
}

func TestNodesHandler_GetAllNodes(t *testing.T) {
	t.Parallel()

	nh, err := NewNodesHandler(
		createMockNodesHandler(),
		&mock.AccountsParserStub{},
	)

	require.Nil(t, err)

	nodes := nh.GetAllNodes()
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

func TestNodesHandler_GetDelegatedNodes(t *testing.T) {
	t.Parallel()

	nh, _ := NewNodesHandler(
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

	nodes := nh.GetDelegatedNodes([]byte("addr6"))
	require.Equal(t, 1, len(nodes))
	assert.Equal(t, "pubkey6", string(nodes[0].PubKeyBytes()))

	nodes = nh.GetDelegatedNodes([]byte("addr1"))
	require.Equal(t, 2, len(nodes))
	//Test the order
	assert.Equal(t, "pubkey1", string(nodes[0].PubKeyBytes()))
	assert.Equal(t, "pubkey3", string(nodes[1].PubKeyBytes()))
}
