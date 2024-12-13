//go:build !race

package process

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/genesis/mock"
	nodeMock "github.com/multiversx/mx-chain-go/node/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	processMock "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	stateAcc "github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	storageCommon "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
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
	arg := createSovereignMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	arg.ShardCoordinator = sharding.NewSovereignShardCoordinator()
	arg.DNSV2Addresses = []string{"00000000000000000500761b8c4a25d3979359223208b412285f635e71300102"}

	trieStorageManagers := createTrieStorageManagers()
	arg.Accounts, _ = createAccountAdapter(
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		arg.RunTypeComponents.AccountsCreator(),
		trieStorageManagers[dataRetriever.UserAccountsUnit.String()],
		&testscommon.PubkeyConverterMock{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	)

	gbc, _ := NewGenesisBlockCreator(arg)
	sgbc, _ := NewSovereignGenesisBlockCreator(gbc)
	return arg, sgbc
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
	arg := createSovereignMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	arg.StartEpochNum = 1
	gbc, _ := NewGenesisBlockCreator(arg)
	sgbc, _ := NewSovereignGenesisBlockCreator(gbc)

	blocks, err := sgbc.CreateGenesisBlocks()
	require.Nil(t, err)
	require.Equal(t, map[uint32]data.HeaderHandler{
		core.SovereignChainShardId: &block.SovereignChainHeader{
			Header: &block.Header{
				ShardID:         core.SovereignChainShardId,
				SoftwareVersion: process.SovereignHeaderVersion,
			},
		},
	}, blocks)
}

func TestSovereignGenesisBlockCreator_CreateGenesisBlocks(t *testing.T) {
	args, sgbc := createSovereignGenesisBlockCreator(t)

	blocks, err := sgbc.CreateGenesisBlocks()
	require.Nil(t, err)
	require.Len(t, blocks, 1)

	sovBlock := blocks[core.SovereignChainShardId].(*block.SovereignChainHeader)
	require.NotNil(t, sovBlock)

	require.Equal(t, sgbc.arg.Core.Hasher().Compute(sgbc.arg.GenesisString), sovBlock.GetPrevHash())
	require.Equal(t, process.SovereignHeaderVersion, sovBlock.GetSoftwareVersion())

	sovBlock.Header = nil
	require.Equal(t, &block.SovereignChainHeader{
		AccumulatedFeesInEpoch: big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
		ValidatorStatsRootHash: make([]byte, 0, 32),
		IsStartOfEpoch:         true,
		EpochStart: block.EpochStartSovereign{
			Economics: block.Economics{
				TotalSupply:       big.NewInt(0).Set(args.Economics.GenesisTotalSupply()),
				TotalToDistribute: big.NewInt(0),
				TotalNewlyMinted:  big.NewInt(0),
				RewardsPerBlock:   big.NewInt(0),
				NodePrice:         big.NewInt(0).Set(args.GenesisNodePrice),
			},
		},
	}, sovBlock)
}

func TestSovereignGenesisBlockCreator_CreateGenesisBaseProcess(t *testing.T) {
	args, sgbc := createSovereignGenesisBlockCreator(t)

	blockStorer := genericMocks.NewStorerMock()
	triggerStorer := genericMocks.NewStorerMock()
	sgbc.arg.Data = &mock.DataComponentsMock{
		Storage: &storageCommon.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				switch unitType {
				case dataRetriever.BlockHeaderUnit:
					return blockStorer, nil
				case dataRetriever.BootstrapUnit:
					return triggerStorer, nil
				}

				return genericMocks.NewStorerMock(), nil
			},
		},
		Blkc:     &testscommon.ChainHandlerStub{},
		DataPool: dataRetrieverMock.NewPoolsHolderMock(),
	}

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

	epochStartID := []byte(core.EpochStartIdentifier(0))
	storedData1, err := blockStorer.Get(epochStartID)
	require.Nil(t, err)
	require.NotNil(t, storedData1)

	storedData2, err := triggerStorer.Get(epochStartID)
	require.Nil(t, err)
	require.NotNil(t, storedData2)

	require.Equal(t, storedData1, storedData2)
	savedSovHdr := &block.SovereignChainHeader{}
	err = args.Core.InternalMarshalizer().Unmarshal(savedSovHdr, storedData1)
	require.Nil(t, err)
	require.Equal(t, process.SovereignHeaderVersion, savedSovHdr.GetSoftwareVersion())
}

func TestSovereignGenesisBlockCreator_initGenesisAccounts(t *testing.T) {
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
				return &mock.UserAccountMock{}, nil
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
	t.Run("should work", func(t *testing.T) {
		_, sgbc := createSovereignGenesisBlockCreator(t)
		contractsToUpdate := map[string]struct{}{
			string(vm.StakingSCAddress):           {},
			string(vm.ValidatorSCAddress):         {},
			string(vm.GovernanceSCAddress):        {},
			string(vm.ESDTSCAddress):              {},
			string(vm.DelegationManagerSCAddress): {},
			string(vm.FirstDelegationSCAddress):   {},
		}

		expectedCodeMetaData := &vmcommon.CodeMetadata{
			Readable: true,
		}

		sgbc.arg.Accounts = &state.AccountsStub{
			SaveAccountCalled: func(account vmcommon.AccountHandler) error {
				if bytes.Equal(account.AddressBytes(), core.SystemAccountAddress) {
					return nil
				}

				addrStr := string(account.AddressBytes())
				_, found := contractsToUpdate[addrStr]
				require.True(t, found)

				userAcc := account.(*state.AccountWrapMock)
				require.NotEmpty(t, userAcc.GetCode())
				require.NotEmpty(t, userAcc.GetOwnerAddress())
				require.Equal(t, expectedCodeMetaData.ToBytes(), userAcc.GetCodeMetadata())

				delete(contractsToUpdate, addrStr)

				return nil
			},
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
				return &state.AccountWrapMock{
					Address: container,
				}, nil
			},
		}

		err := sgbc.initGenesisAccounts()
		require.Nil(t, err)
		require.Empty(t, contractsToUpdate)
	})
}

func TestSovereignGenesisBlockCreator_setSovereignStakedData(t *testing.T) {
	t.Parallel()

	args := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))

	ownerAddr := []byte("owner")
	blsKey := []byte("blsKey")
	acc := &state.AccountWrapMock{
		Balance: big.NewInt(1),
		Address: ownerAddr,
	}
	acc.IncreaseNonce(4)
	args.Accounts = &state.AccountsStub{
		LoadAccountCalled: func(addr []byte) (vmcommon.AccountHandler, error) {
			return acc, nil
		},
	}
	initialNode := &mock.GenesisNodeInfoHandlerMock{
		PubKeyBytesValue: blsKey,
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

	wasSetOwnersCalled := false
	vmExecutionHandler := &processMock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			require.Equal(t, &vmcommon.ContractCallInput{
				VMInput: vmcommon.VMInput{
					CallerAddr: vm.EndOfEpochAddress,
					CallValue:  big.NewInt(0),
					Arguments:  [][]byte{blsKey, ownerAddr},
				},
				RecipientAddr: vm.StakingSCAddress,
				Function:      "setOwnersOnAddresses",
			}, input)

			wasSetOwnersCalled = true
			return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, nil
		},
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
		vmContainer: &processMock.VMContainerMock{
			GetCalled: func(key []byte) (vmcommon.VMExecutionHandler, error) {
				require.Equal(t, factory.SystemVirtualMachine, key)
				return vmExecutionHandler, nil
			},
		},
	}

	txs, err := setSovereignStakedData(args, processors, nodesSpliter)
	require.Nil(t, err)
	require.Equal(t, []data.TransactionHandler{expectedTx}, txs)
	require.True(t, wasSetOwnersCalled)
}

func TestSovereignGenesisBlockCreator_InitSystemAccountCalled(t *testing.T) {
	t.Parallel()

	arg, sgbc := createSovereignGenesisBlockCreator(t)
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

func TestSovereignGenesisBlockCreator_InitSystemSCs(t *testing.T) {
	t.Parallel()

	arg, sgbc := createSovereignGenesisBlockCreator(t)
	require.NotNil(t, sgbc)

	_, err := sgbc.CreateGenesisBlocks()
	require.Nil(t, err)

	accountsDB := arg.Accounts

	// Check delegation manager is initialized
	val := retrieveAccValue(t, accountsDB, vm.DelegationManagerSCAddress, []byte("delegationContracts"))
	delegationList := &systemSmartContracts.DelegationContractList{}
	err = arg.Core.InternalMarshalizer().Unmarshal(delegationList, val)
	require.Nil(t, err)
	require.Equal(t, [][]byte{vm.FirstDelegationSCAddress}, delegationList.Addresses)

	// Check governance manager is initialized
	val = retrieveAccValue(t, accountsDB, vm.GovernanceSCAddress, []byte("owner"))
	require.Equal(t, vm.GovernanceSCAddress, val)
}

func retrieveAccValue(t *testing.T, accountsDB stateAcc.AccountsAdapter, address []byte, key []byte) []byte {
	acc, err := accountsDB.GetExistingAccount(address)
	require.Nil(t, err)

	val, _, err := acc.(data.UserAccountHandler).RetrieveValue(key)
	require.Nil(t, err)
	require.NotEmpty(t, val)

	return val
}
