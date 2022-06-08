package node_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestNode_GetAccountWithOptionsOnFinalBlockShouldWork(t *testing.T) {
	t.Parallel()

	alice, _ := state.NewUserAccount(testscommon.TestPubKeyAlice)
	alice.Balance = big.NewInt(100)

	var calledWithPubkey []byte

	accountsRepostitory := testscommon.NewAccountsRepositoryStub()
	accountsRepostitory.GetAccountOnFinalCalled = func(pubkey []byte) (vmcommon.AccountHandler, state.AccountBlockInfo, error) {
		calledWithPubkey = pubkey
		return alice, &testscommon.AccountBlockInfoStub{
			Nonce:    1,
			Hash:     []byte{0xaa, 0xbb},
			RootHash: []byte{0xbb, 0xaa},
		}, nil
	}

	coreComponents := getDefaultCoreComponents()
	dataComponents := getDefaultDataComponents()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsRepo = accountsRepostitory

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
	)

	account, blockInfo, err := n.GetAccount(testscommon.TestAddressAlice, api.AccountQueryOptions{OnFinalBlock: true})

	require.Nil(t, err)
	require.Equal(t, "100", account.Balance)
	require.Equal(t, testscommon.TestPubKeyAlice, calledWithPubkey)
	require.Equal(t, uint64(1), blockInfo.Nonce)
	require.Equal(t, "aabb", blockInfo.Hash)
	require.Equal(t, "bbaa", blockInfo.RootHash)
}

func TestNode_GetAccountWithOptionsOnCurrentBlockShouldWork(t *testing.T) {
	t.Parallel()

	alice, _ := state.NewUserAccount(testscommon.TestPubKeyAlice)
	alice.Balance = big.NewInt(100)

	var calledWithPubkey []byte

	accountsRepostitory := testscommon.NewAccountsRepositoryStub()
	accountsRepostitory.GetAccountOnCurrentCalled = func(pubkey []byte) (vmcommon.AccountHandler, state.AccountBlockInfo, error) {
		calledWithPubkey = pubkey
		return alice, &testscommon.AccountBlockInfoStub{
			Nonce:    7,
			Hash:     []byte{0xdd, 0xcc},
			RootHash: []byte{0xcc, 0xdd},
		}, nil
	}

	coreComponents := getDefaultCoreComponents()
	dataComponents := getDefaultDataComponents()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsRepo = accountsRepostitory

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
	)

	account, blockInfo, err := n.GetAccount(testscommon.TestAddressAlice, api.AccountQueryOptions{OnFinalBlock: false})

	require.Nil(t, err)
	require.Equal(t, "100", account.Balance)
	require.Equal(t, testscommon.TestPubKeyAlice, calledWithPubkey)
	require.Equal(t, uint64(7), blockInfo.Nonce)
	require.Equal(t, "ddcc", blockInfo.Hash)
	require.Equal(t, "ccdd", blockInfo.RootHash)
}
