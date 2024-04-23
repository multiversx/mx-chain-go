package metachain

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	"github.com/multiversx/mx-chain-go/dataRetriever/dataPool"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/genesis/process/disabled"
	"github.com/multiversx/mx-chain-go/process"
	vmFactory "github.com/multiversx/mx-chain-go/process/factory"
	metaProcess "github.com/multiversx/mx-chain-go/process/factory/metachain"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	disabledState "github.com/multiversx/mx-chain-go/state/disabled"
	"github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/evictionWaitingList"
	"github.com/multiversx/mx-chain-go/storage"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/genesisMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/stakingcommon"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageMock "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts/defaults"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	vmcommonBuiltInFunctions "github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	wasmConfig "github.com/multiversx/mx-chain-vm-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testKeyPair struct {
	walletKey    []byte
	validatorKey []byte
}

func createPhysicalUnit(t *testing.T) (storage.Storer, string) {
	cacheConfig := storageunit.CacheConfig{
		Name:                 "test",
		Type:                 "SizeLRU",
		SizeInBytes:          314572800,
		SizeInBytesPerSender: 0,
		Capacity:             500000,
		SizePerSender:        0,
		Shards:               0,
	}
	dir := t.TempDir()

	dbConfig := config.DBConfig{
		FilePath:          dir,
		Type:              "LvlDBSerial",
		BatchDelaySeconds: 2,
		MaxBatchSize:      45000,
		MaxOpenFiles:      10,
	}

	dbConfigHandler := storageFactory.NewDBConfigHandler(dbConfig)
	persisterFactory, err := storageFactory.NewPersisterFactory(dbConfigHandler)
	assert.Nil(t, err)

	cache, _ := storageunit.NewCache(cacheConfig)
	persist, _ := storageunit.NewDB(persisterFactory, dir)
	unit, _ := storageunit.NewStorageUnit(cache, persist)

	return unit, dir
}

func createMockArgsForSystemSCProcessor() ArgsNewEpochStartSystemSCProcessing {
	return ArgsNewEpochStartSystemSCProcessing{
		SystemVM:                     &mock.VMExecutionHandlerStub{},
		UserAccountsDB:               &stateMock.AccountsStub{},
		PeerAccountsDB:               &stateMock.AccountsStub{},
		Marshalizer:                  &marshallerMock.MarshalizerStub{},
		StartRating:                  0,
		ValidatorInfoCreator:         &testscommon.ValidatorStatisticsProcessorStub{},
		ChanceComputer:               &mock.ChanceComputerStub{},
		ShardCoordinator:             &testscommon.ShardsCoordinatorMock{},
		EndOfEpochCallerAddress:      vm.EndOfEpochAddress,
		StakingSCAddress:             vm.StakingSCAddress,
		ESDTOwnerAddressBytes:        vm.ESDTSCAddress,
		GenesisNodesConfig:           &genesisMocks.NodesSetupStub{},
		EpochNotifier:                &epochNotifier.EpochNotifierStub{},
		NodesConfigProvider:          &shardingMocks.NodesCoordinatorStub{},
		StakingDataProvider:          &stakingcommon.StakingDataProviderStub{},
		AuctionListSelector:          &stakingcommon.AuctionListSelectorStub{},
		MaxNodesChangeConfigProvider: &testscommon.MaxNodesChangeConfigProviderStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
}

func TestNewSystemSCProcessor(t *testing.T) {
	t.Parallel()

	cfg := config.EnableEpochs{
		StakingV2EnableEpoch: 100,
	}
	args, _ := createFullArgumentsForSystemSCProcessing(cfg, testscommon.CreateMemUnit())
	args.Marshalizer = nil
	checkConstructorWithNilArg(t, args, epochStart.ErrNilMarshalizer)

	args, _ = createFullArgumentsForSystemSCProcessing(cfg, testscommon.CreateMemUnit())
	args.PeerAccountsDB = nil
	checkConstructorWithNilArg(t, args, epochStart.ErrNilAccountsDB)

	args, _ = createFullArgumentsForSystemSCProcessing(cfg, testscommon.CreateMemUnit())
	args.SystemVM = nil
	checkConstructorWithNilArg(t, args, epochStart.ErrNilSystemVM)

	args, _ = createFullArgumentsForSystemSCProcessing(cfg, testscommon.CreateMemUnit())
	args.UserAccountsDB = nil
	checkConstructorWithNilArg(t, args, epochStart.ErrNilAccountsDB)

	args, _ = createFullArgumentsForSystemSCProcessing(cfg, testscommon.CreateMemUnit())
	args.ValidatorInfoCreator = nil
	checkConstructorWithNilArg(t, args, epochStart.ErrNilValidatorInfoProcessor)

	args, _ = createFullArgumentsForSystemSCProcessing(cfg, testscommon.CreateMemUnit())
	args.EndOfEpochCallerAddress = nil
	checkConstructorWithNilArg(t, args, epochStart.ErrNilEndOfEpochCallerAddress)

	args, _ = createFullArgumentsForSystemSCProcessing(cfg, testscommon.CreateMemUnit())
	args.StakingSCAddress = nil
	checkConstructorWithNilArg(t, args, epochStart.ErrNilStakingSCAddress)

	args, _ = createFullArgumentsForSystemSCProcessing(cfg, testscommon.CreateMemUnit())
	args.ValidatorInfoCreator = nil
	checkConstructorWithNilArg(t, args, epochStart.ErrNilValidatorInfoProcessor)

	args, _ = createFullArgumentsForSystemSCProcessing(cfg, testscommon.CreateMemUnit())
	args.ChanceComputer = nil
	checkConstructorWithNilArg(t, args, epochStart.ErrNilChanceComputer)

	args, _ = createFullArgumentsForSystemSCProcessing(cfg, testscommon.CreateMemUnit())
	args.GenesisNodesConfig = nil
	checkConstructorWithNilArg(t, args, epochStart.ErrNilGenesisNodesConfig)

	args, _ = createFullArgumentsForSystemSCProcessing(cfg, testscommon.CreateMemUnit())
	args.NodesConfigProvider = nil
	checkConstructorWithNilArg(t, args, epochStart.ErrNilNodesConfigProvider)

	args, _ = createFullArgumentsForSystemSCProcessing(cfg, testscommon.CreateMemUnit())
	args.StakingDataProvider = nil
	checkConstructorWithNilArg(t, args, epochStart.ErrNilStakingDataProvider)

	args, _ = createFullArgumentsForSystemSCProcessing(cfg, testscommon.CreateMemUnit())
	args.EpochNotifier = nil
	checkConstructorWithNilArg(t, args, epochStart.ErrNilEpochStartNotifier)

	args, _ = createFullArgumentsForSystemSCProcessing(cfg, testscommon.CreateMemUnit())
	args.EnableEpochsHandler = nil
	checkConstructorWithNilArg(t, args, epochStart.ErrNilEnableEpochsHandler)

	args, _ = createFullArgumentsForSystemSCProcessing(cfg, testscommon.CreateMemUnit())
	args.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStubWithNoFlagsDefined()
	_, err := NewSystemSCProcessor(args)
	require.True(t, errors.Is(err, core.ErrInvalidEnableEpochsHandler))
}

func checkConstructorWithNilArg(t *testing.T, args ArgsNewEpochStartSystemSCProcessing, expectedErr error) {
	_, err := NewSystemSCProcessor(args)
	require.Equal(t, expectedErr, err)
}

func TestSystemSCProcessor_ProcessSystemSmartContract(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{
		StakingV2EnableEpoch: 1000,
	}, testscommon.CreateMemUnit())
	args.ChanceComputer = &mock.ChanceComputerStub{
		GetChanceCalled: func(rating uint32) uint32 {
			if rating == 0 {
				return 10
			}
			return rating
		},
	}
	s, _ := NewSystemSCProcessor(args)

	prepareStakingContractWithData(
		args.UserAccountsDB,
		[]byte("jailedPubKey0"),
		[]byte("waitingPubKey"),
		args.Marshalizer,
		[]byte("rewardAddress"),
		[]byte("rewardAddress"),
	)
	jailedAcc, _ := args.PeerAccountsDB.LoadAccount([]byte("jailedPubKey0"))
	_ = args.PeerAccountsDB.SaveAccount(jailedAcc)

	validatorsInfo := state.NewShardValidatorsInfoMap()
	vInfo := &state.ValidatorInfo{
		PublicKey:       []byte("jailedPubKey0"),
		ShardId:         0,
		List:            string(common.JailedList),
		TempRating:      1,
		RewardAddress:   []byte("address"),
		AccumulatedFees: big.NewInt(0),
	}
	_ = validatorsInfo.Add(vInfo)
	err := s.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
	require.Nil(t, err)

	require.Len(t, validatorsInfo.GetShardValidatorsInfoMap()[0], 1)
	newValidatorInfo := validatorsInfo.GetShardValidatorsInfoMap()[0][0]
	require.Equal(t, newValidatorInfo.GetList(), string(common.NewList))
}

func TestSystemSCProcessor_JailedNodesShouldNotBeSwappedAllAtOnce(t *testing.T) {
	t.Parallel()

	testSystemSCProcessorJailedNodesShouldNotBeSwappedAllAtOnce(t, 0)
	testSystemSCProcessorJailedNodesShouldNotBeSwappedAllAtOnce(t, 1000)
}

func testSystemSCProcessorJailedNodesShouldNotBeSwappedAllAtOnce(t *testing.T, saveJailedAlwaysEnableEpoch uint32) {
	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{
		StakingV2EnableEpoch:        10000,
		SaveJailedAlwaysEnableEpoch: saveJailedAlwaysEnableEpoch,
	}, testscommon.CreateMemUnit())

	args.ChanceComputer = &mock.ChanceComputerStub{
		GetChanceCalled: func(rating uint32) uint32 {
			if rating == 0 {
				return 10
			}
			return rating
		},
	}
	s, _ := NewSystemSCProcessor(args)
	require.NotNil(t, s)

	numEligible := 9
	numWaiting := 5
	numJailed := 8
	stakingScAcc := stakingcommon.LoadUserAccount(args.UserAccountsDB, vm.StakingSCAddress)
	createEligibleNodes(numEligible, stakingScAcc, args.Marshalizer)
	_ = createWaitingNodes(numWaiting, stakingScAcc, args.UserAccountsDB, args.Marshalizer)
	jailed := createJailedNodes(numJailed, stakingScAcc, args.UserAccountsDB, args.PeerAccountsDB, args.Marshalizer)

	_ = s.userAccountsDB.SaveAccount(stakingScAcc)
	_, _ = s.userAccountsDB.Commit()

	stakingcommon.AddValidatorData(args.UserAccountsDB, []byte("ownerForAll"), [][]byte{[]byte("stakedPubKey1"), []byte("stakedPubKey2"), []byte("stakedPubKey3")}, big.NewInt(900000), args.Marshalizer)

	validatorsInfo := state.NewShardValidatorsInfoMap()
	_ = validatorsInfo.SetValidatorsInShard(0, jailed)

	err := s.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
	require.Nil(t, err)
	for i := 0; i < numWaiting; i++ {
		require.Equal(t, string(common.NewList), validatorsInfo.GetShardValidatorsInfoMap()[0][i].GetList())
	}
	for i := numWaiting; i < numJailed; i++ {
		require.Equal(t, string(common.JailedList), validatorsInfo.GetShardValidatorsInfoMap()[0][i].GetList())
	}

	newJailedNodes := jailed[numWaiting:numJailed]
	checkNodesStatusInSystemSCDataTrie(t, newJailedNodes, args.UserAccountsDB, args.Marshalizer, saveJailedAlwaysEnableEpoch == 0)
}

func checkNodesStatusInSystemSCDataTrie(t *testing.T, nodes []state.ValidatorInfoHandler, accounts state.AccountsAdapter, marshalizer marshal.Marshalizer, jailed bool) {
	account, err := accounts.LoadAccount(vm.StakingSCAddress)
	require.Nil(t, err)

	var buff []byte
	systemScAccount, ok := account.(state.UserAccountHandler)
	require.True(t, ok)
	for _, nodeInfo := range nodes {
		buff, _, err = systemScAccount.RetrieveValue(nodeInfo.GetPublicKey())
		require.Nil(t, err)
		require.True(t, len(buff) > 0)

		stakingData := &systemSmartContracts.StakedDataV1_1{}
		err = marshalizer.Unmarshal(stakingData, buff)
		require.Nil(t, err)

		assert.Equal(t, jailed, stakingData.Jailed)
	}
}

func TestSystemSCProcessor_NobodyToSwapWithStakingV2(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())
	args.ChanceComputer = &mock.ChanceComputerStub{
		GetChanceCalled: func(rating uint32) uint32 {
			if rating == 0 {
				return 10
			}
			return rating
		},
	}
	s, _ := NewSystemSCProcessor(args)
	require.NotNil(t, s)

	owner1 := append([]byte("owner1"), bytes.Repeat([]byte{1}, 26)...)
	blsKeys := [][]byte{
		[]byte("bls key 1"),
		[]byte("bls key 2"),
		[]byte("bls key 3"),
		[]byte("bls key 4"),
	}

	_ = s.initDelegationSystemSC()
	doStake(t, s.systemVM, s.userAccountsDB, owner1, big.NewInt(1000), blsKeys...)
	doUnStake(t, s.systemVM, s.userAccountsDB, owner1, blsKeys[:3]...)
	validatorsInfo := state.NewShardValidatorsInfoMap()
	jailed := &state.ValidatorInfo{
		PublicKey:       blsKeys[0],
		ShardId:         0,
		List:            string(common.JailedList),
		TempRating:      1,
		RewardAddress:   []byte("owner1"),
		AccumulatedFees: big.NewInt(0),
	}
	_ = validatorsInfo.Add(jailed)

	err := s.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
	assert.Nil(t, err)

	for _, vInfo := range validatorsInfo.GetShardValidatorsInfoMap()[0] {
		assert.Equal(t, string(common.JailedList), vInfo.GetList())
	}

	nodesToUnStake, mapOwnersKeys, err := s.stakingDataProvider.ComputeUnQualifiedNodes(validatorsInfo)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(nodesToUnStake))
	assert.Equal(t, 0, len(mapOwnersKeys))
}

func TestSystemSCProcessor_UpdateStakingV2ShouldWork(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{
		StakingV2EnableEpoch: 1000,
	}, testscommon.CreateMemUnit())
	args.ChanceComputer = &mock.ChanceComputerStub{
		GetChanceCalled: func(rating uint32) uint32 {
			if rating == 0 {
				return 10
			}
			return rating
		},
	}
	s, _ := NewSystemSCProcessor(args)
	require.NotNil(t, s)

	owner1 := append([]byte("owner1"), bytes.Repeat([]byte{1}, 26)...)
	owner2 := append([]byte("owner2"), bytes.Repeat([]byte{1}, 26)...)

	blsKeys := [][]byte{
		[]byte("bls key 1"),
		[]byte("bls key 2"),
		[]byte("bls key 3"),
		[]byte("bls key 4"),
	}

	doStake(t, s.systemVM, s.userAccountsDB, owner1, big.NewInt(1000), blsKeys[0], blsKeys[1])
	doStake(t, s.systemVM, s.userAccountsDB, owner2, big.NewInt(1000), blsKeys[2], blsKeys[3])

	args.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: 1000000,
	})

	for _, blsKey := range blsKeys {
		checkOwnerOfBlsKey(t, s.systemVM, blsKey, []byte(""))
	}

	err := s.updateOwnersForBlsKeys()
	assert.Nil(t, err)

	checkOwnerOfBlsKey(t, s.systemVM, blsKeys[0], owner1)
	checkOwnerOfBlsKey(t, s.systemVM, blsKeys[1], owner1)
	checkOwnerOfBlsKey(t, s.systemVM, blsKeys[2], owner2)
	checkOwnerOfBlsKey(t, s.systemVM, blsKeys[3], owner2)
}

func TestSystemSCProcessor_UpdateStakingV2MoreKeysShouldWork(t *testing.T) {
	t.Parallel()

	db, dir := createPhysicalUnit(t)
	require.False(t, check.IfNil(db))

	log.Info("using temporary directory", "path", dir)

	sw := core.NewStopWatch()
	sw.Start("complete test")
	defer func() {
		sw.Stop("complete test")
		log.Info("TestSystemSCProcessor_UpdateStakingV2MoreKeysShouldWork time measurements", sw.GetMeasurements()...)
		_ = db.DestroyUnit()
		_ = os.RemoveAll(dir)
	}()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{
		StakingV2EnableEpoch: 1000,
	}, db)
	args.ChanceComputer = &mock.ChanceComputerStub{
		GetChanceCalled: func(rating uint32) uint32 {
			if rating == 0 {
				return 10
			}
			return rating
		},
	}
	s, _ := NewSystemSCProcessor(args)
	require.NotNil(t, s)

	numKeys := 5000
	keys := make([]*testKeyPair, 0, numKeys)
	for i := 0; i < numKeys; i++ {
		if i%100 == 0 {
			_, err := args.UserAccountsDB.Commit()
			require.Nil(t, err)
		}

		keys = append(keys, createTestKeyPair())
	}

	sw.Start("do stake")
	for _, tkp := range keys {
		doStake(t, s.systemVM, s.userAccountsDB, tkp.walletKey, big.NewInt(1000), tkp.validatorKey)

	}
	sw.Stop("do stake")

	args.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: 1000000,
	})

	sw.Start("initial check")
	for _, tkp := range keys {
		checkOwnerOfBlsKey(t, s.systemVM, tkp.validatorKey, []byte(""))
	}
	sw.Stop("initial check")

	sw.Start("updateOwnersForBlsKeys")
	err := s.updateOwnersForBlsKeys()
	sw.Stop("updateOwnersForBlsKeys")
	assert.Nil(t, err)

	sw.Start("final check")
	for _, tkp := range keys {
		checkOwnerOfBlsKey(t, s.systemVM, tkp.validatorKey, tkp.walletKey)
	}
	sw.Stop("final check")
}

func createTestKeyPair() *testKeyPair {
	tkp := &testKeyPair{
		walletKey:    make([]byte, 32),
		validatorKey: make([]byte, 96),
	}

	_, _ = rand.Read(tkp.walletKey)
	_, _ = rand.Read(tkp.validatorKey)

	return tkp
}

func checkOwnerOfBlsKey(t *testing.T, systemVm vmcommon.VMExecutionHandler, blsKey []byte, expectedOwner []byte) {
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: vm.ValidatorSCAddress,
			Arguments:  [][]byte{blsKey},
			CallValue:  big.NewInt(0),
		},
		RecipientAddr: vm.StakingSCAddress,
		Function:      "getOwner",
	}

	vmOutput, err := systemVm.RunSmartContractCall(vmInput)
	require.Nil(t, err)
	if len(expectedOwner) == 0 {
		require.Equal(t, vmOutput.ReturnCode, vmcommon.UserError)
		return
	}

	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	require.Equal(t, 1, len(vmOutput.ReturnData))

	assert.Equal(t, len(expectedOwner), len(vmOutput.ReturnData[0]))
	if len(expectedOwner) > 0 {
		assert.Equal(t, expectedOwner, vmOutput.ReturnData[0])
	}
}

func doStake(t *testing.T, systemVm vmcommon.VMExecutionHandler, accountsDB state.AccountsAdapter, owner []byte, nodePrice *big.Int, blsKeys ...[]byte) {
	numBlsKeys := len(blsKeys)
	args := [][]byte{{0, byte(numBlsKeys)}}
	for _, blsKey := range blsKeys {
		args = append(args, blsKey)
		args = append(args, []byte("sig"))
	}

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  owner,
			Arguments:   args,
			CallValue:   big.NewInt(0).Mul(big.NewInt(int64(numBlsKeys)), nodePrice),
			GasProvided: math.MaxInt64,
		},
		RecipientAddr: vm.ValidatorSCAddress,
		Function:      "stake",
	}

	vmOutput, err := systemVm.RunSmartContractCall(vmInput)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)

	saveOutputAccounts(t, accountsDB, vmOutput)
}

func doUnStake(t *testing.T, systemVm vmcommon.VMExecutionHandler, accountsDB state.AccountsAdapter, owner []byte, blsKeys ...[]byte) {
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  owner,
			Arguments:   blsKeys,
			CallValue:   big.NewInt(0),
			GasProvided: math.MaxInt64,
		},
		RecipientAddr: vm.ValidatorSCAddress,
		Function:      "unStake",
	}

	vmOutput, err := systemVm.RunSmartContractCall(vmInput)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)

	saveOutputAccounts(t, accountsDB, vmOutput)
}

func createEligibleNodes(numNodes int, stakingSCAcc state.UserAccountHandler, marshalizer marshal.Marshalizer) {
	for i := 0; i < numNodes; i++ {
		stakedData := &systemSmartContracts.StakedDataV2_0{
			Waiting:       false,
			Staked:        true,
			StakedNonce:   0,
			RewardAddress: []byte(fmt.Sprintf("rewardAddress_w%d", i)),
			OwnerAddress:  []byte("ownerForAll"),
			StakeValue:    big.NewInt(100),
		}
		marshaledData, _ := marshalizer.Marshal(stakedData)
		_ = stakingSCAcc.SaveKeyValue([]byte(fmt.Sprintf("waiting_%d", i)), marshaledData)
	}
}

func createJailedNodes(numNodes int, stakingSCAcc state.UserAccountHandler, userAccounts state.AccountsAdapter, peerAccounts state.AccountsAdapter, marshalizer marshal.Marshalizer) []state.ValidatorInfoHandler {
	validatorInfos := make([]state.ValidatorInfoHandler, 0)

	for i := 0; i < numNodes; i++ {
		stakedData := &systemSmartContracts.StakedDataV2_0{
			Staked:        true,
			RewardAddress: []byte(fmt.Sprintf("rewardAddress_j%d", i)),
			StakeValue:    big.NewInt(100),
			OwnerAddress:  []byte("ownerForAll"),
		}
		marshaledData, _ := marshalizer.Marshal(stakedData)
		_ = stakingSCAcc.SaveKeyValue([]byte(fmt.Sprintf("jailed__%d", i)), marshaledData)

		_ = userAccounts.SaveAccount(stakingSCAcc)

		jailedAcc, _ := peerAccounts.LoadAccount([]byte(fmt.Sprintf("jailed__%d", i)))
		_ = peerAccounts.SaveAccount(jailedAcc)

		vInfo := &state.ValidatorInfo{
			PublicKey:       []byte(fmt.Sprintf("jailed__%d", i)),
			ShardId:         0,
			List:            string(common.JailedList),
			TempRating:      1,
			RewardAddress:   []byte("address"),
			AccumulatedFees: big.NewInt(0),
		}

		validatorInfos = append(validatorInfos, vInfo)
	}

	return validatorInfos
}

func addValidatorDataWithUnStakedKey(
	accountsDB state.AccountsAdapter,
	ownerKey []byte,
	registeredKeys [][]byte,
	nodePrice *big.Int,
	marshalizer marshal.Marshalizer,
) {
	stakingAccount := stakingcommon.LoadUserAccount(accountsDB, vm.StakingSCAddress)
	validatorAccount := stakingcommon.LoadUserAccount(accountsDB, vm.ValidatorSCAddress)

	validatorData := &systemSmartContracts.ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   ownerKey,
		TotalStakeValue: big.NewInt(0),
		LockedStake:     big.NewInt(0).Mul(nodePrice, big.NewInt(int64(len(registeredKeys)))),
		TotalUnstaked:   big.NewInt(0).Mul(nodePrice, big.NewInt(int64(len(registeredKeys)))),
		BlsPubKeys:      registeredKeys,
		NumRegistered:   uint32(len(registeredKeys)),
	}

	for _, bls := range registeredKeys {
		validatorData.UnstakedInfo = append(validatorData.UnstakedInfo, &systemSmartContracts.UnstakedValue{
			UnstakedEpoch: 1,
			UnstakedValue: nodePrice,
		})

		stakingData := &systemSmartContracts.StakedDataV2_0{
			RegisterNonce: 0,
			StakedNonce:   0,
			Staked:        false,
			UnStakedNonce: 1,
			UnStakedEpoch: 0,
			RewardAddress: ownerKey,
			StakeValue:    nodePrice,
			JailedRound:   0,
			JailedNonce:   0,
			UnJailedNonce: 0,
			Jailed:        false,
			Waiting:       false,
			NumJailed:     0,
			SlashValue:    big.NewInt(0),
			OwnerAddress:  ownerKey,
		}
		marshaledData, _ := marshalizer.Marshal(stakingData)
		_ = stakingAccount.SaveKeyValue(bls, marshaledData)
	}

	marshaledData, _ := marshalizer.Marshal(validatorData)
	_ = validatorAccount.SaveKeyValue(ownerKey, marshaledData)

	_ = accountsDB.SaveAccount(validatorAccount)
	_ = accountsDB.SaveAccount(stakingAccount)
}

func createWaitingNodes(numNodes int, stakingSCAcc state.UserAccountHandler, userAccounts state.AccountsAdapter, marshalizer marshal.Marshalizer) []*state.ValidatorInfo {
	validatorInfos := make([]*state.ValidatorInfo, 0)
	waitingKeyInList := []byte("waiting")
	for i := 0; i < numNodes; i++ {
		stakedData := &systemSmartContracts.StakedDataV2_0{
			Waiting:       true,
			RewardAddress: []byte(fmt.Sprintf("rewardAddress_w%d", i)),
			OwnerAddress:  []byte("ownerForAll"),
			StakeValue:    big.NewInt(100),
		}
		marshaledData, _ := marshalizer.Marshal(stakedData)
		_ = stakingSCAcc.SaveKeyValue([]byte(fmt.Sprintf("waiting_%d", i)), marshaledData)
		previousKey := string(waitingKeyInList)
		waitingKeyInList = []byte("w_" + fmt.Sprintf("waiting_%d", i))
		waitingListHead := &systemSmartContracts.WaitingList{
			FirstKey: []byte("w_" + fmt.Sprintf("waiting_%d", 0)),
			LastKey:  []byte("w_" + fmt.Sprintf("waiting_%d", numNodes-1)),
			Length:   uint32(numNodes),
		}
		marshaledData, _ = marshalizer.Marshal(waitingListHead)
		_ = stakingSCAcc.SaveKeyValue([]byte("waitingList"), marshaledData)

		waitingListElement := &systemSmartContracts.ElementInList{
			BLSPublicKey: []byte(fmt.Sprintf("waiting_%d", i)),
			PreviousKey:  waitingKeyInList,
			NextKey:      []byte("w_" + fmt.Sprintf("waiting_%d", i+1)),
		}
		if i == numNodes-1 {
			waitingListElement.NextKey = make([]byte, 0)
		}
		if i > 0 {
			waitingListElement.PreviousKey = []byte(previousKey)
		}

		marshaledData, _ = marshalizer.Marshal(waitingListElement)
		_ = stakingSCAcc.SaveKeyValue(waitingKeyInList, marshaledData)

		vInfo := &state.ValidatorInfo{
			PublicKey:       []byte(fmt.Sprintf("waiting_%d", i)),
			ShardId:         0,
			List:            string(common.WaitingList),
			TempRating:      1,
			RewardAddress:   []byte("address"),
			AccumulatedFees: big.NewInt(0),
		}

		validatorInfos = append(validatorInfos, vInfo)
	}

	_ = userAccounts.SaveAccount(stakingSCAcc)

	return validatorInfos
}

func prepareStakingContractWithData(
	accountsDB state.AccountsAdapter,
	stakedKey []byte,
	waitingKey []byte,
	marshalizer marshal.Marshalizer,
	rewardAddress []byte,
	ownerAddress []byte,
) {
	stakingcommon.AddStakingData(accountsDB, ownerAddress, rewardAddress, [][]byte{stakedKey}, marshalizer)
	stakingcommon.AddKeysToWaitingList(accountsDB, [][]byte{waitingKey}, marshalizer, rewardAddress, ownerAddress)
	stakingcommon.AddValidatorData(accountsDB, rewardAddress, [][]byte{stakedKey, waitingKey}, big.NewInt(10000000000), marshalizer)

	_, err := accountsDB.Commit()
	log.LogIfError(err)
}

func createAccountsDB(
	hasher hashing.Hasher,
	marshaller marshal.Marshalizer,
	accountFactory state.AccountFactory,
	trieStorageManager common.StorageManager,
	enableEpochsHandler common.EnableEpochsHandler,
) *state.AccountsDB {
	tr, _ := trie.NewTrie(trieStorageManager, marshaller, hasher, enableEpochsHandler, 5)
	ewlArgs := evictionWaitingList.MemoryEvictionWaitingListArgs{
		RootHashesSize: 100,
		HashesSize:     10000,
	}
	ewl, _ := evictionWaitingList.NewMemoryEvictionWaitingList(ewlArgs)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 10)

	args := state.ArgsAccountsDB{
		Trie:                  tr,
		Hasher:                hasher,
		Marshaller:            marshaller,
		AccountFactory:        accountFactory,
		StoragePruningManager: spm,
		AddressConverter:      &testscommon.PubkeyConverterMock{},
		SnapshotsManager:      disabledState.NewDisabledSnapshotsManager(),
	}
	adb, _ := state.NewAccountsDB(args)
	return adb
}

func createFullArgumentsForSystemSCProcessing(enableEpochsConfig config.EnableEpochs, trieStorer storage.Storer) (ArgsNewEpochStartSystemSCProcessing, vm.SystemSCContainer) {
	hasher := sha256.NewSha256()
	marshalizer := &marshal.GogoProtoMarshalizer{}
	storageManagerArgs := storageMock.GetStorageManagerArgs()
	storageManagerArgs.Marshalizer = marshalizer
	storageManagerArgs.Hasher = hasher
	storageManagerArgs.MainStorer = trieStorer

	trieFactoryManager, _ := trie.CreateTrieStorageManager(storageManagerArgs, storageMock.GetStorageManagerOptions())
	argsAccCreator := factory.ArgsAccountCreator{
		Hasher:              hasher,
		Marshaller:          marshalizer,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	accCreator, _ := factory.NewAccountCreator(argsAccCreator)
	peerAccCreator := factory.NewPeerAccountCreator()
	en := forking.NewGenericEpochNotifier()
	enableEpochsConfig.StakeLimitsEnableEpoch = 10
	enableEpochsConfig.StakingV4Step1EnableEpoch = 444
	enableEpochsConfig.StakingV4Step2EnableEpoch = 445
	epochsConfig := &config.EpochConfig{
		EnableEpochs: enableEpochsConfig,
	}
	enableEpochsHandler, _ := enablers.NewEnableEpochsHandler(epochsConfig.EnableEpochs, en)
	userAccountsDB := createAccountsDB(hasher, marshalizer, accCreator, trieFactoryManager, enableEpochsHandler)
	peerAccountsDB := createAccountsDB(hasher, marshalizer, peerAccCreator, trieFactoryManager, enableEpochsHandler)

	argsValidatorsProcessor := peer.ArgValidatorStatisticsProcessor{
		Marshalizer:                          marshalizer,
		NodesCoordinator:                     &shardingMocks.NodesCoordinatorStub{},
		ShardCoordinator:                     &mock.ShardCoordinatorStub{},
		DataPool:                             &dataRetrieverMock.PoolsHolderStub{},
		StorageService:                       &storageMock.ChainStorerStub{},
		PubkeyConv:                           &testscommon.PubkeyConverterMock{},
		PeerAdapter:                          peerAccountsDB,
		Rater:                                &mock.RaterStub{},
		RewardsHandler:                       &mock.RewardsHandlerStub{},
		NodesSetup:                           &genesisMocks.NodesSetupStub{},
		MaxComputableRounds:                  1,
		MaxConsecutiveRoundsOfRatingDecrease: 2000,
		EnableEpochsHandler:                  enableEpochsHandler,
	}
	vCreator, _ := peer.NewValidatorStatisticsProcessor(argsValidatorsProcessor)

	blockChain, _ := blockchain.NewMetaChain(&statusHandlerMock.AppStatusHandlerStub{})
	gasSchedule := wasmConfig.MakeGasMapForTests()
	gasScheduleNotifier := testscommon.NewGasScheduleNotifierMock(gasSchedule)
	testDataPool := dataRetrieverMock.NewPoolsHolderMock()

	defaults.FillGasMapInternal(gasSchedule, 1)
	signVerifer, _ := disabled.NewMessageSignVerifier(&cryptoMocks.KeyGenStub{})

	nodesSetup := &genesisMocks.NodesSetupStub{}

	argsHook := hooks.ArgBlockChainHook{
		Accounts:                 userAccountsDB,
		PubkeyConv:               &testscommon.PubkeyConverterMock{},
		StorageService:           &storageMock.ChainStorerStub{},
		BlockChain:               blockChain,
		ShardCoordinator:         &mock.ShardCoordinatorStub{},
		Marshalizer:              marshalizer,
		Uint64Converter:          &mock.Uint64ByteSliceConverterMock{},
		NFTStorageHandler:        &testscommon.SimpleNFTStorageHandlerStub{},
		BuiltInFunctions:         vmcommonBuiltInFunctions.NewBuiltInFunctionContainer(),
		DataPool:                 testDataPool,
		GlobalSettingsHandler:    &testscommon.ESDTGlobalSettingsHandlerStub{},
		CompiledSCPool:           testDataPool.SmartContracts(),
		EpochNotifier:            en,
		EnableEpochsHandler:      enableEpochsHandler,
		NilCompiledSCStore:       true,
		GasSchedule:              gasScheduleNotifier,
		Counter:                  &testscommon.BlockChainHookCounterStub{},
		MissingTrieNodesNotifier: &testscommon.MissingTrieNodesNotifierStub{},
	}

	defaults.FillGasMapInternal(gasSchedule, 1)

	blockChainHookImpl, _ := hooks.NewBlockChainHookImpl(argsHook)
	argsNewVMContainerFactory := metaProcess.ArgsNewVMContainerFactory{
		BlockChainHook:      blockChainHookImpl,
		PubkeyConv:          argsHook.PubkeyConv,
		Economics:           stakingcommon.CreateEconomicsData(),
		MessageSignVerifier: signVerifer,
		GasSchedule:         gasScheduleNotifier,
		NodesConfigProvider: nodesSetup,
		Hasher:              hasher,
		Marshalizer:         marshalizer,
		SystemSCConfig: &config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost: "1000",
				OwnerAddress:    "aaaaaa",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				V1: config.GovernanceSystemSCConfigV1{
					ProposalCost: "500",
				},
				Active: config.GovernanceSystemSCConfigActive{
					ProposalCost:     "500",
					MinQuorum:        0.5,
					MinPassThreshold: 0.5,
					MinVetoThreshold: 0.5,
					LostProposalFee:  "1",
				},
				OwnerAddress: "3132333435363738393031323334353637383930313233343536373839303234",
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     "1000",
				UnJailValue:                          "10",
				MinStepValue:                         "10",
				MinStakeValue:                        "1",
				UnBondPeriod:                         1,
				NumRoundsWithoutBleed:                1,
				MaximumPercentageToBleed:             1,
				BleedPercentagePerRound:              1,
				MaxNumberOfNodesForStake:             5,
				ActivateBLSPubKeyMessageVerification: false,
				MinUnstakeTokensValue:                "1",
				StakeLimitPercentage:                 100.0,
				NodeLimitPercentage:                  100.0,
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
			SoftAuctionConfig: config.SoftAuctionConfig{
				TopUpStep:             "10",
				MinTopUp:              "1",
				MaxTopUp:              "32000000",
				MaxNumberOfIterations: 100000,
			},
		},
		ValidatorAccountsDB: peerAccountsDB,
		UserAccountsDB:      userAccountsDB,
		ChanceComputer:      &mock.ChanceComputerStub{},
		ShardCoordinator:    &mock.ShardCoordinatorStub{},
		EnableEpochsHandler: enableEpochsHandler,
		NodesCoordinator:    &shardingMocks.NodesCoordinatorStub{},
		ArgBlockChainHook:   argsHook,
	}
	metaVmFactory, _ := metaProcess.NewVMContainerFactory(argsNewVMContainerFactory)

	vmContainer, _ := metaVmFactory.Create()
	systemVM, _ := vmContainer.Get(vmFactory.SystemVirtualMachine)

	argsStakingDataProvider := StakingDataProviderArgs{
		EnableEpochsHandler: enableEpochsHandler,
		SystemVM:            systemVM,
		MinNodePrice:        "1000",
	}
	stakingSCProvider, _ := NewStakingDataProvider(argsStakingDataProvider)
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(3, core.MetachainShardId)

	nodesConfigProvider, _ := notifier.NewNodesConfigProvider(en, nil)
	auctionCfg := config.SoftAuctionConfig{
		TopUpStep:             "10",
		MinTopUp:              "1",
		MaxTopUp:              "32000000",
		MaxNumberOfIterations: 100000,
	}
	ald, _ := NewAuctionListDisplayer(ArgsAuctionListDisplayer{
		TableDisplayHandler:      NewTableDisplayer(),
		ValidatorPubKeyConverter: &testscommon.PubkeyConverterMock{},
		AddressPubKeyConverter:   &testscommon.PubkeyConverterMock{},
		AuctionConfig:            auctionCfg,
		Denomination:             0,
	})
	argsAuctionListSelector := AuctionListSelectorArgs{
		ShardCoordinator:             shardCoordinator,
		StakingDataProvider:          stakingSCProvider,
		MaxNodesChangeConfigProvider: nodesConfigProvider,
		AuctionListDisplayHandler:    ald,
		SoftAuctionConfig:            auctionCfg,
	}
	als, _ := NewAuctionListSelector(argsAuctionListSelector)

	args := ArgsNewEpochStartSystemSCProcessing{
		SystemVM:                systemVM,
		UserAccountsDB:          userAccountsDB,
		PeerAccountsDB:          peerAccountsDB,
		Marshalizer:             marshalizer,
		StartRating:             5,
		ValidatorInfoCreator:    vCreator,
		EndOfEpochCallerAddress: vm.EndOfEpochAddress,
		StakingSCAddress:        vm.StakingSCAddress,
		ChanceComputer:          &mock.ChanceComputerStub{},
		EpochNotifier:           en,
		GenesisNodesConfig:      nodesSetup,
		StakingDataProvider:     stakingSCProvider,
		AuctionListSelector:     als,
		NodesConfigProvider: &shardingMocks.NodesCoordinatorStub{
			ConsensusGroupSizeCalled: func(shardID uint32) int {
				if shardID == core.MetachainShardId {
					return 400
				}
				return 63
			},
		},
		ShardCoordinator:             shardCoordinator,
		ESDTOwnerAddressBytes:        bytes.Repeat([]byte{1}, 32),
		MaxNodesChangeConfigProvider: nodesConfigProvider,
		EnableEpochsHandler:          enableEpochsHandler,
	}
	return args, metaVmFactory.SystemSmartContractContainer()
}

func TestSystemSCProcessor_ProcessSystemSmartContractInitDelegationMgr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("flag not active", func(t *testing.T) {
		args := createMockArgsForSystemSCProcessor()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				if flag == common.GovernanceFlagInSpecificEpochOnly ||
					flag == common.StakingV4Step1Flag ||
					flag == common.StakingV4Step2Flag ||
					flag == common.SwitchHysteresisForMinNodesFlagInSpecificEpochOnly ||
					flag == common.StakingV2OwnerFlagInSpecificEpochOnly ||
					flag == common.CorrectLastUnJailedFlagInSpecificEpochOnly ||
					flag == common.DelegationSmartContractFlagInSpecificEpochOnly ||
					flag == common.CorrectLastUnJailedFlag ||
					flag == common.SwitchJailWaitingFlag ||
					flag == common.StakingV2Flag ||
					flag == common.ESDTFlagInSpecificEpochOnly {

					return false
				}

				return true
			},
		}
		args.SystemVM = &mock.VMExecutionHandlerStub{
			RunSmartContractCreateCalled: func(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
				assert.Fail(t, "should have not called")

				return nil, fmt.Errorf("should have not called")
			},
		}
		processor, _ := NewSystemSCProcessor(args)
		require.NotNil(t, processor)

		validatorsInfo := state.NewShardValidatorsInfoMap()
		err := processor.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
		require.Nil(t, err)
	})
	t.Run("flag active", func(t *testing.T) {
		args := createMockArgsForSystemSCProcessor()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.DelegationSmartContractFlagInSpecificEpochOnly
			},
		}
		runSmartContractCreateCalled := false
		args.SystemVM = &mock.VMExecutionHandlerStub{
			RunSmartContractCreateCalled: func(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
				runSmartContractCreateCalled = true

				return &vmcommon.VMOutput{}, nil
			},
		}
		processor, _ := NewSystemSCProcessor(args)
		require.NotNil(t, processor)

		validatorsInfo := state.NewShardValidatorsInfoMap()
		err := processor.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
		require.Nil(t, err)
		require.True(t, runSmartContractCreateCalled)
	})
	t.Run("flag active but contract create call errors, should error", func(t *testing.T) {
		args := createMockArgsForSystemSCProcessor()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.DelegationSmartContractFlagInSpecificEpochOnly
			},
		}
		runSmartContractCreateCalled := false
		args.SystemVM = &mock.VMExecutionHandlerStub{
			RunSmartContractCreateCalled: func(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
				runSmartContractCreateCalled = true

				return nil, expectedErr
			},
		}
		processor, _ := NewSystemSCProcessor(args)
		require.NotNil(t, processor)

		validatorsInfo := state.NewShardValidatorsInfoMap()
		err := processor.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
		require.ErrorIs(t, err, expectedErr)
		require.True(t, runSmartContractCreateCalled)
	})
}

func TestSystemSCProcessor_ProcessSystemSmartContractInitGovernance(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("flag not active", func(t *testing.T) {
		args := createMockArgsForSystemSCProcessor()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				if flag == common.GovernanceFlagInSpecificEpochOnly ||
					flag == common.StakingV4Step1Flag ||
					flag == common.StakingV4Step2Flag ||
					flag == common.SwitchHysteresisForMinNodesFlagInSpecificEpochOnly ||
					flag == common.StakingV2OwnerFlagInSpecificEpochOnly ||
					flag == common.CorrectLastUnJailedFlagInSpecificEpochOnly ||
					flag == common.DelegationSmartContractFlagInSpecificEpochOnly ||
					flag == common.CorrectLastUnJailedFlag ||
					flag == common.SwitchJailWaitingFlag ||
					flag == common.StakingV2Flag ||
					flag == common.ESDTFlagInSpecificEpochOnly {

					return false
				}

				return true
			},
		}
		args.SystemVM = &mock.VMExecutionHandlerStub{
			RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
				assert.Fail(t, "should have not called")

				return nil, fmt.Errorf("should have not called")
			},
		}
		processor, _ := NewSystemSCProcessor(args)
		require.NotNil(t, processor)

		validatorsInfo := state.NewShardValidatorsInfoMap()
		err := processor.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
		require.Nil(t, err)
	})
	t.Run("flag active", func(t *testing.T) {
		args := createMockArgsForSystemSCProcessor()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.GovernanceFlagInSpecificEpochOnly
			},
		}
		runSmartContractCreateCalled := false
		args.SystemVM = &mock.VMExecutionHandlerStub{
			RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
				runSmartContractCreateCalled = true

				return &vmcommon.VMOutput{}, nil
			},
		}
		processor, _ := NewSystemSCProcessor(args)
		require.NotNil(t, processor)

		validatorsInfo := state.NewShardValidatorsInfoMap()
		err := processor.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
		require.Nil(t, err)
		require.True(t, runSmartContractCreateCalled)
	})
	t.Run("flag active but contract call errors, should error", func(t *testing.T) {
		args := createMockArgsForSystemSCProcessor()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.GovernanceFlagInSpecificEpochOnly
			},
		}
		runSmartContractCreateCalled := false
		args.SystemVM = &mock.VMExecutionHandlerStub{
			RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
				runSmartContractCreateCalled = true

				return nil, expectedErr
			},
		}
		processor, _ := NewSystemSCProcessor(args)
		require.NotNil(t, processor)

		validatorsInfo := state.NewShardValidatorsInfoMap()
		err := processor.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
		require.ErrorIs(t, err, expectedErr)
		require.Contains(t, err.Error(), "governanceV2")
		require.True(t, runSmartContractCreateCalled)
	})
}

func TestSystemSCProcessor_ProcessDelegationRewardsNothingToExecute(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{
		StakingV2EnableEpoch: 1000,
	}, testscommon.CreateMemUnit())
	s, _ := NewSystemSCProcessor(args)

	localCache := dataPool.NewCurrentBlockTransactionsPool()
	miniBlocks := []*block.MiniBlock{
		{
			SenderShardID:   0,
			ReceiverShardID: 0,
			TxHashes:        [][]byte{[]byte("txHash")},
		},
	}

	err := s.ProcessDelegationRewards(miniBlocks, localCache)
	assert.Nil(t, err)
}

func TestSystemSCProcessor_ProcessDelegationRewardsErrors(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{
		StakingV2EnableEpoch: 1000,
	}, testscommon.CreateMemUnit())
	s, _ := NewSystemSCProcessor(args)

	localCache := dataPool.NewCurrentBlockTransactionsPool()
	miniBlocks := []*block.MiniBlock{
		{
			SenderShardID:   core.MetachainShardId,
			ReceiverShardID: core.MetachainShardId,
			TxHashes:        [][]byte{[]byte("txHash")},
			Type:            block.RewardsBlock,
		},
	}

	err := s.ProcessDelegationRewards(nil, localCache)
	assert.Nil(t, err)

	err = s.ProcessDelegationRewards(miniBlocks, nil)
	assert.Equal(t, err, epochStart.ErrNilLocalTxCache)

	err = s.ProcessDelegationRewards(miniBlocks, localCache)
	assert.Equal(t, err, dataRetriever.ErrTxNotFoundInBlockPool)

	rwdTx := &rewardTx.RewardTx{
		Round:   0,
		Value:   big.NewInt(100),
		RcvAddr: make([]byte, len(vm.StakingSCAddress)),
		Epoch:   0,
	}
	localCache.AddTx([]byte("txHash"), rwdTx)
	copy(rwdTx.RcvAddr, vm.StakingSCAddress)
	err = s.ProcessDelegationRewards(miniBlocks, localCache)
	assert.Equal(t, err, epochStart.ErrSystemDelegationCall)

	rwdTx.RcvAddr[25] = 255
	err = s.ProcessDelegationRewards(miniBlocks, localCache)
	assert.Equal(t, err, vm.ErrUnknownSystemSmartContract)

	rwdTx.RcvAddr = vm.FirstDelegationSCAddress
	err = s.ProcessDelegationRewards(miniBlocks, localCache)
	assert.Equal(t, err, epochStart.ErrSystemDelegationCall)
}

func TestSystemSCProcessor_ProcessDelegationRewards(t *testing.T) {
	t.Parallel()

	args, scContainer := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{
		StakingV2EnableEpoch: 1000,
	}, testscommon.CreateMemUnit())
	s, _ := NewSystemSCProcessor(args)

	localCache := dataPool.NewCurrentBlockTransactionsPool()
	miniBlocks := []*block.MiniBlock{
		{
			SenderShardID:   core.MetachainShardId,
			ReceiverShardID: core.MetachainShardId,
			TxHashes:        [][]byte{[]byte("txHash")},
			Type:            block.RewardsBlock,
		},
	}

	rwdTx := &rewardTx.RewardTx{
		Round:   0,
		Value:   big.NewInt(100),
		RcvAddr: generateSecondDelegationAddress(),
		Epoch:   0,
	}
	localCache.AddTx([]byte("txHash"), rwdTx)

	contract, _ := scContainer.Get(vm.FirstDelegationSCAddress)
	_ = scContainer.Add(rwdTx.RcvAddr, contract)

	err := s.ProcessDelegationRewards(miniBlocks, localCache)
	assert.Nil(t, err)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  vm.EndOfEpochAddress,
			Arguments:   [][]byte{big.NewInt(int64(rwdTx.Epoch)).Bytes()},
			CallValue:   big.NewInt(0),
			GasProvided: 1000000,
		},
		RecipientAddr: rwdTx.RcvAddr,
		Function:      "getRewardData",
	}

	vmOutput, err := args.SystemVM.RunSmartContractCall(vmInput)
	assert.Nil(t, err)
	assert.NotNil(t, vmOutput)

	assert.Equal(t, len(vmOutput.ReturnData), 3)
	assert.True(t, bytes.Equal(vmOutput.ReturnData[0], rwdTx.Value.Bytes()))
}

// generateSecondDelegationAddress will generate the address of the second delegation address (the exact next one after vm.FirstDelegationSCAddress)
// by copying the vm.FirstDelegationSCAddress bytes and altering the corresponding position (28-th) to 2
func generateSecondDelegationAddress() []byte {
	address := make([]byte, len(vm.FirstDelegationSCAddress))
	copy(address, vm.FirstDelegationSCAddress)
	address[28] = 2

	return address
}

func TestSystemSCProcessor_ProcessSystemSmartContractMaxNodesStakedFromQueue(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())
	nodesConfigProvider, _ := notifier.NewNodesConfigProvider(args.EpochNotifier, []config.MaxNodesChangeConfig{{EpochEnable: 0, MaxNumNodes: 10}})
	args.MaxNodesChangeConfigProvider = nodesConfigProvider
	s, _ := NewSystemSCProcessor(args)

	prepareStakingContractWithData(
		args.UserAccountsDB,
		[]byte("stakedPubKey0"),
		[]byte("waitingPubKey"),
		args.Marshalizer,
		[]byte("rewardAddress"),
		[]byte("rewardAddress"),
	)

	validatorsInfo := state.NewShardValidatorsInfoMap()
	err := s.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
	assert.Nil(t, err)

	peerAcc, err := s.getPeerAccount([]byte("waitingPubKey"))
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(peerAcc.AddressBytes(), []byte("waitingPubKey")))
	assert.Equal(t, peerAcc.GetList(), string(common.NewList))
	numRegistered := getTotalNumberOfRegisteredNodes(t, s)
	assert.Equal(t, 1, numRegistered)
}

func getTotalNumberOfRegisteredNodes(t *testing.T, s *systemSCProcessor) int {
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: vm.EndOfEpochAddress,
			CallValue:  big.NewInt(0),
			Arguments:  make([][]byte, 0),
		},
		RecipientAddr: vm.StakingSCAddress,
		Function:      "getTotalNumberOfRegisteredNodes",
	}
	vmOutput, errRun := s.systemVM.RunSmartContractCall(vmInput)
	require.Nil(t, errRun)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	require.Equal(t, 1, len(vmOutput.ReturnData))

	value := big.NewInt(0).SetBytes(vmOutput.ReturnData[0])

	return int(value.Int64())
}

func TestSystemSCProcessor_ProcessSystemSmartContractMaxNodesStakedFromQueueOwnerNotSet(t *testing.T) {
	t.Parallel()

	maxNodesChangeConfig := []config.MaxNodesChangeConfig{{EpochEnable: 10, MaxNumNodes: 10}}
	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{
		MaxNodesChangeEnableEpoch: maxNodesChangeConfig,
		StakingV2EnableEpoch:      10,
	}, testscommon.CreateMemUnit())
	args.MaxNodesChangeConfigProvider, _ = notifier.NewNodesConfigProvider(args.EpochNotifier, maxNodesChangeConfig)
	s, _ := NewSystemSCProcessor(args)

	prepareStakingContractWithData(
		args.UserAccountsDB,
		[]byte("stakedPubKey0"),
		[]byte("waitingPubKey"),
		args.Marshalizer,
		[]byte("rewardAddress"),
		make([]byte, 0),
	)

	args.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: 10,
	})
	validatorsInfo := state.NewShardValidatorsInfoMap()
	err := s.ProcessSystemSmartContract(validatorsInfo, &block.Header{Epoch: 10})
	assert.Nil(t, err)

	peerAcc, err := s.getPeerAccount([]byte("waitingPubKey"))
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(peerAcc.AddressBytes(), []byte("waitingPubKey")))
	assert.Equal(t, peerAcc.GetList(), string(common.NewList))
}

func TestSystemSCProcessor_ESDTInitShouldWork(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{
		ESDTEnableEpoch:              1,
		SwitchJailWaitingEnableEpoch: 1,
	}, testscommon.CreateMemUnit())
	hdr := &block.MetaBlock{
		Epoch: 1,
	}
	args.EpochNotifier.CheckEpoch(hdr)
	s, _ := NewSystemSCProcessor(args)

	initialContractConfig, err := s.extractConfigFromESDTContract()
	require.Nil(t, err)
	require.Equal(t, 4, len(initialContractConfig))
	require.Equal(t, []byte("aaaaaa"), initialContractConfig[0])

	err = s.ProcessSystemSmartContract(state.NewShardValidatorsInfoMap(), &block.Header{Nonce: 1, Epoch: 1})

	require.Nil(t, err)

	updatedContractConfig, err := s.extractConfigFromESDTContract()
	require.Nil(t, err)
	require.Equal(t, 4, len(updatedContractConfig))
	require.Equal(t, args.ESDTOwnerAddressBytes, updatedContractConfig[0])
	// the other config values should be unchanged
	for i := 1; i < len(initialContractConfig); i++ {
		assert.Equal(t, initialContractConfig[i], updatedContractConfig[i])
	}
}

func TestSystemSCProcessor_ProcessSystemSmartContractUnStakeOneNodeStakeOthers(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())
	s, _ := NewSystemSCProcessor(args)

	prepareStakingContractWithData(
		args.UserAccountsDB,
		[]byte("stakedPubKey0"),
		[]byte("waitingPubKey"),
		args.Marshalizer,
		[]byte("rewardAddress"),
		[]byte("rewardAddress"),
	)
	stakingcommon.RegisterValidatorKeys(args.UserAccountsDB,
		[]byte("ownerKey"),
		[]byte("ownerKey"),
		[][]byte{[]byte("stakedPubKey1"), []byte("stakedPubKey2"), []byte("stakedPubKey3")},
		big.NewInt(2000),
		args.Marshalizer,
	)

	validatorsInfo := state.NewShardValidatorsInfoMap()
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey0"),
		List:            string(common.EligibleList),
		RewardAddress:   []byte("rewardAddress"),
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey1"),
		List:            string(common.EligibleList),
		RewardAddress:   []byte("rewardAddress"),
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey2"),
		List:            string(common.EligibleList),
		RewardAddress:   []byte("rewardAddress"),
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey3"),
		List:            string(common.EligibleList),
		RewardAddress:   []byte("rewardAddress"),
		AccumulatedFees: big.NewInt(0),
	})
	for _, vInfo := range validatorsInfo.GetShardValidatorsInfoMap()[0] {
		jailedAcc, _ := args.PeerAccountsDB.LoadAccount(vInfo.GetPublicKey())
		_ = args.PeerAccountsDB.SaveAccount(jailedAcc)
	}

	args.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: 1, // disable stakingV2OwnerFlag
	})
	err := s.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
	assert.Nil(t, err)

	peerAcc, err := s.getPeerAccount([]byte("waitingPubKey"))
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(peerAcc.AddressBytes(), []byte("waitingPubKey")))
	assert.Equal(t, peerAcc.GetList(), string(common.NewList))

	peerAcc, _ = s.getPeerAccount([]byte("stakedPubKey1"))
	assert.Equal(t, peerAcc.GetList(), string(common.LeavingList))

	assert.Equal(t, string(common.LeavingList), validatorsInfo.GetShardValidatorsInfoMap()[0][1].GetList())

	assert.Len(t, validatorsInfo.GetShardValidatorsInfoMap()[0], 5)
	assert.Equal(t, string(common.NewList), validatorsInfo.GetShardValidatorsInfoMap()[0][4].GetList())
}

func TestSystemSCProcessor_ProcessSystemSmartContractUnStakeTheOnlyNodeShouldWork(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())
	s, _ := NewSystemSCProcessor(args)

	prepareStakingContractWithData(
		args.UserAccountsDB,
		[]byte("stakedPubKey0"),
		[]byte("waitingPubKey"),
		args.Marshalizer,
		[]byte("rewardAddress"),
		[]byte("rewardAddress"),
	)

	stakingcommon.AddStakingData(args.UserAccountsDB, []byte("ownerKey"), []byte("ownerKey"), [][]byte{[]byte("stakedPubKey1")}, args.Marshalizer)
	addValidatorDataWithUnStakedKey(args.UserAccountsDB, []byte("ownerKey"), [][]byte{[]byte("stakedPubKey1")}, big.NewInt(1000), args.Marshalizer)
	_, _ = args.UserAccountsDB.Commit()

	validatorsInfo := state.NewShardValidatorsInfoMap()
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey0"),
		List:            string(common.EligibleList),
		RewardAddress:   []byte("rewardAddress"),
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey1"),
		List:            string(common.EligibleList),
		RewardAddress:   []byte("rewardAddress"),
		AccumulatedFees: big.NewInt(0),
	})

	args.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: 1, // disable stakingV2OwnerFlag
	})
	err := s.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
	assert.Nil(t, err)
}

func addDelegationData(
	accountsDB state.AccountsAdapter,
	delegation []byte,
	stakedKeys [][]byte,
	marshalizer marshal.Marshalizer,
) {
	delegatorSC := stakingcommon.LoadUserAccount(accountsDB, delegation)
	dStatus := &systemSmartContracts.DelegationContractStatus{
		StakedKeys:    make([]*systemSmartContracts.NodesData, 0),
		NotStakedKeys: make([]*systemSmartContracts.NodesData, 0),
		UnStakedKeys:  make([]*systemSmartContracts.NodesData, 0),
		NumUsers:      0,
	}

	for _, stakedKey := range stakedKeys {
		dStatus.StakedKeys = append(dStatus.StakedKeys, &systemSmartContracts.NodesData{BLSKey: stakedKey, SignedMsg: stakedKey})
	}

	marshaledData, _ := marshalizer.Marshal(dStatus)
	_ = delegatorSC.SaveKeyValue([]byte("delegationStatus"), marshaledData)
	_ = accountsDB.SaveAccount(delegatorSC)
}

func TestSystemSCProcessor_ProcessSystemSmartContractUnStakeFromDelegationContract(t *testing.T) {
	t.Parallel()

	args, scContainer := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())
	s, _ := NewSystemSCProcessor(args)

	delegationAddr := generateSecondDelegationAddress()

	contract, _ := scContainer.Get(vm.FirstDelegationSCAddress)
	_ = scContainer.Add(delegationAddr, contract)

	stakingcommon.AddStakingData(
		args.UserAccountsDB,
		delegationAddr,
		delegationAddr,
		[][]byte{[]byte("stakedPubKey1"), []byte("stakedPubKey2"), []byte("stakedPubKey3")},
		args.Marshalizer,
	)
	allKeys := [][]byte{[]byte("stakedPubKey0"), []byte("waitingPubKey"), []byte("stakedPubKey1"), []byte("stakedPubKey2"), []byte("stakedPubKey3")}
	stakingcommon.RegisterValidatorKeys(
		args.UserAccountsDB,
		delegationAddr,
		delegationAddr,
		allKeys,
		big.NewInt(3000),
		args.Marshalizer,
	)

	addDelegationData(args.UserAccountsDB, delegationAddr, allKeys, args.Marshalizer)
	_, _ = args.UserAccountsDB.Commit()

	validatorsInfo := state.NewShardValidatorsInfoMap()
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey0"),
		List:            string(common.EligibleList),
		RewardAddress:   delegationAddr,
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey1"),
		List:            string(common.EligibleList),
		RewardAddress:   delegationAddr,
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey2"),
		List:            string(common.WaitingList),
		RewardAddress:   delegationAddr,
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey3"),
		List:            string(common.WaitingList),
		RewardAddress:   delegationAddr,
		AccumulatedFees: big.NewInt(0),
	})
	for _, vInfo := range validatorsInfo.GetShardValidatorsInfoMap()[0] {
		jailedAcc, _ := args.PeerAccountsDB.LoadAccount(vInfo.GetPublicKey())
		_ = args.PeerAccountsDB.SaveAccount(jailedAcc)
	}

	args.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: 1, // disable stakingV2OwnerFlag
	})
	err := s.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
	assert.Nil(t, err)

	for _, vInfo := range validatorsInfo.GetShardValidatorsInfoMap()[0] {
		assert.NotEqual(t, string(common.NewList), vInfo.GetList())
	}

	peerAcc, _ := s.getPeerAccount([]byte("stakedPubKey2"))
	assert.Equal(t, peerAcc.GetList(), string(common.LeavingList))
	assert.Len(t, validatorsInfo.GetShardValidatorsInfoMap()[0], 4)

	delegationSC := stakingcommon.LoadUserAccount(args.UserAccountsDB, delegationAddr)
	marshalledData, _, err := delegationSC.RetrieveValue([]byte("delegationStatus"))
	assert.Nil(t, err)
	dStatus := &systemSmartContracts.DelegationContractStatus{
		StakedKeys:    make([]*systemSmartContracts.NodesData, 0),
		NotStakedKeys: make([]*systemSmartContracts.NodesData, 0),
		UnStakedKeys:  make([]*systemSmartContracts.NodesData, 0),
		NumUsers:      0,
	}
	_ = args.Marshalizer.Unmarshal(dStatus, marshalledData)

	assert.Equal(t, 2, len(dStatus.UnStakedKeys))
	assert.Equal(t, 3, len(dStatus.StakedKeys))
	assert.Equal(t, []byte("stakedPubKey2"), dStatus.UnStakedKeys[1].BLSKey)
}

func TestSystemSCProcessor_ProcessSystemSmartContractShouldUnStakeFromAdditionalQueueOnly(t *testing.T) {
	t.Parallel()

	args, scContainer := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())
	s, _ := NewSystemSCProcessor(args)

	delegationAddr := generateSecondDelegationAddress()

	contract, _ := scContainer.Get(vm.FirstDelegationSCAddress)
	_ = scContainer.Add(delegationAddr, contract)

	listOfKeysInWaiting := [][]byte{[]byte("waitingPubKey"), []byte("waitingPubKe1"), []byte("waitingPubKe2"), []byte("waitingPubKe3"), []byte("waitingPubKe4")}
	allStakedKeys := append(listOfKeysInWaiting, []byte("stakedPubKey0"), []byte("stakedPubKey1"), []byte("stakedPubKey2"), []byte("stakedPubKey3"))

	stakingcommon.RegisterValidatorKeys(args.UserAccountsDB, delegationAddr, delegationAddr, allStakedKeys, big.NewInt(4000), args.Marshalizer)
	stakingcommon.AddKeysToWaitingList(args.UserAccountsDB, listOfKeysInWaiting, args.Marshalizer, delegationAddr, delegationAddr)
	addDelegationData(args.UserAccountsDB, delegationAddr, allStakedKeys, args.Marshalizer)
	_, _ = args.UserAccountsDB.Commit()

	validatorsInfo := state.NewShardValidatorsInfoMap()
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey0"),
		List:            string(common.EligibleList),
		RewardAddress:   delegationAddr,
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey1"),
		List:            string(common.EligibleList),
		RewardAddress:   delegationAddr,
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey2"),
		List:            string(common.EligibleList),
		RewardAddress:   delegationAddr,
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey3"),
		List:            string(common.EligibleList),
		RewardAddress:   delegationAddr,
		AccumulatedFees: big.NewInt(0),
	})
	for _, vInfo := range validatorsInfo.GetShardValidatorsInfoMap()[0] {
		jailedAcc, _ := args.PeerAccountsDB.LoadAccount(vInfo.GetPublicKey())
		_ = args.PeerAccountsDB.SaveAccount(jailedAcc)
	}

	args.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: 1, // disable stakingV2OwnerFlag
	})
	err := s.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
	assert.Nil(t, err)

	for _, vInfo := range validatorsInfo.GetShardValidatorsInfoMap()[0] {
		assert.Equal(t, string(common.EligibleList), vInfo.GetList())
	}

	delegationSC := stakingcommon.LoadUserAccount(args.UserAccountsDB, delegationAddr)
	marshalledData, _, err := delegationSC.RetrieveValue([]byte("delegationStatus"))
	assert.Nil(t, err)
	dStatus := &systemSmartContracts.DelegationContractStatus{
		StakedKeys:    make([]*systemSmartContracts.NodesData, 0),
		NotStakedKeys: make([]*systemSmartContracts.NodesData, 0),
		UnStakedKeys:  make([]*systemSmartContracts.NodesData, 0),
		NumUsers:      0,
	}
	_ = args.Marshalizer.Unmarshal(dStatus, marshalledData)

	assert.Equal(t, 5, len(dStatus.UnStakedKeys))
	assert.Equal(t, 4, len(dStatus.StakedKeys))
}

func TestSystemSCProcessor_ProcessSystemSmartContractUnStakeFromAdditionalQueue(t *testing.T) {
	t.Parallel()

	args, scContainer := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())
	s, _ := NewSystemSCProcessor(args)

	delegationAddr := generateSecondDelegationAddress()

	contract, _ := scContainer.Get(vm.FirstDelegationSCAddress)
	_ = scContainer.Add(delegationAddr, contract)

	prepareStakingContractWithData(
		args.UserAccountsDB,
		[]byte("stakedPubKey0"),
		[]byte("waitingPubKey"),
		args.Marshalizer,
		delegationAddr,
		delegationAddr,
	)

	stakingcommon.AddStakingData(args.UserAccountsDB,
		delegationAddr,
		delegationAddr,
		[][]byte{[]byte("stakedPubKey1"), []byte("stakedPubKey2"), []byte("stakedPubKey3")},
		args.Marshalizer,
	)

	stakingcommon.AddValidatorData(args.UserAccountsDB, delegationAddr, [][]byte{[]byte("stakedPubKey1"), []byte("stakedPubKey2"), []byte("stakedPubKey3"), []byte("waitingPubKey")}, big.NewInt(10000), args.Marshalizer)
	addDelegationData(args.UserAccountsDB, delegationAddr, [][]byte{[]byte("stakedPubKey1"), []byte("stakedPubKey2"), []byte("stakedPubKey3"), []byte("waitingPubKey")}, args.Marshalizer)
	_, _ = args.UserAccountsDB.Commit()

	delegationAddr2 := generateSecondDelegationAddress()
	delegationAddr2[28] = 5 // the fifth delegation address
	_ = scContainer.Add(delegationAddr2, contract)

	listOfKeysInWaiting := [][]byte{[]byte("waitingPubKe1"), []byte("waitingPubKe2"), []byte("waitingPubKe3"), []byte("waitingPubKe4")}
	stakingcommon.AddKeysToWaitingList(args.UserAccountsDB, listOfKeysInWaiting, args.Marshalizer, delegationAddr2, delegationAddr2)
	stakingcommon.AddValidatorData(args.UserAccountsDB, delegationAddr2, listOfKeysInWaiting, big.NewInt(2000), args.Marshalizer)
	addDelegationData(args.UserAccountsDB, delegationAddr2, listOfKeysInWaiting, args.Marshalizer)
	_, _ = args.UserAccountsDB.Commit()

	validatorsInfo := state.NewShardValidatorsInfoMap()
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey0"),
		List:            string(common.EligibleList),
		RewardAddress:   delegationAddr,
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey1"),
		List:            string(common.EligibleList),
		RewardAddress:   delegationAddr,
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey2"),
		List:            string(common.EligibleList),
		RewardAddress:   delegationAddr,
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey3"),
		List:            string(common.EligibleList),
		RewardAddress:   delegationAddr,
		AccumulatedFees: big.NewInt(0),
	})
	for _, vInfo := range validatorsInfo.GetShardValidatorsInfoMap()[0] {
		peerAcc, _ := args.PeerAccountsDB.LoadAccount(vInfo.GetPublicKey())
		_ = args.PeerAccountsDB.SaveAccount(peerAcc)
	}
	args.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: 1, // disable stakingV2OwnerFlag
	})
	err := s.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
	assert.Nil(t, err)

	delegationSC := stakingcommon.LoadUserAccount(args.UserAccountsDB, delegationAddr2)
	marshalledData, _, err := delegationSC.RetrieveValue([]byte("delegationStatus"))
	assert.Nil(t, err)
	dStatus := &systemSmartContracts.DelegationContractStatus{
		StakedKeys:    make([]*systemSmartContracts.NodesData, 0),
		NotStakedKeys: make([]*systemSmartContracts.NodesData, 0),
		UnStakedKeys:  make([]*systemSmartContracts.NodesData, 0),
		NumUsers:      0,
	}
	_ = args.Marshalizer.Unmarshal(dStatus, marshalledData)

	assert.Equal(t, 2, len(dStatus.UnStakedKeys))
	assert.Equal(t, 2, len(dStatus.StakedKeys))
	assert.Equal(t, []byte("waitingPubKe4"), dStatus.UnStakedKeys[0].BLSKey)
	assert.Equal(t, []byte("waitingPubKe3"), dStatus.UnStakedKeys[1].BLSKey)

	stakingSCAcc := stakingcommon.LoadUserAccount(args.UserAccountsDB, vm.StakingSCAddress)
	marshaledData, _, _ := stakingSCAcc.RetrieveValue([]byte("waitingList"))
	waitingListHead := &systemSmartContracts.WaitingList{}
	_ = args.Marshalizer.Unmarshal(waitingListHead, marshaledData)
	assert.Equal(t, uint32(3), waitingListHead.Length)
}

func TestSystemSCProcessor_ProcessSystemSmartContractWrongValidatorInfoShouldBeCleaned(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())
	s, _ := NewSystemSCProcessor(args)

	prepareStakingContractWithData(
		args.UserAccountsDB,
		[]byte("oneAddress1"),
		[]byte("oneAddress2"),
		args.Marshalizer,
		[]byte("oneAddress1"),
		[]byte("oneAddress1"),
	)

	validatorsInfo := state.NewShardValidatorsInfoMap()
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey0"),
		List:            "",
		RewardAddress:   []byte("stakedPubKey0"),
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey1"),
		List:            "",
		RewardAddress:   []byte("stakedPubKey0"),
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey2"),
		List:            "",
		RewardAddress:   []byte("stakedPubKey0"),
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey3"),
		List:            "",
		RewardAddress:   []byte("stakedPubKey0"),
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("oneAddress1"),
		List:            string(common.EligibleList),
		RewardAddress:   []byte("oneAddress1"),
		AccumulatedFees: big.NewInt(0),
	})

	err := s.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
	assert.Nil(t, err)

	assert.Len(t, validatorsInfo.GetShardValidatorsInfoMap()[0], 1)
}

func TestSystemSCProcessor_TogglePauseUnPause(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())
	s, _ := NewSystemSCProcessor(args)

	err := s.ToggleUnStakeUnBond(true)
	assert.Nil(t, err)

	validatorSC := stakingcommon.LoadUserAccount(s.userAccountsDB, vm.ValidatorSCAddress)
	value, _, _ := validatorSC.RetrieveValue([]byte("unStakeUnBondPause"))
	assert.True(t, value[0] == 1)

	err = s.ToggleUnStakeUnBond(false)
	assert.Nil(t, err)

	validatorSC = stakingcommon.LoadUserAccount(s.userAccountsDB, vm.ValidatorSCAddress)
	value, _, _ = validatorSC.RetrieveValue([]byte("unStakeUnBondPause"))
	assert.True(t, value[0] == 0)
}

func TestSystemSCProcessor_ResetUnJailListErrors(t *testing.T) {
	t.Parallel()

	localErr := errors.New("local error")
	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())
	s, _ := NewSystemSCProcessor(args)
	s.systemVM = &mock.VMExecutionHandlerStub{RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
		return nil, localErr
	}}

	err := s.resetLastUnJailed()
	assert.Equal(t, localErr, err)

	s.systemVM = &mock.VMExecutionHandlerStub{RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
		return &vmcommon.VMOutput{ReturnCode: vmcommon.UserError}, nil
	}}

	err = s.resetLastUnJailed()
	assert.Equal(t, epochStart.ErrResetLastUnJailedFromQueue, err)
}

func TestSystemSCProcessor_ProcessSystemSmartContractJailAndUnStake(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())
	s, _ := NewSystemSCProcessor(args)

	stakingcommon.AddStakingData(args.UserAccountsDB,
		[]byte("ownerKey"),
		[]byte("ownerKey"),
		[][]byte{[]byte("stakedPubKey0"), []byte("stakedPubKey1"), []byte("stakedPubKey2"), []byte("stakedPubKey3")},
		args.Marshalizer,
	)
	stakingcommon.AddKeysToWaitingList(args.UserAccountsDB, [][]byte{[]byte("waitingPubKey")}, args.Marshalizer, []byte("ownerKey"), []byte("ownerKey"))
	stakingcommon.AddValidatorData(args.UserAccountsDB, []byte("ownerKey"), [][]byte{[]byte("stakedPubKey0"), []byte("stakedPubKey1"), []byte("stakedPubKey2"), []byte("stakedPubKey3"), []byte("waitingPubKey")}, big.NewInt(0), args.Marshalizer)
	_, _ = args.UserAccountsDB.Commit()

	validatorsInfo := state.NewShardValidatorsInfoMap()
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey0"),
		List:            string(common.EligibleList),
		RewardAddress:   []byte("ownerKey"),
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey1"),
		List:            string(common.EligibleList),
		RewardAddress:   []byte("ownerKey"),
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey2"),
		List:            string(common.EligibleList),
		RewardAddress:   []byte("ownerKey"),
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("stakedPubKey3"),
		List:            string(common.EligibleList),
		RewardAddress:   []byte("ownerKey"),
		AccumulatedFees: big.NewInt(0),
	})

	for _, vInfo := range validatorsInfo.GetShardValidatorsInfoMap()[0] {
		jailedAcc, _ := args.PeerAccountsDB.LoadAccount(vInfo.GetPublicKey())
		_ = args.PeerAccountsDB.SaveAccount(jailedAcc)
	}

	args.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: 1, // disable stakingV2OwnerFlag
	})
	err := s.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
	assert.Nil(t, err)

	_, err = s.peerAccountsDB.GetExistingAccount([]byte("waitingPubKey"))
	assert.NotNil(t, err)

	assert.Len(t, validatorsInfo.GetShardValidatorsInfoMap()[0], 4)
	for _, vInfo := range validatorsInfo.GetShardValidatorsInfoMap()[0] {
		assert.Equal(t, vInfo.GetList(), string(common.LeavingList))
		peerAcc, _ := s.getPeerAccount(vInfo.GetPublicKey())
		assert.Equal(t, peerAcc.GetList(), string(common.LeavingList))
	}
}

func TestSystemSCProcessor_ProcessSystemSmartContractSwapJailedWithWaiting(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())
	args.ChanceComputer = &mock.ChanceComputerStub{
		GetChanceCalled: func(rating uint32) uint32 {
			if rating == 0 {
				return 10
			}
			return rating
		},
	}
	s, _ := NewSystemSCProcessor(args)

	prepareStakingContractWithData(
		args.UserAccountsDB,
		[]byte("jailedPubKey0"),
		[]byte("waitingPubKey"),
		args.Marshalizer,
		[]byte("rewardAddress"),
		[]byte("rewardAddress"),
	)
	jailedAcc, _ := args.PeerAccountsDB.LoadAccount([]byte("jailedPubKey0"))
	_ = args.PeerAccountsDB.SaveAccount(jailedAcc)

	validatorsInfo := state.NewShardValidatorsInfoMap()
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey:       []byte("jailedPubKey0"),
		ShardId:         0,
		List:            string(common.JailedList),
		TempRating:      1,
		RewardAddress:   []byte("address"),
		AccumulatedFees: big.NewInt(0),
	})
	_ = validatorsInfo.Add(&state.ValidatorInfo{
		PublicKey: []byte("waitingPubKey"),
		ShardId:   0,
		List:      string(common.WaitingList),
	})

	err := s.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
	assert.Nil(t, err)

	require.Len(t, validatorsInfo.GetShardValidatorsInfoMap()[0], 2)
	newValidatorInfo := validatorsInfo.GetShardValidatorsInfoMap()[0][0]
	require.Equal(t, newValidatorInfo.GetList(), string(common.NewList))
}

func TestSystemSCProcessor_ProcessSystemSmartContractStakingV4Init(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())
	s, _ := NewSystemSCProcessor(args)

	owner1 := []byte("owner1")
	owner2 := []byte("owner2")
	owner3 := []byte("owner3")

	owner1ListPubKeysWaiting := [][]byte{[]byte("waitingPubKe0"), []byte("waitingPubKe1"), []byte("waitingPubKe2")}
	owner1ListPubKeysStaked := [][]byte{[]byte("stakedPubKey0"), []byte("stakedPubKey1")}
	owner1AllPubKeys := append(owner1ListPubKeysWaiting, owner1ListPubKeysStaked...)

	owner2ListPubKeysWaiting := [][]byte{[]byte("waitingPubKe3"), []byte("waitingPubKe4")}
	owner2ListPubKeysStaked := [][]byte{[]byte("stakedPubKey2")}
	owner2AllPubKeys := append(owner2ListPubKeysWaiting, owner2ListPubKeysStaked...)

	owner3ListPubKeysWaiting := [][]byte{[]byte("waitingPubKe5"), []byte("waitingPubKe6")}

	// Owner1 has 2 staked nodes (one eligible, one waiting) in shard0 + 3 nodes in staking queue.
	// It has enough stake so that all his staking queue nodes will be selected in the auction list
	stakingcommon.AddKeysToWaitingList(args.UserAccountsDB, owner1ListPubKeysWaiting, args.Marshalizer, owner1, owner1)
	stakingcommon.RegisterValidatorKeys(args.UserAccountsDB, owner1, owner1, owner1AllPubKeys, big.NewInt(5000), args.Marshalizer)

	// Owner2 has 1 staked node (eligible) in shard1 + 2 nodes in staking queue.
	// It has enough stake for only ONE node from staking queue to be selected in the auction list
	stakingcommon.AddKeysToWaitingList(args.UserAccountsDB, owner2ListPubKeysWaiting, args.Marshalizer, owner2, owner2)
	stakingcommon.RegisterValidatorKeys(args.UserAccountsDB, owner2, owner2, owner2AllPubKeys, big.NewInt(2500), args.Marshalizer)

	// Owner3 has 0 staked node + 2 nodes in staking queue.
	// It has enough stake so that all his staking queue nodes will be selected in the auction list
	stakingcommon.AddKeysToWaitingList(args.UserAccountsDB, owner3ListPubKeysWaiting, args.Marshalizer, owner3, owner3)
	stakingcommon.RegisterValidatorKeys(args.UserAccountsDB, owner3, owner3, owner3ListPubKeysWaiting, big.NewInt(2000), args.Marshalizer)

	validatorsInfo := state.NewShardValidatorsInfoMap()
	_ = validatorsInfo.Add(createValidatorInfo(owner1ListPubKeysStaked[0], common.EligibleList, "", 0, owner1))
	_ = validatorsInfo.Add(createValidatorInfo(owner1ListPubKeysStaked[1], common.WaitingList, "", 0, owner1))
	_ = validatorsInfo.Add(createValidatorInfo(owner2ListPubKeysStaked[0], common.EligibleList, "", 1, owner2))

	args.EpochNotifier.CheckEpoch(&block.Header{Epoch: stakingV4Step1EnableEpoch})
	err := s.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
	require.Nil(t, err)

	expectedValidatorsInfo := map[uint32][]state.ValidatorInfoHandler{
		0: {
			createValidatorInfo(owner1ListPubKeysStaked[0], common.EligibleList, "", 0, owner1),
			createValidatorInfo(owner1ListPubKeysStaked[1], common.WaitingList, "", 0, owner1),
		},
		1: {
			createValidatorInfo(owner2ListPubKeysStaked[0], common.EligibleList, "", 1, owner2),
		},
	}

	require.Equal(t, expectedValidatorsInfo, validatorsInfo.GetShardValidatorsInfoMap())
}

func TestSystemSCProcessor_ProcessSystemSmartContractStakingV4EnabledCannotPrepareStakingData(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())

	errProcessStakingData := errors.New("error processing staking data")
	args.StakingDataProvider = &stakingcommon.StakingDataProviderStub{
		PrepareStakingDataCalled: func(validatorsMap state.ShardValidatorsInfoMapHandler) error {
			return errProcessStakingData
		},
	}

	owner := []byte("owner")
	ownerStakedKeys := [][]byte{[]byte("pubKey0"), []byte("pubKey1")}
	stakingcommon.RegisterValidatorKeys(args.UserAccountsDB, owner, owner, ownerStakedKeys, big.NewInt(2000), args.Marshalizer)

	validatorsInfo := state.NewShardValidatorsInfoMap()
	_ = validatorsInfo.Add(createValidatorInfo(ownerStakedKeys[0], common.AuctionList, "", 0, owner))
	_ = validatorsInfo.Add(createValidatorInfo(ownerStakedKeys[1], common.AuctionList, "", 0, owner))

	s, _ := NewSystemSCProcessor(args)
	s.EpochConfirmed(stakingV4Step2EnableEpoch, 0)

	err := s.ProcessSystemSmartContract(validatorsInfo, &block.Header{})
	require.Equal(t, errProcessStakingData, err)
}

func TestSystemSCProcessor_ProcessSystemSmartContractStakingV4Enabled(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())
	nodesConfigProvider, _ := notifier.NewNodesConfigProvider(args.EpochNotifier, []config.MaxNodesChangeConfig{{MaxNumNodes: 8}})

	auctionCfg := config.SoftAuctionConfig{
		TopUpStep:             "10",
		MinTopUp:              "1",
		MaxTopUp:              "32000000",
		MaxNumberOfIterations: 100000,
	}
	ald, _ := NewAuctionListDisplayer(ArgsAuctionListDisplayer{
		TableDisplayHandler:      NewTableDisplayer(),
		ValidatorPubKeyConverter: &testscommon.PubkeyConverterMock{},
		AddressPubKeyConverter:   &testscommon.PubkeyConverterMock{},
		AuctionConfig:            auctionCfg,
		Denomination:             0,
	})

	argsAuctionListSelector := AuctionListSelectorArgs{
		ShardCoordinator:             args.ShardCoordinator,
		StakingDataProvider:          args.StakingDataProvider,
		MaxNodesChangeConfigProvider: nodesConfigProvider,
		SoftAuctionConfig:            auctionCfg,
		AuctionListDisplayHandler:    ald,
	}
	als, _ := NewAuctionListSelector(argsAuctionListSelector)
	args.AuctionListSelector = als

	owner1 := []byte("owner1")
	owner2 := []byte("owner2")
	owner3 := []byte("owner3")
	owner4 := []byte("owner4")
	owner5 := []byte("owner5")
	owner6 := []byte("owner6")
	owner7 := []byte("owner7")

	owner1StakedKeys := [][]byte{[]byte("pubKey0"), []byte("pubKey1"), []byte("pubKey2")}
	owner2StakedKeys := [][]byte{[]byte("pubKey3"), []byte("pubKey4"), []byte("pubKey5")}
	owner3StakedKeys := [][]byte{[]byte("pubKey6"), []byte("pubKey7")}
	owner4StakedKeys := [][]byte{[]byte("pubKey8"), []byte("pubKey9"), []byte("pubKe10"), []byte("pubKe11")}
	owner5StakedKeys := [][]byte{[]byte("pubKe12"), []byte("pubKe13")}
	owner6StakedKeys := [][]byte{[]byte("pubKe14"), []byte("pubKe15")}
	owner7StakedKeys := [][]byte{[]byte("pubKe16"), []byte("pubKe17")}

	stakingcommon.RegisterValidatorKeys(args.UserAccountsDB, owner1, owner1, owner1StakedKeys, big.NewInt(6666), args.Marshalizer)
	stakingcommon.RegisterValidatorKeys(args.UserAccountsDB, owner2, owner2, owner2StakedKeys, big.NewInt(5555), args.Marshalizer)
	stakingcommon.RegisterValidatorKeys(args.UserAccountsDB, owner3, owner3, owner3StakedKeys, big.NewInt(4444), args.Marshalizer)
	stakingcommon.RegisterValidatorKeys(args.UserAccountsDB, owner4, owner4, owner4StakedKeys, big.NewInt(6666), args.Marshalizer)
	stakingcommon.RegisterValidatorKeys(args.UserAccountsDB, owner5, owner5, owner5StakedKeys, big.NewInt(1500), args.Marshalizer)
	stakingcommon.RegisterValidatorKeys(args.UserAccountsDB, owner6, owner6, owner6StakedKeys, big.NewInt(1500), args.Marshalizer)
	stakingcommon.RegisterValidatorKeys(args.UserAccountsDB, owner7, owner7, owner7StakedKeys, big.NewInt(1500), args.Marshalizer)

	validatorsInfo := state.NewShardValidatorsInfoMap()
	_ = validatorsInfo.Add(createValidatorInfo(owner1StakedKeys[0], common.EligibleList, "", 0, owner1))
	_ = validatorsInfo.Add(createValidatorInfo(owner1StakedKeys[1], common.WaitingList, "", 0, owner1))
	_ = validatorsInfo.Add(createValidatorInfo(owner1StakedKeys[2], common.AuctionList, "", 0, owner1))

	_ = validatorsInfo.Add(createValidatorInfo(owner2StakedKeys[0], common.EligibleList, "", 1, owner2))
	_ = validatorsInfo.Add(createValidatorInfo(owner2StakedKeys[1], common.AuctionList, "", 1, owner2))
	_ = validatorsInfo.Add(createValidatorInfo(owner2StakedKeys[2], common.AuctionList, "", 1, owner2))

	_ = validatorsInfo.Add(createValidatorInfo(owner3StakedKeys[0], common.LeavingList, "", 1, owner3))
	_ = validatorsInfo.Add(createValidatorInfo(owner3StakedKeys[1], common.AuctionList, "", 1, owner3))

	_ = validatorsInfo.Add(createValidatorInfo(owner4StakedKeys[0], common.JailedList, "", 1, owner4))
	_ = validatorsInfo.Add(createValidatorInfo(owner4StakedKeys[1], common.AuctionList, "", 1, owner4))
	_ = validatorsInfo.Add(createValidatorInfo(owner4StakedKeys[2], common.AuctionList, "", 1, owner4))
	_ = validatorsInfo.Add(createValidatorInfo(owner4StakedKeys[3], common.AuctionList, "", 1, owner4))

	_ = validatorsInfo.Add(createValidatorInfo(owner5StakedKeys[0], common.EligibleList, "", 1, owner5))
	_ = validatorsInfo.Add(createValidatorInfo(owner5StakedKeys[1], common.AuctionList, "", 1, owner5))

	_ = validatorsInfo.Add(createValidatorInfo(owner6StakedKeys[0], common.AuctionList, "", 1, owner6))
	_ = validatorsInfo.Add(createValidatorInfo(owner6StakedKeys[1], common.AuctionList, "", 1, owner6))

	_ = validatorsInfo.Add(createValidatorInfo(owner7StakedKeys[0], common.EligibleList, "", 2, owner7))
	_ = validatorsInfo.Add(createValidatorInfo(owner7StakedKeys[1], common.EligibleList, "", 2, owner7))

	s, _ := NewSystemSCProcessor(args)
	args.EpochNotifier.CheckEpoch(&block.Header{Epoch: stakingV4Step2EnableEpoch})
	err := s.ProcessSystemSmartContract(validatorsInfo, &block.Header{PrevRandSeed: []byte("pubKey7")})
	require.Nil(t, err)

	/*
			- owner5 does not have enough stake for 2 nodes=> his auction node (pubKe13) will be unStaked at the end of the epoch =>
		      will not participate in auction selection
		    - owner6 does not have enough stake for 2 nodes => one of his auction nodes(pubKey14) will be unStaked at the end of the epoch =>
		      his other auction node(pubKey15) will not participate in auction selection
			- MaxNumNodes = 8
			- EligibleBlsKeys = 5 (pubKey0, pubKey1, pubKey3, pubKe13, pubKey17)
			- QualifiedAuctionBlsKeys = 7 (pubKey2, pubKey4, pubKey5, pubKey7, pubKey9, pubKey10, pubKey11)
			We can only select (MaxNumNodes - EligibleBlsKeys = 3) bls keys from AuctionList to be added to NewList

			-> Initial nodes config in auction list is:
		+--------+------------------+------------------+-------------------+--------------+-----------------+---------------------------+
		| Owner  | Num staked nodes | Num active nodes | Num auction nodes | Total top up | Top up per node | Auction list nodes        |
		+--------+------------------+------------------+-------------------+--------------+-----------------+---------------------------+
		| owner3 | 2                | 1                | 1                 | 2444         | 1222            | pubKey7                   |
		| owner4 | 4                | 1                | 3                 | 2666         | 666             | pubKey9, pubKe10, pubKe11 |
		| owner1 | 3                | 2                | 1                 | 3666         | 1222            | pubKey2                   |
		| owner2 | 3                | 1                | 2                 | 2555         | 851             | pubKey4, pubKey5          |
		+--------+------------------+------------------+-------------------+--------------+-----------------+---------------------------+
		  	-> Min possible topUp = 666; max possible topUp = 1333, min required topUp = 1216
			-> Selected nodes config in auction list. For each owner's auction nodes, qualified ones are selected by sorting the bls keys
		+--------+------------------+----------------+--------------+-------------------+-----------------------------+------------------+---------------------------+-----------------------------+
		| Owner  | Num staked nodes | TopUp per node | Total top up | Num auction nodes | Num qualified auction nodes | Num active nodes | Qualified top up per node | Selected auction list nodes |
		+--------+------------------+----------------+--------------+-------------------+-----------------------------+------------------+---------------------------+-----------------------------+
		| owner1 | 3                | 1222           | 3666         | 1                 | 1                           | 2                | 1222                      | pubKey2                     |
		| owner2 | 3                | 851            | 2555         | 2                 | 1                           | 1                | 1277                      | pubKey5                     |
		| owner3 | 2                | 1222           | 2444         | 1                 | 1                           | 1                | 1222                      | pubKey7                     |
		| owner4 | 4                | 666            | 2666         | 3                 | 1                           | 1                | 1333                      | pubKey9                     |
		+--------+------------------+----------------+--------------+-------------------+-----------------------------+------------------+---------------------------+-----------------------------+
			-> Final selected nodes from auction list
		+--------+----------------+--------------------------+
		| Owner  | Registered key | Qualified TopUp per node |
		+--------+----------------+--------------------------+
		| owner4 | pubKey9        | 1333                     |
		| owner2 | pubKey5        | 1277                     |
		| owner1 | pubKey2        | 1222                     |
		+--------+----------------+--------------------------+
		| owner3 | pubKey7        | 1222                     |
		+--------+----------------+--------------------------+

			The following have 1222 top up per node:
			- owner1 with 1 bls key = pubKey2
			- owner3 with 1 bls key = pubKey7

			Since randomness = []byte("pubKey7"), nodes will be sorted based on blsKey XOR randomness, therefore:
			-  XOR1 = []byte("pubKey2") XOR []byte("pubKey7") = [0 0 0 0 0 0 5]
			-  XOR3 = []byte("pubKey7") XOR []byte("pubKey7") = [0 0 0 0 0 0 0]
	*/
	requireTopUpPerNodes(t, s.stakingDataProvider, owner1StakedKeys, big.NewInt(1222))
	requireTopUpPerNodes(t, s.stakingDataProvider, owner2StakedKeys, big.NewInt(851))
	requireTopUpPerNodes(t, s.stakingDataProvider, owner3StakedKeys, big.NewInt(1222))
	requireTopUpPerNodes(t, s.stakingDataProvider, owner4StakedKeys, big.NewInt(666))
	requireTopUpPerNodes(t, s.stakingDataProvider, owner5StakedKeys, big.NewInt(0))
	requireTopUpPerNodes(t, s.stakingDataProvider, owner6StakedKeys, big.NewInt(0))
	requireTopUpPerNodes(t, s.stakingDataProvider, owner7StakedKeys, big.NewInt(0))

	expectedValidatorsInfo := map[uint32][]state.ValidatorInfoHandler{
		0: {
			createValidatorInfo(owner1StakedKeys[0], common.EligibleList, "", 0, owner1),
			createValidatorInfo(owner1StakedKeys[1], common.WaitingList, "", 0, owner1),
			createValidatorInfo(owner1StakedKeys[2], common.SelectedFromAuctionList, common.AuctionList, 0, owner1),
		},
		1: {
			createValidatorInfo(owner2StakedKeys[0], common.EligibleList, "", 1, owner2),
			createValidatorInfo(owner2StakedKeys[1], common.AuctionList, "", 1, owner2),
			createValidatorInfo(owner2StakedKeys[2], common.SelectedFromAuctionList, common.AuctionList, 1, owner2),

			createValidatorInfo(owner3StakedKeys[0], common.LeavingList, "", 1, owner3),
			createValidatorInfo(owner3StakedKeys[1], common.AuctionList, "", 1, owner3),

			createValidatorInfo(owner4StakedKeys[0], common.JailedList, "", 1, owner4),
			createValidatorInfo(owner4StakedKeys[1], common.SelectedFromAuctionList, common.AuctionList, 1, owner4),
			createValidatorInfo(owner4StakedKeys[2], common.AuctionList, "", 1, owner4),
			createValidatorInfo(owner4StakedKeys[3], common.AuctionList, "", 1, owner4),

			createValidatorInfo(owner5StakedKeys[0], common.EligibleList, "", 1, owner5),
			createValidatorInfo(owner5StakedKeys[1], common.LeavingList, common.AuctionList, 1, owner5),

			createValidatorInfo(owner6StakedKeys[0], common.LeavingList, common.AuctionList, 1, owner6),
			createValidatorInfo(owner6StakedKeys[1], common.AuctionList, "", 1, owner6),
		},
		2: {
			createValidatorInfo(owner7StakedKeys[0], common.LeavingList, common.EligibleList, 2, owner7),
			createValidatorInfo(owner7StakedKeys[1], common.EligibleList, "", 2, owner7),
		},
	}

	require.Equal(t, expectedValidatorsInfo, validatorsInfo.GetShardValidatorsInfoMap())
}

func TestSystemSCProcessor_LegacyEpochConfirmedCorrectMaxNumNodesAfterNodeRestart(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())
	nodesConfigEpoch0 := config.MaxNodesChangeConfig{
		EpochEnable:            0,
		MaxNumNodes:            36,
		NodesToShufflePerShard: 4,
	}
	nodesConfigEpoch1 := config.MaxNodesChangeConfig{
		EpochEnable:            1,
		MaxNumNodes:            56,
		NodesToShufflePerShard: 2,
	}
	nodesConfigEpoch6 := config.MaxNodesChangeConfig{
		EpochEnable:            6,
		MaxNumNodes:            48,
		NodesToShufflePerShard: 1,
	}
	nodesConfigProvider, _ := notifier.NewNodesConfigProvider(
		args.EpochNotifier,
		[]config.MaxNodesChangeConfig{
			nodesConfigEpoch0,
			nodesConfigEpoch1,
			nodesConfigEpoch6,
		})
	args.MaxNodesChangeConfigProvider = nodesConfigProvider
	args.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.StakingV2Flag)
	validatorsInfoMap := state.NewShardValidatorsInfoMap()
	s, _ := NewSystemSCProcessor(args)

	args.EpochNotifier.CheckEpoch(&block.Header{Epoch: 0, Nonce: 0})
	require.True(t, s.flagChangeMaxNodesEnabled.IsSet())
	err := s.processLegacy(validatorsInfoMap, 0, 0)
	require.Nil(t, err)
	require.Equal(t, nodesConfigEpoch0.MaxNumNodes, s.maxNodes)

	args.EpochNotifier.CheckEpoch(&block.Header{Epoch: 1, Nonce: 1})
	require.True(t, s.flagChangeMaxNodesEnabled.IsSet())
	err = s.processLegacy(validatorsInfoMap, 1, 1)
	require.Nil(t, err)
	require.Equal(t, nodesConfigEpoch1.MaxNumNodes, s.maxNodes)

	for epoch := uint32(2); epoch <= 5; epoch++ {
		args.EpochNotifier.CheckEpoch(&block.Header{Epoch: epoch, Nonce: uint64(epoch)})
		require.False(t, s.flagChangeMaxNodesEnabled.IsSet())
		err = s.processLegacy(validatorsInfoMap, uint64(epoch), epoch)
		require.Nil(t, err)
		require.Equal(t, nodesConfigEpoch1.MaxNumNodes, s.maxNodes)
	}

	// simulate restart
	args.EpochNotifier.CheckEpoch(&block.Header{Epoch: 0, Nonce: 0})
	args.EpochNotifier.CheckEpoch(&block.Header{Epoch: 5, Nonce: 5})
	require.False(t, s.flagChangeMaxNodesEnabled.IsSet())
	err = s.processLegacy(validatorsInfoMap, 5, 5)
	require.Nil(t, err)
	require.Equal(t, nodesConfigEpoch1.MaxNumNodes, s.maxNodes)

	args.EpochNotifier.CheckEpoch(&block.Header{Epoch: 6, Nonce: 6})
	require.True(t, s.flagChangeMaxNodesEnabled.IsSet())
	err = s.processLegacy(validatorsInfoMap, 6, 6)
	require.Equal(t, epochStart.ErrInvalidMaxNumberOfNodes, err)

	args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub).AddActiveFlags(common.StakingV4StartedFlag)
	err = s.processLegacy(validatorsInfoMap, 6, 6)
	require.Nil(t, err)
	require.Equal(t, nodesConfigEpoch6.MaxNumNodes, s.maxNodes)

	// simulate restart
	args.EpochNotifier.CheckEpoch(&block.Header{Epoch: 0, Nonce: 0})
	args.EpochNotifier.CheckEpoch(&block.Header{Epoch: 6, Nonce: 6})
	require.True(t, s.flagChangeMaxNodesEnabled.IsSet())
	err = s.processLegacy(validatorsInfoMap, 6, 6)
	require.Nil(t, err)
	require.Equal(t, nodesConfigEpoch6.MaxNumNodes, s.maxNodes)

	for epoch := uint32(7); epoch <= 20; epoch++ {
		args.EpochNotifier.CheckEpoch(&block.Header{Epoch: epoch, Nonce: uint64(epoch)})
		require.False(t, s.flagChangeMaxNodesEnabled.IsSet())
		err = s.processLegacy(validatorsInfoMap, uint64(epoch), epoch)
		require.Nil(t, err)
		require.Equal(t, nodesConfigEpoch6.MaxNumNodes, s.maxNodes)
	}

	// simulate restart
	args.EpochNotifier.CheckEpoch(&block.Header{Epoch: 1, Nonce: 1})
	args.EpochNotifier.CheckEpoch(&block.Header{Epoch: 21, Nonce: 21})
	require.False(t, s.flagChangeMaxNodesEnabled.IsSet())
	err = s.processLegacy(validatorsInfoMap, 21, 21)
	require.Nil(t, err)
	require.Equal(t, nodesConfigEpoch6.MaxNumNodes, s.maxNodes)
}

func TestSystemSCProcessor_ProcessSystemSmartContractNilInputValues(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())
	s, _ := NewSystemSCProcessor(args)

	t.Run("nil validators info map, expect error", func(t *testing.T) {
		t.Parallel()

		blockHeader := &block.Header{Nonce: 4}
		err := s.ProcessSystemSmartContract(nil, blockHeader)
		require.True(t, strings.Contains(err.Error(), errNilValidatorsInfoMap.Error()))
		require.True(t, strings.Contains(err.Error(), fmt.Sprintf("%d", blockHeader.GetNonce())))
	})

	t.Run("nil header, expect error", func(t *testing.T) {
		t.Parallel()

		validatorsInfoMap := state.NewShardValidatorsInfoMap()
		err := s.ProcessSystemSmartContract(validatorsInfoMap, nil)
		require.Equal(t, process.ErrNilHeaderHandler, err)
	})
}

func TestLegacySystemSCProcessor_addNewlyStakedNodesToValidatorTrie(t *testing.T) {
	t.Parallel()

	args, _ := createFullArgumentsForSystemSCProcessing(config.EnableEpochs{}, testscommon.CreateMemUnit())
	sysProc, _ := NewSystemSCProcessor(args)

	pubKey := []byte("pubKey")
	existingValidator := &state.ValidatorInfo{
		PublicKey: pubKey,
		List:      "inactive",
	}

	nonce := uint64(4)
	newList := common.AuctionList
	newlyAddedValidator := &state.ValidatorInfo{
		PublicKey:       pubKey,
		List:            string(newList),
		Index:           uint32(nonce),
		TempRating:      sysProc.startRating,
		Rating:          sysProc.startRating,
		RewardAddress:   pubKey,
		AccumulatedFees: big.NewInt(0),
	}

	// Check before stakingV4, we should have both validators
	validatorsInfo := state.NewShardValidatorsInfoMap()
	_ = validatorsInfo.Add(existingValidator)
	args.EpochNotifier.CheckEpoch(&block.Header{Epoch: stakingV4Step1EnableEpoch - 1, Nonce: 1})
	err := sysProc.addNewlyStakedNodesToValidatorTrie(
		validatorsInfo,
		[][]byte{pubKey, pubKey},
		nonce,
		newList,
	)
	require.Nil(t, err)
	require.Equal(t, map[uint32][]state.ValidatorInfoHandler{
		0: {existingValidator, newlyAddedValidator},
	}, validatorsInfo.GetShardValidatorsInfoMap())

	// Check after stakingV4, we should only have the new one
	validatorsInfo = state.NewShardValidatorsInfoMap()
	_ = validatorsInfo.Add(existingValidator)
	args.EpochNotifier.CheckEpoch(&block.Header{Epoch: stakingV4Step1EnableEpoch, Nonce: 1})
	err = sysProc.addNewlyStakedNodesToValidatorTrie(
		validatorsInfo,
		[][]byte{pubKey, pubKey},
		nonce,
		newList,
	)
	require.Nil(t, err)
	require.Equal(t, map[uint32][]state.ValidatorInfoHandler{
		0: {newlyAddedValidator},
	}, validatorsInfo.GetShardValidatorsInfoMap())
}

func requireTopUpPerNodes(t *testing.T, s epochStart.StakingDataProvider, stakedPubKeys [][]byte, topUp *big.Int) {
	for _, pubKey := range stakedPubKeys {
		owner, err := s.GetBlsKeyOwner(pubKey)
		require.Nil(t, err)

		totalTopUp := s.GetOwnersData()[owner].TotalTopUp
		topUpPerNode := big.NewInt(0).Div(totalTopUp, big.NewInt(int64(len(stakedPubKeys))))
		require.Equal(t, topUp, topUpPerNode)
	}
}

// This func sets rating and temp rating with the start rating value used in createFullArgumentsForSystemSCProcessing
func createValidatorInfo(pubKey []byte, list common.PeerType, previousList common.PeerType, shardID uint32, owner []byte) *state.ValidatorInfo {
	rating := uint32(5)

	return &state.ValidatorInfo{
		PublicKey:       pubKey,
		List:            string(list),
		PreviousList:    string(previousList),
		ShardId:         shardID,
		RewardAddress:   owner,
		AccumulatedFees: zero,
		Rating:          rating,
		TempRating:      rating,
	}
}
