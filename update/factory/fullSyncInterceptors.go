package factory

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	disabledDataRetriever "github.com/multiversx/mx-chain-go/dataRetriever/disabled"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/dataValidators"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	interceptorFactory "github.com/multiversx/mx-chain-go/process/interceptors/factory"
	"github.com/multiversx/mx-chain-go/process/interceptors/processor"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/update/disabled"
)

var _ process.InterceptorsContainerFactory = (*fullSyncInterceptorsContainerFactory)(nil)

const numGoRoutines = 2000

// fullSyncInterceptorsContainerFactory will handle the creation the interceptors container for shards
type fullSyncInterceptorsContainerFactory struct {
	mainContainer          process.InterceptorsContainer
	fullArchiveContainer   process.InterceptorsContainer
	shardCoordinator       sharding.Coordinator
	accounts               state.AccountsAdapter
	store                  dataRetriever.StorageService
	dataPool               dataRetriever.PoolsHolder
	mainMessenger          process.TopicHandler
	fullArchiveMessenger   process.TopicHandler
	nodesCoordinator       nodesCoordinator.NodesCoordinator
	blockBlackList         process.TimeCacher
	argInterceptorFactory  *interceptorFactory.ArgInterceptedDataFactory
	globalThrottler        process.InterceptorThrottler
	maxTxNonceDeltaAllowed int
	addressPubkeyConv      core.PubkeyConverter
	whiteListHandler       update.WhiteListHandler
	whiteListerVerifiedTxs update.WhiteListHandler
	antifloodHandler       process.P2PAntifloodHandler
	preferredPeersHolder   update.PreferredPeersHolderHandler
	nodeOperationMode      common.NodeOperation
}

// ArgsNewFullSyncInterceptorsContainerFactory holds the arguments needed for fullSyncInterceptorsContainerFactory
type ArgsNewFullSyncInterceptorsContainerFactory struct {
	CoreComponents                   process.CoreComponentsHolder
	CryptoComponents                 process.CryptoComponentsHolder
	Accounts                         state.AccountsAdapter
	ShardCoordinator                 sharding.Coordinator
	NodesCoordinator                 nodesCoordinator.NodesCoordinator
	MainMessenger                    process.TopicHandler
	FullArchiveMessenger             process.TopicHandler
	Store                            dataRetriever.StorageService
	DataPool                         dataRetriever.PoolsHolder
	MaxTxNonceDeltaAllowed           int
	TxFeeHandler                     process.FeeHandler
	BlockBlackList                   process.TimeCacher
	HeaderSigVerifier                process.InterceptedHeaderSigVerifier
	HeaderIntegrityVerifier          process.HeaderIntegrityVerifier
	SizeCheckDelta                   uint32
	ValidityAttester                 process.ValidityAttester
	EpochStartTrigger                process.EpochStartTriggerHandler
	WhiteListHandler                 update.WhiteListHandler
	WhiteListerVerifiedTxs           update.WhiteListHandler
	MainInterceptorsContainer        process.InterceptorsContainer
	FullArchiveInterceptorsContainer process.InterceptorsContainer
	AntifloodHandler                 process.P2PAntifloodHandler
	NodeOperationMode                common.NodeOperation
}

// NewFullSyncInterceptorsContainerFactory is responsible for creating a new interceptors factory object
func NewFullSyncInterceptorsContainerFactory(
	args ArgsNewFullSyncInterceptorsContainerFactory,
) (*fullSyncInterceptorsContainerFactory, error) {
	err := checkBaseParams(
		args.CoreComponents,
		args.CryptoComponents,
		args.ShardCoordinator,
		args.Accounts,
		args.Store,
		args.DataPool,
		args.MainMessenger,
		args.FullArchiveMessenger,
		args.NodesCoordinator,
		args.BlockBlackList,
		args.WhiteListerVerifiedTxs,
	)
	if err != nil {
		return nil, err
	}

	if args.SizeCheckDelta > 0 {
		m := marshal.NewSizeCheckUnmarshalizer(args.CoreComponents.InternalMarshalizer(), args.SizeCheckDelta)
		err = args.CoreComponents.SetInternalMarshalizer(m)
		if err != nil {
			return nil, err
		}
	}
	if check.IfNil(args.TxFeeHandler) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.HeaderSigVerifier) {
		return nil, process.ErrNilHeaderSigVerifier
	}
	if check.IfNil(args.HeaderIntegrityVerifier) {
		return nil, process.ErrNilHeaderIntegrityVerifier
	}
	if check.IfNil(args.ValidityAttester) {
		return nil, process.ErrNilValidityAttester
	}
	if check.IfNil(args.EpochStartTrigger) {
		return nil, process.ErrNilEpochStartTrigger
	}
	if check.IfNil(args.MainInterceptorsContainer) {
		return nil, fmt.Errorf("%w on main network", update.ErrNilInterceptorsContainer)
	}
	if check.IfNil(args.FullArchiveInterceptorsContainer) {
		return nil, fmt.Errorf("%w on full archive network", update.ErrNilInterceptorsContainer)
	}
	if check.IfNil(args.WhiteListHandler) {
		return nil, update.ErrNilWhiteListHandler
	}
	if check.IfNil(args.AntifloodHandler) {
		return nil, process.ErrNilAntifloodHandler
	}

	argInterceptorFactory := &interceptorFactory.ArgInterceptedDataFactory{
		CoreComponents:          args.CoreComponents,
		CryptoComponents:        args.CryptoComponents,
		ShardCoordinator:        args.ShardCoordinator,
		NodesCoordinator:        args.NodesCoordinator,
		FeeHandler:              args.TxFeeHandler,
		HeaderSigVerifier:       args.HeaderSigVerifier,
		HeaderIntegrityVerifier: args.HeaderIntegrityVerifier,
		ValidityAttester:        args.ValidityAttester,
		EpochStartTrigger:       args.EpochStartTrigger,
		WhiteListerVerifiedTxs:  args.WhiteListerVerifiedTxs,
		ArgsParser:              smartContract.NewArgumentParser(),
	}

	icf := &fullSyncInterceptorsContainerFactory{
		mainContainer:          args.MainInterceptorsContainer,
		fullArchiveContainer:   args.FullArchiveInterceptorsContainer,
		accounts:               args.Accounts,
		shardCoordinator:       args.ShardCoordinator,
		mainMessenger:          args.MainMessenger,
		fullArchiveMessenger:   args.FullArchiveMessenger,
		store:                  args.Store,
		dataPool:               args.DataPool,
		nodesCoordinator:       args.NodesCoordinator,
		argInterceptorFactory:  argInterceptorFactory,
		blockBlackList:         args.BlockBlackList,
		maxTxNonceDeltaAllowed: args.MaxTxNonceDeltaAllowed,
		whiteListHandler:       args.WhiteListHandler,
		whiteListerVerifiedTxs: args.WhiteListerVerifiedTxs,
		antifloodHandler:       args.AntifloodHandler,
		//TODO: inject the real peers holder once we have the peers mapping before epoch bootstrap finishes
		preferredPeersHolder: disabled.NewPreferredPeersHolder(),
		nodeOperationMode:    args.NodeOperationMode,
	}

	icf.globalThrottler, err = throttler.NewNumGoRoutinesThrottler(numGoRoutines)
	if err != nil {
		return nil, err
	}

	return icf, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (ficf *fullSyncInterceptorsContainerFactory) Create() (process.InterceptorsContainer, process.InterceptorsContainer, error) {
	err := ficf.generateTxInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = ficf.generateUnsignedTxsInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = ficf.generateRewardTxInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = ficf.generateMiniBlocksInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = ficf.generateMetachainHeaderInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = ficf.generateShardHeaderInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = ficf.generateTrieNodesInterceptors()
	if err != nil {
		return nil, nil, err
	}

	return ficf.mainContainer, ficf.fullArchiveContainer, nil
}

func checkBaseParams(
	coreComponents process.CoreComponentsHolder,
	cryptoComponents process.CryptoComponentsHolder,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
	store dataRetriever.StorageService,
	dataPool dataRetriever.PoolsHolder,
	mainMessenger process.TopicHandler,
	fullArchiveMessenger process.TopicHandler,
	nodesCoordinator nodesCoordinator.NodesCoordinator,
	blockBlackList process.TimeCacher,
	whiteListerVerifiedTxs update.WhiteListHandler,
) error {
	if check.IfNil(coreComponents) {
		return process.ErrNilCoreComponentsHolder
	}
	if check.IfNil(cryptoComponents) {
		return process.ErrNilCryptoComponentsHolder
	}
	if check.IfNil(coreComponents.AddressPubKeyConverter()) {
		return process.ErrNilPubkeyConverter
	}
	if check.IfNil(coreComponents.InternalMarshalizer()) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(coreComponents.Hasher()) {
		return process.ErrNilHasher
	}
	if check.IfNil(coreComponents.Uint64ByteSliceConverter()) {
		return process.ErrNilUint64Converter
	}
	if len(coreComponents.ChainID()) == 0 {
		return process.ErrInvalidChainID
	}
	if check.IfNil(cryptoComponents.TxSignKeyGen()) {
		return process.ErrNilKeyGen
	}
	if check.IfNil(cryptoComponents.TxSingleSigner()) {
		return process.ErrNilSingleSigner
	}
	if check.IfNil(cryptoComponents.BlockSignKeyGen()) {
		return process.ErrNilKeyGen
	}
	if check.IfNil(cryptoComponents.BlockSigner()) {
		return process.ErrNilSingleSigner
	}
	multiSigner, err := cryptoComponents.GetMultiSigner(0)
	if err != nil {
		return err
	}
	if check.IfNil(multiSigner) {
		return process.ErrNilMultiSigVerifier
	}
	if check.IfNil(shardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(mainMessenger) {
		return fmt.Errorf("%w on main network", process.ErrNilMessenger)
	}
	if check.IfNil(fullArchiveMessenger) {
		return fmt.Errorf("%w on full archive network", process.ErrNilMessenger)
	}
	if check.IfNil(store) {
		return process.ErrNilStore
	}
	if check.IfNil(dataPool) {
		return process.ErrNilDataPoolHolder
	}
	if check.IfNil(nodesCoordinator) {
		return process.ErrNilNodesCoordinator
	}
	if check.IfNil(accounts) {
		return process.ErrNilAccountsAdapter
	}
	if check.IfNil(blockBlackList) {
		return update.ErrNilTimeCache
	}
	if check.IfNil(whiteListerVerifiedTxs) {
		return process.ErrNilWhiteListHandler
	}

	return nil
}

func (ficf *fullSyncInterceptorsContainerFactory) checkIfInterceptorExists(identifier string) bool {
	_, err := ficf.mainContainer.Get(identifier)
	return err == nil
}

func (ficf *fullSyncInterceptorsContainerFactory) generateShardHeaderInterceptors() error {
	numShards := ficf.shardCoordinator.NumberOfShards()
	tmpSC, err := sharding.NewMultiShardCoordinator(numShards, core.MetachainShardId)
	if err != nil {
		return err
	}

	keys := make([]string, numShards)
	interceptorsSlice := make([]process.Interceptor, numShards)

	//wire up to topics: shardBlocks_0_META, shardBlocks_1_META ...
	for idx := uint32(0); idx < numShards; idx++ {
		identifierHeader := factory.ShardBlocksTopic + tmpSC.CommunicationIdentifier(idx)
		if ficf.checkIfInterceptorExists(identifierHeader) {
			continue
		}

		interceptor, errCreate := ficf.createOneShardHeaderInterceptor(identifierHeader)
		if errCreate != nil {
			return errCreate
		}

		keys[int(idx)] = identifierHeader
		interceptorsSlice[int(idx)] = interceptor
	}

	return ficf.addInterceptorsToContainers(keys, interceptorsSlice)
}

func (ficf *fullSyncInterceptorsContainerFactory) createOneShardHeaderInterceptor(topic string) (process.Interceptor, error) {
	hdrFactory, err := interceptorFactory.NewInterceptedShardHeaderDataFactory(ficf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	argProcessor := &processor.ArgHdrInterceptorProcessor{
		Headers:        ficf.dataPool.Headers(),
		BlockBlackList: ficf.blockBlackList,
	}
	hdrProcessor, err := processor.NewHdrInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	interceptor, err := interceptors.NewSingleDataInterceptor(
		interceptors.ArgSingleDataInterceptor{
			Topic:            topic,
			DataFactory:      hdrFactory,
			Processor:        hdrProcessor,
			Throttler:        ficf.globalThrottler,
			AntifloodHandler: ficf.antifloodHandler,
			WhiteListRequest: ficf.whiteListHandler,
			CurrentPeerId:    ficf.mainMessenger.ID(),
		},
	)
	if err != nil {
		return nil, err
	}

	return ficf.createTopicAndAssignHandler(topic, interceptor, true)
}

func (ficf *fullSyncInterceptorsContainerFactory) generateUnsignedTxsInterceptors() error {
	shardC := ficf.shardCoordinator

	numShards := shardC.NumberOfShards()

	keys := make([]string, numShards+1)
	interceptorsSlice := make([]process.Interceptor, numShards+1)

	for idx := uint32(0); idx < numShards; idx++ {
		identifierScr := factory.UnsignedTransactionTopic + shardC.CommunicationIdentifier(idx)
		if ficf.checkIfInterceptorExists(identifierScr) {
			continue
		}

		interceptor, err := ficf.createOneUnsignedTxInterceptor(identifierScr)
		if err != nil {
			return err
		}

		keys[int(idx)] = identifierScr
		interceptorsSlice[int(idx)] = interceptor
	}

	identifierScr := factory.UnsignedTransactionTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	if !ficf.checkIfInterceptorExists(identifierScr) {
		interceptor, err := ficf.createOneUnsignedTxInterceptor(identifierScr)
		if err != nil {
			return err
		}

		keys[numShards] = identifierScr
		interceptorsSlice[numShards] = interceptor
	}

	return ficf.addInterceptorsToContainers(keys, interceptorsSlice)
}

func (ficf *fullSyncInterceptorsContainerFactory) generateTrieNodesInterceptors() error {
	numShards := ficf.shardCoordinator.NumberOfShards()

	keys := make([]string, 0)
	trieInterceptors := make([]process.Interceptor, 0)

	for i := uint32(0); i < numShards; i++ {
		identifierTrieNodes := factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(i, core.MetachainShardId)
		if ficf.checkIfInterceptorExists(identifierTrieNodes) {
			continue
		}

		interceptor, err := ficf.createOneTrieNodesInterceptor(identifierTrieNodes)
		if err != nil {
			return err
		}

		keys = append(keys, identifierTrieNodes)
		trieInterceptors = append(trieInterceptors, interceptor)
	}

	identifierTrieNodes := factory.ValidatorTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	if !ficf.checkIfInterceptorExists(identifierTrieNodes) {
		interceptor, err := ficf.createOneTrieNodesInterceptor(identifierTrieNodes)
		if err != nil {
			return err
		}

		keys = append(keys, identifierTrieNodes)
		trieInterceptors = append(trieInterceptors, interceptor)
	}

	identifierTrieNodes = factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	if !ficf.checkIfInterceptorExists(identifierTrieNodes) {
		interceptor, err := ficf.createOneTrieNodesInterceptor(identifierTrieNodes)
		if err != nil {
			return err
		}

		keys = append(keys, identifierTrieNodes)
		trieInterceptors = append(trieInterceptors, interceptor)
	}

	return ficf.addInterceptorsToContainers(keys, trieInterceptors)
}

func (ficf *fullSyncInterceptorsContainerFactory) createTopicAndAssignHandler(
	topic string,
	interceptor process.Interceptor,
	createChannel bool,
) (process.Interceptor, error) {

	err := createTopicAndAssignHandlerOnMessenger(topic, interceptor, createChannel, ficf.mainMessenger)
	if err != nil {
		return nil, err
	}

	if ficf.nodeOperationMode == common.FullArchiveMode {
		err = createTopicAndAssignHandlerOnMessenger(topic, interceptor, createChannel, ficf.fullArchiveMessenger)
		if err != nil {
			return nil, err
		}
	}

	return interceptor, nil
}

func createTopicAndAssignHandlerOnMessenger(
	topic string,
	interceptor process.Interceptor,
	createChannel bool,
	messenger process.TopicHandler,
) error {

	err := messenger.CreateTopic(topic, createChannel)
	if err != nil {
		return err
	}

	return messenger.RegisterMessageProcessor(topic, common.HardforkInterceptorsIdentifier, interceptor)
}

func (ficf *fullSyncInterceptorsContainerFactory) generateTxInterceptors() error {
	shardC := ficf.shardCoordinator

	numShards := shardC.NumberOfShards()

	keys := make([]string, numShards)
	interceptorSlice := make([]process.Interceptor, numShards)

	for idx := uint32(0); idx < numShards; idx++ {
		identifierTx := factory.TransactionTopic + shardC.CommunicationIdentifier(idx)
		if ficf.checkIfInterceptorExists(identifierTx) {
			continue
		}

		interceptor, err := ficf.createOneTxInterceptor(identifierTx)
		if err != nil {
			return err
		}

		keys[int(idx)] = identifierTx
		interceptorSlice[int(idx)] = interceptor
	}

	//tx interceptor for metachain topic
	identifierTx := factory.TransactionTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	if !ficf.checkIfInterceptorExists(identifierTx) {
		interceptor, err := ficf.createOneTxInterceptor(identifierTx)
		if err != nil {
			return err
		}

		keys = append(keys, identifierTx)
		interceptorSlice = append(interceptorSlice, interceptor)
	}

	return ficf.addInterceptorsToContainers(keys, interceptorSlice)
}

func (ficf *fullSyncInterceptorsContainerFactory) createOneTxInterceptor(topic string) (process.Interceptor, error) {
	txValidator, err := dataValidators.NewTxValidator(
		ficf.accounts,
		ficf.shardCoordinator,
		ficf.whiteListHandler,
		ficf.addressPubkeyConv,
		ficf.argInterceptorFactory.CoreComponents.TxVersionChecker(),
		ficf.maxTxNonceDeltaAllowed,
	)
	if err != nil {
		return nil, err
	}

	argProcessor := &processor.ArgTxInterceptorProcessor{
		ShardedDataCache: ficf.dataPool.Transactions(),
		UserShardedPool:  ficf.dataPool.UserTransactions(),
		TxValidator:      txValidator,
		ShardCoordinator: ficf.shardCoordinator,
	}
	txProcessor, err := processor.NewTxInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	txFactory, err := interceptorFactory.NewInterceptedTxDataFactory(ficf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	interceptor, err := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:                topic,
			Marshalizer:          ficf.argInterceptorFactory.CoreComponents.InternalMarshalizer(),
			DataFactory:          txFactory,
			Processor:            txProcessor,
			Throttler:            ficf.globalThrottler,
			AntifloodHandler:     ficf.antifloodHandler,
			WhiteListRequest:     ficf.whiteListHandler,
			CurrentPeerId:        ficf.mainMessenger.ID(),
			PreferredPeersHolder: ficf.preferredPeersHolder,
		},
	)
	if err != nil {
		return nil, err
	}

	return ficf.createTopicAndAssignHandler(topic, interceptor, true)
}

func (ficf *fullSyncInterceptorsContainerFactory) createOneUnsignedTxInterceptor(topic string) (process.Interceptor, error) {
	argProcessor := &processor.ArgTxInterceptorProcessor{
		ShardedDataCache: ficf.dataPool.UnsignedTransactions(),
		UserShardedPool:  disabledDataRetriever.NewShardedDataCacherNotifier(),
		TxValidator:      dataValidators.NewDisabledTxValidator(),
		ShardCoordinator: ficf.shardCoordinator,
	}
	txProcessor, err := processor.NewTxInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	txFactory, err := interceptorFactory.NewInterceptedUnsignedTxDataFactory(ficf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	interceptor, err := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:                topic,
			Marshalizer:          ficf.argInterceptorFactory.CoreComponents.InternalMarshalizer(),
			DataFactory:          txFactory,
			Processor:            txProcessor,
			Throttler:            ficf.globalThrottler,
			AntifloodHandler:     ficf.antifloodHandler,
			WhiteListRequest:     ficf.whiteListHandler,
			CurrentPeerId:        ficf.mainMessenger.ID(),
			PreferredPeersHolder: ficf.preferredPeersHolder,
		},
	)
	if err != nil {
		return nil, err
	}

	return ficf.createTopicAndAssignHandler(topic, interceptor, true)
}

func (ficf *fullSyncInterceptorsContainerFactory) createOneRewardTxInterceptor(topic string) (process.Interceptor, error) {
	argProcessor := &processor.ArgTxInterceptorProcessor{
		ShardedDataCache: ficf.dataPool.RewardTransactions(),
		UserShardedPool:  disabledDataRetriever.NewShardedDataCacherNotifier(),
		TxValidator:      dataValidators.NewDisabledTxValidator(),
		ShardCoordinator: ficf.shardCoordinator,
	}
	txProcessor, err := processor.NewTxInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	txFactory, err := interceptorFactory.NewInterceptedRewardTxDataFactory(ficf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	interceptor, err := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:                topic,
			Marshalizer:          ficf.argInterceptorFactory.CoreComponents.InternalMarshalizer(),
			DataFactory:          txFactory,
			Processor:            txProcessor,
			Throttler:            ficf.globalThrottler,
			AntifloodHandler:     ficf.antifloodHandler,
			WhiteListRequest:     ficf.whiteListHandler,
			CurrentPeerId:        ficf.mainMessenger.ID(),
			PreferredPeersHolder: ficf.preferredPeersHolder,
		},
	)
	if err != nil {
		return nil, err
	}

	return ficf.createTopicAndAssignHandler(topic, interceptor, true)
}

func (ficf *fullSyncInterceptorsContainerFactory) generateMiniBlocksInterceptors() error {
	shardC := ficf.shardCoordinator
	numShards := shardC.NumberOfShards()
	keys := make([]string, numShards+1)
	interceptorsSlice := make([]process.Interceptor, numShards+1)

	for idx := uint32(0); idx < numShards; idx++ {
		identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(idx)
		if ficf.checkIfInterceptorExists(identifierMiniBlocks) {
			continue
		}

		interceptor, err := ficf.createOneMiniBlocksInterceptor(identifierMiniBlocks)
		if err != nil {
			return err
		}

		keys[int(idx)] = identifierMiniBlocks
		interceptorsSlice[int(idx)] = interceptor
	}

	identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	if !ficf.checkIfInterceptorExists(identifierMiniBlocks) {
		interceptor, err := ficf.createOneMiniBlocksInterceptor(identifierMiniBlocks)
		if err != nil {
			return err
		}

		keys[numShards] = identifierMiniBlocks
		interceptorsSlice[numShards] = interceptor
	}

	return ficf.addInterceptorsToContainers(keys, interceptorsSlice)
}

func (ficf *fullSyncInterceptorsContainerFactory) createOneMiniBlocksInterceptor(topic string) (process.Interceptor, error) {
	argProcessor := &processor.ArgMiniblockInterceptorProcessor{
		MiniblockCache:   ficf.dataPool.MiniBlocks(),
		Marshalizer:      ficf.argInterceptorFactory.CoreComponents.InternalMarshalizer(),
		Hasher:           ficf.argInterceptorFactory.CoreComponents.Hasher(),
		ShardCoordinator: ficf.shardCoordinator,
		WhiteListHandler: ficf.whiteListHandler,
	}
	txBlockBodyProcessor, err := processor.NewMiniblockInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	txFactory, err := interceptorFactory.NewInterceptedMiniblockDataFactory(ficf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	interceptor, err := interceptors.NewSingleDataInterceptor(
		interceptors.ArgSingleDataInterceptor{
			Topic:                topic,
			DataFactory:          txFactory,
			Processor:            txBlockBodyProcessor,
			Throttler:            ficf.globalThrottler,
			AntifloodHandler:     ficf.antifloodHandler,
			WhiteListRequest:     ficf.whiteListHandler,
			CurrentPeerId:        ficf.mainMessenger.ID(),
			PreferredPeersHolder: ficf.preferredPeersHolder,
		},
	)
	if err != nil {
		return nil, err
	}

	return ficf.createTopicAndAssignHandler(topic, interceptor, true)
}

func (ficf *fullSyncInterceptorsContainerFactory) generateMetachainHeaderInterceptors() error {
	identifierHdr := factory.MetachainBlocksTopic
	if ficf.checkIfInterceptorExists(identifierHdr) {
		return nil
	}

	hdrFactory, err := interceptorFactory.NewInterceptedMetaHeaderDataFactory(ficf.argInterceptorFactory)
	if err != nil {
		return err
	}

	argProcessor := &processor.ArgHdrInterceptorProcessor{
		Headers:        ficf.dataPool.Headers(),
		BlockBlackList: ficf.blockBlackList,
	}
	hdrProcessor, err := processor.NewHdrInterceptorProcessor(argProcessor)
	if err != nil {
		return err
	}

	//only one metachain header topic
	interceptor, err := interceptors.NewSingleDataInterceptor(
		interceptors.ArgSingleDataInterceptor{
			Topic:                identifierHdr,
			DataFactory:          hdrFactory,
			Processor:            hdrProcessor,
			Throttler:            ficf.globalThrottler,
			AntifloodHandler:     ficf.antifloodHandler,
			WhiteListRequest:     ficf.whiteListHandler,
			CurrentPeerId:        ficf.mainMessenger.ID(),
			PreferredPeersHolder: ficf.preferredPeersHolder,
		},
	)
	if err != nil {
		return err
	}

	_, err = ficf.createTopicAndAssignHandler(identifierHdr, interceptor, true)
	if err != nil {
		return err
	}

	return ficf.addInterceptorsToContainers([]string{identifierHdr}, []process.Interceptor{interceptor})
}

func (ficf *fullSyncInterceptorsContainerFactory) createOneTrieNodesInterceptor(topic string) (process.Interceptor, error) {
	trieNodesProcessor, err := processor.NewTrieNodesInterceptorProcessor(ficf.dataPool.TrieNodes())
	if err != nil {
		return nil, err
	}

	trieNodesFactory, err := interceptorFactory.NewInterceptedTrieNodeDataFactory(ficf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	interceptor, err := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:                topic,
			Marshalizer:          ficf.argInterceptorFactory.CoreComponents.InternalMarshalizer(),
			DataFactory:          trieNodesFactory,
			Processor:            trieNodesProcessor,
			Throttler:            ficf.globalThrottler,
			AntifloodHandler:     ficf.antifloodHandler,
			WhiteListRequest:     ficf.whiteListHandler,
			CurrentPeerId:        ficf.mainMessenger.ID(),
			PreferredPeersHolder: ficf.preferredPeersHolder,
		},
	)
	if err != nil {
		return nil, err
	}

	return ficf.createTopicAndAssignHandler(topic, interceptor, true)
}

func (ficf *fullSyncInterceptorsContainerFactory) generateRewardTxInterceptors() error {
	numShards := ficf.shardCoordinator.NumberOfShards()

	tmpSC, err := sharding.NewMultiShardCoordinator(numShards, core.MetachainShardId)
	if err != nil {
		return err
	}

	keys := make([]string, numShards)
	interceptorSlice := make([]process.Interceptor, numShards)

	for idx := uint32(0); idx < numShards; idx++ {
		identifierScr := factory.RewardsTransactionTopic + tmpSC.CommunicationIdentifier(idx)
		if ficf.checkIfInterceptorExists(identifierScr) {
			return nil
		}

		var interceptor process.Interceptor
		interceptor, err = ficf.createOneRewardTxInterceptor(identifierScr)
		if err != nil {
			return err
		}

		keys[int(idx)] = identifierScr
		interceptorSlice[int(idx)] = interceptor
	}

	return ficf.addInterceptorsToContainers(keys, interceptorSlice)
}

func (ficf *fullSyncInterceptorsContainerFactory) addInterceptorsToContainers(keys []string, interceptors []process.Interceptor) error {
	err := ficf.mainContainer.AddMultiple(keys, interceptors)
	if err != nil {
		return err
	}

	if ficf.nodeOperationMode != common.FullArchiveMode {
		return nil
	}

	return ficf.fullArchiveContainer.AddMultiple(keys, interceptors)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ficf *fullSyncInterceptorsContainerFactory) IsInterfaceNil() bool {
	return ficf == nil
}
