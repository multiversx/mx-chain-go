//go:build !race

package process

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	factoryRunType "github.com/multiversx/mx-chain-go/factory/runType"
	"github.com/multiversx/mx-chain-go/genesis/mock"
	nodeMock "github.com/multiversx/mx-chain-go/node/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	stateAcc "github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

var (
	esdtKeyPrefix        = []byte(core.ProtectedKeyPrefix + core.ESDTKeyIdentifier)
	sovereignNativeToken = "WEGLD-bd4d79"
)

func createGenesisBlockCreator(t *testing.T) *genesisBlockCreator {
	arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	gbc, _ := NewGenesisBlockCreator(arg)
	return gbc
}

func createSovereignGenesisBlockCreator(t *testing.T) (ArgsGenesisBlockCreator, *sovereignGenesisBlockCreator) {
	arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	arg.ShardCoordinator = sharding.NewSovereignShardCoordinator(core.SovereignChainShardId)
	arg.DNSV2Addresses = []string{"00000000000000000500761b8c4a25d3979359223208b412285f635e71300102"}
	arg.Config.SovereignConfig.GenesisConfig.NativeESDT = sovereignNativeToken

	sovRunTypeComps := createSovRunTypeComps(t)
	arg.RunTypeComponents = sovRunTypeComps

	trieStorageManagers := createTrieStorageManagers()
	arg.Accounts, _ = createAccountAdapter(
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		sovRunTypeComps.AccountsCreator(),
		trieStorageManagers[dataRetriever.UserAccountsUnit.String()],
		&testscommon.PubkeyConverterMock{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	)

	gbc, _ := NewGenesisBlockCreator(arg)
	sgbc, _ := NewSovereignGenesisBlockCreator(gbc)
	return arg, sgbc
}

func createSovRunTypeComps(t *testing.T) runTypeComponentsHandler {
	runTypeFactory, err := factoryRunType.NewRunTypeComponentsFactory(&factory.CoreComponentsHolderMock{
		HasherCalled: func() hashing.Hasher {
			return &hashingMocks.HasherMock{}
		},
		InternalMarshalizerCalled: func() marshal.Marshalizer {
			return &mock.MarshalizerMock{}
		},
		EnableEpochsHandlerCalled: func() common.EnableEpochsHandler {
			return &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
		},
	})
	require.Nil(t, err)

	sovRunTypeFactory, err := factoryRunType.NewSovereignRunTypeComponentsFactory(runTypeFactory,
		config.SovereignConfig{
			GenesisConfig: config.GenesisConfig{
				NativeESDT: sovereignNativeToken,
			},
		},
	)
	require.Nil(t, err)
	sovRunTypeComp, err := factoryRunType.NewManagedRunTypeComponents(sovRunTypeFactory)

	err = sovRunTypeComp.Create()
	require.Nil(t, err)

	return sovRunTypeComp
}

func requireTokenExists(
	t *testing.T,
	account vmcommon.AccountHandler,
	tokenName []byte,
	expectedValue *big.Int,
	marshaller marshal.Marshalizer,
) {
	marshaledData := getAccTokenMarshalledData(t, tokenName, account)
	esdtData := &esdt.ESDigitalToken{}
	err := marshaller.Unmarshal(esdtData, marshaledData)
	require.Nil(t, err)
	require.Equal(t, &esdt.ESDigitalToken{
		Type:  uint32(core.Fungible),
		Value: expectedValue,
	}, esdtData)
}

func getAccTokenMarshalledData(t *testing.T, tokenName []byte, account vmcommon.AccountHandler) []byte {
	tokenId := append(esdtKeyPrefix, tokenName...)
	esdtNFTTokenKey := computeESDTNFTTokenKey(tokenId, 0)

	marshaledData, _, err := account.(vmcommon.UserAccountHandler).AccountDataHandler().RetrieveValue(esdtNFTTokenKey)
	require.Nil(t, err)

	return marshaledData
}

func computeESDTNFTTokenKey(esdtTokenKey []byte, nonce uint64) []byte {
	return append(esdtTokenKey, big.NewInt(0).SetUint64(nonce).Bytes()...)
}

func getAccount(t *testing.T, accountsDb stateAcc.AccountsAdapter, address string) vmcommon.AccountHandler {
	addressBytes, err := hex.DecodeString(address)
	require.Nil(t, err)

	account, err := accountsDb.LoadAccount(addressBytes)
	require.Nil(t, err)

	return account
}

func TestNewSovereignGenesisBlockCreator(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		gbc := createGenesisBlockCreator(t)
		sgbc, err := NewSovereignGenesisBlockCreator(gbc)
		require.Nil(t, err)
		require.NotNil(t, sgbc)
	})

	t.Run("nil genesis block creator, should return error", func(t *testing.T) {
		t.Parallel()

		sgbc, err := NewSovereignGenesisBlockCreator(nil)
		require.Equal(t, errNilGenesisBlockCreator, err)
		require.Nil(t, sgbc)
	})
}

func TestSovereignGenesisBlockCreator_CreateGenesisBlocksEmptyBlocks(t *testing.T) {
	arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	arg.StartEpochNum = 1
	gbc, _ := NewGenesisBlockCreator(arg)
	sgbc, _ := NewSovereignGenesisBlockCreator(gbc)

	blocks, err := sgbc.CreateGenesisBlocks()
	require.Nil(t, err)
	require.Equal(t, map[uint32]data.HeaderHandler{
		core.SovereignChainShardId: &block.SovereignChainHeader{
			Header: &block.Header{
				ShardID: core.SovereignChainShardId,
			},
		},
	}, blocks)
}

func TestSovereignGenesisBlockCreator_CreateGenesisBaseProcess(t *testing.T) {
	args, sgbc := createSovereignGenesisBlockCreator(t)

	blocks, err := sgbc.CreateGenesisBlocks()
	require.Nil(t, err)
	require.Len(t, blocks, 1)
	require.Contains(t, blocks, core.SovereignChainShardId)

	indexingData := sgbc.GetIndexingData()
	require.Len(t, indexingData, 1)

	numDNSTypeScTxs := 2 * 256 // there are 2 contracts in testdata/smartcontracts.json
	numDefaultTypeScTxs := 1
	reqNumDeployInitialScTxs := numDNSTypeScTxs + numDefaultTypeScTxs

	sovereignIdxData := indexingData[core.SovereignChainShardId]
	require.Len(t, sovereignIdxData.DeployInitialScTxs, reqNumDeployInitialScTxs)
	require.Len(t, sovereignIdxData.DeploySystemScTxs, 4)
	require.Len(t, sovereignIdxData.DelegationTxs, 3)
	require.Len(t, sovereignIdxData.StakingTxs, 0)
	require.Greater(t, len(sovereignIdxData.ScrsTxs), 3)

	addr1 := "a00102030405060708090001020304050607080900010203040506070809000a"
	addr2 := "b00102030405060708090001020304050607080900010203040506070809000b"
	addr3 := "c00102030405060708090001020304050607080900010203040506070809000c"

	accountsDB := args.Accounts
	acc1 := getAccount(t, accountsDB, addr1)
	acc2 := getAccount(t, accountsDB, addr2)
	acc3 := getAccount(t, accountsDB, addr3)

	balance1 := big.NewInt(5000)
	balance2 := big.NewInt(2000)
	balance3 := big.NewInt(0)
	requireTokenExists(t, acc1, []byte(sovereignNativeToken), balance1, args.Core.InternalMarshalizer())
	requireTokenExists(t, acc2, []byte(sovereignNativeToken), balance2, args.Core.InternalMarshalizer())
	requireTokenExists(t, acc3, []byte(sovereignNativeToken), balance3, args.Core.InternalMarshalizer())
}

func TestSovereignGenesisBlockCreator_initGenesisAccountsErrorCases(t *testing.T) {
	t.Parallel()

	localErr := errors.New("local error")
	t.Run("cannot load system account, should return error", func(t *testing.T) {
		_, sgbc := createSovereignGenesisBlockCreator(t)
		sgbc.arg.Accounts = &state.AccountsStub{
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
				if bytes.Equal(container, core.SystemAccountAddress) {
					return nil, localErr
				}
				return nil, nil
			},
		}

		err := sgbc.initGenesisAccounts()
		require.Equal(t, localErr, err)
	})
	t.Run("cannot save system account, should return error", func(t *testing.T) {
		_, sgbc := createSovereignGenesisBlockCreator(t)
		sgbc.arg.Accounts = &state.AccountsStub{
			SaveAccountCalled: func(account vmcommon.AccountHandler) error {
				if bytes.Equal(account.AddressBytes(), core.SystemAccountAddress) {
					return localErr
				}
				return nil
			},
		}

		err := sgbc.initGenesisAccounts()
		require.Equal(t, localErr, err)
	})
	t.Run("cannot load esdt sc account, should return error", func(t *testing.T) {
		_, sgbc := createSovereignGenesisBlockCreator(t)
		sgbc.arg.Accounts = &state.AccountsStub{
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
				if bytes.Equal(container, core.ESDTSCAddress) {
					return nil, localErr
				}
				return nil, nil
			},
		}

		err := sgbc.initGenesisAccounts()
		require.Equal(t, localErr, err)
	})
	t.Run("cannot save esdt sc account, should return error", func(t *testing.T) {
		_, sgbc := createSovereignGenesisBlockCreator(t)
		sgbc.arg.Accounts = &state.AccountsStub{
			SaveAccountCalled: func(account vmcommon.AccountHandler) error {
				if bytes.Equal(account.AddressBytes(), core.ESDTSCAddress) {
					return localErr
				}
				return nil
			},
		}

		err := sgbc.initGenesisAccounts()
		require.Equal(t, localErr, err)
	})
}

func TestSovereignGenesisBlockCreator_setSovereignStakedData(t *testing.T) {
	t.Parallel()

	args := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))

	acc := &state.AccountWrapMock{
		Balance: big.NewInt(1),
	}
	acc.IncreaseNonce(4)
	args.Accounts = &state.AccountsStub{
		LoadAccountCalled: func(addr []byte) (vmcommon.AccountHandler, error) {
			return acc, nil
		},
	}
	initialNode := &sharding.InitialNode{
		Address: "addr",
	}
	nodesSpliter := &mock.NodesListSplitterStub{
		GetAllNodesCalled: func() []nodesCoordinator.GenesisNodeInfoHandler {
			return []nodesCoordinator.GenesisNodeInfoHandler{initialNode}
		},
	}
	expectedTx := &transaction.Transaction{
		Nonce:     acc.GetNonce(),
		Value:     new(big.Int).Set(args.GenesisNodePrice),
		RcvAddr:   vm.ValidatorSCAddress,
		SndAddr:   initialNode.AddressBytes(),
		GasPrice:  0,
		GasLimit:  math.MaxUint64,
		Data:      []byte("stake@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString(initialNode.PubKeyBytes()) + "@" + hex.EncodeToString([]byte("genesis"))),
		Signature: nil,
	}
	processors := &genesisProcessors{
		txProcessor: &testscommon.TxProcessorStub{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				require.Equal(t, expectedTx, transaction)

				return vmcommon.Ok, nil
			},
		},
		queryService: &nodeMock.SCQueryServiceStub{
			ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, common.BlockInfo, error) {
				require.Equal(t, &process.SCQuery{
					ScAddress: vm.StakingSCAddress,
					FuncName:  "isStaked",
					Arguments: [][]byte{initialNode.PubKeyBytes()}}, query)

				return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, holders.NewBlockInfo(nil, 0, nil), nil
			},
		},
	}

	txs, err := setSovereignStakedData(args, processors, nodesSpliter)
	require.Nil(t, err)
	require.Equal(t, []data.TransactionHandler{expectedTx}, txs)
}

func TestSovereignGenesisBlockCreator_InitSystemAccountCalled(t *testing.T) {
	t.Parallel()

	arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	arg.ShardCoordinator = sharding.NewSovereignShardCoordinator(core.SovereignChainShardId)
	arg.DNSV2Addresses = []string{"00000000000000000500761b8c4a25d3979359223208b412285f635e71300102"}

	gbc, _ := NewGenesisBlockCreator(arg)
	sgbc, _ := NewSovereignGenesisBlockCreator(gbc)
	require.NotNil(t, sgbc)

	acc, err := arg.Accounts.GetExistingAccount(core.SystemAccountAddress)
	require.Nil(t, acc)
	require.Equal(t, err, stateAcc.ErrAccNotFound)

	_, err = sgbc.CreateGenesisBlocks()
	require.Nil(t, err)

	acc, err = arg.Accounts.GetExistingAccount(core.SystemAccountAddress)
	require.NotNil(t, acc)
	require.Nil(t, err)
}
