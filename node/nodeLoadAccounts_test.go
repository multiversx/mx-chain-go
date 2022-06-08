package node_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestNode_GetAccountWithOptionsShouldWork(t *testing.T) {
	t.Parallel()

	aliceOnFinal, _ := state.NewUserAccount(testscommon.TestPubKeyAlice)
	aliceOnFinal.Balance = big.NewInt(100)

	aliceOnCurrent, _ := state.NewUserAccount(testscommon.TestPubKeyAlice)
	aliceOnCurrent.Balance = big.NewInt(150)

	accountsRepostitory := testscommon.NewAccountsRepositoryStub()
	accountsRepostitory.GetAccountOnFinalCalled = func(pubkey []byte) (vmcommon.AccountHandler, state.AccountBlockInfo, error) {
		if bytes.Equal(pubkey, testscommon.TestPubKeyAlice) {
			return aliceOnFinal, &testscommon.AccountBlockInfoStub{
				Nonce:    1,
				Hash:     []byte{0xaa},
				RootHash: []byte{0xbb},
			}, nil
		}

		return nil, nil, state.ErrAccNotFound
	}
	accountsRepostitory.GetAccountOnCurrentCalled = func(pubkey []byte) (vmcommon.AccountHandler, state.AccountBlockInfo, error) {
		if bytes.Equal(pubkey, testscommon.TestPubKeyAlice) {
			return aliceOnCurrent, &testscommon.AccountBlockInfoStub{
				Nonce:    7,
				Hash:     []byte{0xcc},
				RootHash: []byte{0xdd},
			}, nil
		}

		return nil, nil, state.ErrAccNotFound
	}

	coreComponents := getDefaultCoreComponents()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsRepo = accountsRepostitory

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	// Get on final
	account, blockInfo, err := n.GetAccount(testscommon.TestAddressAlice, api.AccountQueryOptions{OnFinalBlock: true})
	require.Nil(t, err)
	require.Equal(t, "100", account.Balance)
	require.Equal(t, uint64(1), blockInfo.Nonce)
	require.Equal(t, "aa", blockInfo.Hash)
	require.Equal(t, "bb", blockInfo.RootHash)

	account, blockInfo, err = n.GetAccount(testscommon.TestAddressBob, api.AccountQueryOptions{OnFinalBlock: true})
	require.Nil(t, err)
	require.Equal(t, "0", account.Balance)
	require.Empty(t, blockInfo)

	// Get on current
	account, blockInfo, err = n.GetAccount(testscommon.TestAddressAlice, api.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, "150", account.Balance)
	require.Equal(t, uint64(7), blockInfo.Nonce)
	require.Equal(t, "cc", blockInfo.Hash)
	require.Equal(t, "dd", blockInfo.RootHash)

	account, blockInfo, err = n.GetAccount(testscommon.TestAddressBob, api.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, "0", account.Balance)
	require.Empty(t, blockInfo)
}

func TestNode_GetCodeWithOptionsShouldWork(t *testing.T) {
	t.Parallel()

	goodCodeHashOnFinal := []byte("code hash on final")
	goodCodeHashOnCurrent := []byte("code hash on current")
	codeOnFinal := []byte("code on final")
	codeOnCurrent := []byte("code on current")

	accountsRepostitory := testscommon.NewAccountsRepositoryStub()
	accountsRepostitory.GetCodeOnFinalCalled = func(codeHash []byte) ([]byte, state.AccountBlockInfo) {
		if bytes.Equal(codeHash, goodCodeHashOnFinal) {
			return codeOnFinal, &testscommon.AccountBlockInfoStub{
				Nonce:    1,
				Hash:     []byte{0xaa},
				RootHash: []byte{0xbb},
			}
		}

		return nil, nil
	}
	accountsRepostitory.GetCodeOnCurrentCalled = func(codeHash []byte) ([]byte, state.AccountBlockInfo) {
		if bytes.Equal(codeHash, goodCodeHashOnCurrent) {
			return codeOnCurrent, &testscommon.AccountBlockInfoStub{
				Nonce:    1,
				Hash:     []byte{0xcc},
				RootHash: []byte{0xdd},
			}
		}

		return nil, nil
	}

	coreComponents := getDefaultCoreComponents()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsRepo = accountsRepostitory

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	// Get on final
	code, blockInfo := n.GetCode(goodCodeHashOnFinal, api.AccountQueryOptions{OnFinalBlock: true})
	require.Equal(t, codeOnFinal, code)
	require.Equal(t, uint64(1), blockInfo.Nonce)
	require.Equal(t, "aa", blockInfo.Hash)
	require.Equal(t, "bb", blockInfo.RootHash)

	code, blockInfo = n.GetCode([]byte("missing"), api.AccountQueryOptions{OnFinalBlock: true})
	require.Nil(t, code)
	require.Empty(t, blockInfo)

	// Get on current
	code, blockInfo = n.GetCode(goodCodeHashOnCurrent, api.AccountQueryOptions{})
	require.Equal(t, codeOnCurrent, code)
	require.Equal(t, uint64(1), blockInfo.Nonce)
	require.Equal(t, "cc", blockInfo.Hash)
	require.Equal(t, "dd", blockInfo.RootHash)

	code, blockInfo = n.GetCode([]byte("missing"), api.AccountQueryOptions{})
	require.Nil(t, code)
	require.Empty(t, blockInfo)
}
