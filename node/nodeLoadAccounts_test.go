package node_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestNode_GetAccountWithOptionsOnFinalBlockShouldWork(t *testing.T) {
	t.Parallel()

	alice, _ := state.NewUserAccount(testscommon.PubKeyOfAlice)
	alice.Balance = big.NewInt(100)

	var calledWithPubkey []byte
	var calledWithRootHash []byte

	accountsRepostitory := testscommon.NewAccountsRepositoryStub()
	accountsRepostitory.GetExistingAccountCalled = func(pubkey []byte, rootHash []byte) (vmcommon.AccountHandler, error) {
		calledWithPubkey = pubkey
		calledWithRootHash = rootHash
		return alice, nil
	}

	chainHandler := &testscommon.ChainHandlerStub{}
	chainHandler.SetFinalBlockInfo(1, []byte{0xaa, 0xbb}, []byte{0xbb, 0xaa})

	coreComponents := getDefaultCoreComponents()
	dataComponents := getDefaultDataComponents()
	dataComponents.BlockChain = chainHandler
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsRepo = accountsRepostitory

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
	)

	account, blockInfo, err := n.GetAccount(testscommon.AddressOfAlice, api.AccountQueryOptions{OnFinalBlock: true})

	require.Nil(t, err)
	require.Equal(t, "100", account.Balance)
	require.Equal(t, testscommon.PubKeyOfAlice, calledWithPubkey)
	require.Equal(t, []byte{0xbb, 0xaa}, calledWithRootHash)
	require.Equal(t, uint64(1), blockInfo.Nonce)
	require.Equal(t, "aabb", blockInfo.Hash)
	require.Equal(t, "bbaa", blockInfo.RootHash)
}

func TestNode_GetAccountWithOptionsOnCurrentBlockShouldWork(t *testing.T) {
	t.Parallel()

	alice, _ := state.NewUserAccount(testscommon.PubKeyOfAlice)
	alice.Balance = big.NewInt(100)

	var calledWithPubkey []byte
	var calledWithRootHash []byte

	accountsRepostitory := testscommon.NewAccountsRepositoryStub()
	accountsRepostitory.GetExistingAccountCalled = func(pubkey []byte, rootHash []byte) (vmcommon.AccountHandler, error) {
		calledWithPubkey = pubkey
		calledWithRootHash = rootHash
		return alice, nil
	}

	chainHandler := &testscommon.ChainHandlerStub{}
	_ = chainHandler.SetCurrentBlockHeaderAndRootHash(&block.Header{Nonce: 7}, []byte{0xcc, 0xdd})
	chainHandler.SetCurrentBlockHeaderHash([]byte{0xdd, 0xcc})

	coreComponents := getDefaultCoreComponents()
	dataComponents := getDefaultDataComponents()
	dataComponents.BlockChain = chainHandler
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsRepo = accountsRepostitory

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
	)

	account, blockInfo, err := n.GetAccount(testscommon.AddressOfAlice, api.AccountQueryOptions{OnFinalBlock: false})

	require.Nil(t, err)
	require.Equal(t, "100", account.Balance)
	require.Equal(t, testscommon.PubKeyOfAlice, calledWithPubkey)
	require.Equal(t, []byte{0xcc, 0xdd}, calledWithRootHash)
	require.Equal(t, uint64(7), blockInfo.Nonce)
	require.Equal(t, "ddcc", blockInfo.Hash)
	require.Equal(t, "ccdd", blockInfo.RootHash)
}
