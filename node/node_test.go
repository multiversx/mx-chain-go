package node_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	atomicCore "github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/core/versioning"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/guardians"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext/esdtSupply"
	"github.com/multiversx/mx-chain-go/factory"
	factoryMock "github.com/multiversx/mx-chain-go/factory/mock"
	heartbeatData "github.com/multiversx/mx-chain-go/heartbeat/data"
	integrationTestsMock "github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/node"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/mock"
	nodeMockFactory "github.com/multiversx/mx-chain-go/node/mock/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/state/trackableDataTrie"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	factoryTests "github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/guardianMocks"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	mockStorage "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/testscommon/txsSenderMock"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testBlockInfo struct {
}

func (t testBlockInfo) apiResult() api.BlockInfo {
	return api.BlockInfo{
		Nonce:    37,
		Hash:     hex.EncodeToString([]byte("hash")),
		RootHash: hex.EncodeToString([]byte("root")),
	}
}

func (t testBlockInfo) forProcessing() common.BlockInfo {
	hash := []byte("hash")
	rHash := []byte("root")
	return holders.NewBlockInfo(hash, 37, rHash)
}

var dummyBlockInfo = testBlockInfo{}

func createMockPubkeyConverter() *testscommon.PubkeyConverterMock {
	return testscommon.NewPubkeyConverterMock(32)
}

func createAcc(address []byte) state.UserAccountHandler {
	dtlp, _ := parsers.NewDataTrieLeafParser(address, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
	dtt, _ := trackableDataTrie.NewTrackableDataTrie(address, &testscommon.HasherStub{}, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
	acc, _ := accounts.NewUserAccount(address, dtt, dtlp)

	return acc
}

func getAccAdapter(balance *big.Int) *stateMock.AccountsStub {
	accDB := &stateMock.AccountsStub{}
	accDB.GetExistingAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		acc := createAcc(address)
		_ = acc.AddToBalance(balance)
		acc.IncreaseNonce(1)

		return acc, nil
	}
	accDB.RecreateTrieCalled = func(_ []byte) error {
		return nil
	}

	return accDB
}

func getPrivateKey() *mock.PrivateKeyStub {
	return &mock.PrivateKeyStub{}
}

func getMessenger() *p2pmocks.MessengerStub {
	messenger := &p2pmocks.MessengerStub{
		CloseCalled: func() error {
			return nil
		},
		BootstrapCalled: func() error {
			return nil
		},
		BroadcastCalled: func(topic string, buff []byte) {
		},
	}

	return messenger
}

func getMarshalizer() marshal.Marshalizer {
	return &mock.MarshalizerFake{}
}

func getHasher() hashing.Hasher {
	return &mock.HasherMock{}
}

func TestNewNode(t *testing.T) {
	n, err := node.NewNode()

	assert.Nil(t, err)
	assert.NotNil(t, n)
}

func TestNewNode_NilOptionShouldError(t *testing.T) {
	t.Parallel()

	_, err := node.NewNode(node.WithCoreComponents(nil))
	assert.NotNil(t, err)
}

func TestNewNode_ApplyNilOptionShouldError(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()
	err := n.ApplyOptions(node.WithCoreComponents(nil))
	assert.NotNil(t, err)
}

func TestGetBalance_GetAccountFailsShouldError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("error")

	accountsRepository := &stateMock.AccountsRepositoryStub{}
	accountsRepository.GetAccountWithBlockInfoCalled = func(address []byte, options api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return nil, nil, expectedErr
	}

	dataComponents := getDefaultDataComponents()
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsRepo = accountsRepository

	n, _ := node.NewNode(
		node.WithDataComponents(dataComponents),
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)
	_, _, err := n.GetBalance(testscommon.TestAddressAlice, api.AccountQueryOptions{})
	assert.Equal(t, expectedErr, err)
}

func createDummyHexAddress(hexChars int) string {
	if hexChars < 1 {
		return ""
	}

	buff := make([]byte, hexChars/2)
	_, _ = rand.Reader.Read(buff)

	return hex.EncodeToString(buff)
}

func TestGetBalance_AccountNotFoundShouldReturnZeroBalance(t *testing.T) {
	t.Parallel()

	accountsRepository := &stateMock.AccountsRepositoryStub{}

	dataComponents := getDefaultDataComponents()
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsRepo = accountsRepository

	n, _ := node.NewNode(
		node.WithDataComponents(dataComponents),
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)
	balance, _, err := n.GetBalance(testscommon.TestAddressAlice, api.AccountQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(0), balance)
}

func TestNode_GetBalanceAccNotFoundShouldReturnEmpty(t *testing.T) {
	t.Parallel()

	accDB := &stateMock.AccountsStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
			return nil, nil, state.NewErrAccountNotFoundAtBlock(dummyBlockInfo.forProcessing())
		},
		RecreateTrieCalled: func(_ []byte) error {
			return nil
		},
	}

	dataComponents := getDefaultDataComponents()
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()

	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)

	n, _ := node.NewNode(
		node.WithDataComponents(dataComponents),
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	balance, bInfo, err := n.GetBalance(testscommon.TestAddressAlice, api.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, dummyBlockInfo.apiResult(), bInfo)
	require.Empty(t, balance)
}

func TestGetBalance(t *testing.T) {
	t.Parallel()

	testAccount, _ := accounts.NewUserAccount(testscommon.TestPubKeyAlice, &trieMock.DataTrieTrackerStub{}, &trieMock.TrieLeafParserStub{})
	testAccount.Balance = big.NewInt(100)

	accountsRepository := &stateMock.AccountsRepositoryStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
			return testAccount, nil, nil
		},
	}

	dataComponents := getDefaultDataComponents()
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsRepo = accountsRepository

	n, _ := node.NewNode(
		node.WithDataComponents(dataComponents),
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	balance, _, err := n.GetBalance(testscommon.TestAddressAlice, api.AccountQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(100), balance)
}

func TestGetUsername(t *testing.T) {
	t.Parallel()

	expectedUsername := []byte("elrond")

	testAccount, _ := accounts.NewUserAccount(testscommon.TestPubKeyAlice, &trieMock.DataTrieTrackerStub{}, &trieMock.TrieLeafParserStub{})
	testAccount.UserName = expectedUsername
	accountsRepository := &stateMock.AccountsRepositoryStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
			return testAccount, nil, nil
		},
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()

	dataComponents := getDefaultDataComponents()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsRepo = accountsRepository

	n, _ := node.NewNode(
		node.WithDataComponents(dataComponents),
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	username, _, err := n.GetUsername(testscommon.TestAddressAlice, api.AccountQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, string(expectedUsername), username)
}

func TestNode_GetCodeHashAccNotFoundShouldReturnEmpty(t *testing.T) {
	t.Parallel()

	accDB := &stateMock.AccountsStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
			return nil, nil, state.NewErrAccountNotFoundAtBlock(dummyBlockInfo.forProcessing())
		},
		RecreateTrieCalled: func(_ []byte) error {
			return nil
		},
	}

	dataComponents := getDefaultDataComponents()
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()

	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)

	n, _ := node.NewNode(
		node.WithDataComponents(dataComponents),
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	codeHash, bInfo, err := n.GetCodeHash("erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th", api.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, dummyBlockInfo.apiResult(), bInfo)
	require.Empty(t, codeHash)
}

func TestGetCodeHash(t *testing.T) {
	t.Parallel()

	expectedCodeHash := []byte("hash")

	testAccount, _ := accounts.NewUserAccount(testscommon.TestPubKeyAlice, &trieMock.DataTrieTrackerStub{}, &trieMock.TrieLeafParserStub{})
	testAccount.CodeHash = expectedCodeHash
	accountsRepository := &stateMock.AccountsRepositoryStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
			return testAccount, nil, nil
		},
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()

	dataComponents := getDefaultDataComponents()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsRepo = accountsRepository

	n, _ := node.NewNode(
		node.WithDataComponents(dataComponents),
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	codeHash, _, err := n.GetCodeHash(testscommon.TestAddressAlice, api.AccountQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, expectedCodeHash, codeHash)
}

func TestNode_GetKeyValuePairsAccNotFoundShouldReturnEmpty(t *testing.T) {
	t.Parallel()

	accDB := &stateMock.AccountsStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
			return nil, nil, state.NewErrAccountNotFoundAtBlock(dummyBlockInfo.forProcessing())
		},
		RecreateTrieCalled: func(_ []byte) error {
			return nil
		},
	}

	dataComponents := getDefaultDataComponents()
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()

	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)

	n, _ := node.NewNode(
		node.WithDataComponents(dataComponents),
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	pairs, bInfo, err := n.GetKeyValuePairs(testscommon.TestAddressAlice, api.AccountQueryOptions{}, context.Background())
	require.Nil(t, err)
	require.Equal(t, dummyBlockInfo.apiResult(), bInfo)
	require.Len(t, pairs, 0)
}

func TestNode_GetKeyValuePairs(t *testing.T) {
	t.Parallel()

	acc := createAcc([]byte("newaddress"))

	k1, v1 := []byte("key1"), []byte("value1")
	k2, v2 := []byte("key2"), []byte("value2")

	accDB := &stateMock.AccountsStub{}
	acc.SetDataTrie(
		&trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.KeyBuilder, tlp common.TrieLeafParser) error {
				go func() {
					suffix := append(k1, acc.AddressBytes()...)
					trieLeaf, _ := tlp.ParseLeaf(k1, append(v1, suffix...), core.NotSpecified)
					leavesChannels.LeavesChan <- trieLeaf

					suffix = append(k2, acc.AddressBytes()...)
					trieLeaf2, _ := tlp.ParseLeaf(k2, append(v2, suffix...), core.NotSpecified)
					leavesChannels.LeavesChan <- trieLeaf2
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
				}()

				return nil
			},
			RootCalled: func() ([]byte, error) {
				return nil, nil
			},
		})

	accDB.GetAccountWithBlockInfoCalled = func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return acc, nil, nil
	}
	accDB.RecreateTrieCalled = func(rootHash []byte) error {
		return nil
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	dataComponents := getDefaultDataComponents()
	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithDataComponents(dataComponents),
	)

	pairs, _, err := n.GetKeyValuePairs(createDummyHexAddress(64), api.AccountQueryOptions{}, context.Background())
	assert.Nil(t, err)
	resV1, ok := pairs[hex.EncodeToString(k1)]
	assert.True(t, ok)
	assert.Equal(t, hex.EncodeToString(v1), resV1)

	resV2, ok := pairs[hex.EncodeToString(k2)]
	assert.True(t, ok)
	assert.Equal(t, hex.EncodeToString(v2), resV2)
}

func TestNode_GetKeyValuePairs_GetAllLeavesShouldFail(t *testing.T) {
	t.Parallel()

	acc := createAcc([]byte("newaddress"))

	accDB := &stateMock.AccountsStub{}

	expectedErr := errors.New("expected err")
	acc.SetDataTrie(
		&trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.KeyBuilder, _ common.TrieLeafParser) error {
				go func() {
					leavesChannels.ErrChan.WriteInChanNonBlocking(expectedErr)
					close(leavesChannels.LeavesChan)
				}()

				return nil
			},
			RootCalled: func() ([]byte, error) {
				return nil, nil
			},
		})

	accDB.GetAccountWithBlockInfoCalled = func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return acc, nil, nil
	}
	accDB.RecreateTrieCalled = func(rootHash []byte) error {
		return nil
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	dataComponents := getDefaultDataComponents()
	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithDataComponents(dataComponents),
	)

	pairs, blockInfo, err := n.GetKeyValuePairs(createDummyHexAddress(64), api.AccountQueryOptions{}, context.Background())
	assert.Nil(t, pairs)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, api.BlockInfo{}, blockInfo)
}

func TestNode_GetKeyValuePairsContextShouldTimeout(t *testing.T) {
	t.Parallel()

	acc := createAcc([]byte("newaddress"))

	accDB := &stateMock.AccountsStub{}
	acc.SetDataTrie(
		&trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.KeyBuilder, _ common.TrieLeafParser) error {
				go func() {
					time.Sleep(time.Second)
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
				}()

				return nil
			},
			RootCalled: func() ([]byte, error) {
				return nil, nil
			},
		})

	accDB.GetAccountWithBlockInfoCalled = func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return acc, nil, nil
	}
	accDB.RecreateTrieCalled = func(rootHash []byte) error {
		return nil
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	dataComponents := getDefaultDataComponents()
	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithDataComponents(dataComponents),
	)

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	pairs, _, err := n.GetKeyValuePairs(createDummyHexAddress(64), api.AccountQueryOptions{}, ctxWithTimeout)
	assert.Nil(t, pairs)
	assert.Equal(t, node.ErrTrieOperationsTimeout, err)
}

func TestNode_GetValueForKeyAccNotFoundShouldReturnEmpty(t *testing.T) {
	t.Parallel()

	accDB := &stateMock.AccountsStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
			return nil, nil, state.NewErrAccountNotFoundAtBlock(dummyBlockInfo.forProcessing())
		},
		RecreateTrieCalled: func(_ []byte) error {
			return nil
		},
	}

	dataComponents := getDefaultDataComponents()
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()

	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)

	n, _ := node.NewNode(
		node.WithDataComponents(dataComponents),
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	value, bInfo, err := n.GetValueForKey(testscommon.TestAddressAlice, "0a0a", api.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, dummyBlockInfo.apiResult(), bInfo)
	require.Empty(t, value)
}

func TestNode_GetValueForKey(t *testing.T) {
	t.Parallel()

	acc := createAcc([]byte("newaddress"))

	k1, v1 := []byte("key1"), []byte("value1")
	_ = acc.SaveKeyValue(k1, v1)

	accDB := &stateMock.AccountsStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
			return acc, nil, nil
		},
		RecreateTrieCalled: func(_ []byte) error {
			return nil
		},
	}

	dataComponents := getDefaultDataComponents()
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)

	n, _ := node.NewNode(
		node.WithDataComponents(dataComponents),
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	value, _, err := n.GetValueForKey(createDummyHexAddress(64), hex.EncodeToString(k1), api.AccountQueryOptions{})
	assert.NoError(t, err)
	assert.Equal(t, hex.EncodeToString(v1), value)
}

func TestNode_GetESDTDataAccNotFoundShouldReturnEmpty(t *testing.T) {
	t.Parallel()

	esdtToken := "newToken"

	accDB := &stateMock.AccountsStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
			return nil, nil, state.NewErrAccountNotFoundAtBlock(dummyBlockInfo.forProcessing())
		},
		RecreateTrieCalled: func(_ []byte) error {
			return nil
		},
	}

	dataComponents := getDefaultDataComponents()
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()

	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)

	n, _ := node.NewNode(
		node.WithDataComponents(dataComponents),
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	esdtTokenData, bInfo, err := n.GetESDTData(testscommon.TestAddressAlice, esdtToken, 0, api.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, dummyBlockInfo.apiResult(), bInfo)
	require.Equal(t, "0", esdtTokenData.Value.String())
}

func TestNode_GetESDTData(t *testing.T) {
	t.Parallel()

	acc := createAcc(testscommon.TestPubKeyAlice)
	esdtToken := "newToken"

	esdtData := &esdt.ESDigitalToken{Value: big.NewInt(10)}

	accDB := &stateMock.AccountsStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
			return acc, nil, nil
		},
		RecreateTrieCalled: func(_ []byte) error {
			return nil
		},
	}

	esdtStorageStub := &testscommon.EsdtStorageHandlerStub{
		GetESDTNFTTokenOnDestinationWithCustomSystemAccountCalled: func(acnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64, _ vmcommon.UserAccountHandler) (*esdt.ESDigitalToken, bool, error) {
			return esdtData, false, nil
		},
	}

	dataComponents := getDefaultDataComponents()
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()

	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)

	n, _ := node.NewNode(
		node.WithDataComponents(dataComponents),
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithESDTNFTStorageHandler(esdtStorageStub),
	)

	esdtTokenData, _, err := n.GetESDTData(testscommon.TestAddressAlice, esdtToken, 0, api.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, esdtData.Value.String(), esdtTokenData.Value.String())
}

func TestNode_GetESDTDataForNFT(t *testing.T) {
	t.Parallel()

	acc := createAcc(testscommon.TestPubKeyAlice)
	esdtToken := "newToken"
	nonce := int64(100)

	esdtData := &esdt.ESDigitalToken{Value: big.NewInt(10)}

	esdtStorageStub := &testscommon.EsdtStorageHandlerStub{
		GetESDTNFTTokenOnDestinationWithCustomSystemAccountCalled: func(acnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64, _ vmcommon.UserAccountHandler) (*esdt.ESDigitalToken, bool, error) {
			return esdtData, false, nil
		},
	}
	accDB := &stateMock.AccountsStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
			return acc, nil, nil
		},
		RecreateTrieCalled: func(_ []byte) error {
			return nil
		},
	}

	dataComponents := getDefaultDataComponents()
	coreComponents := getDefaultCoreComponents()
	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)

	n, _ := node.NewNode(
		node.WithDataComponents(dataComponents),
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithESDTNFTStorageHandler(esdtStorageStub),
	)

	esdtTokenData, _, err := n.GetESDTData(testscommon.TestAddressAlice, esdtToken, uint64(nonce), api.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, esdtData.Value.String(), esdtTokenData.Value.String())
}

func TestNode_GetAllESDTTokens(t *testing.T) {
	t.Parallel()

	acc := createAcc(testscommon.TestPubKeyAlice)
	esdtToken := "newToken"
	esdtKey := []byte(core.ProtectedKeyPrefix + core.ESDTKeyIdentifier + esdtToken)

	esdtData := &esdt.ESDigitalToken{Value: big.NewInt(10)}

	esdtStorageStub := &testscommon.EsdtStorageHandlerStub{
		GetESDTNFTTokenOnDestinationWithCustomSystemAccountCalled: func(acnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64, _ vmcommon.UserAccountHandler) (*esdt.ESDigitalToken, bool, error) {
			return esdtData, false, nil
		},
	}

	acc.SetDataTrie(
		&trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.KeyBuilder, _ common.TrieLeafParser) error {
				go func() {
					trieLeaf := keyValStorage.NewKeyValStorage(esdtKey, nil)
					leavesChannels.LeavesChan <- trieLeaf
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
				}()

				return nil
			},
			RootCalled: func() ([]byte, error) {
				return nil, nil
			},
		})

	accDB := &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	accDB.GetAccountWithBlockInfoCalled = func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return acc, nil, nil
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()

	dataComponents := getDefaultDataComponents()
	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithDataComponents(dataComponents),
		node.WithESDTNFTStorageHandler(esdtStorageStub),
	)

	value, _, err := n.GetAllESDTTokens(testscommon.TestAddressAlice, api.AccountQueryOptions{}, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(value))
	assert.Equal(t, esdtData, value[esdtToken])
}

func TestNode_GetAllESDTTokens_GetAllLeavesShouldFail(t *testing.T) {
	t.Parallel()

	acc := createAcc(testscommon.TestPubKeyAlice)

	expectedErr := errors.New("expected error")
	acc.SetDataTrie(
		&trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.KeyBuilder, _ common.TrieLeafParser) error {
				go func() {
					leavesChannels.ErrChan.WriteInChanNonBlocking(expectedErr)
					close(leavesChannels.LeavesChan)
				}()

				return nil
			},
			RootCalled: func() ([]byte, error) {
				return nil, nil
			},
		})

	accDB := &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	accDB.GetAccountWithBlockInfoCalled = func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return acc, nil, nil
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()

	dataComponents := getDefaultDataComponents()
	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithDataComponents(dataComponents),
	)

	value, blockInfo, err := n.GetAllESDTTokens(testscommon.TestAddressAlice, api.AccountQueryOptions{}, context.Background())
	assert.Nil(t, value)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, api.BlockInfo{}, blockInfo)
}

func TestNode_GetAllESDTTokensContextShouldTimeout(t *testing.T) {
	t.Parallel()

	acc := createAcc(testscommon.TestPubKeyAlice)

	acc.SetDataTrie(
		&trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.KeyBuilder, _ common.TrieLeafParser) error {
				go func() {
					time.Sleep(time.Second)
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
				}()

				return nil
			},
			RootCalled: func() ([]byte, error) {
				return nil, nil
			},
		})

	accDB := &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	accDB.GetAccountWithBlockInfoCalled = func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return acc, nil, nil
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()

	dataComponents := getDefaultDataComponents()
	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithDataComponents(dataComponents),
	)

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	value, _, err := n.GetAllESDTTokens(testscommon.TestAddressAlice, api.AccountQueryOptions{}, ctxWithTimeout)
	assert.Nil(t, value)
	assert.Equal(t, node.ErrTrieOperationsTimeout, err)
}

func TestNode_GetAllESDTsAccNotFoundShouldReturnEmpty(t *testing.T) {
	t.Parallel()

	accDB := &stateMock.AccountsStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
			return nil, nil, state.NewErrAccountNotFoundAtBlock(dummyBlockInfo.forProcessing())
		},
		RecreateTrieCalled: func(_ []byte) error {
			return nil
		},
	}

	dataComponents := getDefaultDataComponents()
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()

	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)

	n, _ := node.NewNode(
		node.WithDataComponents(dataComponents),
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	tokens, bInfo, err := n.GetAllESDTTokens(testscommon.TestAddressAlice, api.AccountQueryOptions{}, context.Background())
	require.Nil(t, err)
	require.Equal(t, dummyBlockInfo.apiResult(), bInfo)
	require.Len(t, tokens, 0)
}

func TestNode_GetAllESDTTokensShouldReturnEsdtAndFormattedNft(t *testing.T) {
	t.Parallel()

	acc := createAcc(testscommon.TestPubKeyAlice)

	esdtToken := "TKKR-7q8w9e"
	esdtKey := []byte(core.ProtectedKeyPrefix + core.ESDTKeyIdentifier + esdtToken)

	esdtData := &esdt.ESDigitalToken{Value: big.NewInt(10)}
	marshalledData, _ := getMarshalizer().Marshal(esdtData)
	suffix := append(esdtKey, acc.AddressBytes()...)

	nftToken := "TCKR-67tgv3"
	nftNonce := big.NewInt(1)
	nftKey := []byte(core.ProtectedKeyPrefix + core.ESDTKeyIdentifier + nftToken)
	nftKeyWithBytes := append(nftKey, nftNonce.Bytes()...)
	nftSuffix := append(nftKeyWithBytes, acc.AddressBytes()...)

	nftMetaData := &esdt.MetaData{Nonce: nftNonce.Uint64(), Creator: []byte("12345678901234567890123456789012")}
	nftData := &esdt.ESDigitalToken{Type: uint32(core.NonFungible), Value: big.NewInt(10), TokenMetaData: nftMetaData}
	marshalledNftData, _ := getMarshalizer().Marshal(nftData)

	esdtStorageStub := &testscommon.EsdtStorageHandlerStub{
		GetESDTNFTTokenOnDestinationWithCustomSystemAccountCalled: func(acnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64, _ vmcommon.UserAccountHandler) (*esdt.ESDigitalToken, bool, error) {
			switch string(esdtTokenKey) {
			case string(esdtKey):
				return esdtData, false, nil
			case string(nftKey):
				return nftData, false, nil
			default:
				return nil, false, nil
			}
		},
	}
	acc.SetDataTrie(
		&trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.KeyBuilder, _ common.TrieLeafParser) error {
				wg := &sync.WaitGroup{}
				wg.Add(1)
				go func() {
					trieLeaf := keyValStorage.NewKeyValStorage(esdtKey, append(marshalledData, suffix...))
					leavesChannels.LeavesChan <- trieLeaf

					trieLeaf = keyValStorage.NewKeyValStorage(nftKey, append(marshalledNftData, nftSuffix...))
					leavesChannels.LeavesChan <- trieLeaf
					wg.Done()
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
				}()

				wg.Wait()

				return nil
			},
			RootCalled: func() ([]byte, error) {
				return nil, nil
			},
		})

	accDB := &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	accDB.GetAccountWithBlockInfoCalled = func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return acc, nil, nil
	}

	coreComponents := getDefaultCoreComponents()
	dataComponents := getDefaultDataComponents()
	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithESDTNFTStorageHandler(esdtStorageStub),
	)

	tokens, _, err := n.GetAllESDTTokens(testscommon.TestAddressAlice, api.AccountQueryOptions{}, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 2, len(tokens))
	assert.Equal(t, esdtData, tokens[esdtToken])

	// check that the NFT was formatted correctly
	expectedNftFormattedKey := "TCKR-67tgv3-01"
	assert.NotNil(t, tokens[expectedNftFormattedKey])
	assert.Equal(t, uint64(1), tokens[expectedNftFormattedKey].TokenMetaData.Nonce)
}

func testNodeGetAllIssuedESDTs(t *testing.T, nodeFactory node.NodeFactory, shardId uint32) {
	acc := createAcc([]byte("newaddress"))
	esdtToken := []byte("TCK-RANDOM")
	sftToken := []byte("SFT-RANDOM")
	nftToken := []byte("NFT-RANDOM")

	esdtData := &systemSmartContracts.ESDTDataV2{TokenName: []byte("fungible"), TokenType: []byte(core.FungibleESDT)}
	marshalledData, _ := getMarshalizer().Marshal(esdtData)
	_ = acc.SaveKeyValue(esdtToken, marshalledData)

	sftData := &systemSmartContracts.ESDTDataV2{TokenName: []byte("semi fungible"), TokenType: []byte(core.SemiFungibleESDT)}
	sftMarshalledData, _ := getMarshalizer().Marshal(sftData)
	_ = acc.SaveKeyValue(sftToken, sftMarshalledData)

	nftData := &systemSmartContracts.ESDTDataV2{TokenName: []byte("non fungible"), TokenType: []byte(core.NonFungibleESDT)}
	nftMarshalledData, _ := getMarshalizer().Marshal(nftData)
	_ = acc.SaveKeyValue(nftToken, nftMarshalledData)

	esdtSuffix := append(esdtToken, acc.AddressBytes()...)
	nftSuffix := append(nftToken, acc.AddressBytes()...)
	sftSuffix := append(sftToken, acc.AddressBytes()...)

	acc.SetDataTrie(
		&trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.KeyBuilder, tlp common.TrieLeafParser) error {
				go func() {
					trieLeaf, _ := tlp.ParseLeaf(esdtToken, append(marshalledData, esdtSuffix...), core.NotSpecified)
					leavesChannels.LeavesChan <- trieLeaf

					trieLeaf, _ = tlp.ParseLeaf(sftToken, append(sftMarshalledData, sftSuffix...), core.NotSpecified)
					leavesChannels.LeavesChan <- trieLeaf

					trieLeaf, _ = tlp.ParseLeaf(nftToken, append(nftMarshalledData, nftSuffix...), core.NotSpecified)
					leavesChannels.LeavesChan <- trieLeaf
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
				}()

				return nil
			},
			RootCalled: func() ([]byte, error) {
				return nil, nil
			},
		})

	accDB := &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	accDB.GetAccountWithBlockInfoCalled = func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return acc, nil, nil
	}

	coreComponents := getDefaultCoreComponents()
	dataComponents := getDefaultDataComponents()
	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = &mock.ShardCoordinatorMock{
		SelfShardId: shardId,
	}

	n, _ := nodeFactory.CreateNewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
	)

	value, err := n.GetAllIssuedESDTs(core.FungibleESDT, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(value))
	assert.Equal(t, string(esdtToken), value[0])

	value, err = n.GetAllIssuedESDTs(core.SemiFungibleESDT, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(value))
	assert.Equal(t, string(sftToken), value[0])

	value, err = n.GetAllIssuedESDTs(core.NonFungibleESDT, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(value))
	assert.Equal(t, string(nftToken), value[0])

	value, err = n.GetAllIssuedESDTs("", context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 3, len(value))
}

func TestNode_GetAllESDTTokens_ShouldWork(t *testing.T) {
	t.Parallel()

	testNodeGetAllIssuedESDTs(t, node.NewNodeFactory(), core.MetachainShardId)
}

func TestNode_GetAllESDTTokens_WrongChainId(t *testing.T) {
	t.Parallel()

	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = &mock.ShardCoordinatorMock{
		SelfShardId: 0,
	}
	n, err := node.NewNode(node.WithProcessComponents(processComponents))
	require.Nil(t, err)

	value, err := n.GetAllIssuedESDTs("", context.Background())
	require.Nil(t, value)
	require.Equal(t, node.ErrMetachainOnlyEndpoint, err)
}

func testNodeGetESDTsWithRole(t *testing.T, nodeFactory node.NodeFactory, shardId uint32) {
	addrBytes := testscommon.TestPubKeyAlice
	acc := createAcc(addrBytes)
	esdtToken := []byte("TCK-RANDOM")

	specialRoles := []*systemSmartContracts.ESDTRoles{
		{
			Address: addrBytes,
			Roles:   [][]byte{[]byte(core.ESDTRoleNFTAddQuantity), []byte(core.ESDTRoleLocalMint)},
		},
	}

	esdtData := &systemSmartContracts.ESDTDataV2{TokenName: []byte("fungible"), TokenType: []byte(core.FungibleESDT), SpecialRoles: specialRoles}
	marshalledData, _ := getMarshalizer().Marshal(esdtData)
	_ = acc.SaveKeyValue(esdtToken, marshalledData)

	esdtSuffix := append(esdtToken, acc.AddressBytes()...)

	acc.SetDataTrie(
		&trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.KeyBuilder, tlp common.TrieLeafParser) error {
				go func() {
					trieLeaf, _ := tlp.ParseLeaf(esdtToken, append(marshalledData, esdtSuffix...), core.NotSpecified)
					leavesChannels.LeavesChan <- trieLeaf
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
				}()

				return nil
			},
			RootCalled: func() ([]byte, error) {
				return nil, nil
			},
		})

	accDB := &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	accDB.GetAccountWithBlockInfoCalled = func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return acc, nil, nil
	}
	coreComponents := getDefaultCoreComponents()
	dataComponents := getDefaultDataComponents()
	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = &mock.ShardCoordinatorMock{
		SelfShardId: shardId,
	}
	n, _ := nodeFactory.CreateNewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
	)

	tokenResult, _, err := n.GetESDTsWithRole(testscommon.TestAddressAlice, core.ESDTRoleNFTAddQuantity, api.AccountQueryOptions{}, context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, len(tokenResult))
	require.Equal(t, string(esdtToken), tokenResult[0])

	tokenResult, _, err = n.GetESDTsWithRole(testscommon.TestAddressAlice, core.ESDTRoleLocalMint, api.AccountQueryOptions{}, context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, len(tokenResult))
	require.Equal(t, string(esdtToken), tokenResult[0])

	tokenResult, _, err = n.GetESDTsWithRole(testscommon.TestAddressAlice, core.ESDTRoleNFTCreate, api.AccountQueryOptions{}, context.Background())
	require.NoError(t, err)
	require.Len(t, tokenResult, 0)
}

func TestNode_GetESDTsWithRole_ShouldWork(t *testing.T) {
	t.Parallel()

	testNodeGetESDTsWithRole(t, node.NewNodeFactory(), core.MetachainShardId)
}

func TestNode_GetESDTsWithRole_WrongChainId(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = &mock.ShardCoordinatorMock{
		SelfShardId: 0,
	}
	n, err := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithProcessComponents(processComponents))
	require.Nil(t, err)

	testResult, _, err := n.GetESDTsWithRole(testscommon.TestAddressAlice, core.ESDTRoleNFTCreate, api.AccountQueryOptions{}, context.Background())
	require.Nil(t, testResult)
	require.Equal(t, node.ErrMetachainOnlyEndpoint, err)
}

func testNodeGetESDTsRoles(t *testing.T, nodeFactory node.NodeFactory, shardId uint32) {
	addrBytes := testscommon.TestPubKeyAlice
	acc := createAcc(addrBytes)
	esdtToken := []byte("TCK-RANDOM")

	specialRoles := []*systemSmartContracts.ESDTRoles{
		{
			Address: addrBytes,
			Roles:   [][]byte{[]byte(core.ESDTRoleNFTAddQuantity), []byte(core.ESDTRoleLocalMint)},
		},
	}

	esdtData := &systemSmartContracts.ESDTDataV2{TokenName: []byte("fungible"), TokenType: []byte(core.FungibleESDT), SpecialRoles: specialRoles}
	marshalledData, _ := getMarshalizer().Marshal(esdtData)

	esdtSuffix := append(esdtToken, acc.AddressBytes()...)

	acc.SetDataTrie(
		&trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.KeyBuilder, tlp common.TrieLeafParser) error {
				go func() {
					trieLeaf, _ := tlp.ParseLeaf(esdtToken, append(marshalledData, esdtSuffix...), core.NotSpecified)
					leavesChannels.LeavesChan <- trieLeaf
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
				}()

				return nil
			},
			RootCalled: func() ([]byte, error) {
				return nil, nil
			},
		})

	accDB := &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	accDB.GetAccountWithBlockInfoCalled = func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return acc, nil, nil
	}
	coreComponents := getDefaultCoreComponents()
	stateComponents := getDefaultStateComponents()
	dataComponents := getDefaultDataComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = &mock.ShardCoordinatorMock{
		SelfShardId: shardId,
	}
	n, _ := nodeFactory.CreateNewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
	)

	tokenResult, _, err := n.GetESDTsRoles(testscommon.TestAddressAlice, api.AccountQueryOptions{}, context.Background())
	require.NoError(t, err)
	require.Equal(t, map[string][]string{
		string(esdtToken): {core.ESDTRoleNFTAddQuantity, core.ESDTRoleLocalMint},
	}, tokenResult)
}

func TestNode_GetESDTsRoles_ShouldWork(t *testing.T) {
	t.Parallel()

	testNodeGetESDTsWithRole(t, node.NewNodeFactory(), core.MetachainShardId)
}

func TestNode_GetESDTsRoles_WrongChainId(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = &mock.ShardCoordinatorMock{
		SelfShardId: 0,
	}
	n, err := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithProcessComponents(processComponents))
	require.Nil(t, err)

	tokenResult, _, err := n.GetESDTsRoles(testscommon.TestAddressAlice, api.AccountQueryOptions{}, context.Background())
	require.Nil(t, tokenResult)
	require.Equal(t, node.ErrMetachainOnlyEndpoint, err)
}

func testNodeGetNFTTokenIDsRegisteredByAddress(t *testing.T, nodeFactory node.NodeFactory, shardId uint32) {
	addrBytes := testscommon.TestPubKeyAlice
	acc := createAcc(addrBytes)
	esdtToken := []byte("TCK-RANDOM")

	esdtData := &systemSmartContracts.ESDTDataV2{TokenName: []byte("fungible"), TokenType: []byte(core.SemiFungibleESDT), OwnerAddress: addrBytes}
	marshalledData, _ := getMarshalizer().Marshal(esdtData)
	_ = acc.SaveKeyValue(esdtToken, marshalledData)

	esdtSuffix := append(esdtToken, acc.AddressBytes()...)

	acc.SetDataTrie(
		&trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.KeyBuilder, tlp common.TrieLeafParser) error {
				go func() {
					trieLeaf, _ := tlp.ParseLeaf(esdtToken, append(marshalledData, esdtSuffix...), core.NotSpecified)
					leavesChannels.LeavesChan <- trieLeaf
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
				}()

				return nil
			},
			RootCalled: func() ([]byte, error) {
				return nil, nil
			},
		},
	)

	accDB := &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	accDB.GetAccountWithBlockInfoCalled = func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return acc, nil, nil
	}
	coreComponents := getDefaultCoreComponents()
	stateComponents := getDefaultStateComponents()
	dataComponents := getDefaultDataComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = &mock.ShardCoordinatorMock{
		SelfShardId: shardId,
	}
	n, _ := nodeFactory.CreateNewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
	)

	tokenResult, _, err := n.GetNFTTokenIDsRegisteredByAddress(testscommon.TestAddressAlice, api.AccountQueryOptions{}, context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, len(tokenResult))
	require.Equal(t, string(esdtToken), tokenResult[0])
}

func TestNode_GetNFTTokenIDsRegisteredByAddress_ShouldWork(t *testing.T) {
	t.Parallel()

	testNodeGetNFTTokenIDsRegisteredByAddress(t, node.NewNodeFactory(), core.MetachainShardId)
}

func TestNode_GetNFTTokenIDsRegisteredByAddress_WrongChainId(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = &mock.ShardCoordinatorMock{
		SelfShardId: 0,
	}
	n, err := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithProcessComponents(processComponents))
	require.Nil(t, err)

	tokenResult, _, err := n.GetNFTTokenIDsRegisteredByAddress(testscommon.TestAddressAlice, api.AccountQueryOptions{}, context.Background())
	require.Nil(t, tokenResult)
	require.Equal(t, node.ErrMetachainOnlyEndpoint, err)
}

func TestNode_GetNFTTokenIDsRegisteredByAddressContextShouldTimeout(t *testing.T) {
	t.Parallel()

	addrBytes := testscommon.TestPubKeyAlice
	acc := createAcc(addrBytes)

	acc.SetDataTrie(
		&trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.KeyBuilder, _ common.TrieLeafParser) error {
				go func() {
					time.Sleep(time.Second)
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
				}()

				return nil
			},
			RootCalled: func() ([]byte, error) {
				return nil, nil
			},
		},
	)

	accDB := &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	accDB.GetAccountWithBlockInfoCalled = func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return acc, nil, nil
	}
	coreComponents := getDefaultCoreComponents()
	stateComponents := getDefaultStateComponents()
	dataComponents := getDefaultDataComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = &mock.ShardCoordinatorMock{
		SelfShardId: core.MetachainShardId,
	}
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
	)

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	tokens, _, err := n.GetNFTTokenIDsRegisteredByAddress(testscommon.TestAddressAlice, api.AccountQueryOptions{}, ctxWithTimeout)
	require.Nil(t, tokens)
	require.Equal(t, node.ErrTrieOperationsTimeout, err)
}

// ------- GenerateTransaction

func TestGenerateTransaction_NoAddrConverterShouldError(t *testing.T) {
	t.Parallel()

	privateKey := getPrivateKey()
	coreComponents := getDefaultCoreComponents()
	coreComponents.AddrPubKeyConv = nil
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_NoAccAdapterShouldError(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)
	stateComponents.AccountsAPI = nil
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", &mock.PrivateKeyStub{}, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_NoPrivateKeyShouldError(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", nil, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_CreateAddressFailsShouldError(t *testing.T) {
	t.Parallel()

	accAdapter := getAccAdapter(big.NewInt(0))
	privateKey := getPrivateKey()

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = accAdapter

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_GetAccountFailsShouldError(t *testing.T) {
	t.Parallel()

	accAdapter := &stateMock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return nil, nil
		},
	}
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = accAdapter
	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = &mock.SinglesignMock{}

	privateKey := getPrivateKey()
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithCryptoComponents(cryptoComponents),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_GetAccountReturnsNilShouldWork(t *testing.T) {
	t.Parallel()

	accAdapter := &stateMock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return createAcc(address), nil
		},
	}
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = accAdapter
	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = singleSigner

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithCryptoComponents(cryptoComponents),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.Nil(t, err)
}

func TestGenerateTransaction_GetExistingAccountShouldWork(t *testing.T) {
	t.Parallel()

	accAdapter := getAccAdapter(big.NewInt(0))
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = accAdapter
	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = singleSigner

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithCryptoComponents(cryptoComponents),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.Nil(t, err)
}

func TestGenerateTransaction_MarshalErrorsShouldError(t *testing.T) {
	t.Parallel()

	accAdapter := getAccAdapter(big.NewInt(0))
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}
	marshalizer := &mock.MarshalizerMock{
		MarshalHandler: func(obj interface{}) ([]byte, error) {
			return nil, errors.New("error")
		},
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = marshalizer
	coreComponents.VmMarsh = marshalizer
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = accAdapter
	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = singleSigner

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithCryptoComponents(cryptoComponents),
	)
	_, err := n.GenerateTransaction("sender", "receiver", big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_SignTxErrorsShouldError(t *testing.T) {
	t.Parallel()

	accAdapter := getAccAdapter(big.NewInt(0))
	privateKey := &mock.PrivateKeyStub{}
	singleSigner := &mock.SinglesignFailMock{}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = accAdapter
	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = singleSigner

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithCryptoComponents(cryptoComponents),
	)
	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.NotNil(t, err)
}

func TestGenerateTransaction_ShouldSetCorrectSignature(t *testing.T) {
	t.Parallel()

	accAdapter := getAccAdapter(big.NewInt(0))
	signature := []byte("signed")
	privateKey := &mock.PrivateKeyStub{}
	singleSigner := &mock.SinglesignMock{}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = accAdapter
	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = singleSigner

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithCryptoComponents(cryptoComponents),
	)

	tx, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.Nil(t, err)
	assert.Equal(t, signature, tx.Signature)
}

func TestGenerateTransaction_ShouldSetCorrectNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(7)
	accAdapter := &stateMock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			acc := createAcc(address)
			_ = acc.AddToBalance(big.NewInt(0))
			acc.IncreaseNonce(nonce)

			return acc, nil
		},
	}

	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = accAdapter
	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = singleSigner

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithCryptoComponents(cryptoComponents),
	)

	tx, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.Nil(t, err)
	assert.Equal(t, nonce, tx.Nonce)
}

func TestGenerateTransaction_CorrectParamsShouldNotError(t *testing.T) {
	t.Parallel()

	accAdapter := getAccAdapter(big.NewInt(0))
	privateKey := getPrivateKey()
	singleSigner := &mock.SinglesignMock{}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = accAdapter
	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = singleSigner

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithCryptoComponents(cryptoComponents),
	)

	_, err := n.GenerateTransaction(createDummyHexAddress(64), createDummyHexAddress(64), big.NewInt(10), "code", privateKey, []byte("chainID"), 1)
	assert.Nil(t, err)
}

func getDefaultTransactionArgs() *external.ArgsCreateTransaction {
	return &external.ArgsCreateTransaction{
		Nonce:            uint64(0),
		Value:            new(big.Int).SetInt64(10).String(),
		Receiver:         "rcv",
		ReceiverUsername: []byte("rcvrUsername"),
		Sender:           "snd",
		SenderUsername:   []byte("sndrUsername"),
		GasPrice:         uint64(10),
		GasLimit:         uint64(20),
		DataField:        []byte("-"),
		SignatureHex:     hex.EncodeToString(bytes.Repeat([]byte{0}, 10)),
		ChainID:          "chainID",
		Version:          1,
		Options:          0,
		Guardian:         "",
		GuardianSigHex:   "",
	}
}

func TestCreateTransaction_NilArgsShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	tx, txHash, err := n.CreateTransaction(nil)

	assert.Nil(t, tx)
	assert.Nil(t, txHash)
	assert.Equal(t, node.ErrNilCreateTransactionArgs, err)
}

func TestCreateTransaction_NilAddrConverterShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	coreComponents.AddrPubKeyConv = nil
	txArgs := getDefaultTransactionArgs()
	tx, txHash, err := n.CreateTransaction(txArgs)

	assert.Nil(t, tx)
	assert.Nil(t, txHash)
	assert.Equal(t, node.ErrNilPubkeyConverter, err)
}

func TestCreateTransaction_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
	}

	stateComponents := getDefaultStateComponents()
	processComponents := getDefaultProcessComponents()

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
	)

	stateComponents.AccountsAPI = nil

	txArgs := getDefaultTransactionArgs()
	tx, txHash, err := n.CreateTransaction(txArgs)

	assert.Nil(t, tx)
	assert.Nil(t, txHash)
	assert.Equal(t, node.ErrNilAccountsAdapter, err)
}

func TestCreateTransaction_InvalidSignatureShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
	}
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	txArgs := getDefaultTransactionArgs()
	txArgs.SignatureHex = "-"
	tx, txHash, err := n.CreateTransaction(txArgs)

	assert.Nil(t, tx)
	assert.Nil(t, txHash)
	assert.NotNil(t, err)
}

func TestCreateTransaction_ChainIDFieldChecks(t *testing.T) {
	t.Parallel()

	chainID := "chain id"
	expectedHash := []byte("expected hash")
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = mock.HasherMock{
		ComputeCalled: func(s string) []byte {
			return expectedHash
		},
	}
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
		EncodeCalled: func(pkBytes []byte) (string, error) {
			return string(pkBytes), nil
		},
		LenCalled: func() int {
			return 3
		},
	}

	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithAddressSignatureSize(10),
	)

	signature := hex.EncodeToString([]byte(strings.Repeat("s", 10)))
	emptyChainID := ""
	txArgs := getDefaultTransactionArgs()
	txArgs.SignatureHex = signature
	txArgs.ChainID = emptyChainID
	_, _, err := n.CreateTransaction(txArgs)
	assert.Equal(t, node.ErrInvalidChainIDInTransaction, err)

	for i := 1; i < len(chainID); i++ {
		newChainID := strings.Repeat("c", i)
		txArgs = getDefaultTransactionArgs()
		txArgs.SignatureHex = signature
		txArgs.ChainID = newChainID
		_, _, err = n.CreateTransaction(txArgs)
		assert.NoError(t, err)
	}

	newChainID := chainID + "additional text"
	txArgs = getDefaultTransactionArgs()
	txArgs.SignatureHex = signature
	txArgs.ChainID = newChainID

	_, _, err = n.CreateTransaction(txArgs)
	assert.Equal(t, node.ErrInvalidChainIDInTransaction, err)
}

func TestCreateTransaction_InvalidTxVersionShouldErr(t *testing.T) {
	t.Parallel()

	expectedHash := []byte("expected hash")
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = mock.HasherMock{
		ComputeCalled: func(s string) []byte {
			return expectedHash
		},
	}
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
		EncodeCalled: func(pkBytes []byte) (string, error) {
			return string(pkBytes), nil
		},
		LenCalled: func() int {
			return 3
		},
	}
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	txArgs := getDefaultTransactionArgs()
	txArgs.Version = 0
	txArgs.ChainID = ""
	txArgs.SignatureHex = "617eff4f"

	_, _, err := n.CreateTransaction(txArgs)
	assert.Equal(t, node.ErrInvalidTransactionVersion, err)
}

func TestCreateTransaction_SenderShardIdIsInDifferentShardShouldNotValidate(t *testing.T) {
	t.Parallel()

	expectedHash := []byte("expected hash")
	crtShardID := uint32(1)
	chainID := []byte("chain ID")
	version := uint32(1)

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = mock.HasherMock{
		ComputeCalled: func(s string) []byte {
			return expectedHash
		},
	}
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
		EncodeCalled: func(pkBytes []byte) (string, error) {
			return string(pkBytes), nil
		},
		LenCalled: func() int {
			return 3
		},
	}
	coreComponents.ChainIdCalled = func() string {
		return string(chainID)
	}
	coreComponents.MinTransactionVersionCalled = func() uint32 {
		return version
	}
	coreComponents.EconomicsHandler = &economicsmocks.EconomicsHandlerMock{
		CheckValidityTxValuesCalled: func(tx data.TransactionWithFeeHandler) error {
			return nil
		},
	}

	stateComponents := getDefaultStateComponents()

	shardCoordinator := &mock.ShardCoordinatorMock{
		ComputeIdCalled: func(i []byte) uint32 {
			return crtShardID + 1
		},
		SelfShardId: crtShardID,
	}
	bootstrapComponents := getDefaultBootstrapComponents()
	bootstrapComponents.ShCoordinator = shardCoordinator

	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = shardCoordinator

	cryptoComponents := getDefaultCryptoComponents()

	n, _ := node.NewNode(
		node.WithBootstrapComponents(bootstrapComponents),
		node.WithCoreComponents(coreComponents),
		node.WithCryptoComponents(cryptoComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
		node.WithAddressSignatureSize(10),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := "rcv"

	txArgs := getDefaultTransactionArgs()
	txArgs.Version = version
	txArgs.Nonce = nonce
	txArgs.Value = value.String()
	txArgs.Receiver = receiver

	tx, txHash, err := n.CreateTransaction(txArgs)

	assert.NotNil(t, tx)
	assert.Equal(t, expectedHash, txHash)
	assert.Nil(t, err)
	assert.Equal(t, nonce, tx.Nonce)
	assert.Equal(t, value, tx.Value)
	assert.True(t, bytes.Equal([]byte(receiver), tx.RcvAddr))

	err = n.ValidateTransaction(tx)
	assert.True(t, errors.Is(err, node.ErrDifferentSenderShardId))
}

func TestCreateTransaction_SignatureLengthChecks(t *testing.T) {
	t.Parallel()

	maxValueLength := 7
	signatureLength := 10
	chainID := "chain id"
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.EconomicsHandler = &economicsmocks.EconomicsHandlerStub{
		GenesisTotalSupplyCalled: func() *big.Int {
			str := strings.Repeat("1", maxValueLength)
			bi := big.NewInt(0)
			bi.SetString(str, 10)
			return bi
		},
	}
	coreComponents.ChainIdCalled = func() string {
		return chainID
	}
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
		EncodeCalled: func(pkBytes []byte) (string, error) {
			return string(pkBytes), nil
		},
		LenCalled: func() int {
			return 3
		},
	}

	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithAddressSignatureSize(signatureLength),
	)

	value := "1" + strings.Repeat("0", maxValueLength)
	txArgs := getDefaultTransactionArgs()
	txArgs.Value = value

	for i := 0; i <= signatureLength; i++ {
		signatureBytes := []byte(strings.Repeat("a", i))
		signatureHex := hex.EncodeToString(signatureBytes)

		txArgs.SignatureHex = signatureHex

		tx, _, err := n.CreateTransaction(txArgs)
		assert.NotNil(t, tx)
		assert.NoError(t, err)
		assert.Equal(t, signatureBytes, tx.Signature)
	}

	signature := hex.EncodeToString([]byte(strings.Repeat("a", signatureLength+1)))
	txArgs.SignatureHex = signature

	tx, txHash, err := n.CreateTransaction(txArgs)
	assert.Nil(t, tx)
	assert.Empty(t, txHash)
	assert.Equal(t, node.ErrInvalidSignatureLength, err)
}

func TestCreateTransaction_SenderLengthChecks(t *testing.T) {
	t.Parallel()

	maxLength := 7
	chainID := "chain id"
	encodedAddressLen := 5
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.ChainIdCalled = func() string {
		return chainID
	}
	coreComponents.EconomicsHandler = &economicsmocks.EconomicsHandlerStub{
		GenesisTotalSupplyCalled: func() *big.Int {
			str := strings.Repeat("1", maxLength)
			bi := big.NewInt(0)
			bi.SetString(str, 10)
			return bi
		},
	}
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
		EncodeCalled: func(pkBytes []byte) (string, error) {
			return string(pkBytes), nil
		},
		LenCalled: func() int {
			return encodedAddressLen
		},
	}

	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithAddressSignatureSize(10),
	)

	txArgs := getDefaultTransactionArgs()
	txArgs.ChainID = chainID

	for i := 0; i <= encodedAddressLen; i++ {
		txArgs.Sender = strings.Repeat("s", i)

		_, _, err := n.CreateTransaction(txArgs)
		assert.NoError(t, err)
	}

	txArgs.Sender = strings.Repeat("s", encodedAddressLen) + "additional"

	tx, txHash, err := n.CreateTransaction(txArgs)
	assert.Nil(t, tx)
	assert.Empty(t, txHash)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, node.ErrInvalidAddressLength))
}

func TestCreateTransaction_ReceiverLengthChecks(t *testing.T) {
	t.Parallel()

	maxLength := 7
	chainID := "chain id"
	encodedAddressLen := 5
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.ChainIdCalled = func() string {
		return chainID
	}
	coreComponents.EconomicsHandler = &economicsmocks.EconomicsHandlerStub{
		GenesisTotalSupplyCalled: func() *big.Int {
			str := strings.Repeat("1", maxLength)
			bi := big.NewInt(0)
			bi.SetString(str, 10)
			return bi
		},
	}
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
		EncodeCalled: func(pkBytes []byte) (string, error) {
			return string(pkBytes), nil
		},
		LenCalled: func() int {
			return encodedAddressLen
		},
	}

	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithAddressSignatureSize(10),
	)

	txArgs := getDefaultTransactionArgs()
	txArgs.ChainID = chainID

	for i := 0; i <= encodedAddressLen; i++ {
		txArgs.Receiver = strings.Repeat("r", i)

		_, _, err := n.CreateTransaction(txArgs)
		assert.NoError(t, err)
	}

	txArgs.Receiver = strings.Repeat("r", encodedAddressLen) + "additional"

	tx, txHash, err := n.CreateTransaction(txArgs)
	assert.Nil(t, tx)
	assert.Empty(t, txHash)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, node.ErrInvalidAddressLength))
}

func TestCreateTransaction_TooBigSenderUsernameShouldErr(t *testing.T) {
	t.Parallel()

	maxLength := 7
	chainID := "chain id"
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.ChainIdCalled = func() string {
		return chainID
	}
	coreComponents.EconomicsHandler = &economicsmocks.EconomicsHandlerStub{
		GenesisTotalSupplyCalled: func() *big.Int {
			str := strings.Repeat("1", maxLength)
			bi := big.NewInt(0)
			bi.SetString(str, 10)
			return bi
		},
	}
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
		EncodeCalled: func(pkBytes []byte) (string, error) {
			return string(pkBytes), nil
		},
		LenCalled: func() int {
			return 3
		},
	}

	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithAddressSignatureSize(10),
	)

	txArgs := getDefaultTransactionArgs()
	txArgs.Value = "1" + strings.Repeat("0", maxLength+1)
	txArgs.ChainID = chainID
	txArgs.SenderUsername = bytes.Repeat([]byte{0}, core.MaxUserNameLength+1)

	tx, txHash, err := n.CreateTransaction(txArgs)
	assert.Nil(t, tx)
	assert.Empty(t, txHash)
	assert.Error(t, err)
	assert.Equal(t, node.ErrInvalidSenderUsernameLength, err)
}

func TestCreateTransaction_TooBigReceiverUsernameShouldErr(t *testing.T) {
	t.Parallel()

	maxLength := 7
	chainID := "chain id"
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.ChainIdCalled = func() string {
		return chainID
	}
	coreComponents.EconomicsHandler = &economicsmocks.EconomicsHandlerStub{
		GenesisTotalSupplyCalled: func() *big.Int {
			str := strings.Repeat("1", maxLength)
			bi := big.NewInt(0)
			bi.SetString(str, 10)
			return bi
		},
	}
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
		EncodeCalled: func(pkBytes []byte) (string, error) {
			return string(pkBytes), nil
		},
		LenCalled: func() int {
			return 3
		},
	}

	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithAddressSignatureSize(10),
	)

	txArgs := getDefaultTransactionArgs()
	txArgs.ChainID = chainID
	txArgs.ReceiverUsername = bytes.Repeat([]byte{0}, core.MaxUserNameLength+1)
	txArgs.Value = "1" + strings.Repeat("0", maxLength+1)

	tx, txHash, err := n.CreateTransaction(txArgs)
	assert.Nil(t, tx)
	assert.Empty(t, txHash)
	assert.Error(t, err)
	assert.Equal(t, node.ErrInvalidReceiverUsernameLength, err)
}

func TestCreateTransaction_DataFieldSizeExceedsMaxShouldErr(t *testing.T) {
	t.Parallel()

	maxLength := 7
	chainID := "chain id"
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.ChainIdCalled = func() string {
		return chainID
	}
	coreComponents.EconomicsHandler = &economicsmocks.EconomicsHandlerStub{
		GenesisTotalSupplyCalled: func() *big.Int {
			str := strings.Repeat("1", maxLength)
			bi := big.NewInt(0)
			bi.SetString(str, 10)
			return bi
		},
	}
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
		EncodeCalled: func(pkBytes []byte) (string, error) {
			return string(pkBytes), nil
		},
		LenCalled: func() int {
			return 3
		},
	}

	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithAddressSignatureSize(10),
	)

	txArgs := getDefaultTransactionArgs()
	txArgs.ChainID = chainID
	txArgs.DataField = bytes.Repeat([]byte{0}, core.MegabyteSize+1)
	txArgs.Value = "1" + strings.Repeat("0", maxLength+1)

	tx, txHash, err := n.CreateTransaction(txArgs)
	assert.Nil(t, tx)
	assert.Empty(t, txHash)
	assert.Error(t, err)
	assert.Equal(t, node.ErrDataFieldTooBig, err)
}

func TestCreateTransaction_TooLargeValueFieldShouldErr(t *testing.T) {
	t.Parallel()

	maxLength := 7
	chainID := "chain id"
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.ChainIdCalled = func() string {
		return chainID
	}
	coreComponents.EconomicsHandler = &economicsmocks.EconomicsHandlerStub{
		GenesisTotalSupplyCalled: func() *big.Int {
			str := strings.Repeat("1", maxLength)
			bi := big.NewInt(0)
			bi.SetString(str, 10)
			return bi
		},
	}
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
		EncodeCalled: func(pkBytes []byte) (string, error) {
			return string(pkBytes), nil
		},
		LenCalled: func() int {
			return 3
		},
	}

	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithAddressSignatureSize(10),
	)

	txArgs := getDefaultTransactionArgs()
	txArgs.ChainID = chainID
	txArgs.Value = "1" + strings.Repeat("0", maxLength+1)

	tx, txHash, err := n.CreateTransaction(txArgs)
	assert.Nil(t, tx)
	assert.Empty(t, txHash)
	assert.Error(t, err)
	assert.Equal(t, node.ErrTransactionValueLengthTooBig, err)
}

func TestCreateTransaction_InvalidGuardianSigShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
	}
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithAddressSignatureSize(16),
	)

	txArgs := getDefaultTransactionArgs()
	txArgs.SignatureHex = hex.EncodeToString(bytes.Repeat([]byte{0}, 1))
	txArgs.GuardianSigHex = hex.EncodeToString(bytes.Repeat([]byte{0}, 32))

	tx, txHash, err := n.CreateTransaction(txArgs)

	assert.Nil(t, tx)
	assert.Nil(t, txHash)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, node.ErrInvalidSignatureLength))
}

func TestCreateTransaction_InvalidGuardianAddressLenShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()

	encodedAddressLen := 8
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
		LenCalled: func() int {
			return encodedAddressLen
		},
	}
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithAddressSignatureSize(16),
	)

	txArgs := getDefaultTransactionArgs()
	txArgs.SignatureHex = hex.EncodeToString(bytes.Repeat([]byte{0}, 8))
	txArgs.GuardianSigHex = hex.EncodeToString(bytes.Repeat([]byte{0}, 8))
	txArgs.Guardian = strings.Repeat("g", encodedAddressLen) + "additional"

	tx, txHash, err := n.CreateTransaction(txArgs)

	assert.Nil(t, tx)
	assert.Nil(t, txHash)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, node.ErrInvalidAddressLength))
}

func TestCreateTransaction_AddressPubKeyConverterDecode(t *testing.T) {
	t.Parallel()

	minAddrLen := 4
	encodedAddressLen := 8
	addrPubKeyConverter := &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			if len(hexAddress) < minAddrLen {
				return nil, errors.New("decode error")
			}
			return []byte(hexAddress), nil
		},
		LenCalled: func() int {
			return encodedAddressLen
		},
	}

	guardianSig := hex.EncodeToString(bytes.Repeat([]byte{0}, 8))
	guardian := strings.Repeat("g", encodedAddressLen)

	t.Run("fail to decode receiver", func(t *testing.T) {
		t.Parallel()

		coreComponents := getDefaultCoreComponents()
		coreComponents.IntMarsh = getMarshalizer()
		coreComponents.VmMarsh = getMarshalizer()
		coreComponents.TxMarsh = getMarshalizer()
		coreComponents.Hash = getHasher()

		coreComponents.AddrPubKeyConv = addrPubKeyConverter
		stateComponents := getDefaultStateComponents()
		stateComponents.AccountsAPI = &stateMock.AccountsStub{}

		n, _ := node.NewNode(
			node.WithCoreComponents(coreComponents),
			node.WithStateComponents(stateComponents),
			node.WithAddressSignatureSize(16),
		)

		txArgs := getDefaultTransactionArgs()
		txArgs.Guardian = guardian
		txArgs.GuardianSigHex = guardianSig

		tx, txHash, err := n.CreateTransaction(txArgs)

		assert.Nil(t, tx)
		assert.Nil(t, txHash)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "receiver address"))
	})

	t.Run("fail to decode sender", func(t *testing.T) {
		t.Parallel()

		coreComponents := getDefaultCoreComponents()
		coreComponents.IntMarsh = getMarshalizer()
		coreComponents.VmMarsh = getMarshalizer()
		coreComponents.TxMarsh = getMarshalizer()
		coreComponents.Hash = getHasher()

		coreComponents.AddrPubKeyConv = addrPubKeyConverter
		stateComponents := getDefaultStateComponents()
		stateComponents.AccountsAPI = &stateMock.AccountsStub{}

		n, _ := node.NewNode(
			node.WithCoreComponents(coreComponents),
			node.WithStateComponents(stateComponents),
			node.WithAddressSignatureSize(16),
		)

		txArgs := getDefaultTransactionArgs()
		txArgs.Guardian = guardian
		txArgs.GuardianSigHex = guardianSig
		txArgs.Receiver = strings.Repeat("r", minAddrLen+1)

		tx, txHash, err := n.CreateTransaction(txArgs)

		assert.Nil(t, tx)
		assert.Nil(t, txHash)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "sender address"))
	})
}

func TestCreateTransaction_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	version := uint32(1)
	expectedHash := []byte("expected hash")
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = mock.HasherMock{
		ComputeCalled: func(s string) []byte {
			return expectedHash
		},
	}
	coreComponents.TxVersionCheckHandler = versioning.NewTxVersionChecker(version)
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
		EncodeCalled: func(pkBytes []byte) (string, error) {
			return string(pkBytes), nil
		},
		LenCalled: func() int {
			return 3
		},
	}
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
			return createAcc([]byte("address")), nil
		},
	}

	processComponents := getDefaultProcessComponents()
	processComponents.EpochTrigger = &mock.EpochStartTriggerStub{
		EpochCalled: func() uint32 {
			return 1
		},
	}

	networkComponents := getDefaultNetworkComponents()
	cryptoComponents := getDefaultCryptoComponents()
	bootstrapComponents := getDefaultBootstrapComponents()
	bootstrapComponents.ShCoordinator = processComponents.ShardCoordinator()
	bootstrapComponents.HdrIntegrityVerifier = processComponents.HeaderIntegrVerif
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
		node.WithNetworkComponents(networkComponents),
		node.WithCryptoComponents(cryptoComponents),
		node.WithBootstrapComponents(bootstrapComponents),
		node.WithAddressSignatureSize(10),
	)

	nonce := uint64(0)
	value := new(big.Int).SetInt64(10)
	receiver := "rcv"

	txArgs := getDefaultTransactionArgs()
	txArgs.Receiver = receiver
	txArgs.Nonce = nonce
	txArgs.Value = value.String()
	txArgs.ChainID = coreComponents.ChainID()
	txArgs.Version = coreComponents.MinTransactionVersion()

	tx, txHash, err := n.CreateTransaction(txArgs)
	assert.NotNil(t, tx)
	assert.Equal(t, expectedHash, txHash)
	assert.Nil(t, err)
	assert.Equal(t, nonce, tx.Nonce)
	assert.Equal(t, value, tx.Value)
	assert.True(t, bytes.Equal([]byte(receiver), tx.RcvAddr))

	err = n.ValidateTransaction(tx)
	assert.Nil(t, err)
}

func TestCreateTransaction_TxSignedWithHashShouldErrVersionShoudBe2(t *testing.T) {
	t.Parallel()

	expectedHash := []byte("expected hash")
	crtShardID := uint32(1)
	chainID := "chain ID"
	version := uint32(1)

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = mock.HasherMock{
		ComputeCalled: func(s string) []byte {
			return expectedHash
		},
	}
	coreComponents.MinTransactionVersionCalled = func() uint32 {
		return version
	}
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
		EncodeCalled: func(pkBytes []byte) (string, error) {
			return string(pkBytes), nil
		},
		LenCalled: func() int {
			return 3
		},
	}
	coreComponents.TxSignHasherField = &mock.HasherMock{}
	coreComponents.ChainIdCalled = func() string {
		return chainID
	}

	feeHandler := &economicsmocks.EconomicsHandlerStub{
		CheckValidityTxValuesCalled: func(tx data.TransactionWithFeeHandler) error {
			return nil
		},
	}
	coreComponents.EconomicsHandler = feeHandler
	coreComponents.TxVersionCheckHandler = versioning.NewTxVersionChecker(version)

	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}

	bootstrapComponents := getDefaultBootstrapComponents()
	bootstrapComponents.ShCoordinator = &mock.ShardCoordinatorMock{
		ComputeIdCalled: func(i []byte) uint32 {
			return crtShardID
		},
		SelfShardId: crtShardID,
	}

	processComponents := getDefaultProcessComponents()
	processComponents.EpochTrigger = &mock.EpochStartTriggerStub{
		EpochCalled: func() uint32 {
			return 1
		},
	}
	processComponents.WhiteListerVerifiedTxsInternal = &testscommon.WhiteListHandlerStub{}
	processComponents.WhiteListHandlerInternal = &testscommon.WhiteListHandlerStub{}

	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = &mock.SingleSignerMock{}
	cryptoComponents.TxKeyGen = &mock.KeyGenMock{}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithBootstrapComponents(bootstrapComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
		node.WithCryptoComponents(cryptoComponents),
		node.WithEnableSignTxWithHashEpoch(2),
		node.WithAddressSignatureSize(10),
	)

	options := transaction.MaskSignedWithHash

	txArgs := getDefaultTransactionArgs()
	txArgs.ChainID = chainID
	txArgs.Version = version
	txArgs.Options = options

	tx, _, err := n.CreateTransaction(txArgs)
	require.Nil(t, err)
	err = n.ValidateTransaction(tx)
	assert.Equal(t, process.ErrInvalidTransactionVersion, err)
}

func TestCreateTransaction_TxSignedWithHashNoEnabledShouldErr(t *testing.T) {
	t.Parallel()

	expectedHash := []byte("expected hash")
	crtShardID := uint32(1)
	chainID := "chain ID"
	version := uint32(1)
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = mock.HasherMock{
		ComputeCalled: func(s string) []byte {
			return expectedHash
		},
	}
	coreComponents.TxSignHasherField = mock.HasherMock{}
	coreComponents.ChainIdCalled = func() string {
		return chainID
	}
	coreComponents.MinTransactionVersionCalled = func() uint32 {
		return version
	}
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return []byte(hexAddress), nil
		},
		EncodeCalled: func(pkBytes []byte) (string, error) {
			return string(pkBytes), nil
		},
		LenCalled: func() int {
			return 3
		},
	}

	feeHandler := &economicsmocks.EconomicsHandlerStub{
		CheckValidityTxValuesCalled: func(tx data.TransactionWithFeeHandler) error {
			return nil
		},
	}
	coreComponents.EconomicsHandler = feeHandler
	coreComponents.TxVersionCheckHandler = versioning.NewTxVersionChecker(version)

	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}

	bootstrapComponents := getDefaultBootstrapComponents()
	bootstrapComponents.ShCoordinator = &mock.ShardCoordinatorMock{
		ComputeIdCalled: func(i []byte) uint32 {
			return crtShardID
		},
		SelfShardId: crtShardID,
	}

	processComponents := getDefaultProcessComponents()
	processComponents.EpochTrigger = &mock.EpochStartTriggerStub{
		EpochCalled: func() uint32 {
			return 1
		},
	}
	processComponents.WhiteListerVerifiedTxsInternal = &testscommon.WhiteListHandlerStub{
		IsWhiteListedCalled: func(interceptedData process.InterceptedData) bool {
			return false
		},
	}
	processComponents.WhiteListHandlerInternal = &testscommon.WhiteListHandlerStub{}

	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxSig = &mock.SingleSignerMock{}
	cryptoComponents.TxKeyGen = &mock.KeyGenMock{
		PublicKeyFromByteArrayMock: func(b []byte) (crypto.PublicKey, error) {
			return nil, nil
		},
	}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithBootstrapComponents(bootstrapComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
		node.WithCryptoComponents(cryptoComponents),
		node.WithEnableSignTxWithHashEpoch(2),
		node.WithAddressSignatureSize(10),
	)

	options := transaction.MaskSignedWithHash

	txArgs := getDefaultTransactionArgs()
	txArgs.ChainID = chainID
	txArgs.Version = version + 1
	txArgs.Options = options

	tx, _, _ := n.CreateTransaction(txArgs)

	err := n.ValidateTransaction(tx)
	assert.Equal(t, process.ErrTransactionSignedWithHashIsNotEnabled, err)
}

func TestCreateShardedStores_NilShardCoordinatorShouldError(t *testing.T) {
	messenger := getMessenger()
	dataPool := dataRetrieverMock.NewPoolsHolderStub()
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}
	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = messenger
	dataComponents := getDefaultDataComponents()
	dataComponents.DataPool = dataPool
	processComponents := getDefaultProcessComponents()

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
		node.WithNetworkComponents(networkComponents),
		node.WithDataComponents(dataComponents),
		node.WithProcessComponents(processComponents),
	)

	processComponents.ShardCoord = nil
	err := n.CreateShardedStores()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil shard coordinator")
}

func TestCreateShardedStores_NilDataPoolShouldError(t *testing.T) {
	messenger := getMessenger()
	shardCoordinator := mock.NewOneShardCoordinatorMock()

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = shardCoordinator
	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = messenger
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}
	dataComponents := getDefaultDataComponents()

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithProcessComponents(processComponents),
		node.WithNetworkComponents(networkComponents),
		node.WithStateComponents(stateComponents),
		node.WithDataComponents(dataComponents),
	)

	dataComponents.DataPool = nil
	err := n.CreateShardedStores()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil data pool")
}

func TestCreateShardedStores_NilTransactionDataPoolShouldError(t *testing.T) {
	messenger := getMessenger()
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	dataPool := dataRetrieverMock.NewPoolsHolderStub()
	dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return nil
	}
	dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
		return &testscommon.HeadersCacherStub{}
	}
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	dataComponents := getDefaultDataComponents()
	dataComponents.DataPool = dataPool
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = shardCoordinator
	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = messenger

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
		node.WithNetworkComponents(networkComponents),
	)

	err := n.CreateShardedStores()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil transaction sharded data store")
}

func TestCreateShardedStores_NilHeaderDataPoolShouldError(t *testing.T) {
	messenger := getMessenger()
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	dataPool := dataRetrieverMock.NewPoolsHolderStub()
	dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}

	dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
		return nil
	}
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	dataComponents := getDefaultDataComponents()
	dataComponents.DataPool = dataPool
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = shardCoordinator
	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = messenger

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
		node.WithNetworkComponents(networkComponents),
	)

	err := n.CreateShardedStores()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "nil header sharded data store")
}

func TestCreateShardedStores_ReturnsSuccessfully(t *testing.T) {
	messenger := getMessenger()
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	numOfShards := uint32(2)
	shardCoordinator.SetNoShards(numOfShards)

	dataPool := dataRetrieverMock.NewPoolsHolderStub()
	dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}
	dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
		return &testscommon.HeadersCacherStub{}
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.TxMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	dataComponents := getDefaultDataComponents()
	dataComponents.DataPool = dataPool
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = shardCoordinator
	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = messenger

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
		node.WithNetworkComponents(networkComponents),
	)

	err := n.CreateShardedStores()
	assert.Nil(t, err)
}

func TestNode_ValidatorStatisticsApi(t *testing.T) {
	t.Parallel()

	initialPubKeys := make(map[uint32][]string)
	keys := [][]string{{"key0"}, {"key1"}, {"key2"}}
	initialPubKeys[0] = keys[0]
	initialPubKeys[1] = keys[1]
	initialPubKeys[2] = keys[2]

	validatorsInfo := make(map[uint32][]*state.ValidatorInfo)

	for shardId, pubkeysPerShard := range initialPubKeys {
		validatorsInfo[shardId] = make([]*state.ValidatorInfo, 0)
		for _, pubKey := range pubkeysPerShard {
			validatorsInfo[shardId] = append(validatorsInfo[shardId], &state.ValidatorInfo{
				PublicKey:                  []byte(pubKey),
				ShardId:                    shardId,
				List:                       "",
				Index:                      0,
				TempRating:                 0,
				Rating:                     0,
				RewardAddress:              nil,
				LeaderSuccess:              0,
				LeaderFailure:              0,
				ValidatorSuccess:           0,
				ValidatorFailure:           0,
				NumSelectedInSuccessBlocks: 0,
				AccumulatedFees:            nil,
				TotalLeaderSuccess:         0,
				TotalLeaderFailure:         0,
				TotalValidatorSuccess:      0,
				TotalValidatorFailure:      0,
			})
		}
	}

	vsp := &mock.ValidatorStatisticsProcessorStub{
		RootHashCalled: func() (i []byte, err error) {
			return []byte("hash"), nil
		},
		GetValidatorInfoForRootHashCalled: func(rootHash []byte) (m map[uint32][]*state.ValidatorInfo, err error) {
			return validatorsInfo, nil
		},
	}

	validatorProvider := &mock.ValidatorsProviderStub{GetLatestValidatorsCalled: func() map[string]*accounts.ValidatorApiResponse {
		apiResponses := make(map[string]*accounts.ValidatorApiResponse)

		for _, vis := range validatorsInfo {
			for _, vi := range vis {
				apiResponses[hex.EncodeToString(vi.GetPublicKey())] = &accounts.ValidatorApiResponse{}
			}
		}

		return apiResponses
	},
	}

	processComponents := getDefaultProcessComponents()
	processComponents.ValidatorProvider = validatorProvider
	processComponents.ValidatorStatistics = vsp

	n, _ := node.NewNode(
		node.WithInitialNodesPubKeys(initialPubKeys),
		node.WithProcessComponents(processComponents),
	)

	expectedData := &accounts.ValidatorApiResponse{}
	validatorsData, err := n.ValidatorStatisticsApi()
	require.Equal(t, expectedData, validatorsData[hex.EncodeToString([]byte(keys[2][0]))])
	require.Nil(t, err)
}

// ------- GetAccount

func TestNode_GetAccountPubkeyConverterFailsShouldErr(t *testing.T) {
	t.Parallel()

	accDB := &stateMock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			return nil, state.ErrAccNotFound
		},
	}

	errExpected := errors.New("expected error")
	coreComponents := getDefaultCoreComponents()
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(hexAddress string) ([]byte, error) {
			return nil, errExpected
		},
	}
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = accDB

	n, _ := node.NewNode(
		node.WithStateComponents(stateComponents),
		node.WithCoreComponents(coreComponents),
	)

	recovAccnt, _, err := n.GetAccount(createDummyHexAddress(64), api.AccountQueryOptions{})

	assert.Empty(t, recovAccnt)
	assert.ErrorIs(t, err, errExpected)
}

func TestNode_GetAccountAccountDoesNotExistsShouldRetEmpty(t *testing.T) {
	t.Parallel()

	accountsRepostitory := &stateMock.AccountsRepositoryStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
			blockInfo := holders.NewBlockInfo([]byte{0xaa}, 7, []byte{0xbb})
			return nil, nil, state.NewErrAccountNotFoundAtBlock(blockInfo)
		},
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

	account, blockInfo, err := n.GetAccount(testscommon.TestAddressAlice, api.AccountQueryOptions{})

	require.Nil(t, err)
	require.Equal(t, uint64(0), account.Nonce)
	require.Equal(t, "0", account.Balance)
	require.Equal(t, "0", account.DeveloperReward)
	require.Nil(t, account.CodeHash)
	require.Nil(t, account.RootHash)
	require.Equal(t, uint64(7), blockInfo.Nonce)
	require.Equal(t, "aa", blockInfo.Hash)
	require.Equal(t, "bb", blockInfo.RootHash)
}

func TestNode_GetAccountAccountsRepositoryFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")

	accountsRepostitory := &stateMock.AccountsRepositoryStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
			return nil, nil, errExpected
		},
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

	recovAccnt, _, err := n.GetAccount(testscommon.TestAddressAlice, api.AccountQueryOptions{})

	assert.Empty(t, recovAccnt)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, errExpected)
}

func TestNode_GetAccountAccNotFoundShouldReturnEmpty(t *testing.T) {
	t.Parallel()

	accDB := &stateMock.AccountsStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
			return nil, nil, state.NewErrAccountNotFoundAtBlock(dummyBlockInfo.forProcessing())
		},
		RecreateTrieCalled: func(_ []byte) error {
			return nil
		},
	}

	dataComponents := getDefaultDataComponents()
	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()

	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)

	n, _ := node.NewNode(
		node.WithDataComponents(dataComponents),
		node.WithCoreComponents(coreComponents),
		node.WithStateComponents(stateComponents),
	)

	acc, bInfo, err := n.GetAccount(testscommon.TestAddressAlice, api.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, dummyBlockInfo.apiResult(), bInfo)
	require.Equal(t, api.AccountResponse{Address: testscommon.TestAddressAlice, Balance: "0", DeveloperReward: "0"}, acc)
}

func TestNode_GetAccountAccountExistsShouldReturn(t *testing.T) {
	t.Parallel()

	accnt := createAcc(testscommon.TestPubKeyBob)
	_ = accnt.AddToBalance(big.NewInt(1))
	accnt.IncreaseNonce(2)
	accnt.SetRootHash([]byte("root hash"))
	accnt.SetCodeHash([]byte("code hash"))
	accnt.AddToDeveloperReward(big.NewInt(37))
	accnt.SetCodeMetadata([]byte("metadata"))
	accnt.SetOwnerAddress(testscommon.TestPubKeyAlice)

	accDB := &stateMock.AccountsStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
			return accnt, nil, nil
		},
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}

	coreComponents := getDefaultCoreComponents()
	dataComponents := getDefaultDataComponents()
	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
	)

	recovAccnt, _, err := n.GetAccount(testscommon.TestAddressBob, api.AccountQueryOptions{})

	require.Nil(t, err)
	require.Equal(t, uint64(2), recovAccnt.Nonce)
	require.Equal(t, "1", recovAccnt.Balance)
	require.Equal(t, []byte("root hash"), recovAccnt.RootHash)
	require.Equal(t, []byte("code hash"), recovAccnt.CodeHash)
	require.Equal(t, []byte("metadata"), recovAccnt.CodeMetadata)
	require.Equal(t, testscommon.TestAddressAlice, recovAccnt.OwnerAddress)
}

func TestNode_AppStatusHandlersShouldIncrement(t *testing.T) {
	t.Parallel()

	metricKey := common.MetricCurrentRound
	incrementCalled := make(chan bool, 1)

	appStatusHandlerStub := &statusHandlerMock.AppStatusHandlerStub{
		IncrementHandler: func(key string) {
			incrementCalled <- true
		},
	}

	statusCoreComp := &factoryTests.StatusCoreComponentsStub{
		AppStatusHandlerField: appStatusHandlerStub,
	}

	_, _ = node.NewNode(
		node.WithStatusCoreComponents(statusCoreComp))

	appStatusHandlerStub.Increment(metricKey)

	select {
	case <-incrementCalled:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timeout - function not called")
	}
}

func TestNode_AppStatusHandlerShouldDecrement(t *testing.T) {
	t.Parallel()

	metricKey := common.MetricCurrentRound
	decrementCalled := make(chan bool, 1)

	appStatusHandlerStub := &statusHandlerMock.AppStatusHandlerStub{
		DecrementHandler: func(key string) {
			decrementCalled <- true
		},
	}

	statusCoreComp := &factoryTests.StatusCoreComponentsStub{
		AppStatusHandlerField: appStatusHandlerStub,
	}

	_, _ = node.NewNode(
		node.WithStatusCoreComponents(statusCoreComp))

	appStatusHandlerStub.Decrement(metricKey)

	select {
	case <-decrementCalled:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timeout - function not called")
	}
}

func TestNode_AppStatusHandlerShouldSetInt64Value(t *testing.T) {
	t.Parallel()

	metricKey := common.MetricCurrentRound
	setInt64ValueCalled := make(chan bool, 1)

	appStatusHandlerStub := &statusHandlerMock.AppStatusHandlerStub{
		SetInt64ValueHandler: func(key string, value int64) {
			setInt64ValueCalled <- true
		},
	}

	statusCoreComp := &factoryTests.StatusCoreComponentsStub{
		AppStatusHandlerField: appStatusHandlerStub,
	}

	_, _ = node.NewNode(
		node.WithStatusCoreComponents(statusCoreComp))

	appStatusHandlerStub.SetInt64Value(metricKey, int64(1))

	select {
	case <-setInt64ValueCalled:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timeout - function not called")
	}
}

func TestNode_AppStatusHandlerShouldSetUInt64Value(t *testing.T) {
	t.Parallel()

	metricKey := common.MetricCurrentRound
	setUInt64ValueCalled := make(chan bool, 1)

	appStatusHandlerStub := &statusHandlerMock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
			setUInt64ValueCalled <- true
		},
	}

	statusCoreComp := &factoryTests.StatusCoreComponentsStub{
		AppStatusHandlerField: appStatusHandlerStub,
	}

	_, _ = node.NewNode(
		node.WithStatusCoreComponents(statusCoreComp))

	appStatusHandlerStub.SetUInt64Value(metricKey, uint64(1))

	select {
	case <-setUInt64ValueCalled:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timeout - function not called")
	}
}

func TestNode_EncodeDecodeAddressPubkey(t *testing.T) {
	t.Parallel()

	buff := []byte("abcdefg")

	coreComponents := getDefaultCoreComponents()
	coreComponents.AddrPubKeyConv = testscommon.NewPubkeyConverterMock(32)
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
	)
	encoded, err := n.EncodeAddressPubkey(buff)
	assert.Nil(t, err)

	recoveredBytes, err := n.DecodeAddressPubkey(encoded)

	assert.Nil(t, err)
	assert.Equal(t, buff, recoveredBytes)
}

func TestNode_EncodeDecodeAddressPubkeyWithNilConverterShouldErr(t *testing.T) {
	t.Parallel()

	buff := []byte("abcdefg")

	coreComponents := getDefaultCoreComponents()
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents))

	coreComponents.AddrPubKeyConv = nil
	encoded, err := n.EncodeAddressPubkey(buff)

	assert.Empty(t, encoded)
	assert.True(t, errors.Is(err, node.ErrNilPubkeyConverter))
}

func TestNode_DecodeAddressPubkeyWithNilConverterShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
	)

	coreComponents.AddrPubKeyConv = nil
	recoveredBytes, err := n.DecodeAddressPubkey("")

	assert.True(t, errors.Is(err, node.ErrNilPubkeyConverter))
	assert.Nil(t, recoveredBytes)
}

func TestNode_DirectTrigger(t *testing.T) {
	t.Parallel()

	wasCalled := false
	epoch := uint32(47839)
	recoveredEpoch := uint32(0)
	recoveredWithEarlyEndOfEpoch := atomicCore.Flag{}
	hardforkTrigger := &testscommon.HardforkTriggerStub{
		TriggerCalled: func(epoch uint32, withEarlyEndOfEpoch bool) error {
			wasCalled = true
			atomic.StoreUint32(&recoveredEpoch, epoch)
			recoveredWithEarlyEndOfEpoch.SetValue(withEarlyEndOfEpoch)

			return nil
		},
	}

	processComponents := &integrationTestsMock.ProcessComponentsStub{
		HardforkTriggerField: hardforkTrigger,
	}

	n, _ := node.NewNode(
		node.WithProcessComponents(processComponents),
	)

	err := n.DirectTrigger(epoch, true)

	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, epoch, recoveredEpoch)
	assert.True(t, recoveredWithEarlyEndOfEpoch.IsSet())
}

func TestNode_IsSelfTrigger(t *testing.T) {
	t.Parallel()

	wasCalled := false
	hardforkTrigger := &testscommon.HardforkTriggerStub{
		IsSelfTriggerCalled: func() bool {
			wasCalled = true

			return true
		},
	}

	processComponents := &integrationTestsMock.ProcessComponentsStub{
		HardforkTriggerField: hardforkTrigger,
	}

	n, _ := node.NewNode(
		node.WithProcessComponents(processComponents),
	)

	isSelf := n.IsSelfTrigger()

	assert.True(t, isSelf)
	assert.True(t, wasCalled)
}

// ------- Query handlers

func TestNode_AddQueryHandlerNilHandlerShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()

	err := n.AddQueryHandler("handler", nil)

	assert.True(t, errors.Is(err, node.ErrNilQueryHandler))
}

func TestNode_AddQueryHandlerEmptyNameShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()

	err := n.AddQueryHandler("", &mock.QueryHandlerStub{})

	assert.True(t, errors.Is(err, node.ErrEmptyQueryHandlerName))
}

func TestNode_AddQueryHandlerExistsShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()

	err := n.AddQueryHandler("handler", &mock.QueryHandlerStub{})
	assert.Nil(t, err)

	err = n.AddQueryHandler("handler", &mock.QueryHandlerStub{})

	assert.True(t, errors.Is(err, node.ErrQueryHandlerAlreadyExists))
}

func TestNode_GetQueryHandlerNotExistsShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()

	qh, err := n.GetQueryHandler("handler")

	assert.True(t, check.IfNil(qh))
	assert.True(t, errors.Is(err, node.ErrNilQueryHandler))
}

func TestNode_GetQueryHandlerShouldWork(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()

	qh := &mock.QueryHandlerStub{}
	handler := "handler"
	_ = n.AddQueryHandler(handler, &mock.QueryHandlerStub{})

	qhRecovered, err := n.GetQueryHandler(handler)

	assert.Equal(t, qhRecovered, qh)
	assert.Nil(t, err)
}

func TestNode_GetPeerInfoUnknownPeerShouldErr(t *testing.T) {
	t.Parallel()

	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = &p2pmocks.MessengerStub{
		PeersCalled: func() []core.PeerID {
			return make([]core.PeerID, 0)
		},
	}

	n, _ := node.NewNode(
		node.WithNetworkComponents(networkComponents),
	)

	pid := "pid"
	vals, err := n.GetPeerInfo(pid)

	assert.Nil(t, vals)
	assert.True(t, errors.Is(err, node.ErrUnknownPeerID))
}

func TestNode_ShouldWork(t *testing.T) {
	t.Parallel()

	pid1 := "pid1"
	pid2 := "pid2"

	processComponents := getDefaultProcessComponents()
	processComponents.MainPeerMapper = &p2pmocks.NetworkShardingCollectorStub{
		GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
			return core.P2PPeerInfo{
				PeerType: 0,
				ShardID:  0,
				PkBytes:  pid.Bytes(),
			}
		},
	}

	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = &p2pmocks.MessengerStub{
		PeersCalled: func() []core.PeerID {
			// return them unsorted
			return []core.PeerID{core.PeerID(pid2), core.PeerID(pid1)}
		},
		PeerAddressesCalled: func(pid core.PeerID) []string {
			return []string{"addr" + string(pid)}
		},
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.ValPubKeyConv = testscommon.NewPubkeyConverterMock(32)

	n, _ := node.NewNode(
		node.WithNetworkComponents(networkComponents),
		node.WithProcessComponents(processComponents),
		node.WithCoreComponents(coreComponents),
		node.WithPeerDenialEvaluator(&mock.PeerDenialEvaluatorStub{
			IsDeniedCalled: func(pid core.PeerID) bool {
				return pid == core.PeerID(pid1)
			},
		}),
	)

	vals, err := n.GetPeerInfo("3sf1k") // will return both pids, sorted

	assert.Nil(t, err)
	require.Equal(t, 2, len(vals))

	expected := []core.QueryP2PPeerInfo{
		{
			Pid:           core.PeerID(pid1).Pretty(),
			Addresses:     []string{"addr" + pid1},
			Pk:            hex.EncodeToString([]byte(pid1)),
			IsBlacklisted: true,
			PeerType:      core.UnknownPeer.String(),
			PeerSubType:   core.RegularPeer.String(),
		},
		{
			Pid:           core.PeerID(pid2).Pretty(),
			Addresses:     []string{"addr" + pid2},
			Pk:            hex.EncodeToString([]byte(pid2)),
			IsBlacklisted: false,
			PeerType:      core.UnknownPeer.String(),
			PeerSubType:   core.RegularPeer.String(),
		},
	}

	assert.Equal(t, expected, vals)
}

func TestNode_ValidateTransactionForSimulation_CheckSignatureFalse(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = testscommon.NewPubkeyConverterMock(3)
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{}

	bootstrapComponents := getDefaultBootstrapComponents()
	bootstrapComponents.ShCoordinator = &mock.ShardCoordinatorMock{}

	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = bootstrapComponents.ShCoordinator
	processComponents.WhiteListHandlerInternal = &testscommon.WhiteListHandlerStub{}
	processComponents.WhiteListerVerifiedTxsInternal = &testscommon.WhiteListHandlerStub{}
	processComponents.EpochTrigger = &mock.EpochStartTriggerStub{}

	cryptoComponents := getDefaultCryptoComponents()
	cryptoComponents.TxKeyGen = &mock.KeyGenMock{
		PublicKeyFromByteArrayMock: func(b []byte) (crypto.PublicKey, error) {
			return nil, nil
		},
	}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithProcessComponents(processComponents),
		node.WithBootstrapComponents(bootstrapComponents),
		node.WithStateComponents(stateComponents),
		node.WithCryptoComponents(cryptoComponents),
	)

	tx := &transaction.Transaction{
		Nonce:     11,
		Value:     big.NewInt(25),
		RcvAddr:   []byte("rec"),
		SndAddr:   []byte("snd"),
		GasPrice:  6,
		GasLimit:  12,
		Data:      []byte(""),
		Signature: []byte("sig1"),
		ChainID:   []byte(coreComponents.ChainID()),
	}

	err := n.ValidateTransactionForSimulation(tx, false)
	require.NoError(t, err)
}

func TestGetKeyValuePairs_CannotDecodeAddress(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local err")
	coreComponents := getDefaultCoreComponents()
	coreComponents.AddrPubKeyConv = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(humanReadable string) ([]byte, error) {
			return nil, expectedErr
		},
	}

	dataComponents := getDefaultDataComponents()
	dataComponents.BlockChain = &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}

	n, _ := node.NewNode(
		node.WithStateComponents(getDefaultStateComponents()),
		node.WithDataComponents(dataComponents),
		node.WithCoreComponents(coreComponents),
	)

	res, _, err := n.GetKeyValuePairs("addr", api.AccountQueryOptions{}, context.Background())
	require.Nil(t, res)
	require.True(t, strings.Contains(fmt.Sprintf("%v", err), expectedErr.Error()))
}

func TestNode_Close(t *testing.T) {
	t.Parallel()

	n, err := node.NewNode()
	require.Nil(t, err)

	closerCalledOrder := make([]*mock.CloserStub, 0)
	c1 := &mock.CloserStub{}
	c1.CloseCalled = func() error {
		closerCalledOrder = append(closerCalledOrder, c1)
		return nil
	}

	c2 := &mock.CloserStub{}
	c2.CloseCalled = func() error {
		closerCalledOrder = append(closerCalledOrder, c2)
		return nil
	}

	c3 := &mock.CloserStub{}
	c3.CloseCalled = func() error {
		closerCalledOrder = append(closerCalledOrder, c3)
		return nil
	}

	queryCalled := make(map[string]*mock.QueryHandlerStub)
	q1 := &mock.QueryHandlerStub{}
	q1.CloseCalled = func() error {
		queryCalled["q1"] = q1
		return nil
	}
	q2 := &mock.QueryHandlerStub{}
	q2.CloseCalled = func() error {
		queryCalled["q2"] = q2
		return nil
	}

	n.AddClosableComponents(c1, c2, c3)
	_ = n.AddQueryHandler("q1", q1)
	_ = n.AddQueryHandler("q2", q2)

	err = n.Close()
	assert.Nil(t, err)
	require.Equal(t, 3, len(closerCalledOrder))
	assert.True(t, c3 == closerCalledOrder[0]) // pointer testing
	assert.True(t, c2 == closerCalledOrder[1]) // pointer testing
	assert.True(t, c1 == closerCalledOrder[2]) // pointer testing

	require.Equal(t, 2, len(queryCalled))
	require.True(t, queryCalled["q1"] == q1) // pointer testing
	require.True(t, queryCalled["q2"] == q2) // pointer testing
}

func TestNode_getClosableComponentName(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	n := &node.Node{}
	assert.Equal(t, coreComponents.String(), n.GetClosableComponentName(coreComponents, 0))

	component := &struct {
		factory.Closer
	}{}

	index := 45
	componentName := n.GetClosableComponentName(component, index)
	assert.True(t, strings.Contains(componentName, fmt.Sprintf("n.closableComponents[%d] - ", index)))
}

func TestNode_GetProofInvalidRootHash(t *testing.T) {
	t.Parallel()

	stateComponents := getDefaultStateComponents()
	n, _ := node.NewNode(node.WithStateComponents(stateComponents))

	response, err := n.GetProof("invalidRootHash", "0123")
	assert.Nil(t, response)
	assert.NotNil(t, err)
}

func TestNode_GetProofInvalidKey(t *testing.T) {
	t.Parallel()

	stateComponents := getDefaultStateComponents()
	n, _ := node.NewNode(
		node.WithStateComponents(stateComponents),
		node.WithCoreComponents(getDefaultCoreComponents()),
	)

	response, err := n.GetProof("deadbeef", "key")
	assert.Nil(t, response)
	assert.NotNil(t, err)
}

func TestNode_GetProofShouldWork(t *testing.T) {
	t.Parallel()

	trieKey := "0123"
	value := []byte("value")
	proof := [][]byte{[]byte("valid"), []byte("proof")}
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{
		GetTrieCalled: func(_ []byte) (common.Trie, error) {
			return &trieMock.TrieStub{
				GetProofCalled: func(key []byte) ([][]byte, []byte, error) {
					assert.Equal(t, trieKey, hex.EncodeToString(key))
					return proof, value, nil
				},
			}, nil
		},
	}
	n, _ := node.NewNode(
		node.WithStateComponents(stateComponents),
		node.WithCoreComponents(getDefaultCoreComponents()),
	)

	rootHash := "deadbeef"
	response, err := n.GetProof(rootHash, trieKey)
	assert.Nil(t, err)
	assert.Equal(t, proof, response.Proof)
	assert.Equal(t, value, response.Value)
	assert.Equal(t, rootHash, response.RootHash)
}

func TestNode_getProofTrieNotPresent(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("expected err")
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{
		GetTrieCalled: func(_ []byte) (common.Trie, error) {
			return nil, expectedErr
		},
	}
	n, _ := node.NewNode(
		node.WithStateComponents(stateComponents),
		node.WithCoreComponents(getDefaultCoreComponents()),
	)

	response, err := n.ComputeProof([]byte("deadbeef"), []byte("0123"))
	assert.Nil(t, response)
	assert.Equal(t, expectedErr, err)
}

func TestNode_getProofErrWhenComputingProof(t *testing.T) {
	t.Parallel()

	dataComponents := getDefaultDataComponents()
	expectedErr := fmt.Errorf("expected err")
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{
		GetTrieCalled: func(_ []byte) (common.Trie, error) {
			return &trieMock.TrieStub{
				GetProofCalled: func(_ []byte) ([][]byte, []byte, error) {
					return nil, nil, expectedErr
				},
			}, nil
		},
		RecreateTrieCalled: func(_ []byte) error {
			return nil
		},
	}
	n, _ := node.NewNode(
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithCoreComponents(getDefaultCoreComponents()),
	)

	response, err := n.ComputeProof([]byte("deadbeef"), []byte("0123"))
	assert.Nil(t, response)
	assert.Equal(t, expectedErr, err)
}

func TestNode_GetProofDataTrieInvalidRootHash(t *testing.T) {
	t.Parallel()

	stateComponents := getDefaultStateComponents()
	n, _ := node.NewNode(node.WithStateComponents(stateComponents))

	responseMainTrie, responseDataTrie, err := n.GetProofDataTrie("invalidRootHash", "0123", "4567")
	assert.Nil(t, responseMainTrie)
	assert.Nil(t, responseDataTrie)
	assert.NotNil(t, err)
}

func TestNode_GetProofDataTrieInvalidAddress(t *testing.T) {
	t.Parallel()

	stateComponents := getDefaultStateComponents()
	n, _ := node.NewNode(
		node.WithStateComponents(stateComponents),
		node.WithCoreComponents(getDefaultCoreComponents()),
	)

	responseMainTrie, responseDataTrie, err := n.GetProofDataTrie("deadbeef", "address", "4567")
	assert.Nil(t, responseMainTrie)
	assert.Nil(t, responseDataTrie)
	assert.NotNil(t, err)
}

func TestNode_GetProofDataTrieInvalidKey(t *testing.T) {
	t.Parallel()

	stateComponents := getDefaultStateComponents()
	n, _ := node.NewNode(
		node.WithStateComponents(stateComponents),
		node.WithCoreComponents(getDefaultCoreComponents()),
	)

	responseMainTrie, responseDataTrie, err := n.GetProofDataTrie("deadbeef", "0123", "key")
	assert.Nil(t, responseMainTrie)
	assert.Nil(t, responseDataTrie)
	assert.NotNil(t, err)
}

func TestNode_GetProofDataTrieShouldWork(t *testing.T) {
	t.Parallel()

	mainTrieKey := "0123"
	dataTrieKey := "4567"
	mainTrieValue := []byte("mainValue")
	dataTrieValue := []byte("dataTrieValue")
	mainTrieProof := [][]byte{[]byte("valid"), []byte("proof"), []byte("mainTrie")}
	dataTrieProof := [][]byte{[]byte("valid"), []byte("proof"), []byte("dataTrie")}
	dataTrieRootHash := []byte("dataTrieRoot")
	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{
		GetTrieCalled: func(_ []byte) (common.Trie, error) {
			return &trieMock.TrieStub{
				GetProofCalled: func(key []byte) ([][]byte, []byte, error) {
					if hex.EncodeToString(key) == mainTrieKey {
						return mainTrieProof, mainTrieValue, nil
					}
					if hex.EncodeToString(key) == dataTrieKey {
						return dataTrieProof, dataTrieValue, nil
					}

					return nil, nil, fmt.Errorf("key not found")
				},
			}, nil
		},
		GetAccountFromBytesCalled: func(address []byte, accountBytes []byte) (vmcommon.AccountHandler, error) {
			acc := &stateMock.AccountWrapMock{}
			acc.SetTrackableDataTrie(&trieMock.DataTrieTrackerStub{
				RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
					assert.Equal(t, dataTrieKey, hex.EncodeToString(key))
					return dataTrieValue, 0, nil
				},
			})
			acc.SetRootHash(dataTrieRootHash)
			return acc, nil
		},
	}
	n, _ := node.NewNode(
		node.WithStateComponents(stateComponents),
		node.WithCoreComponents(getDefaultCoreComponents()),
	)

	rootHash := "deadbeef"
	mainTrieResponse, dataTrieResponse, err := n.GetProofDataTrie(rootHash, mainTrieKey, dataTrieKey)
	assert.Nil(t, err)
	assert.Equal(t, mainTrieProof, mainTrieResponse.Proof)
	assert.Equal(t, mainTrieValue, mainTrieResponse.Value)
	assert.Equal(t, rootHash, mainTrieResponse.RootHash)

	assert.Equal(t, dataTrieProof, dataTrieResponse.Proof)
	assert.Equal(t, dataTrieValue, dataTrieResponse.Value)
	assert.Equal(t, hex.EncodeToString(dataTrieRootHash), dataTrieResponse.RootHash)
}

func TestNode_VerifyProofInvalidRootHash(t *testing.T) {
	t.Parallel()

	stateComponents := getDefaultStateComponents()
	n, _ := node.NewNode(node.WithStateComponents(stateComponents))

	response, err := n.VerifyProof("invalidRootHash", "0123", [][]byte{})
	assert.False(t, response)
	assert.NotNil(t, err)
}

func TestNode_VerifyProofInvalidAddress(t *testing.T) {
	t.Parallel()

	stateComponents := getDefaultStateComponents()
	stateComponents.AccountsAPI = &stateMock.AccountsStub{
		GetTrieCalled: func(_ []byte) (common.Trie, error) {
			return &trieMock.TrieStub{}, nil
		},
	}
	n, _ := node.NewNode(
		node.WithStateComponents(stateComponents),
		node.WithCoreComponents(getDefaultCoreComponents()),
	)

	response, err := n.VerifyProof("deadbeef", "address", [][]byte{})
	assert.False(t, response)
	assert.NotNil(t, err)
}

func TestNode_VerifyProof(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	coreComponents.Hash = sha256.NewSha256()
	coreComponents.IntMarsh = &marshal.GogoProtoMarshalizer{}
	n, _ := node.NewNode(
		node.WithStateComponents(getDefaultStateComponents()),
		node.WithCoreComponents(coreComponents),
	)

	rootHash := "bc2e549d98c31ffe6e9419b933d03b37e84f74c42601412302799d277651a6d8"
	address := "bf42213747697e9dec4211ef50ba6061b54729b53ba0c4994948cab478af8854"
	p, _ := hex.DecodeString("0a41040508080f0a0807040b0a0c080409040909040c000a0b03050b09020704050b010600060a0b00050f0e010102040c0e0d090e07090607040703010202040f0b10124c1202000022206182d14320be95434f5508acad9478d3b6cf837bfce7ebfe47c2e860d1b98ca72a20bf42213747697e9dec4211ef50ba6061b54729b53ba0c4994948cab478af88543202000001")
	proof := [][]byte{p}

	response, err := n.VerifyProof(rootHash, address, proof)
	assert.True(t, response)
	assert.Nil(t, err)
}

func TestNode_IsDataTrieMigrated(t *testing.T) {
	t.Parallel()

	t.Run("invalid address", func(t *testing.T) {
		t.Parallel()

		n, _ := node.NewNode(
			node.WithStateComponents(getDefaultStateComponents()),
			node.WithCoreComponents(getDefaultCoreComponents()),
		)

		isMigrated, err := n.IsDataTrieMigrated("invalid address", api.AccountQueryOptions{})
		assert.False(t, isMigrated)
		assert.NotNil(t, err)
	})

	t.Run("load account err", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("load account error")
		stateComponents := getDefaultStateComponents()
		stateComponents.AccountsRepo = &stateMock.AccountsRepositoryStub{
			GetAccountWithBlockInfoCalled: func(_ []byte, _ api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
				return nil, nil, expectedErr
			},
		}

		n, _ := node.NewNode(
			node.WithStateComponents(stateComponents),
			node.WithCoreComponents(getDefaultCoreComponents()),
		)

		isMigrated, err := n.IsDataTrieMigrated("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l", api.AccountQueryOptions{})
		assert.False(t, isMigrated)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("wrong type assertion", func(t *testing.T) {
		t.Parallel()

		stateComponents := getDefaultStateComponents()
		stateComponents.AccountsRepo = &stateMock.AccountsRepositoryStub{
			GetAccountWithBlockInfoCalled: func(_ []byte, _ api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
				return &stateMock.AccountWrapMock{}, nil, nil
			},
		}

		n, _ := node.NewNode(
			node.WithStateComponents(stateComponents),
			node.WithCoreComponents(getDefaultCoreComponents()),
		)

		isMigrated, err := n.IsDataTrieMigrated("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l", api.AccountQueryOptions{})
		assert.False(t, isMigrated)
		assert.True(t, strings.Contains(err.Error(), "wrong type assertion"))
	})

	t.Run("should work and return false", func(t *testing.T) {
		t.Parallel()

		acc := createAcc([]byte("000000000000000000010000000000000000000000000000000000000001ffff"))
		acc.SetDataTrie(&trieMock.TrieStub{
			IsMigratedToLatestVersionCalled: func() (bool, error) {
				return false, nil
			},
		})

		stateComponents := getDefaultStateComponents()
		stateComponents.AccountsRepo = &stateMock.AccountsRepositoryStub{
			GetAccountWithBlockInfoCalled: func(_ []byte, _ api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
				return acc, nil, nil
			},
		}

		n, _ := node.NewNode(
			node.WithStateComponents(stateComponents),
			node.WithCoreComponents(getDefaultCoreComponents()),
		)

		isMigrated, err := n.IsDataTrieMigrated("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l", api.AccountQueryOptions{})
		assert.False(t, isMigrated)
		assert.Nil(t, err)
	})

	t.Run("should work and return true", func(t *testing.T) {
		t.Parallel()

		acc := createAcc([]byte("000000000000000000010000000000000000000000000000000000000001ffff"))
		acc.SetDataTrie(&trieMock.TrieStub{
			IsMigratedToLatestVersionCalled: func() (bool, error) {
				return true, nil
			},
		})

		stateComponents := getDefaultStateComponents()
		stateComponents.AccountsRepo = &stateMock.AccountsRepositoryStub{
			GetAccountWithBlockInfoCalled: func(_ []byte, _ api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
				return acc, nil, nil
			},
		}

		n, _ := node.NewNode(
			node.WithStateComponents(stateComponents),
			node.WithCoreComponents(getDefaultCoreComponents()),
		)

		isMigrated, err := n.IsDataTrieMigrated("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l", api.AccountQueryOptions{})
		assert.True(t, isMigrated)
		assert.Nil(t, err)
	})
}

func TestGetESDTSupplyError(t *testing.T) {
	t.Parallel()

	localErr := errors.New("local error")
	historyProc := &dblookupext.HistoryRepositoryStub{
		GetESDTSupplyCalled: func(token string) (*esdtSupply.SupplyESDT, error) {
			return nil, localErr
		},
	}
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.HistoryRepositoryInternal = historyProc

	n, _ := node.NewNode(
		node.WithProcessComponents(processComponentsMock),
	)

	_, err := n.GetTokenSupply("my-token")
	require.Equal(t, localErr, err)
}

func TestGetESDTSupply(t *testing.T) {
	t.Parallel()

	historyProc := &dblookupext.HistoryRepositoryStub{
		GetESDTSupplyCalled: func(token string) (*esdtSupply.SupplyESDT, error) {
			return &esdtSupply.SupplyESDT{
				Supply: big.NewInt(100),
				Minted: big.NewInt(15),
			}, nil
		},
	}
	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.HistoryRepositoryInternal = historyProc

	n, _ := node.NewNode(
		node.WithProcessComponents(processComponentsMock),
	)

	supply, err := n.GetTokenSupply("my-token")
	require.Nil(t, err)

	require.Equal(t, &api.ESDTSupply{
		Supply: "100",
		Burned: "0",
		Minted: "15",
	}, supply)
}

func TestNode_SendBulkTransactions(t *testing.T) {
	t.Parallel()

	flag := atomicCore.Flag{}
	expectedNoOfTxs := uint64(444)
	tx1 := &transaction.Transaction{Nonce: 123}
	tx2 := &transaction.Transaction{Nonce: 321}
	expectedTxs := []*transaction.Transaction{tx1, tx2}
	txsSender := &txsSenderMock.TxsSenderHandlerMock{
		SendBulkTransactionsCalled: func(txs []*transaction.Transaction) (uint64, error) {
			flag.SetValue(true)
			require.Equal(t, expectedTxs, txs)
			return expectedNoOfTxs, nil
		},
	}

	processComponentsMock := getDefaultProcessComponents()
	processComponentsMock.TxsSenderHandlerField = txsSender
	n, err := node.NewNode(node.WithProcessComponents(processComponentsMock))
	require.Nil(t, err)

	actualNoOfTxs, err := n.SendBulkTransactions(expectedTxs)
	require.True(t, flag.IsSet())
	require.Equal(t, expectedNoOfTxs, actualNoOfTxs)
	require.Nil(t, err)
}

func TestNode_GetHeartbeats(t *testing.T) {
	t.Parallel()

	numMessages := 5
	providedMessages := make([]heartbeatData.PubKeyHeartbeat, numMessages)
	for i := 0; i < numMessages; i++ {
		providedMessages[i] = createHeartbeatMessage("v2", i, true)
	}

	heartbeatV2Components := createMockHeartbeatV2Components(providedMessages)

	n, err := node.NewNode(node.WithHeartbeatV2Components(heartbeatV2Components))
	require.Nil(t, err)

	receivedMessages := n.GetHeartbeats()
	assert.True(t, sameMessages(providedMessages, receivedMessages))
}

func TestNode_Getters(t *testing.T) {
	t.Parallel()

	coreComponents := getDefaultCoreComponents()
	statusCoreComponents := &factoryTests.StatusCoreComponentsStub{
		AppStatusHandlerField: &statusHandlerMock.AppStatusHandlerStub{},
	}
	cryptoComponents := getDefaultCryptoComponents()
	stateComponents := getDefaultStateComponents()
	bootstrapComponents := getDefaultBootstrapComponents()
	dataComponents := getDefaultDataComponents()
	heartbeatComponents := &factoryMock.HeartbeatV2ComponentsStub{}
	networkComponents := getDefaultNetworkComponents()
	processComponents := getDefaultProcessComponents()
	consensusGroupSize := 10

	n, err := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithStatusCoreComponents(statusCoreComponents),
		node.WithCryptoComponents(cryptoComponents),
		node.WithStateComponents(stateComponents),
		node.WithBootstrapComponents(bootstrapComponents),
		node.WithDataComponents(dataComponents),
		node.WithHeartbeatV2Components(heartbeatComponents),
		node.WithNetworkComponents(networkComponents),
		node.WithProcessComponents(processComponents),
		node.WithConsensusGroupSize(consensusGroupSize),
		node.WithImportMode(true),
	)
	require.Nil(t, err)

	//pointer testing
	assert.True(t, n.GetCoreComponents() == coreComponents)
	assert.True(t, n.GetStatusCoreComponents() == statusCoreComponents)
	assert.True(t, n.GetCryptoComponents() == cryptoComponents)
	assert.True(t, n.GetStateComponents() == stateComponents)
	assert.True(t, n.GetBootstrapComponents() == bootstrapComponents)
	assert.True(t, n.GetDataComponents() == dataComponents)
	assert.True(t, n.GetHeartbeatV2Components() == heartbeatComponents)
	assert.True(t, n.GetNetworkComponents() == networkComponents)
	assert.True(t, n.GetProcessComponents() == processComponents)
	assert.Equal(t, consensusGroupSize, n.GetConsensusGroupSize())
	assert.True(t, n.IsInImportMode())
}

func TestNode_GetEpochStartDataAPI(t *testing.T) {
	t.Parallel()

	prevHash := []byte("prevHash")
	rootHash := []byte("rootHash")
	accumulatedFees := big.NewInt(100)
	developerFees := big.NewInt(200)

	dataComponents := getDefaultDataComponents()
	blockchain := dataComponents.BlockChain.(*testscommon.ChainHandlerStub)
	timestamp := uint64(778899)
	shardID := uint32(2)
	blockchain.GetGenesisHeaderCalled = func() data.HeaderHandler {
		return &block.Header{
			TimeStamp:       timestamp,
			ShardID:         shardID,
			PrevHash:        prevHash,
			RootHash:        rootHash,
			AccumulatedFees: accumulatedFees,
			DeveloperFees:   developerFees,
		}
	}

	bootstrapComponents := getDefaultBootstrapComponents()
	shardCoordinator := bootstrapComponents.ShardCoordinator().(*mock.ShardCoordinatorMock)

	coreComponents := getDefaultCoreComponents()

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithBootstrapComponents(bootstrapComponents),
	)
	epoch := uint32(37)
	nonce := uint64(112233)
	round := uint64(445566)

	t.Run("genesis block should work", func(t *testing.T) {
		result, err := n.GetEpochStartDataAPI(0)
		assert.Nil(t, err)
		expectedResult := &common.EpochStartDataAPI{
			Nonce:             0,
			Round:             0,
			Timestamp:         int64(timestamp),
			Epoch:             0,
			Shard:             shardID,
			PrevBlockHash:     hex.EncodeToString(prevHash),
			StateRootHash:     hex.EncodeToString(rootHash),
			ScheduledRootHash: "",
			AccumulatedFees:   accumulatedFees.String(),
			DeveloperFees:     developerFees.String(),
		}
		assert.Equal(t, expectedResult, result)
	})
	t.Run("should work for metachain", func(t *testing.T) {
		shardCoordinator.SelfShardId = core.MetachainShardId

		returnedHeader := &block.MetaBlock{
			Nonce:           nonce,
			Epoch:           epoch,
			Round:           round,
			TimeStamp:       timestamp,
			PrevHash:        prevHash,
			RootHash:        rootHash,
			AccumulatedFees: accumulatedFees,
			DeveloperFees:   developerFees,
		}

		headerBytes, err := coreComponents.IntMarsh.Marshal(returnedHeader)
		require.Nil(t, err)

		unit := &mockStorage.StorerStub{
			GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
				expectedIdentifier := core.EpochStartIdentifier(epoch)
				require.Equal(t, expectedIdentifier, string(key))

				return headerBytes, nil
			},
		}

		storageService := dataComponents.StorageService().(*mockStorage.ChainStorerStub)
		storageService.GetStorerCalled = func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			require.Equal(t, dataRetriever.MetaBlockUnit, unitType)
			return unit, nil
		}

		result, err := n.GetEpochStartDataAPI(epoch)
		assert.Nil(t, err)

		expectedResult := &common.EpochStartDataAPI{
			Nonce:             nonce,
			Round:             round,
			Timestamp:         int64(timestamp),
			Epoch:             epoch,
			Shard:             core.MetachainShardId,
			PrevBlockHash:     hex.EncodeToString(prevHash),
			StateRootHash:     hex.EncodeToString(rootHash),
			ScheduledRootHash: "",
			AccumulatedFees:   accumulatedFees.String(),
			DeveloperFees:     developerFees.String(),
		}
		assert.Equal(t, expectedResult, result)
	})
	t.Run("should work for shard chain", func(t *testing.T) {
		shardCoordinator.SelfShardId = 0

		returnedHeader := &block.Header{
			Nonce:           nonce,
			Epoch:           epoch,
			Round:           round,
			ShardID:         shardID,
			TimeStamp:       timestamp,
			PrevHash:        prevHash,
			RootHash:        rootHash,
			AccumulatedFees: accumulatedFees,
			DeveloperFees:   developerFees,
		}

		headerBytes, err := coreComponents.IntMarsh.Marshal(returnedHeader)
		require.Nil(t, err)

		unit := &mockStorage.StorerStub{
			GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
				expectedIdentifier := core.EpochStartIdentifier(epoch)
				require.Equal(t, expectedIdentifier, string(key))

				return headerBytes, nil
			},
		}

		storageService := dataComponents.StorageService().(*mockStorage.ChainStorerStub)
		storageService.GetStorerCalled = func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			require.Equal(t, dataRetriever.BlockHeaderUnit, unitType)
			return unit, nil
		}

		result, err := n.GetEpochStartDataAPI(epoch)
		assert.Nil(t, err)

		expectedResult := &common.EpochStartDataAPI{
			Nonce:             nonce,
			Round:             round,
			Timestamp:         int64(timestamp),
			Epoch:             epoch,
			Shard:             shardID,
			PrevBlockHash:     hex.EncodeToString(prevHash),
			StateRootHash:     hex.EncodeToString(rootHash),
			ScheduledRootHash: "",
			AccumulatedFees:   accumulatedFees.String(),
			DeveloperFees:     developerFees.String(),
		}
		assert.Equal(t, expectedResult, result)
	})
}

func createMockHeartbeatV2Components(providedMessages []heartbeatData.PubKeyHeartbeat) *factoryMock.HeartbeatV2ComponentsStub {
	heartbeatV2Components := &factoryMock.HeartbeatV2ComponentsStub{}
	heartbeatV2Components.MonitorField = &integrationTestsMock.HeartbeatMonitorStub{
		GetHeartbeatsCalled: func() []heartbeatData.PubKeyHeartbeat {
			return providedMessages
		},
	}

	return heartbeatV2Components
}

func sameMessages(provided, received []heartbeatData.PubKeyHeartbeat) bool {
	providedLen, receivedLen := len(provided), len(received)
	if receivedLen != providedLen {
		return false
	}

	areEqual := true
	for i := 0; i < providedLen; i++ {
		p := provided[i]
		r := received[i]
		areEqual = areEqual &&
			(p.PublicKey == r.PublicKey) &&
			(p.TimeStamp == r.TimeStamp) &&
			(p.IsActive == r.IsActive) &&
			(p.ReceivedShardID == r.ReceivedShardID) &&
			(p.ComputedShardID == r.ComputedShardID) &&
			(p.VersionNumber == r.VersionNumber) &&
			(p.Identity == r.Identity) &&
			(p.PeerType == r.PeerType) &&
			(p.Nonce == r.Nonce) &&
			(p.NumInstances == r.NumInstances) &&
			(p.PeerSubType == r.PeerSubType) &&
			(p.PidString == r.PidString)

		if !areEqual {
			return false
		}
	}

	return true
}

func createHeartbeatMessage(prefix string, idx int, isActive bool) heartbeatData.PubKeyHeartbeat {
	return heartbeatData.PubKeyHeartbeat{
		PublicKey:       fmt.Sprintf("%d%spk", idx, prefix),
		TimeStamp:       time.Now(),
		IsActive:        isActive,
		ReceivedShardID: 0,
		ComputedShardID: 0,
		VersionNumber:   "v01",
		NodeDisplayName: fmt.Sprintf("%d%s", idx, "node"),
		Identity:        "identity",
		PeerType:        core.ValidatorPeer.String(),
		Nonce:           10,
		NumInstances:    1,
		PeerSubType:     1,
		PidString:       fmt.Sprintf("%d%spid", idx, prefix),
	}
}

func TestNode_setTxGuardianData(t *testing.T) {
	t.Parallel()
	lenPubKey := 32
	coreComponents := getDefaultCoreComponents()
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
	)
	guardianPubKey := bytes.Repeat([]byte{1}, lenPubKey)
	guardian, _ := coreComponents.AddrPubKeyConv.Encode(guardianPubKey)
	guardianSig := []byte("guardian sig")
	guardianSigHex := hex.EncodeToString(guardianSig)

	t.Run("invalid guardian address should err", func(t *testing.T) {
		tx := &transaction.Transaction{}
		tx.Options |= transaction.MaskGuardedTransaction

		err := n.SetTxGuardianData("invalid guardian address", guardianSigHex, tx)
		require.NotNil(t, err)
		require.Nil(t, tx.GuardianAddr)
		require.Nil(t, tx.GuardianSignature)
	})
	t.Run("invalid guardian sig hex should err", func(t *testing.T) {
		tx := &transaction.Transaction{}
		tx.Options |= transaction.MaskGuardedTransaction

		err := n.SetTxGuardianData(guardian, "invalid guardian sig hex", tx)
		require.NotNil(t, err)
		require.Nil(t, tx.GuardianAddr)
		require.Nil(t, tx.GuardianSignature)
	})
	t.Run("no guardian option set on tx should err", func(t *testing.T) {
		tx := &transaction.Transaction{}

		err := n.SetTxGuardianData(guardian, guardianSigHex, tx)
		require.NotNil(t, err)
		require.Nil(t, tx.GuardianAddr)
		require.Nil(t, tx.GuardianSignature)
	})
	t.Run("setTxGuardianData ok", func(t *testing.T) {
		tx := &transaction.Transaction{}
		tx.Options |= transaction.MaskGuardedTransaction

		err := n.SetTxGuardianData(guardian, guardianSigHex, tx)
		require.Nil(t, err)
		require.Equal(t, guardianPubKey, tx.GuardianAddr)
		require.Equal(t, guardianSig, tx.GuardianSignature)
	})
}

func TestNode_GetGuardianData(t *testing.T) {
	userAddressBytes := bytes.Repeat([]byte{3}, 32)

	testAccount, _ := accounts.NewUserAccount(userAddressBytes, &trieMock.DataTrieTrackerStub{}, &trieMock.TrieLeafParserStub{})
	testAccountsDB := &stateMock.AccountsStub{
		GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
			return testAccount, nil, nil
		},
		RecreateTrieCalled: func(_ []byte) error {
			return nil
		},
	}
	coreComponents := getDefaultCoreComponents()
	dataComponents := getDefaultDataComponents()
	coreComponents.IntMarsh = getMarshalizer()
	coreComponents.VmMarsh = getMarshalizer()
	coreComponents.Hash = getHasher()
	coreComponents.AddrPubKeyConv = createMockPubkeyConverter()
	testStateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      testAccountsDB,
		CurrentStateAccountsWrapper:    testAccountsDB,
		HistoricalStateAccountsWrapper: testAccountsDB,
	}
	testStateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)
	userAddress, _ := coreComponents.AddressPubKeyConverter().Encode(userAddressBytes)
	g1 := &guardians.Guardian{
		Address:         bytes.Repeat([]byte{1}, 32),
		ActivationEpoch: 0,
	}
	g2 := &guardians.Guardian{
		Address:         bytes.Repeat([]byte{2}, 32),
		ActivationEpoch: 1,
	}
	addressG1, _ := coreComponents.AddressPubKeyConverter().Encode(g1.Address)
	apiG1 := &api.Guardian{
		Address:         addressG1,
		ActivationEpoch: g1.ActivationEpoch,
	}
	addressG2, _ := coreComponents.AddressPubKeyConverter().Encode(g2.Address)
	apiG2 := &api.Guardian{
		Address:         addressG2,
		ActivationEpoch: g2.ActivationEpoch,
	}
	t.Run("error on loadUserAccountHandlerByAddress", func(t *testing.T) {
		accDB := &stateMock.AccountsStub{
			GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
				return testAccount, nil, nil
			},
			RecreateTrieCalled: func(_ []byte) error {
				return nil
			},
		}
		stateComponents := getDefaultStateComponents()
		argsLocal := state.ArgsAccountsRepository{
			FinalStateAccountsWrapper:      accDB,
			CurrentStateAccountsWrapper:    accDB,
			HistoricalStateAccountsWrapper: accDB,
		}
		stateComponents.AccountsRepo, _ = state.NewAccountsRepository(argsLocal)
		n, _ := node.NewNode(
			node.WithDataComponents(dataComponents),
			node.WithCoreComponents(coreComponents),
			node.WithStateComponents(stateComponents),
		)
		guardianData, blockInfo, err := n.GetGuardianData("address", api.AccountQueryOptions{})
		require.Equal(t, api.GuardianData{}, guardianData)
		require.Equal(t, api.BlockInfo{}, blockInfo)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), "invalid address"))
	})
	t.Run("error on loadUserAccountHandlerByAddress but account is new", func(t *testing.T) {
		providedBlockInfo := holders.NewBlockInfo([]byte{0xaa}, 7, []byte{0xbb})
		accDB := &stateMock.AccountsStub{
			GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
				return nil, nil, state.NewErrAccountNotFoundAtBlock(providedBlockInfo)
			},
			RecreateTrieCalled: func(_ []byte) error {
				return nil
			},
		}
		stateComponents := getDefaultStateComponents()
		argsLocal := state.ArgsAccountsRepository{
			FinalStateAccountsWrapper:      accDB,
			CurrentStateAccountsWrapper:    accDB,
			HistoricalStateAccountsWrapper: accDB,
		}
		stateComponents.AccountsRepo, _ = state.NewAccountsRepository(argsLocal)
		n, _ := node.NewNode(
			node.WithDataComponents(dataComponents),
			node.WithCoreComponents(coreComponents),
			node.WithStateComponents(stateComponents),
		)
		guardianData, blockInfo, err := n.GetGuardianData(userAddress, api.AccountQueryOptions{})
		require.Equal(t, api.GuardianData{}, guardianData)
		expectedBlockInfo := api.BlockInfo{
			Nonce:    providedBlockInfo.GetNonce(),
			Hash:     hex.EncodeToString(providedBlockInfo.GetHash()),
			RootHash: hex.EncodeToString(providedBlockInfo.GetRootHash()),
		}
		require.Equal(t, expectedBlockInfo, blockInfo)
		require.Nil(t, err)
	})
	t.Run("getPendingAndActiveGuardians with error", func(t *testing.T) {
		expectedError := errors.New("expected error")
		bootstrapComponents := getDefaultBootstrapComponents()
		bootstrapComponents.GuardedAccountHandlerField = &guardianMocks.GuardedAccountHandlerStub{
			GetConfiguredGuardiansCalled: func(uah state.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error) {
				return nil, nil, expectedError
			},
		}
		n, _ := node.NewNode(
			node.WithDataComponents(dataComponents),
			node.WithCoreComponents(coreComponents),
			node.WithStateComponents(testStateComponents),
			node.WithBootstrapComponents(bootstrapComponents),
		)
		guardianData, blockInfo, err := n.GetGuardianData(userAddress, api.AccountQueryOptions{})
		require.Equal(t, api.GuardianData{}, guardianData)
		require.Equal(t, api.BlockInfo{}, blockInfo)
		require.Equal(t, expectedError, err)
	})
	t.Run("one active", func(t *testing.T) {
		bootstrapComponents := getDefaultBootstrapComponents()
		bootstrapComponents.GuardedAccountHandlerField = &guardianMocks.GuardedAccountHandlerStub{
			GetConfiguredGuardiansCalled: func(uah state.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error) {
				return g1, nil, nil
			},
		}
		n, _ := node.NewNode(
			node.WithDataComponents(dataComponents),
			node.WithCoreComponents(coreComponents),
			node.WithStateComponents(testStateComponents),
			node.WithBootstrapComponents(bootstrapComponents),
		)
		guardianData, blockInfo, err := n.GetGuardianData(userAddress, api.AccountQueryOptions{})
		require.Equal(t, api.GuardianData{
			ActiveGuardian:  apiG1,
			PendingGuardian: nil,
			Guarded:         false,
		}, guardianData)
		require.Equal(t, api.BlockInfo{}, blockInfo)
		require.Nil(t, err)
	})
	t.Run("one pending", func(t *testing.T) {
		bootstrapComponents := getDefaultBootstrapComponents()
		bootstrapComponents.GuardedAccountHandlerField = &guardianMocks.GuardedAccountHandlerStub{
			GetConfiguredGuardiansCalled: func(uah state.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error) {
				return nil, g1, nil
			},
		}
		n, _ := node.NewNode(
			node.WithDataComponents(dataComponents),
			node.WithCoreComponents(coreComponents),
			node.WithStateComponents(testStateComponents),
			node.WithBootstrapComponents(bootstrapComponents),
		)
		guardianData, blockInfo, err := n.GetGuardianData(userAddress, api.AccountQueryOptions{})
		require.Equal(t, api.GuardianData{
			ActiveGuardian:  nil,
			PendingGuardian: apiG1,
			Guarded:         false,
		}, guardianData)
		require.Equal(t, api.BlockInfo{}, blockInfo)
		require.Nil(t, err)
	})
	t.Run("one active and one pending", func(t *testing.T) {
		bootstrapComponents := getDefaultBootstrapComponents()
		bootstrapComponents.GuardedAccountHandlerField = &guardianMocks.GuardedAccountHandlerStub{
			GetConfiguredGuardiansCalled: func(uah state.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error) {
				return g1, g2, nil
			},
		}
		n, _ := node.NewNode(
			node.WithDataComponents(dataComponents),
			node.WithCoreComponents(coreComponents),
			node.WithStateComponents(testStateComponents),
			node.WithBootstrapComponents(bootstrapComponents),
		)
		guardianData, blockInfo, err := n.GetGuardianData(userAddress, api.AccountQueryOptions{})
		require.Equal(t, api.GuardianData{
			ActiveGuardian:  apiG1,
			PendingGuardian: apiG2,
			Guarded:         false,
		}, guardianData)
		require.Equal(t, api.BlockInfo{}, blockInfo)
		require.Nil(t, err)
	})
	t.Run("one active and one pending and account guarded", func(t *testing.T) {
		acc, _ := accounts.NewUserAccount(userAddressBytes, &trieMock.DataTrieTrackerStub{}, &trieMock.TrieLeafParserStub{})
		acc.CodeMetadata = (&vmcommon.CodeMetadata{Guarded: true}).ToBytes()
		accDB := &stateMock.AccountsStub{
			GetAccountWithBlockInfoCalled: func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
				return acc, nil, nil
			},
			RecreateTrieCalled: func(_ []byte) error {
				return nil
			},
		}
		stateComponents := getDefaultStateComponents()
		argsLocal := state.ArgsAccountsRepository{
			FinalStateAccountsWrapper:      accDB,
			CurrentStateAccountsWrapper:    accDB,
			HistoricalStateAccountsWrapper: accDB,
		}
		stateComponents.AccountsRepo, _ = state.NewAccountsRepository(argsLocal)
		bootstrapComponents := getDefaultBootstrapComponents()
		bootstrapComponents.GuardedAccountHandlerField = &guardianMocks.GuardedAccountHandlerStub{
			GetConfiguredGuardiansCalled: func(uah state.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error) {
				return g1, g2, nil
			},
		}
		n, _ := node.NewNode(
			node.WithDataComponents(dataComponents),
			node.WithCoreComponents(coreComponents),
			node.WithStateComponents(stateComponents),
			node.WithBootstrapComponents(bootstrapComponents),
		)
		guardianData, blockInfo, err := n.GetGuardianData(userAddress, api.AccountQueryOptions{})
		require.Equal(t, api.GuardianData{
			ActiveGuardian:  apiG1,
			PendingGuardian: apiG2,
			Guarded:         true,
		}, guardianData)
		require.Equal(t, api.BlockInfo{}, blockInfo)
		require.Nil(t, err)
	})
}

func TestNode_getPendingAndActiveGuardians(t *testing.T) {
	coreComponents := getDefaultCoreComponents()
	bootstrapComponents := getDefaultBootstrapComponents()
	expectedErr := errors.New("expected err")
	g1PubKey := bytes.Repeat([]byte{1}, 32)
	g2PubKey := bytes.Repeat([]byte{2}, 32)
	g1 := &guardians.Guardian{
		Address:         g1PubKey,
		ActivationEpoch: 10,
	}
	g2 := &guardians.Guardian{
		Address:         g2PubKey,
		ActivationEpoch: 1,
	}

	addressG1, _ := coreComponents.AddrPubKeyConv.Encode(g1.Address)
	expectedG1 := &api.Guardian{
		Address:         addressG1,
		ActivationEpoch: g1.ActivationEpoch,
	}
	addressG2, _ := coreComponents.AddrPubKeyConv.Encode(g2.Address)
	expectedG2 := &api.Guardian{
		Address:         addressG2,
		ActivationEpoch: g2.ActivationEpoch,
	}

	t.Run("get configured guardians with error should propagate error", func(t *testing.T) {
		bootstrapComponents.GuardedAccountHandlerField = &guardianMocks.GuardedAccountHandlerStub{
			GetConfiguredGuardiansCalled: func(uah state.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error) {
				return nil, nil, expectedErr
			},
		}
		n, _ := node.NewNode(
			node.WithCoreComponents(coreComponents),
			node.WithBootstrapComponents(bootstrapComponents),
		)

		activeGuardian, pendingGuardian, err := n.GetPendingAndActiveGuardians(&stateMock.UserAccountStub{})
		require.Nil(t, activeGuardian)
		require.Nil(t, pendingGuardian)
		require.Equal(t, expectedErr, err)
	})
	t.Run("no pending and no active but no error", func(t *testing.T) {
		bootstrapComponents.GuardedAccountHandlerField = &guardianMocks.GuardedAccountHandlerStub{
			GetConfiguredGuardiansCalled: func(uah state.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error) {
				return nil, nil, nil
			},
		}
		n, _ := node.NewNode(
			node.WithCoreComponents(coreComponents),
			node.WithBootstrapComponents(bootstrapComponents),
		)
		activeGuardian, pendingGuardian, err := n.GetPendingAndActiveGuardians(&stateMock.UserAccountStub{})
		require.Nil(t, activeGuardian)
		require.Nil(t, pendingGuardian)
		require.Nil(t, err)
	})
	t.Run("one active", func(t *testing.T) {
		bootstrapComponents.GuardedAccountHandlerField = &guardianMocks.GuardedAccountHandlerStub{
			GetConfiguredGuardiansCalled: func(uah state.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error) {
				return g1, nil, nil
			},
		}
		n, _ := node.NewNode(
			node.WithCoreComponents(coreComponents),
			node.WithBootstrapComponents(bootstrapComponents),
		)
		activeGuardian, pendingGuardian, err := n.GetPendingAndActiveGuardians(&stateMock.UserAccountStub{})
		require.NotNil(t, activeGuardian)

		require.Equal(t, expectedG1, activeGuardian)
		require.Nil(t, pendingGuardian)
		require.Nil(t, err)
	})
	t.Run("one pending", func(t *testing.T) {
		bootstrapComponents.GuardedAccountHandlerField = &guardianMocks.GuardedAccountHandlerStub{
			GetConfiguredGuardiansCalled: func(uah state.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error) {
				return nil, g1, nil
			},
		}
		n, _ := node.NewNode(
			node.WithCoreComponents(coreComponents),
			node.WithBootstrapComponents(bootstrapComponents),
		)
		activeGuardian, pendingGuardian, err := n.GetPendingAndActiveGuardians(&stateMock.UserAccountStub{})
		require.NotNil(t, pendingGuardian)
		require.Equal(t, expectedG1, pendingGuardian)
		require.Nil(t, activeGuardian)
		require.Nil(t, err)
	})
	t.Run("one active one pending", func(t *testing.T) {
		bootstrapComponents.GuardedAccountHandlerField = &guardianMocks.GuardedAccountHandlerStub{
			GetConfiguredGuardiansCalled: func(uah state.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error) {
				return g1, g2, nil
			},
		}
		n, _ := node.NewNode(
			node.WithCoreComponents(coreComponents),
			node.WithBootstrapComponents(bootstrapComponents),
		)

		activeGuardian, pendingGuardian, err := n.GetPendingAndActiveGuardians(&stateMock.UserAccountStub{})
		require.NotNil(t, activeGuardian)
		require.NotNil(t, pendingGuardian)
		require.Equal(t, expectedG2, pendingGuardian)
		require.Equal(t, expectedG1, activeGuardian)
		require.Nil(t, err)
	})
}

func getDefaultCoreComponents() *nodeMockFactory.CoreComponentsMock {
	return &nodeMockFactory.CoreComponentsMock{
		IntMarsh:            &marshallerMock.MarshalizerMock{},
		TxMarsh:             &marshallerMock.MarshalizerMock{},
		VmMarsh:             &marshallerMock.MarshalizerMock{},
		TxSignHasherField:   &testscommon.HasherStub{},
		Hash:                &testscommon.HasherStub{},
		UInt64ByteSliceConv: testscommon.NewNonceHashConverterMock(),
		AddrPubKeyConv:      testscommon.RealWorldBech32PubkeyConverter,
		ValPubKeyConv:       testscommon.NewPubkeyConverterMock(32),
		PathHdl:             &testscommon.PathManagerStub{},
		ChainIdCalled: func() string {
			return "chainID"
		},
		MinTransactionVersionCalled: func() uint32 {
			return 1
		},
		WDTimer:               &testscommon.WatchdogMock{},
		Alarm:                 &testscommon.AlarmSchedulerStub{},
		NtpTimer:              &testscommon.SyncTimerStub{},
		RoundHandlerField:     &testscommon.RoundHandlerMock{},
		EconomicsHandler:      &economicsmocks.EconomicsHandlerMock{},
		APIEconomicsHandler:   &economicsmocks.EconomicsHandlerMock{},
		RatingsConfig:         &testscommon.RatingsInfoMock{},
		RatingHandler:         &testscommon.RaterMock{},
		NodesConfig:           &testscommon.NodesSetupStub{},
		StartTime:             time.Time{},
		EpochChangeNotifier:   &epochNotifier.EpochNotifierStub{},
		TxVersionCheckHandler: versioning.NewTxVersionChecker(0),
		ChanStopProcess:       make(chan endProcess.ArgEndProcess, 1),
	}
}

func getDefaultProcessComponents() *factoryMock.ProcessComponentsMock {
	return &factoryMock.ProcessComponentsMock{
		NodesCoord: &shardingMocks.NodesCoordinatorMock{},
		ShardCoord: &testscommon.ShardsCoordinatorMock{
			NoShards:     1,
			CurrentShard: 0,
		},
		IntContainer:                         &testscommon.InterceptorsContainerStub{},
		ReqFinder:                            &dataRetrieverMock.RequestersFinderStub{},
		RoundHandlerField:                    &testscommon.RoundHandlerMock{},
		EpochTrigger:                         &testscommon.EpochStartTriggerStub{},
		EpochNotifier:                        &mock.EpochStartNotifierStub{},
		ForkDetect:                           &mock.ForkDetectorMock{},
		BlockProcess:                         &testscommon.BlockProcessorStub{},
		BlackListHdl:                         &testscommon.TimeCacheStub{},
		BootSore:                             &mock.BootstrapStorerMock{},
		HeaderSigVerif:                       &mock.HeaderSigVerifierStub{},
		HeaderIntegrVerif:                    &mock.HeaderIntegrityVerifierStub{},
		ValidatorStatistics:                  &mock.ValidatorStatisticsProcessorMock{},
		ValidatorProvider:                    &mock.ValidatorsProviderStub{},
		BlockTrack:                           &mock.BlockTrackerStub{},
		PendingMiniBlocksHdl:                 &mock.PendingMiniBlocksHandlerStub{},
		ReqHandler:                           &testscommon.RequestHandlerStub{},
		RequestedItemsHandlerInternal:        &testscommon.RequestedItemsHandlerStub{},
		TxLogsProcess:                        &mock.TxLogProcessorMock{},
		HeaderConstructValidator:             &mock.HeaderValidatorStub{},
		MainPeerMapper:                       &p2pmocks.NetworkShardingCollectorStub{},
		FullArchivePeerMapper:                &p2pmocks.NetworkShardingCollectorStub{},
		WhiteListHandlerInternal:             &testscommon.WhiteListHandlerStub{},
		WhiteListerVerifiedTxsInternal:       &testscommon.WhiteListHandlerStub{},
		TxsSenderHandlerField:                &txsSenderMock.TxsSenderHandlerMock{},
		ScheduledTxsExecutionHandlerInternal: &testscommon.ScheduledTxsExecutionStub{},
		HistoryRepositoryInternal:            &dblookupext.HistoryRepositoryStub{},
		ESDTDataStorageHandlerForAPIInternal: &testscommon.EsdtStorageHandlerStub{},
		ResContainer:                         &dataRetrieverMock.ResolversContainerStub{},
	}
}

func getDefaultDataComponents() *nodeMockFactory.DataComponentsMock {
	chainHandler := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 42}
		},
		GetCurrentBlockHeaderHashCalled: func() []byte {
			return []byte("header hash")
		},
		GetCurrentBlockRootHashCalled: func() []byte {
			return []byte("root hash")
		},
	}
	dataPool := dataRetrieverMock.NewPoolsHolderMock()

	return &nodeMockFactory.DataComponentsMock{
		BlockChain: chainHandler,
		Store:      &mockStorage.ChainStorerStub{},
		DataPool:   dataPool,
		MbProvider: &mock.MiniBlocksProviderStub{},
	}
}

func getDefaultBootstrapComponents() *mainFactoryMocks.BootstrapComponentsStub {
	return &mainFactoryMocks.BootstrapComponentsStub{
		Bootstrapper: &bootstrapMocks.EpochStartBootstrapperStub{
			TrieHolder:      &trieMock.TriesHolderStub{},
			StorageManagers: map[string]common.StorageManager{"0": &storageManager.StorageManagerStub{}},
			BootstrapCalled: nil,
		},
		BootstrapParams:            &bootstrapMocks.BootstrapParamsHandlerMock{},
		NodeRole:                   "",
		ShCoordinator:              &mock.ShardCoordinatorMock{},
		HdrIntegrityVerifier:       &mock.HeaderIntegrityVerifierStub{},
		GuardedAccountHandlerField: &guardianMocks.GuardedAccountHandlerStub{},
	}
}

func TestNode_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var n *node.Node
	require.True(t, n.IsInterfaceNil())

	n, _ = node.NewNode()
	require.False(t, n.IsInterfaceNil())
}
