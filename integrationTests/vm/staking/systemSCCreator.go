package staking

import (
	"bytes"
	"strconv"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/metachain"
	epochStartMock "github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/genesis/process/disabled"
	"github.com/multiversx/mx-chain-go/process"
	metaProcess "github.com/multiversx/mx-chain-go/process/factory/metachain"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/smartContract/builtInFunctions"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	vmcommonMock "github.com/multiversx/mx-chain-vm-common-go/mock"
)

func createSystemSCProcessor(
	nc nodesCoordinator.NodesCoordinator,
	coreComponents factory.CoreComponentsHolder,
	stateComponents factory.StateComponentsHandler,
	shardCoordinator sharding.Coordinator,
	maxNodesConfig []config.MaxNodesChangeConfig,
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor,
	systemVM vmcommon.VMExecutionHandler,
	stakingDataProvider epochStart.StakingDataProvider,
) process.EpochStartSystemSCProcessor {
	maxNodesChangeConfigProvider, _ := notifier.NewNodesConfigProvider(
		coreComponents.EpochNotifier(),
		maxNodesConfig,
	)
	argsAuctionListSelector := metachain.AuctionListSelectorArgs{
		ShardCoordinator:             shardCoordinator,
		StakingDataProvider:          stakingDataProvider,
		MaxNodesChangeConfigProvider: maxNodesChangeConfigProvider,
		SoftAuctionConfig: config.SoftAuctionConfig{
			TopUpStep: "10",
			MinTopUp:  "1",
			MaxTopUp:  "32000000",
		},
	}
	auctionListSelector, _ := metachain.NewAuctionListSelector(argsAuctionListSelector)

	args := metachain.ArgsNewEpochStartSystemSCProcessing{
		SystemVM:                systemVM,
		UserAccountsDB:          stateComponents.AccountsAdapter(),
		PeerAccountsDB:          stateComponents.PeerAccounts(),
		Marshalizer:             coreComponents.InternalMarshalizer(),
		StartRating:             initialRating,
		ValidatorInfoCreator:    validatorStatisticsProcessor,
		EndOfEpochCallerAddress: vm.EndOfEpochAddress,
		StakingSCAddress:        vm.StakingSCAddress,
		ChanceComputer:          &epochStartMock.ChanceComputerStub{},
		EpochNotifier:           coreComponents.EpochNotifier(),
		GenesisNodesConfig:      &mock.NodesSetupStub{},
		StakingDataProvider:     stakingDataProvider,
		NodesConfigProvider:     nc,
		ShardCoordinator:        shardCoordinator,
		ESDTOwnerAddressBytes:   bytes.Repeat([]byte{1}, 32),
		EpochConfig: config.EpochConfig{
			EnableEpochs: config.EnableEpochs{
				StakingV4InitEnableEpoch:  stakingV4InitEpoch,
				StakingV4EnableEpoch:      stakingV4EnableEpoch,
				MaxNodesChangeEnableEpoch: maxNodesConfig,
			},
		},
		MaxNodesChangeConfigProvider: maxNodesChangeConfigProvider,
		AuctionListSelector:          auctionListSelector,
	}

	systemSCProcessor, _ := metachain.NewSystemSCProcessor(args)
	return systemSCProcessor
}

func createStakingDataProvider(
	epochNotifier process.EpochNotifier,
	systemVM vmcommon.VMExecutionHandler,
) epochStart.StakingDataProvider {
	argsStakingDataProvider := metachain.StakingDataProviderArgs{
		EpochNotifier:            epochNotifier,
		SystemVM:                 systemVM,
		MinNodePrice:             strconv.Itoa(nodePrice),
		StakingV4InitEnableEpoch: stakingV4InitEpoch,
		StakingV4EnableEpoch:     stakingV4EnableEpoch,
	}
	stakingSCProvider, _ := metachain.NewStakingDataProvider(argsStakingDataProvider)

	return stakingSCProvider
}

func createValidatorStatisticsProcessor(
	dataComponents factory.DataComponentsHolder,
	coreComponents factory.CoreComponentsHolder,
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
		RewardsHandler:                       &epochStartMock.RewardsHandlerStub{},
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
	dataComponents factory.DataComponentsHolder,
	coreComponents factory.CoreComponentsHolder,
	accountsAdapter state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	gasScheduleNotifier core.GasScheduleNotifier,
) process.BlockChainHookHandler {
	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule:      gasScheduleNotifier,
		MapDNSAddresses:  make(map[string]struct{}),
		Marshalizer:      coreComponents.InternalMarshalizer(),
		Accounts:         accountsAdapter,
		ShardCoordinator: shardCoordinator,
		EpochNotifier:    coreComponents.EpochNotifier(),
	}
	builtInFunctionsContainer, _, _, _ := builtInFunctions.CreateBuiltInFuncContainerAndNFTStorageHandler(argsBuiltIn)

	argsHook := hooks.ArgBlockChainHook{
		Accounts:              accountsAdapter,
		PubkeyConv:            coreComponents.AddressPubKeyConverter(),
		StorageService:        dataComponents.StorageService(),
		BlockChain:            dataComponents.Blockchain(),
		ShardCoordinator:      shardCoordinator,
		Marshalizer:           coreComponents.InternalMarshalizer(),
		Uint64Converter:       coreComponents.Uint64ByteSliceConverter(),
		NFTStorageHandler:     &testscommon.SimpleNFTStorageHandlerStub{},
		BuiltInFunctions:      builtInFunctionsContainer,
		DataPool:              dataComponents.Datapool(),
		CompiledSCPool:        dataComponents.Datapool().SmartContracts(),
		EpochNotifier:         coreComponents.EpochNotifier(),
		GlobalSettingsHandler: &vmcommonMock.GlobalSettingsHandlerStub{},
		NilCompiledSCStore:    true,
	}

	blockChainHook, _ := hooks.NewBlockChainHookImpl(argsHook)
	return blockChainHook
}

func createVMContainerFactory(
	coreComponents factory.CoreComponentsHolder,
	gasScheduleNotifier core.GasScheduleNotifier,
	blockChainHook process.BlockChainHookHandler,
	peerAccounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	nc nodesCoordinator.NodesCoordinator,
	maxNumNodes uint32,
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
				GenesisNodePrice:                     strconv.Itoa(nodePrice),
				UnJailValue:                          "10",
				MinStepValue:                         "10",
				MinStakeValue:                        "1",
				UnBondPeriod:                         1,
				NumRoundsWithoutBleed:                1,
				MaximumPercentageToBleed:             1,
				BleedPercentagePerRound:              1,
				MaxNumberOfNodesForStake:             uint64(maxNumNodes),
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
		ChanceComputer:      coreComponents.Rater(),
		EpochNotifier:       coreComponents.EpochNotifier(),
		EpochConfig: &config.EpochConfig{
			EnableEpochs: config.EnableEpochs{
				StakingV4InitEnableEpoch: stakingV4InitEpoch,
				StakingV4EnableEpoch:     stakingV4EnableEpoch,
			},
		},
		ShardCoordinator: shardCoordinator,
		NodesCoordinator: nc,
	}

	metaVmFactory, _ := metaProcess.NewVMContainerFactory(argsNewVMContainerFactory)
	return metaVmFactory
}
