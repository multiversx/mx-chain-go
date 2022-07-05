package node_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/holders"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	mockState "github.com/ElrondNetwork/elrond-go/testscommon/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestNode_GetAccountWithOptionsShouldWork(t *testing.T) {
	t.Parallel()

	alice, _ := state.NewUserAccount(testscommon.TestPubKeyAlice)
	alice.Balance = big.NewInt(100)

	accountsRepostitory := &mockState.AccountsRepositoryStub{}
	accountsRepostitory.GetAccountWithBlockInfoCalled = func(pubkey []byte, options api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
		if bytes.Equal(pubkey, testscommon.TestPubKeyAlice) {
			return alice, holders.NewBlockInfo([]byte{0xaa}, 1, []byte{0xbb}), nil
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

	account, blockInfo, err := n.GetAccount(testscommon.TestAddressAlice, api.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, "100", account.Balance)
	require.Equal(t, uint64(1), blockInfo.Nonce)
	require.Equal(t, "aa", blockInfo.Hash)
	require.Equal(t, "bb", blockInfo.RootHash)
}

func TestNode_GetCodeWithOptionsShouldWork(t *testing.T) {
	t.Parallel()

	goodCodeHash := []byte("code hash")
	code := []byte("code")

	accountsRepostitory := &mockState.AccountsRepositoryStub{}
	accountsRepostitory.GetCodeWithBlockInfoCalled = func(codeHash []byte, options api.AccountQueryOptions) ([]byte, common.BlockInfo, error) {
		if bytes.Equal(codeHash, goodCodeHash) {
			return code, holders.NewBlockInfo([]byte{0xaa}, 1, []byte{0xbb}), nil
		}

		return nil, nil, nil
	}

	coreComponents := getDefaultCoreComponents()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsRepo = accountsRepostitory

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	code, blockInfo := n.GetCode(goodCodeHash, api.AccountQueryOptions{})
	require.Equal(t, code, code)
	require.Equal(t, uint64(1), blockInfo.Nonce)
	require.Equal(t, "aa", blockInfo.Hash)
	require.Equal(t, "bb", blockInfo.RootHash)
}
