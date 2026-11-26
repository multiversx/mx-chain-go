package node_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node"
	"github.com/multiversx/mx-chain-go/node/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestTrieListingAPIsRejectIncompleteResults(t *testing.T) {
	calls := map[string]func(*node.Node, context.Context) (interface{}, error){
		"keys": func(n *node.Node, ctx context.Context) (interface{}, error) {
			v, _, err := n.GetKeyValuePairs(testscommon.TestAddressAlice, api.AccountQueryOptions{}, ctx)
			return v, err
		},
		"account with keys": func(n *node.Node, ctx context.Context) (interface{}, error) {
			v, _, err := n.GetAccountWithKeys(testscommon.TestAddressAlice, api.AccountQueryOptions{WithKeys: true}, ctx)
			return v.Pairs, err
		},
		"tokens": func(n *node.Node, ctx context.Context) (interface{}, error) {
			v, _, err := n.GetAllESDTTokens(testscommon.TestAddressAlice, api.AccountQueryOptions{}, ctx)
			return v, err
		},
		"issued": func(n *node.Node, ctx context.Context) (interface{}, error) { return n.GetAllIssuedESDTs("", ctx) },
		"roles": func(n *node.Node, ctx context.Context) (interface{}, error) {
			v, _, err := n.GetESDTsRoles(testscommon.TestAddressAlice, api.AccountQueryOptions{}, ctx)
			return v, err
		},
		"with role": func(n *node.Node, ctx context.Context) (interface{}, error) {
			v, _, err := n.GetESDTsWithRole(testscommon.TestAddressAlice, core.ESDTRoleLocalMint, api.AccountQueryOptions{}, ctx)
			return v, err
		},
		"registered": func(n *node.Node, ctx context.Context) (interface{}, error) {
			v, _, err := n.GetNFTTokenIDsRegisteredByAddress(testscommon.TestAddressAlice, api.AccountQueryOptions{}, ctx)
			return v, err
		},
	}
	for name, call := range calls {
		for _, cancelled := range []bool{false, true} {
			label := name + "/limit"
			if cancelled {
				label = name + "/cancel"
			}
			t.Run(label, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				acc := createAcc(testscommon.TestPubKeyAlice)
				acc.SetDataTrie(&trieMock.TrieStub{RootCalled: func() ([]byte, error) { return nil, nil }, GetAllLeavesOnChannelCalled: func(ch *common.TrieIteratorChannels, _ context.Context, _ []byte, _ common.KeyBuilder, _ common.TrieLeafParser) error {
					go func() {
						defer close(ch.LeavesChan)
						ch.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("unrelated"), []byte("value"))
						if cancelled {
							cancel()
							return
						}
						ch.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("other"), []byte("value"))
					}()
					return nil
				}})
				n := getNodeWithAccount(&stateMock.AccountsStub{GetAccountWithBlockInfoCalled: func([]byte, common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
					return acc, nil, nil
				}})
				proc := getDefaultProcessComponents()
				proc.ShardCoord = &mock.ShardCoordinatorMock{SelfShardId: core.MetachainShardId}
				require.NoError(t, n.ApplyOptions(node.WithProcessComponents(proc), node.WithTrieOperations(config.TrieOperationsConfig{MaxLeaves: 1})))
				result, err := call(n, ctx)
				want := node.ErrTrieOperationsLimit
				if cancelled {
					want = node.ErrTrieOperationsTimeout
				}
				require.ErrorIs(t, err, want)
				require.Empty(t, result)
			})
		}
	}
}

func TestTrieListingChargesExternalNFTMetadata(t *testing.T) {
	acc := createAcc(testscommon.TestPubKeyAlice)
	acc.SetDataTrie(&trieMock.TrieStub{RootCalled: func() ([]byte, error) { return nil, nil }, GetAllLeavesOnChannelCalled: func(ch *common.TrieIteratorChannels, _ context.Context, _ []byte, _ common.KeyBuilder, _ common.TrieLeafParser) error {
		ch.LeavesChan <- keyValStorage.NewKeyValStorage([]byte(core.ProtectedKeyPrefix+core.ESDTKeyIdentifier+"TOKEN-abcdef"), nil)
		close(ch.LeavesChan)
		return nil
	}})
	n := getNodeWithAccount(&stateMock.AccountsStub{GetAccountWithBlockInfoCalled: func([]byte, common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return acc, nil, nil
	}})
	loaded := false
	storage := &testscommon.EsdtStorageHandlerStub{GetESDTNFTTokenOnDestinationWithCustomSystemAccountCalled: func(vmcommon.UserAccountHandler, []byte, uint64, vmcommon.UserAccountHandler) (*esdt.ESDigitalToken, bool, error) {
		loaded = true
		return &esdt.ESDigitalToken{Value: big.NewInt(1), TokenMetaData: &esdt.MetaData{Attributes: make([]byte, 1024)}}, false, nil
	}}
	require.NoError(t, n.ApplyOptions(node.WithTrieOperations(config.TrieOperationsConfig{MaxBytes: 1000}), node.WithESDTNFTStorageHandler(storage)))
	result, _, err := n.GetAllESDTTokens(testscommon.TestAddressAlice, api.AccountQueryOptions{}, context.Background())
	require.True(t, loaded)
	require.ErrorIs(t, err, node.ErrTrieOperationsLimit)
	require.Nil(t, result)
}
