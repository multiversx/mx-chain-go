package runType

import (
	"fmt"
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	sovereignFactory "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	requesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap"
	"github.com/multiversx/mx-chain-go/errors"
	factoryVm "github.com/multiversx/mx-chain-go/factory/vm"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/parsing"
	processComp "github.com/multiversx/mx-chain-go/genesis/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/factory/interceptorscontainer"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/processorV2"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/sharding"
	nodesCoord "github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state/factory"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

type ArgsSovereignRunTypeComponents struct {
	RunTypeComponentsFactory *runTypeComponentsFactory
	Config                   config.SovereignConfig
	DataCodec                sovereign.DataCodecHandler
	TopicsChecker            sovereign.TopicsCheckerHandler
	ExtraVerifier            process.ExtraHeaderSigVerifierHandler
}

type sovereignRunTypeComponentsFactory struct {
	*runTypeComponentsFactory
	sovConfig     config.SovereignConfig
	dataCodec     sovereign.DataCodecHandler
	topicsChecker sovereign.TopicsCheckerHandler
	extraVerifier process.ExtraHeaderSigVerifierHandler
}

// NewSovereignRunTypeComponentsFactory will return a new instance of runTypeComponentsFactory
func NewSovereignRunTypeComponentsFactory(args ArgsSovereignRunTypeComponents) (*sovereignRunTypeComponentsFactory, error) {
	if check.IfNil(args.RunTypeComponentsFactory) {
		return nil, errors.ErrNilRunTypeComponentsFactory
	}
	if check.IfNil(args.DataCodec) {
		return nil, errors.ErrNilDataCodec
	}
	if check.IfNil(args.TopicsChecker) {
		return nil, errors.ErrNilTopicsChecker
	}
	if check.IfNil(args.ExtraVerifier) {
		return nil, errors.ErrNilExtraSubRoundSigner
	}

	return &sovereignRunTypeComponentsFactory{
		runTypeComponentsFactory: args.RunTypeComponentsFactory,
		sovConfig:                args.Config,
		dataCodec:                args.DataCodec,
		topicsChecker:            args.TopicsChecker,
		extraVerifier:            args.ExtraVerifier,
	}, nil
}

// Create creates the runType components
func (rcf *sovereignRunTypeComponentsFactory) Create() (*runTypeComponents, error) {
	rtc, err := rcf.runTypeComponentsFactory.Create()
	if err != nil {
		return nil, err
	}

	blockChainHookHandlerFactory, err := hooks.NewSovereignBlockChainHookFactory(rtc.blockChainHookHandlerCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignBlockChainHookFactory failed: %w", err)
	}

	epochStartBootstrapperFactory, err := bootstrap.NewSovereignEpochStartBootstrapperFactory(rtc.epochStartBootstrapperCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignEpochStartBootstrapperFactory failed: %w", err)
	}

	bootstrapperFromStorageFactory, err := storageBootstrap.NewSovereignShardStorageBootstrapperFactory(rtc.bootstrapperFromStorageCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignShardStorageBootstrapperFactory failed: %w", err)
	}

	bootstrapperFactory, err := storageBootstrap.NewSovereignShardBootstrapFactory(rtc.bootstrapperCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignShardBootstrapFactory failed: %w", err)
	}

	blockProcessorFactory, err := block.NewSovereignBlockProcessorFactory(rtc.blockProcessorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignBlockProcessorFactory failed: %w", err)
	}

	forkDetectorFactory, err := sync.NewSovereignForkDetectorFactory(rtc.forkDetectorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignForkDetectorFactory failed: %w", err)
	}

	blockTrackerFactory, err := track.NewSovereignBlockTrackerFactory(rtc.blockTrackerCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignBlockTrackerFactory failed: %w", err)
	}

	requestHandlerFactory, err := requestHandlers.NewSovereignResolverRequestHandlerFactory(rtc.requestHandlerCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignResolverRequestHandlerFactory failed: %w", err)
	}

	headerValidatorFactory, err := block.NewSovereignHeaderValidatorFactory(rtc.headerValidatorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignHeaderValidatorFactory failed: %w", err)
	}

	scheduledTxsExecutionFactory, err := preprocess.NewSovereignScheduledTxsExecutionFactory()
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignScheduledTxsExecutionFactory failed: %w", err)
	}

	transactionCoordinatorFactory, err := coordinator.NewSovereignTransactionCoordinatorFactory(rtc.transactionCoordinatorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignTransactionCoordinatorFactory failed: %w", err)
	}

	validatorStatisticsProcessorFactory, err := peer.NewSovereignValidatorStatisticsProcessorFactory(rtc.validatorStatisticsProcessorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignValidatorStatisticsProcessorFactory failed: %w", err)
	}

	additionalStorageServiceCreator, err := storageFactory.NewSovereignAdditionalStorageServiceFactory()
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignAdditionalStorageServiceFactory failed: %w", err)
	}

	scProcessorCreator, err := processorV2.NewSovereignSCProcessFactory(rtc.scProcessorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignSCProcessFactory failed: %w", err)
	}

	scResultPreProcessorCreator, err := preprocess.NewSovereignSmartContractResultPreProcessorFactory(rtc.scResultPreProcessorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignSmartContractResultPreProcessorFactory failed: %w", err)
	}

	vmContainerShardCreator, err := factoryVm.NewSovereignVmContainerShardFactory(blockChainHookHandlerFactory, rtc.vmContainerMetaFactory, rtc.vmContainerShardFactory)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignVmContainerShardFactory failed: %w", err)
	}

	totalSupply, ok := big.NewInt(0).SetString(rcf.configs.EconomicsConfig.GlobalSettings.GenesisTotalSupply, 10)
	if !ok {
		return nil, fmt.Errorf("can not parse total suply from economics.toml, %s is not a valid value",
			rcf.configs.EconomicsConfig.GlobalSettings.GenesisTotalSupply)
	}

	accountsParserArgs := genesis.AccountsParserArgs{
		InitialAccounts: rcf.initialAccounts,
		EntireSupply:    totalSupply,
		MinterAddress:   rcf.configs.EconomicsConfig.GlobalSettings.GenesisMintingSenderAddress,
		PubkeyConverter: rcf.coreComponents.AddressPubKeyConverter(),
		KeyGenerator:    rcf.cryptoComponents.TxSignKeyGen(),
		Hasher:          rcf.coreComponents.Hasher(),
		Marshalizer:     rcf.coreComponents.InternalMarshalizer(),
	}
	accountsParser, err := parsing.NewAccountsParser(accountsParserArgs)
	if err != nil {
		return nil, fmt.Errorf("runTypeComponentsFactory - NewAccountsParser failed: %w", err)
	}
	sovereignAccountsParser, err := parsing.NewSovereignAccountsParser(accountsParser)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignAccountsParser failed: %w", err)
	}

	accountsCreator, err := factory.NewSovereignAccountCreator(factory.ArgsSovereignAccountCreator{
		ArgsAccountCreator: factory.ArgsAccountCreator{
			Hasher:              rcf.coreComponents.Hasher(),
			Marshaller:          rcf.coreComponents.InternalMarshalizer(),
			EnableEpochsHandler: rcf.coreComponents.EnableEpochsHandler(),
		},
		BaseTokenID: rcf.sovConfig.GenesisConfig.NativeESDT,
	})
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignAccountCreator failed: %w", err)
	}

	expiryTime := time.Second * time.Duration(rcf.sovConfig.OutgoingSubscribedEvents.TimeToWaitForUnconfirmedOutGoingOperationInSeconds)

	txPreProcessorCreator := preprocess.NewSovereignTxPreProcessorCreator()

	err = rtc.extraHeaderSigVerifierHandler.RegisterExtraHeaderSigVerifier(rcf.extraVerifier)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - RegisterExtraHeaderSigVerifier failed: %w", err)
	}

	return &runTypeComponents{
		blockChainHookHandlerCreator:            blockChainHookHandlerFactory,
		epochStartBootstrapperCreator:           epochStartBootstrapperFactory,
		bootstrapperFromStorageCreator:          bootstrapperFromStorageFactory,
		bootstrapperCreator:                     bootstrapperFactory,
		blockProcessorCreator:                   blockProcessorFactory,
		forkDetectorCreator:                     forkDetectorFactory,
		blockTrackerCreator:                     blockTrackerFactory,
		requestHandlerCreator:                   requestHandlerFactory,
		headerValidatorCreator:                  headerValidatorFactory,
		scheduledTxsExecutionCreator:            scheduledTxsExecutionFactory,
		transactionCoordinatorCreator:           transactionCoordinatorFactory,
		validatorStatisticsProcessorCreator:     validatorStatisticsProcessorFactory,
		additionalStorageServiceCreator:         additionalStorageServiceCreator,
		scProcessorCreator:                      scProcessorCreator,
		scResultPreProcessorCreator:             scResultPreProcessorCreator,
		consensusModel:                          consensus.ConsensusModelV2,
		vmContainerMetaFactory:                  rtc.vmContainerMetaFactory,
		vmContainerShardFactory:                 vmContainerShardCreator,
		accountsParser:                          sovereignAccountsParser,
		accountsCreator:                         accountsCreator,
		outGoingOperationsPoolHandler:           sovereignFactory.NewOutGoingOperationPool(expiryTime),
		dataCodecHandler:                        rcf.dataCodec,
		topicsCheckerHandler:                    rcf.topicsChecker,
		shardCoordinatorCreator:                 sharding.NewSovereignShardCoordinatorFactory(),
		nodesCoordinatorWithRaterFactoryCreator: nodesCoord.NewSovereignIndexHashedNodesCoordinatorWithRaterFactory(),
		requestersContainerFactoryCreator:       requesterscontainer.NewSovereignShardRequestersContainerFactoryCreator(),
		interceptorsContainerFactoryCreator:     interceptorscontainer.NewSovereignShardInterceptorsContainerFactoryCreator(),
		shardResolversContainerFactoryCreator:   resolverscontainer.NewSovereignShardResolversContainerFactoryCreator(),
		txPreProcessorCreator:                   txPreProcessorCreator,
		extraHeaderSigVerifierHandler:           rtc.extraHeaderSigVerifierHandler,
		genesisBlockCreatorFactory:              processComp.NewSovereignGenesisBlockCreatorFactory(),
		genesisMetaBlockCheckerCreator:          processComp.NewSovereignGenesisMetaBlockChecker(),
	}, nil
}
