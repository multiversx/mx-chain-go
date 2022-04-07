package staking

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/config"
	mock3 "github.com/ElrondNetwork/elrond-go/epochStart/mock"
	factory2 "github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/genesis/process/disabled"
	"github.com/ElrondNetwork/elrond-go/process"
	metaProcess "github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/peer"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

func createValidatorStatisticsProcessor(
	dataComponents factory2.DataComponentsHolder,
	coreComponents factory2.CoreComponentsHolder,
	nc nodesCoordinator.NodesCoordinator,
	shardCoordinator sharding.Coordinator,
	peerAccounts state.AccountsAdapter,
) process.ValidatorStatisticsProcessor {
	argsValidatorsProcessor := peer.ArgValidatorStatisticsProcessor{
		Marshalizer:                          coreComponents.InternalMarshalizer(),
		NodesCoordinator:                     nc,
		ShardCoordinator:                     shardCoordinator,
		DataPool:                             dataComponents.Datapool(),
		StorageService:                       dataComponents.StorageService(),
		PubkeyConv:                           coreComponents.AddressPubKeyConverter(),
		PeerAdapter:                          peerAccounts,
		Rater:                                coreComponents.Rater(),
		RewardsHandler:                       &mock3.RewardsHandlerStub{},
		NodesSetup:                           &mock.NodesSetupStub{},
		MaxComputableRounds:                  1,
		MaxConsecutiveRoundsOfRatingDecrease: 2000,
		EpochNotifier:                        coreComponents.EpochNotifier(),
		StakingV2EnableEpoch:                 0,
		StakingV4EnableEpoch:                 stakingV4EnableEpoch,
	}
	validatorStatisticsProcessor, _ := peer.NewValidatorStatisticsProcessor(argsValidatorsProcessor)
	return validatorStatisticsProcessor
}

func createBlockChainHook(
	dataComponents factory2.DataComponentsHolder,
	coreComponents factory2.CoreComponentsHolder,
	accountsAdapter state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	builtInFunctionsContainer vmcommon.BuiltInFunctionContainer,
) process.BlockChainHookHandler {
	argsHook := hooks.ArgBlockChainHook{
		Accounts:           accountsAdapter,
		PubkeyConv:         coreComponents.AddressPubKeyConverter(),
		StorageService:     dataComponents.StorageService(),
		BlockChain:         dataComponents.Blockchain(),
		ShardCoordinator:   shardCoordinator,
		Marshalizer:        coreComponents.InternalMarshalizer(),
		Uint64Converter:    coreComponents.Uint64ByteSliceConverter(),
		NFTStorageHandler:  &testscommon.SimpleNFTStorageHandlerStub{},
		BuiltInFunctions:   builtInFunctionsContainer,
		DataPool:           dataComponents.Datapool(),
		CompiledSCPool:     dataComponents.Datapool().SmartContracts(),
		EpochNotifier:      coreComponents.EpochNotifier(),
		NilCompiledSCStore: true,
	}

	blockChainHook, _ := hooks.NewBlockChainHookImpl(argsHook)
	return blockChainHook
}

func createVMContainerFactory(
	coreComponents factory2.CoreComponentsHolder,
	gasScheduleNotifier core.GasScheduleNotifier,
	blockChainHook process.BlockChainHookHandler,
	peerAccounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	nc nodesCoordinator.NodesCoordinator,
) process.VirtualMachinesContainerFactory {
	signVerifer, _ := disabled.NewMessageSignVerifier(&cryptoMocks.KeyGenStub{})

	argsNewVMContainerFactory := metaProcess.ArgsNewVMContainerFactory{
		BlockChainHook:      blockChainHook,
		PubkeyConv:          coreComponents.AddressPubKeyConverter(),
		Economics:           coreComponents.EconomicsData(),
		MessageSignVerifier: signVerifer,
		GasSchedule:         gasScheduleNotifier,
		NodesConfigProvider: &mock.NodesSetupStub{},
		Hasher:              coreComponents.Hasher(),
		Marshalizer:         coreComponents.InternalMarshalizer(),
		SystemSCConfig: &config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost:  "1000",
				OwnerAddress:     "aaaaaa",
				DelegationTicker: "DEL",
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
				GenesisNodePrice:                     "1000",
				UnJailValue:                          "10",
				MinStepValue:                         "10",
				MinStakeValue:                        "1",
				UnBondPeriod:                         1,
				NumRoundsWithoutBleed:                1,
				MaximumPercentageToBleed:             1,
				BleedPercentagePerRound:              1,
				MaxNumberOfNodesForStake:             24, // TODO HERE ADD MAX NUM NODES
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
		},
		ValidatorAccountsDB: peerAccounts,
		ChanceComputer:      &mock3.ChanceComputerStub{},
		EpochNotifier:       coreComponents.EpochNotifier(),
		EpochConfig: &config.EpochConfig{
			EnableEpochs: config.EnableEpochs{
				StakingV2EnableEpoch:               0,
				StakeEnableEpoch:                   0,
				DelegationManagerEnableEpoch:       0,
				DelegationSmartContractEnableEpoch: 0,
				StakeLimitsEnableEpoch:             10,
				StakingV4InitEnableEpoch:           stakingV4InitEpoch,
				StakingV4EnableEpoch:               stakingV4EnableEpoch,
			},
		},
		ShardCoordinator: shardCoordinator,
		NodesCoordinator: nc,
	}

	metaVmFactory, _ := metaProcess.NewVMContainerFactory(argsNewVMContainerFactory)
	return metaVmFactory
}
