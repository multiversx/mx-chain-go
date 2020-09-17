package metachain

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/genesis/process/disabled"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/marshal"
	economics2 "github.com/ElrondNetwork/elrond-go/process/economics"
	vmFactory "github.com/ElrondNetwork/elrond-go/process/factory"
	metaProcess "github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/peer"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts/defaults"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemSCProcessor_ProcessSystemSmartContract(t *testing.T) {
	t.Parallel()

	args := createFullArgumentsForSystemSCProcessing()
	args.ChanceComputer = &mock.ChanceComputerStub{
		GetChanceCalled: func(rating uint32) uint32 {
			if rating == 0 {
				return 10
			}
			return rating
		},
	}
	s, _ := NewSystemSCProcessor(args)

	prepareStakingContractWithData(args.UserAccountsDB, []byte("jailedPubKey0"), []byte("waitingPubKey"), args.Marshalizer)
	jailedAcc, _ := args.PeerAccountsDB.LoadAccount([]byte("jailedPubKey0"))
	_ = args.PeerAccountsDB.SaveAccount(jailedAcc)

	validatorInfos := make(map[uint32][]*state.ValidatorInfo)
	vInfo := &state.ValidatorInfo{
		PublicKey:       []byte("jailedPubKey0"),
		ShardId:         0,
		List:            string(core.JailedList),
		TempRating:      1,
		RewardAddress:   []byte("address"),
		AccumulatedFees: big.NewInt(0),
	}
	validatorInfos[0] = append(validatorInfos[0], vInfo)
	err := s.ProcessSystemSmartContract(validatorInfos)
	assert.Nil(t, err)

	assert.Equal(t, len(validatorInfos[0]), 1)
	newValidatorInfo := validatorInfos[0][0]
	assert.Equal(t, newValidatorInfo.List, string(core.NewList))
}

func TestSystemSCProcessor_JailedNodesShouldNotBeSwappedAllAtOnce(t *testing.T) {
	t.Parallel()

	args := createFullArgumentsForSystemSCProcessing()
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
	stakingScAcc := createStakingScAcc(args.UserAccountsDB)
	createEligibleNodes(numEligible, stakingScAcc, args.Marshalizer)
	_ = createWaitingNodes(numWaiting, stakingScAcc, args.UserAccountsDB, args.Marshalizer)
	jailed := createJailedNodes(numJailed, stakingScAcc, args.UserAccountsDB, args.PeerAccountsDB, args.Marshalizer)
	validatorsInfo := make(map[uint32][]*state.ValidatorInfo)
	validatorsInfo[0] = append(validatorsInfo[0], jailed...)

	err := s.ProcessSystemSmartContract(validatorsInfo)
	assert.Nil(t, err)
	for i := 0; i < numWaiting; i++ {
		assert.Equal(t, string(core.NewList), validatorsInfo[0][i].List)
	}
	for i := numWaiting; i < numJailed; i++ {
		assert.Equal(t, string(core.JailedList), validatorsInfo[0][i].List)
	}
}

func createStakingScAcc(accountsDB state.AccountsAdapter) state.UserAccountHandler {
	acc, _ := accountsDB.LoadAccount(vm.StakingSCAddress)
	stakingSCAcc := acc.(state.UserAccountHandler)

	return stakingSCAcc
}

func createEligibleNodes(numNodes int, stakingSCAcc state.UserAccountHandler, marshalizer marshal.Marshalizer) {
	for i := 0; i < numNodes; i++ {
		stakedData := &systemSmartContracts.StakedDataV2{
			Waiting:       false,
			Staked:        true,
			StakedNonce:   0,
			RewardAddress: []byte(fmt.Sprintf("rewardAddress_w%d", i)),
			StakeValue:    big.NewInt(100),
		}
		marshaledData, _ := marshalizer.Marshal(stakedData)
		stakingSCAcc.DataTrieTracker().SaveKeyValue([]byte(fmt.Sprintf("waiting_%d", i)), marshaledData)
	}
}

func createJailedNodes(numNodes int, stakingSCAcc state.UserAccountHandler, userAccounts state.AccountsAdapter, peerAccounts state.AccountsAdapter, marshalizer marshal.Marshalizer) []*state.ValidatorInfo {
	validatorInfos := make([]*state.ValidatorInfo, 0)

	for i := 0; i < numNodes; i++ {
		stakedData := &systemSmartContracts.StakedDataV2{
			Staked:        true,
			RewardAddress: []byte(fmt.Sprintf("rewardAddress_j%d", i)),
			StakeValue:    big.NewInt(100),
		}
		marshaledData, _ := marshalizer.Marshal(stakedData)
		stakingSCAcc.DataTrieTracker().SaveKeyValue([]byte(fmt.Sprintf("jailed__%d", i)), marshaledData)

		_ = userAccounts.SaveAccount(stakingSCAcc)

		jailedAcc, _ := peerAccounts.LoadAccount([]byte(fmt.Sprintf("jailed__%d", i)))
		_ = peerAccounts.SaveAccount(jailedAcc)

		vInfo := &state.ValidatorInfo{
			PublicKey:       []byte(fmt.Sprintf("jailed__%d", i)),
			ShardId:         0,
			List:            string(core.JailedList),
			TempRating:      1,
			RewardAddress:   []byte("address"),
			AccumulatedFees: big.NewInt(0),
		}

		validatorInfos = append(validatorInfos, vInfo)
	}

	return validatorInfos
}

func createWaitingNodes(numNodes int, stakingSCAcc state.UserAccountHandler, userAccounts state.AccountsAdapter, marshalizer marshal.Marshalizer) []*state.ValidatorInfo {
	validatorInfos := make([]*state.ValidatorInfo, 0)

	for i := 0; i < numNodes; i++ {
		stakedData := &systemSmartContracts.StakedDataV2{
			Waiting:       true,
			RewardAddress: []byte(fmt.Sprintf("rewardAddress_w%d", i)),
			StakeValue:    big.NewInt(100),
		}
		marshaledData, _ := marshalizer.Marshal(stakedData)
		stakingSCAcc.DataTrieTracker().SaveKeyValue([]byte(fmt.Sprintf("waiting_%d", i)), marshaledData)

		waitingKeyInList := []byte("w_" + fmt.Sprintf("waiting_%d", i))
		waitingListHead := &systemSmartContracts.WaitingList{
			FirstKey: []byte("w_" + fmt.Sprintf("waiting_%d", 0)),
			LastKey:  []byte("w_" + fmt.Sprintf("waiting_%d", numNodes-1)),
			Length:   uint32(numNodes),
		}
		marshaledData, _ = marshalizer.Marshal(waitingListHead)
		stakingSCAcc.DataTrieTracker().SaveKeyValue([]byte("waitingList"), marshaledData)

		waitingListElement := &systemSmartContracts.ElementInList{
			BLSPublicKey: []byte(fmt.Sprintf("waiting_%d", i)),
			PreviousKey:  waitingKeyInList,
			NextKey:      []byte("w_" + fmt.Sprintf("waiting_%d", i+1)),
		}
		marshaledData, _ = marshalizer.Marshal(waitingListElement)
		stakingSCAcc.DataTrieTracker().SaveKeyValue(waitingKeyInList, marshaledData)

		vInfo := &state.ValidatorInfo{
			PublicKey:       []byte(fmt.Sprintf("waiting_%d", i)),
			ShardId:         0,
			List:            string(core.WaitingList),
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
) {
	stakingSCAcc := createStakingScAcc(accountsDB)

	stakedData := &systemSmartContracts.StakedDataV2{
		Staked:        true,
		RewardAddress: []byte("rewardAddress"),
		StakeValue:    big.NewInt(100),
	}
	marshaledData, _ := marshalizer.Marshal(stakedData)
	stakingSCAcc.DataTrieTracker().SaveKeyValue(stakedKey, marshaledData)

	stakedData = &systemSmartContracts.StakedDataV2{
		Waiting:       true,
		RewardAddress: []byte("rewardAddress"),
		StakeValue:    big.NewInt(100),
	}
	marshaledData, _ = marshalizer.Marshal(stakedData)
	stakingSCAcc.DataTrieTracker().SaveKeyValue(waitingKey, marshaledData)

	waitingKeyInList := []byte("w_" + string(waitingKey))
	waitingListHead := &systemSmartContracts.WaitingList{
		FirstKey: waitingKeyInList,
		LastKey:  waitingKeyInList,
		Length:   1,
	}
	marshaledData, _ = marshalizer.Marshal(waitingListHead)
	stakingSCAcc.DataTrieTracker().SaveKeyValue([]byte("waitingList"), marshaledData)

	waitingListElement := &systemSmartContracts.ElementInList{
		BLSPublicKey: waitingKey,
		PreviousKey:  waitingKeyInList,
		NextKey:      make([]byte, 0),
	}
	marshaledData, _ = marshalizer.Marshal(waitingListElement)
	stakingSCAcc.DataTrieTracker().SaveKeyValue(waitingKeyInList, marshaledData)

	_ = accountsDB.SaveAccount(stakingSCAcc)
}

func createAccountsDB(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	accountFactory state.AccountFactory,
	trieStorageManager data.StorageManager,
) *state.AccountsDB {
	tr, _ := trie.NewTrie(trieStorageManager, marshalizer, hasher, 5)
	adb, _ := state.NewAccountsDB(tr, hasher, marshalizer, accountFactory)
	return adb
}

func createFullArgumentsForSystemSCProcessing() ArgsNewEpochStartSystemSCProcessing {
	hasher := sha256.Sha256{}
	marshalizer := &marshal.GogoProtoMarshalizer{}
	trieFactoryManager, _ := trie.NewTrieStorageManagerWithoutPruning(createMemUnit())
	userAccountsDB := createAccountsDB(hasher, marshalizer, factory.NewAccountCreator(), trieFactoryManager)
	peerAccountsDB := createAccountsDB(hasher, marshalizer, factory.NewPeerAccountCreator(), trieFactoryManager)

	argsValidatorsProcessor := peer.ArgValidatorStatisticsProcessor{
		Marshalizer:         marshalizer,
		NodesCoordinator:    &mock.NodesCoordinatorStub{},
		ShardCoordinator:    &mock.ShardCoordinatorStub{},
		DataPool:            &testscommon.PoolsHolderStub{},
		StorageService:      &mock.ChainStorerStub{},
		PubkeyConv:          &mock.PubkeyConverterMock{},
		PeerAdapter:         peerAccountsDB,
		Rater:               &mock.RaterStub{},
		RewardsHandler:      &mock.RewardsHandlerStub{},
		NodesSetup:          &mock.NodesSetupStub{},
		MaxComputableRounds: 1,
	}
	vCreator, _ := peer.NewValidatorStatisticsProcessor(argsValidatorsProcessor)

	blockChain := blockchain.NewMetaChain()
	argsHook := hooks.ArgBlockChainHook{
		Accounts:         userAccountsDB,
		PubkeyConv:       &mock.PubkeyConverterMock{},
		StorageService:   &mock.ChainStorerStub{},
		BlockChain:       blockChain,
		ShardCoordinator: &mock.ShardCoordinatorStub{},
		Marshalizer:      marshalizer,
		Uint64Converter:  &mock.Uint64ByteSliceConverterMock{},
		BuiltInFunctions: builtInFunctions.NewBuiltInFunctionContainer(),
	}

	gasSchedule := make(map[string]map[string]uint64)
	defaults.FillGasMapInternal(gasSchedule, 1)
	signVerifer, _ := disabled.NewMessageSignVerifier(&mock.KeyGenMock{})
	metaVmFactory, _ := metaProcess.NewVMContainerFactory(
		argsHook,
		createEconomicsData(),
		signVerifer,
		gasSchedule,
		&mock.NodesSetupStub{},
		hasher,
		marshalizer,
		&config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost: "1000",
				OwnerAddress:    "aaaaaa",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				ProposalCost:     "500",
				NumNodes:         100,
				MinQuorum:        50,
				MinPassThreshold: 50,
				MinVetoThreshold: 50,
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     "1000",
				UnJailValue:                          "10",
				MinStepValue:                         "10",
				MinStakeValue:                        "1",
				UnBondPeriod:                         1,
				AuctionEnableNonce:                   1000000,
				StakeEnableNonce:                     0,
				NumRoundsWithoutBleed:                1,
				MaximumPercentageToBleed:             1,
				BleedPercentagePerRound:              1,
				MaxNumberOfNodesForStake:             100,
				NodesToSelectInAuction:               100,
				ActivateBLSPubKeyMessageVerification: false,
			},
		},
		peerAccountsDB,
		&mock.ChanceComputerStub{},
	)

	vmContainer, _ := metaVmFactory.Create()
	systemVM, _ := vmContainer.Get(vmFactory.SystemVirtualMachine)
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
	}
	return args
}

func createEconomicsData() *economics2.EconomicsData {
	maxGasLimitPerBlock := strconv.FormatUint(1500000000, 10)
	minGasPrice := strconv.FormatUint(10, 10)
	minGasLimit := strconv.FormatUint(10, 10)

	argsNewEconomicsData := economics2.ArgsNewEconomicsData{
		Economics: &config.EconomicsConfig{
			GlobalSettings: config.GlobalSettings{
				GenesisTotalSupply: "2000000000000000000000",
				MinimumInflation:   0,
				YearSettings: []*config.YearSetting{
					{
						Year:             0,
						MaximumInflation: 0.01,
					},
				},
			},
			RewardsSettings: config.RewardsSettings{
				LeaderPercentage:              0.1,
				DeveloperPercentage:           0.1,
				ProtocolSustainabilityAddress: "protocol",
			},
			FeeSettings: config.FeeSettings{
				MaxGasLimitPerBlock:     maxGasLimitPerBlock,
				MaxGasLimitPerMetaBlock: maxGasLimitPerBlock,
				MinGasPrice:             minGasPrice,
				MinGasLimit:             minGasLimit,
				GasPerDataByte:          "1",
				DataLimitForBaseCalc:    "10000",
			},
		},
		PenalizedTooMuchGasEnableEpoch: 0,
		EpochNotifier:                  &mock.EpochNotifierStub{},
	}
	economicsData, _ := economics2.NewEconomicsData(argsNewEconomicsData)
	return economicsData
}
