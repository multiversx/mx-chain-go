package node_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-go/api/groups"
	apiMock "github.com/multiversx/mx-chain-go/api/mock"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	stateErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/node"
	"github.com/multiversx/mx-chain-go/node/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/disabled"
	stateFactory "github.com/multiversx/mx-chain-go/state/factory"
	pruningDisabled "github.com/multiversx/mx-chain-go/state/storagePruningManager/disabled"
	"github.com/multiversx/mx-chain-go/state/triesHolder"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/integrationtests"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie/leavesRetriever"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
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
				dataTrie := &trieMock.TrieStub{RootCalled: func() ([]byte, error) { return nil, nil }, GetAllLeavesOnChannelCalled: func(ch *common.TrieIteratorChannels, _ context.Context, _ []byte, _ common.KeyBuilder, _ common.TrieLeafParser) error {
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
				}}
				acc := createAccWithDth(testscommon.TestPubKeyAlice, &trieMock.TriesHolderStub{
					GetCalled: func(_ []byte) common.Trie { return dataTrie },
				})
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
	dataTrie := &trieMock.TrieStub{RootCalled: func() ([]byte, error) { return nil, nil }, GetAllLeavesOnChannelCalled: func(ch *common.TrieIteratorChannels, _ context.Context, _ []byte, _ common.KeyBuilder, _ common.TrieLeafParser) error {
		ch.LeavesChan <- keyValStorage.NewKeyValStorage([]byte(core.ProtectedKeyPrefix+core.ESDTKeyIdentifier+"TOKEN-abcdef"), nil)
		close(ch.LeavesChan)
		return nil
	}}
	acc := createAccWithDth(testscommon.TestPubKeyAlice, &trieMock.TriesHolderStub{
		GetCalled: func(_ []byte) common.Trie { return dataTrie },
	})
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

func TestDataTrieListingLoadFailure(t *testing.T) {
	for _, version := range []struct {
		name        string
		autoBalance bool
	}{{name: "legacy"}, {name: "auto-balanced", autoBalance: true}} {
		t.Run(version.name, func(t *testing.T) {
			testDataTrieListingLoadFailure(t, version.autoBalance)
		})
	}
}

func testDataTrieListingLoadFailure(t *testing.T, autoBalance bool) {
	t.Helper()
	db := testscommon.CreateMemUnit()
	flags := enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.FixDataTrieLeafValidationFlag)
	if autoBalance {
		flags.AddActiveFlags(common.AutoBalanceDataTriesFlag)
	}
	writer := integrationtests.CreateAccountsDB(db, flags)
	t.Cleanup(func() { require.NoError(t, writer.Close()) })
	pubkey := bytes.Repeat([]byte{1}, 32)
	wallet := bytes.Repeat([]byte{2}, 32)
	conv := testscommon.RealWorldBech32PubkeyConverter
	address, err := conv.Encode(pubkey)
	require.NoError(t, err)
	walletAddress, err := conv.Encode(wallet)
	require.NoError(t, err)
	const tokenID = "TOKEN-abcdef"
	plainKey := []byte("key")
	storedValue := []byte("value")
	tokenKey := []byte(core.ProtectedKeyPrefix + core.ESDTKeyIdentifier + tokenID)
	marshaller := integrationtests.TestMarshalizer
	token, err := marshaller.Marshal(&systemSmartContracts.ESDTDataV2{
		OwnerAddress: pubkey, TokenType: []byte(core.NonFungibleESDT),
		SpecialRoles: []*systemSmartContracts.ESDTRoles{{Address: pubkey, Roles: [][]byte{[]byte(core.ESDTRoleNFTCreate)}}},
	})
	require.NoError(t, err)

	for _, entry := range []struct {
		address []byte
		values  map[string][]byte
	}{
		{address: pubkey, values: map[string][]byte{string(plainKey): storedValue, string(tokenKey): []byte("token balance")}},
		{address: vm.ESDTSCAddress, values: map[string][]byte{tokenID: token}},
		{address: core.SystemAccountAddress},
		{address: wallet},
	} {
		acc, errLoad := writer.LoadAccount(entry.address)
		require.NoError(t, errLoad)
		user := acc.(state.UserAccountHandler)
		user.IncreaseNonce(1)
		for key, value := range entry.values {
			require.NoError(t, user.SaveKeyValue([]byte(key), value))
		}
		require.NoError(t, writer.SaveAccount(user))
	}
	root, err := writer.Commit()
	require.NoError(t, err)
	account, err := writer.GetExistingAccount(pubkey)
	require.NoError(t, err)
	dataRoot := account.(state.UserAccountHandler).GetRootHash()
	account, err = writer.GetExistingAccount(vm.ESDTSCAddress)
	require.NoError(t, err)
	tokenRoot := account.(state.UserAccountHandler).GetRootHash()

	n, repo := newDataTrieListingNode(t, writer, flags, root)
	router := gin.New()
	group, err := groups.NewAddressGroup(&apiMock.FacadeStub{GetKeyValuePairsCalled: func(ctx context.Context, address string, opts api.AccountQueryOptions) (map[string]string, api.BlockInfo, error) {
		return n.GetKeyValuePairs(address, opts, ctx)
	}})
	require.NoError(t, err)
	group.RegisterRoutes(router.Group("/address"), config.ApiRoutesConfig{APIPackages: map[string]config.APIPackageConfig{
		"address": {Routes: []config.RouteConfig{{Name: "/:address/keys", Open: true}}},
	}})
	getKeysHTTP := func(addr string) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/address/"+addr+"/keys", nil))
		return response
	}

	// Simulate local storage loss before any API view has cached the data roots.
	removed := make(map[string][]byte)
	for _, key := range [][]byte{dataRoot, tokenRoot} {
		encoded, errGet := db.Get(key)
		require.NoError(t, errGet)
		removed[string(key)] = encoded
		require.NoError(t, db.Remove(key))
	}

	ctx := t.Context()
	opts := api.AccountQueryOptions{}
	t.Run("corrupt root", func(t *testing.T) {
		require.NoError(t, db.Put(dataRoot, []byte{255}))
		defer func() { require.NoError(t, db.Remove(dataRoot)) }()
		keys, _, errRead := n.GetKeyValuePairs(address, opts, ctx)
		require.ErrorIs(t, errRead, state.ErrDataTrieLoadFailed)
		require.Nil(t, core.UnwrapGetNodeFromDBErr(errRead))
		require.Empty(t, keys)
	})
	calls := []struct {
		name string
		root []byte
		call func() (any, error)
	}{
		{"keys/current", dataRoot, func() (any, error) { keys, _, err := n.GetKeyValuePairs(address, opts, ctx); return keys, err }},
		{"keys/final", dataRoot, func() (any, error) {
			keys, _, err := n.GetKeyValuePairs(address, api.AccountQueryOptions{OnFinalBlock: true}, ctx)
			return keys, err
		}},
		{"keys/historical", dataRoot, func() (any, error) {
			keys, _, err := n.GetKeyValuePairs(address, api.AccountQueryOptions{BlockRootHash: root}, ctx)
			return keys, err
		}},
		{"iterate", dataRoot, func() (any, error) { keys, _, _, err := n.IterateKeys(address, 10, nil, opts, ctx); return keys, err }},
		{"account with keys", dataRoot, func() (any, error) {
			acc, _, err := n.GetAccountWithKeys(address, api.AccountQueryOptions{WithKeys: true}, ctx)
			return acc.Pairs, err
		}},
		{"balances", dataRoot, func() (any, error) { tokens, _, err := n.GetAllESDTTokens(address, opts, ctx); return tokens, err }},
		{"issued", tokenRoot, func() (any, error) { return n.GetAllIssuedESDTs("", ctx) }},
		{"registered", tokenRoot, func() (any, error) {
			tokens, _, err := n.GetNFTTokenIDsRegisteredByAddress(address, opts, ctx)
			return tokens, err
		}},
		{"with role", tokenRoot, func() (any, error) {
			tokens, _, err := n.GetESDTsWithRole(address, core.ESDTRoleNFTCreate, opts, ctx)
			return tokens, err
		}},
		{"roles", tokenRoot, func() (any, error) { roles, _, err := n.GetESDTsRoles(address, opts, ctx); return roles, err }},
	}
	for _, test := range calls {
		t.Run(test.name, func(t *testing.T) {
			result, errCall := test.call()
			requireDataTrieLoadFailure(t, errCall, test.root)
			require.Empty(t, result)
		})
	}
	t.Run("HTTP missing root", func(t *testing.T) {
		response := getKeysHTTP(address)
		require.Equal(t, http.StatusInternalServerError, response.Code)
		require.Contains(t, response.Body.String(), hex.EncodeToString(dataRoot))
	})
	t.Run("iteration and migration preserve cause", func(t *testing.T) {
		account, _, errAccount := repo.GetAccountWithBlockInfo(pubkey, opts)
		require.NoError(t, errAccount)
		requireDataTrieLoadFailure(t, account.(state.UserAccountHandler).GetAllLeaves(nil, ctx), dataRoot)
		_, errMigration := n.IsDataTrieMigrated(address, opts)
		requireDataTrieLoadFailure(t, errMigration, dataRoot)
	})

	response := getKeysHTTP(walletAddress)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), `"pairs":{}`)
	missingAddress, err := conv.Encode(bytes.Repeat([]byte{3}, 32))
	require.NoError(t, err)
	response = getKeysHTTP(missingAddress)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), `"pairs":{}`)
	_, err = n.IsDataTrieMigrated(walletAddress, opts)
	require.ErrorIs(t, err, stateErrors.ErrNilTrie)

	for key, value := range removed {
		require.NoError(t, db.Put([]byte(key), value))
	}
	for _, test := range calls {
		t.Run("repaired/"+test.name, func(t *testing.T) {
			result, errCall := test.call()
			require.NoError(t, errCall)
			require.NotEmpty(t, result)
		})
	}
	response = getKeysHTTP(address)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), hex.EncodeToString(storedValue))

	t.Run("missing descendant", func(t *testing.T) {
		requireListingReportsMissingDescendant(t, writer, db, flags, root, dataRoot, address)
	})
}

func requireListingReportsMissingDescendant(t *testing.T, writer *state.AccountsDB, db storage.Storer, flags common.EnableEpochsHandler, root, dataRoot []byte, address string) {
	t.Helper()
	dataTrie, err := writer.GetTrie(dataRoot)
	require.NoError(t, err)
	hashes, err := dataTrie.GetAllHashes()
	require.NoError(t, err)
	require.Greater(t, len(hashes), 1)
	for _, hash := range hashes {
		if !bytes.Equal(hash, dataRoot) {
			require.NoError(t, db.Remove(hash))
		}
	}
	n, _ := newDataTrieListingNode(t, writer, flags, root)
	keys, _, err := n.GetKeyValuePairs(address, api.AccountQueryOptions{}, t.Context())
	require.Error(t, err)
	require.NotErrorIs(t, err, state.ErrDataTrieLoadFailed)
	require.Empty(t, keys)
}

func requireDataTrieLoadFailure(t *testing.T, err error, root []byte) {
	t.Helper()
	require.ErrorIs(t, err, state.ErrDataTrieLoadFailed)
	cause := core.UnwrapGetNodeFromDBErr(err)
	require.NotNil(t, cause)
	require.Equal(t, root, cause.GetKey())
}

func newDataTrieListingNode(t *testing.T, writer *state.AccountsDB, flags common.EnableEpochsHandler, root []byte) (*node.Node, state.AccountsRepository) {
	t.Helper()
	mainTrie, err := writer.GetTrie(root)
	require.NoError(t, err)
	chain := &testscommon.ChainHandlerStub{
		GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler { return &block.Header{Nonce: 1} },
		GetCurrentBlockRootHashCalled:    func() []byte { return root },
		GetCurrentBlockHeaderHashCalled:  func() []byte { return []byte("block") },
		GetFinalBlockInfoCalled:          func() (uint64, []byte, []byte) { return 1, []byte("block"), root },
	}
	args := func() state.ArgsAccountsDB {
		collector := disabled.NewDisabledStateAccessesCollector()
		holder, errHolder := triesHolder.NewReadOnlyDataTriesHolder(common.TenMbSize)
		require.NoError(t, errHolder)
		creator, errCreator := stateFactory.NewAccountCreator(stateFactory.ArgsAccountCreator{
			Hasher: integrationtests.TestHasher, Marshaller: integrationtests.TestMarshalizer,
			EnableEpochsHandler: flags, StateAccessesCollector: collector, DataTriesHolder: holder, DataTrieCreator: mainTrie,
		})
		require.NoError(t, errCreator)
		return state.ArgsAccountsDB{
			Trie: mainTrie, Hasher: integrationtests.TestHasher, Marshaller: integrationtests.TestMarshalizer,
			AccountFactory: creator, StateAccessesCollector: collector, DataTriesHolder: holder,
			StoragePruningManager: pruningDisabled.NewDisabledStoragePruningManager(),
			AddressConverter:      testscommon.RealWorldBech32PubkeyConverter, SnapshotsManager: disabled.NewDisabledSnapshotsManager(),
		}
	}
	current, err := stateFactory.CreateAccountsAdapterAPIOnCurrent(args(), chain)
	require.NoError(t, err)
	final, err := stateFactory.CreateAccountsAdapterAPIOnFinal(args(), chain)
	require.NoError(t, err)
	historical, err := stateFactory.CreateAccountsAdapterAPIOnHistorical(args())
	require.NoError(t, err)
	repo, err := state.NewAccountsRepository(state.ArgsAccountsRepository{
		CurrentStateAccountsWrapper: current, FinalStateAccountsWrapper: final, HistoricalStateAccountsWrapper: historical,
	})
	require.NoError(t, err)
	retriever, err := leavesRetriever.NewLeavesRetriever(mainTrie.GetStorageManager(), integrationtests.TestMarshalizer, integrationtests.TestHasher, common.TenMbSize)
	require.NoError(t, err)
	coreComponents := getDefaultCoreComponents()
	coreComponents.AddrPubKeyConv = testscommon.RealWorldBech32PubkeyConverter
	coreComponents.IntMarsh = integrationtests.TestMarshalizer
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsRepo = repo
	stateComponents.LeavesRetriever = retriever
	dataComponents := getDefaultDataComponents()
	dataComponents.BlockChain = chain
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = &testscommon.ShardsCoordinatorMock{NoShards: 1, CurrentShard: core.MetachainShardId}
	n, err := node.NewNode(
		node.WithCoreComponents(coreComponents), node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents), node.WithProcessComponents(processComponents),
		node.WithESDTNFTStorageHandler(&testscommon.EsdtStorageHandlerStub{
			GetESDTNFTTokenOnDestinationWithCustomSystemAccountCalled: func(vmcommon.UserAccountHandler, []byte, uint64, vmcommon.UserAccountHandler) (*esdt.ESDigitalToken, bool, error) {
				return &esdt.ESDigitalToken{Value: big.NewInt(7)}, false, nil
			},
		}),
	)
	require.NoError(t, err)
	return n, repo
}
