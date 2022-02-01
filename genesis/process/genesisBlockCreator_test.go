package process

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math"
	"math/big"
	"testing"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/config"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/mock"
	"github.com/ElrondNetwork/elrond-go/genesis/parsing"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	factoryState "github.com/ElrondNetwork/elrond-go/state/factory"
	"github.com/ElrondNetwork/elrond-go/storage"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/testscommon/economicsmocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/trie/factory"
	"github.com/ElrondNetwork/elrond-go/update"
	updateMock "github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts/defaults"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var nodePrice = big.NewInt(5000)

// TODO improve code coverage of this package
func createMockArgument(
	t *testing.T,
	genesisFilename string,
	initialNodes genesis.InitialNodesHandler,
	entireSupply *big.Int,
) ArgsGenesisBlockCreator {

	memDBMock := mock.NewMemDbMock()
	storageManager, _ := trie.NewTrieStorageManagerWithoutPruning(memDBMock)

	trieStorageManagers := make(map[string]common.StorageManager)
	trieStorageManagers[factory.UserAccountTrie] = storageManager
	trieStorageManagers[factory.PeerAccountTrie] = storageManager

	arg := ArgsGenesisBlockCreator{
		GenesisTime:   0,
		StartEpochNum: 0,
		Core: &mock.CoreComponentsMock{
			IntMarsh:            &mock.MarshalizerMock{},
			TxMarsh:             &mock.MarshalizerMock{},
			Hash:                &hashingMocks.HasherMock{},
			UInt64ByteSliceConv: &mock.Uint64ByteSliceConverterMock{},
			AddrPubKeyConv:      mock.NewPubkeyConverterMock(32),
			Chain:               "chainID",
			MinTxVersion:        1,
		},
		Data: &mock.DataComponentsMock{
			Storage: &mock.ChainStorerStub{
				GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
					return mock.NewStorerMock()
				},
			},
			Blkc:     &mock.BlockChainStub{},
			DataPool: dataRetrieverMock.NewPoolsHolderMock(),
		},
		InitialNodesSetup: &mock.InitialNodesSetupHandlerStub{},
		TxLogsProcessor:   &mock.TxLogProcessorMock{},
		VirtualMachineConfig: config.VirtualMachineConfig{
			ArwenVersions: []config.ArwenVersionByEpoch{
				{StartEpoch: 0, Version: "*"},
			},
		},
		HardForkConfig: config.HardforkConfig{},
		SystemSCConfig: config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost: "5000000000000000000000",
				OwnerAddress:    "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				Active: config.GovernanceSystemSCConfigActive{
					ProposalCost:     "500",
					MinQuorum:        "50",
					MinPassThreshold: "50",
					MinVetoThreshold: "50",
				},
				FirstWhitelistedAddress: "3132333435363738393031323334353637383930313233343536373839303234",
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     nodePrice.Text(10),
				UnJailValue:                          "10",
				MinStepValue:                         "10",
				MinStakeValue:                        "1",
				UnBondPeriod:                         1,
				NumRoundsWithoutBleed:                1,
				MaximumPercentageToBleed:             1,
				BleedPercentagePerRound:              1,
				MaxNumberOfNodesForStake:             10,
				ActivateBLSPubKeyMessageVerification: false,
				MinUnstakeTokensValue:                "1",
			},
			DelegationManagerSystemSCConfig: config.DelegationManagerSystemSCConfig{
				MinCreationDeposit:  "100",
				MinStakeAmount:      "100",
				ConfigChangeAddress: "3132333435363738393031323334353637383930313233343536373839303234",
			},
			DelegationSystemSCConfig: config.DelegationSystemSCConfig{
				MinServiceFee: 0,
				MaxServiceFee: 100,
			},
		},
		TrieStorageManagers: trieStorageManagers,
		BlockSignKeyGen:     &mock.KeyGenMock{},
		ImportStartHandler:  &mock.ImportStartHandlerStub{},
		GenesisNodePrice:    nodePrice,
		EpochConfig: &config.EpochConfig{
			EnableEpochs: config.EnableEpochs{
				BuiltInFunctionsEnableEpoch:    0,
				SCDeployEnableEpoch:            0,
				RelayedTransactionsEnableEpoch: 0,
				PenalizedTooMuchGasEnableEpoch: 0,
			},
		},
	}

	arg.ShardCoordinator = &mock.ShardCoordinatorMock{
		NumOfShards: 2,
		SelfShardId: 0,
	}

	var err error
	arg.Accounts, err = createAccountAdapter(
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		factoryState.NewAccountCreator(),
		trieStorageManagers[factory.UserAccountTrie],
	)
	require.Nil(t, err)

	arg.ValidatorAccounts = &stateMock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return make([]byte, 0), nil
		},
		CommitCalled: func() ([]byte, error) {
			return make([]byte, 0), nil
		},
		SaveAccountCalled: func(account vmcommon.AccountHandler) error {
			return nil
		},
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return state.NewEmptyPeerAccount(), nil
		},
	}

	gasMap := arwenConfig.MakeGasMapForTests()
	defaults.FillGasMapInternal(gasMap, 1)
	arg.GasSchedule = mock.NewGasScheduleNotifierMock(gasMap)
	ted := &economicsmocks.EconomicsHandlerStub{
		GenesisTotalSupplyCalled: func() *big.Int {
			return entireSupply
		},
		MaxGasLimitPerBlockCalled: func(shardID uint32) uint64 {
			return math.MaxUint64
		},
	}
	arg.Economics = ted

	args := genesis.AccountsParserArgs{
		GenesisFilePath: genesisFilename,
		EntireSupply:    arg.Economics.GenesisTotalSupply(),
		MinterAddress:   "",
		PubkeyConverter: arg.Core.AddressPubKeyConverter(),
		KeyGenerator:    &mock.KeyGeneratorStub{},
		Hasher:          &hashingMocks.HasherMock{},
		Marshalizer:     &mock.MarshalizerMock{},
	}

	arg.AccountsParser, err = parsing.NewAccountsParser(args)
	require.Nil(t, err)

	arg.SmartContractParser, err = parsing.NewSmartContractsParser(
		"testdata/smartcontracts.json",
		arg.Core.AddressPubKeyConverter(),
		&mock.KeyGeneratorStub{},
	)
	require.Nil(t, err)

	arg.InitialNodesSetup = initialNodes

	return arg
}

func TestGenesisBlockCreator_CreateGenesisBlockAfterHardForkShouldCreateSCResultingAddresses(t *testing.T) {
	// TODO reinstate test after Arwen pointer fix
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Arwen fix")
	}

	scAddressBytes, _ := hex.DecodeString("00000000000000000500761b8c4a25d3979359223208b412285f635e71300102")
	initialNodesSetup := &mock.InitialNodesHandlerStub{
		InitialNodesInfoCalled: func() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler) {
			return map[uint32][]sharding.GenesisNodeInfoHandler{
				0: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{1}, 96),
					},
				},
				1: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{3}, 96),
					},
				},
			}, make(map[uint32][]sharding.GenesisNodeInfoHandler)
		},
		MinNumberOfNodesCalled: func() uint32 {
			return 1
		},
	}
	arg := createMockArgument(
		t,
		"testdata/genesisTest1.json",
		initialNodesSetup,
		big.NewInt(22000),
	)
	gbc, err := NewGenesisBlockCreator(arg)
	require.Nil(t, err)

	blocks, err := gbc.CreateGenesisBlocks()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(blocks))

	mapAddressesWithDeploy, err := arg.SmartContractParser.GetDeployedSCAddresses(genesis.DNSType)
	assert.Nil(t, err)
	assert.Equal(t, len(mapAddressesWithDeploy), core.MaxNumShards)

	newArgs := createMockArgument(
		t,
		"testdata/genesisTest1.json",
		initialNodesSetup,
		big.NewInt(22000),
	)
	hardForkGbc, err := NewGenesisBlockCreator(newArgs)
	assert.Nil(t, err)
	err = hardForkGbc.computeDNSAddresses(gbc.arg.EpochConfig.EnableEpochs)
	assert.Nil(t, err)

	mapAfterHardForkAddresses, err := newArgs.SmartContractParser.GetDeployedSCAddresses(genesis.DNSType)
	assert.Nil(t, err)
	assert.Equal(t, len(mapAfterHardForkAddresses), core.MaxNumShards)
	for address := range mapAddressesWithDeploy {
		_, ok := mapAfterHardForkAddresses[address]
		assert.True(t, ok)
	}
}

func TestGenesisBlockCreator_CreateGenesisBlocksJustDelegationShouldWorkAndDNS(t *testing.T) {
	// TODO reinstate test after Arwen pointer fix
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Arwen fix")
	}

	scAddressBytes, _ := hex.DecodeString("00000000000000000500761b8c4a25d3979359223208b412285f635e71300102")
	stakedAddr, _ := hex.DecodeString("b00102030405060708090001020304050607080900010203040506070809000b")
	initialNodesSetup := &mock.InitialNodesHandlerStub{
		InitialNodesInfoCalled: func() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler) {
			return map[uint32][]sharding.GenesisNodeInfoHandler{
				0: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{1}, 96),
					},
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: stakedAddr,
						PubKeyBytesValue:  bytes.Repeat([]byte{2}, 96),
					},
				},
				1: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{3}, 96),
					},
				},
			}, make(map[uint32][]sharding.GenesisNodeInfoHandler)
		},
		MinNumberOfNodesCalled: func() uint32 {
			return 1
		},
	}
	arg := createMockArgument(
		t,
		"testdata/genesisTest1.json",
		initialNodesSetup,
		big.NewInt(22000),
	)

	gbc, err := NewGenesisBlockCreator(arg)
	require.Nil(t, err)

	blocks, err := gbc.CreateGenesisBlocks()

	assert.Nil(t, err)
	assert.Equal(t, 3, len(blocks))
}

func TestGenesisBlockCreator_CreateGenesisBlocksStakingAndDelegationShouldWorkAndDNS(t *testing.T) {
	// TODO reinstate test after Arwen pointer fix
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Arwen fix")
	}

	scAddressBytes, _ := hex.DecodeString("00000000000000000500761b8c4a25d3979359223208b412285f635e71300102")
	stakedAddr, _ := hex.DecodeString("b00102030405060708090001020304050607080900010203040506070809000b")
	stakedAddr2, _ := hex.DecodeString("d00102030405060708090001020304050607080900010203040506070809000d")
	initialNodesSetup := &mock.InitialNodesHandlerStub{
		InitialNodesInfoCalled: func() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler) {
			return map[uint32][]sharding.GenesisNodeInfoHandler{
				0: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{1}, 96),
					},
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: stakedAddr,
						PubKeyBytesValue:  bytes.Repeat([]byte{2}, 96),
					},
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{3}, 96),
					},
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: stakedAddr2,
						PubKeyBytesValue:  bytes.Repeat([]byte{8}, 96),
					},
				},
				1: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{4}, 96),
					},
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{5}, 96),
					},
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: stakedAddr2,
						PubKeyBytesValue:  bytes.Repeat([]byte{6}, 96),
					},
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: stakedAddr2,
						PubKeyBytesValue:  bytes.Repeat([]byte{7}, 96),
					},
				},
			}, make(map[uint32][]sharding.GenesisNodeInfoHandler)
		},
		MinNumberOfNodesCalled: func() uint32 {
			return 1
		},
	}
	arg := createMockArgument(
		t,
		"testdata/genesisTest2.json",
		initialNodesSetup,
		big.NewInt(47000),
	)
	gbc, err := NewGenesisBlockCreator(arg)
	require.Nil(t, err)

	blocks, err := gbc.CreateGenesisBlocks()

	assert.Nil(t, err)
	assert.Equal(t, 3, len(blocks))
}

func TestGenesisBlockCreator_GetIndexingDataShouldWork(t *testing.T) {
	// TODO reinstate test after Arwen pointer fix
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Arwen fix")
	}

	scAddressBytes, _ := hex.DecodeString("00000000000000000500761b8c4a25d3979359223208b412285f635e71300102")
	stakedAddr, _ := hex.DecodeString("b00102030405060708090001020304050607080900010203040506070809000b")
	stakedAddr2, _ := hex.DecodeString("d00102030405060708090001020304050607080900010203040506070809000d")
	initialGenesisNodes := map[uint32][]sharding.GenesisNodeInfoHandler{
		0: {
			&mock.GenesisNodeInfoHandlerMock{
				AddressBytesValue: scAddressBytes,
				PubKeyBytesValue:  bytes.Repeat([]byte{1}, 96),
			},
			&mock.GenesisNodeInfoHandlerMock{
				AddressBytesValue: stakedAddr,
				PubKeyBytesValue:  bytes.Repeat([]byte{2}, 96),
			},
			&mock.GenesisNodeInfoHandlerMock{
				AddressBytesValue: scAddressBytes,
				PubKeyBytesValue:  bytes.Repeat([]byte{3}, 96),
			},
			&mock.GenesisNodeInfoHandlerMock{
				AddressBytesValue: stakedAddr2,
				PubKeyBytesValue:  bytes.Repeat([]byte{8}, 96),
			},
		},
		1: {
			&mock.GenesisNodeInfoHandlerMock{
				AddressBytesValue: scAddressBytes,
				PubKeyBytesValue:  bytes.Repeat([]byte{4}, 96),
			},
			&mock.GenesisNodeInfoHandlerMock{
				AddressBytesValue: scAddressBytes,
				PubKeyBytesValue:  bytes.Repeat([]byte{5}, 96),
			},
			&mock.GenesisNodeInfoHandlerMock{
				AddressBytesValue: stakedAddr2,
				PubKeyBytesValue:  bytes.Repeat([]byte{6}, 96),
			},
			&mock.GenesisNodeInfoHandlerMock{
				AddressBytesValue: stakedAddr2,
				PubKeyBytesValue:  bytes.Repeat([]byte{7}, 96),
			},
		},
	}
	initialNodesSetup := &mock.InitialNodesHandlerStub{
		InitialNodesInfoCalled: func() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler) {
			return initialGenesisNodes, make(map[uint32][]sharding.GenesisNodeInfoHandler)
		},
		MinNumberOfNodesCalled: func() uint32 {
			return 1
		},
	}
	arg := createMockArgument(
		t,
		"testdata/genesisTest2.json",
		initialNodesSetup,
		big.NewInt(47000),
	)
	gbc, err := NewGenesisBlockCreator(arg)
	require.Nil(t, err)

	blocks, err := gbc.CreateGenesisBlocks()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(blocks))

	indexingData := gbc.GetIndexingData()

	numDNSTypeScTxs := 256
	numDefaultTypeScTxs := 1
	numSystemSC := 4

	numInitialNodes := 0
	for k := range initialGenesisNodes {
		numInitialNodes += len(initialGenesisNodes[k])
	}

	reqNumDeployInitialScTxs := numDNSTypeScTxs + numDefaultTypeScTxs
	reqNumScrs := getRequiredNumScrsTxs(indexingData, 0)
	reqNumDelegationTxs := 4
	assert.Equal(t, reqNumDeployInitialScTxs, len(indexingData[0].DeployInitialScTxs))
	assert.Equal(t, 0, len(indexingData[0].DeploySystemScTxs))
	assert.Equal(t, reqNumDelegationTxs, len(indexingData[0].DelegationTxs))
	assert.Equal(t, 0, len(indexingData[0].StakingTxs))
	assert.Equal(t, reqNumScrs, len(indexingData[0].ScrsTxs))

	reqNumDeployInitialScTxs = numDNSTypeScTxs
	reqNumScrs = getRequiredNumScrsTxs(indexingData, 1)
	assert.Equal(t, reqNumDeployInitialScTxs, len(indexingData[1].DeployInitialScTxs))
	assert.Equal(t, 0, len(indexingData[1].DeploySystemScTxs))
	assert.Equal(t, 0, len(indexingData[1].DelegationTxs))
	assert.Equal(t, 0, len(indexingData[1].StakingTxs))
	assert.Equal(t, reqNumScrs, len(indexingData[1].ScrsTxs))

	reqNumScrs = getRequiredNumScrsTxs(indexingData, core.MetachainShardId)
	assert.Equal(t, 0, len(indexingData[core.MetachainShardId].DeployInitialScTxs))
	assert.Equal(t, numSystemSC, len(indexingData[core.MetachainShardId].DeploySystemScTxs))
	assert.Equal(t, 0, len(indexingData[core.MetachainShardId].DelegationTxs))
	assert.Equal(t, numInitialNodes, len(indexingData[core.MetachainShardId].StakingTxs))
	assert.Equal(t, reqNumScrs, len(indexingData[core.MetachainShardId].ScrsTxs))
}

func getRequiredNumScrsTxs(idata map[uint32]*genesis.IndexingData, shardId uint32) int {
	n := 2 * (len(idata[shardId].DeployInitialScTxs) + len(idata[shardId].DeploySystemScTxs) + len(idata[shardId].DelegationTxs))
	n += 3 * len(idata[shardId].StakingTxs)
	return n
}

func TestCreateArgsGenesisBlockCreator_ShouldErrWhenGetNewArgForShardFails(t *testing.T) {
	scAddressBytes, _ := hex.DecodeString("00000000000000000500761b8c4a25d3979359223208b412285f635e71300102")
	shardIDs := []uint32{0, 1}
	mapArgsGenesisBlockCreator := make(map[uint32]ArgsGenesisBlockCreator)
	initialNodesSetup := createDummyNodesHandler(scAddressBytes)
	arg := createMockArgument(
		t,
		"testdata/genesisTest1.json",
		initialNodesSetup,
		big.NewInt(22000),
	)

	arg.ShardCoordinator = &mock.ShardCoordinatorMock{SelfShardId: 1}
	arg.TrieStorageManagers = make(map[string]common.StorageManager)
	gbc, err := NewGenesisBlockCreator(arg)
	require.Nil(t, err)

	err = gbc.createArgsGenesisBlockCreator(shardIDs, mapArgsGenesisBlockCreator)
	assert.True(t, errors.Is(err, trie.ErrNilTrieStorage))
}

func TestCreateArgsGenesisBlockCreator_ShouldWork(t *testing.T) {
	shardIDs := []uint32{0, 1}
	mapArgsGenesisBlockCreator := make(map[uint32]ArgsGenesisBlockCreator)
	scAddressBytes, _ := hex.DecodeString("00000000000000000500761b8c4a25d3979359223208b412285f635e71300102")
	initialNodesSetup := &mock.InitialNodesHandlerStub{
		InitialNodesInfoCalled: func() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler) {
			return map[uint32][]sharding.GenesisNodeInfoHandler{
				0: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{1}, 96),
					},
				},
				1: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{3}, 96),
					},
				},
			}, make(map[uint32][]sharding.GenesisNodeInfoHandler)
		},
		MinNumberOfNodesCalled: func() uint32 {
			return 1
		},
	}
	arg := createMockArgument(
		t,
		"testdata/genesisTest1.json",
		initialNodesSetup,
		big.NewInt(22000),
	)
	gbc, err := NewGenesisBlockCreator(arg)
	require.Nil(t, err)

	err = gbc.createArgsGenesisBlockCreator(shardIDs, mapArgsGenesisBlockCreator)
	assert.Nil(t, err)
	require.Equal(t, 2, len(mapArgsGenesisBlockCreator))
	assert.Equal(t, uint32(0), mapArgsGenesisBlockCreator[0].ShardCoordinator.SelfId())
	assert.Equal(t, uint32(1), mapArgsGenesisBlockCreator[1].ShardCoordinator.SelfId())
}

func TestCreateHardForkBlockProcessors_ShouldWork(t *testing.T) {
	selfShardID := uint32(0)
	shardIDs := []uint32{1, core.MetachainShardId}
	mapArgsGenesisBlockCreator := make(map[uint32]ArgsGenesisBlockCreator)
	mapHardForkBlockProcessor := make(map[uint32]update.HardForkBlockProcessor)
	scAddressBytes, _ := hex.DecodeString("00000000000000000500761b8c4a25d3979359223208b412285f635e71300102")
	initialNodesSetup := &mock.InitialNodesHandlerStub{
		InitialNodesInfoCalled: func() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler) {
			return map[uint32][]sharding.GenesisNodeInfoHandler{
				0: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{1}, 96),
					},
				},
				1: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{3}, 96),
					},
				},
			}, make(map[uint32][]sharding.GenesisNodeInfoHandler)
		},
		MinNumberOfNodesCalled: func() uint32 {
			return 1
		},
	}
	arg := createMockArgument(
		t,
		"testdata/genesisTest1.json",
		initialNodesSetup,
		big.NewInt(22000),
	)
	arg.importHandler = &updateMock.ImportHandlerStub{
		GetAccountsDBForShardCalled: func(shardID uint32) state.AccountsAdapter {
			return &stateMock.AccountsStub{}
		},
	}
	gbc, err := NewGenesisBlockCreator(arg)
	require.Nil(t, err)

	_ = gbc.createArgsGenesisBlockCreator(shardIDs, mapArgsGenesisBlockCreator)

	err = createHardForkBlockProcessors(selfShardID, shardIDs, mapArgsGenesisBlockCreator, mapHardForkBlockProcessor)
	assert.Nil(t, err)
	require.Equal(t, 2, len(mapHardForkBlockProcessor))
}

func createDummyNodesHandler(scAddressBytes []byte) genesis.InitialNodesHandler {
	return &mock.InitialNodesHandlerStub{
		InitialNodesInfoCalled: func() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler) {
			return map[uint32][]sharding.GenesisNodeInfoHandler{
				0: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{1}, 96),
					},
				},
				1: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{3}, 96),
					},
				},
			}, make(map[uint32][]sharding.GenesisNodeInfoHandler)
		},
		MinNumberOfNodesCalled: func() uint32 {
			return 1
		},
	}
}
