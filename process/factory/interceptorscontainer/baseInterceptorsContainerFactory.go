package interceptorscontainer

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/disabled"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/dataValidators"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	interceptorFactory "github.com/multiversx/mx-chain-go/process/interceptors/factory"
	"github.com/multiversx/mx-chain-go/process/interceptors/processor"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
)

const (
	numGoRoutines                   = 100
	chunksProcessorRequestInterval  = time.Millisecond * 400
	minTimespanDurationInSec        = int64(1)
	errorOnMainNetworkString        = "on main network"
	errorOnFullArchiveNetworkString = "on full archive network"
)

type baseInterceptorsContainerFactory struct {
	mainContainer              process.InterceptorsContainer
	fullArchiveContainer       process.InterceptorsContainer
	shardCoordinator           sharding.Coordinator
	accounts                   state.AccountsAdapter
	store                      dataRetriever.StorageService
	dataPool                   dataRetriever.PoolsHolder
	mainMessenger              process.TopicHandler
	fullArchiveMessenger       process.TopicHandler
	nodesCoordinator           nodesCoordinator.NodesCoordinator
	blockBlackList             process.TimeCacher
	argInterceptorFactory      *interceptorFactory.ArgInterceptedDataFactory
	globalThrottler            process.InterceptorThrottler
	maxTxNonceDeltaAllowed     int
	antifloodHandler           process.P2PAntifloodHandler
	whiteListHandler           process.WhiteListHandler
	whiteListerVerifiedTxs     process.WhiteListHandler
	preferredPeersHolder       process.PreferredPeersHolderHandler
	hasher                     hashing.Hasher
	requestHandler             process.RequestHandler
	mainPeerShardMapper        process.PeerShardMapper
	fullArchivePeerShardMapper process.PeerShardMapper
	hardforkTrigger            heartbeat.HardforkTrigger
	nodeOperationMode          common.NodeOperation
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
	blackList process.TimeCacher,
	antifloodHandler process.P2PAntifloodHandler,
	whiteListHandler process.WhiteListHandler,
	whiteListerVerifiedTxs process.WhiteListHandler,
	preferredPeersHolder process.PreferredPeersHolderHandler,
	requestHandler process.RequestHandler,
	mainPeerShardMapper process.PeerShardMapper,
	fullArchivePeerShardMapper process.PeerShardMapper,
	hardforkTrigger heartbeat.HardforkTrigger,
) error {
	if check.IfNil(coreComponents) {
		return process.ErrNilCoreComponentsHolder
	}
	if check.IfNil(cryptoComponents) {
		return process.ErrNilCryptoComponentsHolder
	}
	if check.IfNil(shardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(mainMessenger) {
		return fmt.Errorf("%w %s", process.ErrNilMessenger, errorOnMainNetworkString)
	}
	if check.IfNil(fullArchiveMessenger) {
		return fmt.Errorf("%w %s", process.ErrNilMessenger, errorOnFullArchiveNetworkString)
	}
	if check.IfNil(store) {
		return process.ErrNilStore
	}
	if check.IfNil(coreComponents.InternalMarshalizer()) || check.IfNil(coreComponents.TxMarshalizer()) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(coreComponents.Hasher()) {
		return process.ErrNilHasher
	}
	if check.IfNil(coreComponents.TxSignHasher()) {
		return process.ErrNilHasher
	}
	if check.IfNil(coreComponents.TxVersionChecker()) {
		return process.ErrNilTransactionVersionChecker
	}
	if check.IfNil(coreComponents.EpochNotifier()) {
		return process.ErrNilEpochNotifier
	}
	if len(coreComponents.ChainID()) == 0 {
		return process.ErrInvalidChainID
	}
	if coreComponents.MinTransactionVersion() == 0 {
		return process.ErrInvalidTransactionVersion
	}
	multiSigner, err := cryptoComponents.GetMultiSigner(0)
	if err != nil {
		return err
	}
	if check.IfNil(multiSigner) {
		return process.ErrNilMultiSigVerifier
	}
	if check.IfNil(cryptoComponents.BlockSignKeyGen()) {
		return process.ErrNilKeyGen
	}
	if check.IfNil(cryptoComponents.BlockSigner()) {
		return process.ErrNilSingleSigner
	}
	if check.IfNil(cryptoComponents.TxSignKeyGen()) {
		return process.ErrNilKeyGen
	}
	if check.IfNil(cryptoComponents.TxSingleSigner()) {
		return process.ErrNilSingleSigner
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
	if check.IfNil(blackList) {
		return process.ErrNilBlackListCacher
	}
	if check.IfNil(antifloodHandler) {
		return process.ErrNilAntifloodHandler
	}
	if check.IfNil(whiteListHandler) {
		return process.ErrNilWhiteListHandler
	}
	if check.IfNil(whiteListerVerifiedTxs) {
		return process.ErrNilWhiteListHandler
	}
	if check.IfNil(coreComponents.AddressPubKeyConverter()) {
		return process.ErrNilPubkeyConverter
	}
	if check.IfNil(preferredPeersHolder) {
		return process.ErrNilPreferredPeersHolder
	}
	if check.IfNil(requestHandler) {
		return process.ErrNilRequestHandler
	}
	if check.IfNil(mainPeerShardMapper) {
		return fmt.Errorf("%w %s", process.ErrNilPeerShardMapper, errorOnMainNetworkString)
	}
	if check.IfNil(fullArchivePeerShardMapper) {
		return fmt.Errorf("%w %s", process.ErrNilPeerShardMapper, errorOnFullArchiveNetworkString)
	}
	if check.IfNil(hardforkTrigger) {
		return process.ErrNilHardforkTrigger
	}

	return nil
}

func (bicf *baseInterceptorsContainerFactory) createTopicAndAssignHandler(
	topic string,
	interceptor process.Interceptor,
	createChannel bool,
) (process.Interceptor, error) {

	err := createTopicAndAssignHandlerOnMessenger(topic, interceptor, createChannel, bicf.mainMessenger)
	if err != nil {
		return nil, err
	}

	if bicf.nodeOperationMode == common.FullArchiveMode {
		err = createTopicAndAssignHandlerOnMessenger(topic, interceptor, createChannel, bicf.fullArchiveMessenger)
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

	return messenger.RegisterMessageProcessor(topic, common.DefaultInterceptorsIdentifier, interceptor)
}

// ------- Tx interceptors

func (bicf *baseInterceptorsContainerFactory) generateTxInterceptors() error {
	shardC := bicf.shardCoordinator

	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards)
	interceptorSlice := make([]process.Interceptor, noOfShards)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := factory.TransactionTopic + shardC.CommunicationIdentifier(idx)

		interceptor, err := bicf.createOneTxInterceptor(identifierTx)
		if err != nil {
			return err
		}

		keys[int(idx)] = identifierTx
		interceptorSlice[int(idx)] = interceptor
	}

	// tx interceptor for metachain topic
	identifierTx := factory.TransactionTopic + shardC.CommunicationIdentifier(core.MetachainShardId)

	interceptor, err := bicf.createOneTxInterceptor(identifierTx)
	if err != nil {
		return err
	}

	keys = append(keys, identifierTx)
	interceptorSlice = append(interceptorSlice, interceptor)

	return bicf.addInterceptorsToContainers(keys, interceptorSlice)
}

func (bicf *baseInterceptorsContainerFactory) createOneTxInterceptor(topic string) (process.Interceptor, error) {
	if bicf.argInterceptorFactory == nil {
		return nil, process.ErrNilArgumentStruct
	}
	if check.IfNil(bicf.argInterceptorFactory.CoreComponents) {
		return nil, process.ErrNilCoreComponentsHolder
	}

	addrPubKeyConverter := bicf.argInterceptorFactory.CoreComponents.AddressPubKeyConverter()

	txValidator, err := dataValidators.NewTxValidator(
		bicf.accounts,
		bicf.shardCoordinator,
		bicf.whiteListHandler,
		addrPubKeyConverter,
		bicf.argInterceptorFactory.CoreComponents.TxVersionChecker(),
		bicf.maxTxNonceDeltaAllowed,
	)
	if err != nil {
		return nil, err
	}

	argProcessor := &processor.ArgTxInterceptorProcessor{
		ShardedDataCache: bicf.dataPool.Transactions(),
		UserShardedPool:  bicf.dataPool.UserTransactions(),
		TxValidator:      txValidator,
	}
	txProcessor, err := processor.NewTxInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	txFactory, err := interceptorFactory.NewInterceptedTxDataFactory(bicf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	internalMarshaller := bicf.argInterceptorFactory.CoreComponents.InternalMarshalizer()
	interceptor, err := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:                topic,
			Marshalizer:          internalMarshaller,
			DataFactory:          txFactory,
			Processor:            txProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			CurrentPeerId:        bicf.mainMessenger.ID(),
			PreferredPeersHolder: bicf.preferredPeersHolder,
		},
	)
	if err != nil {
		return nil, err
	}

	return bicf.createTopicAndAssignHandler(topic, interceptor, true)
}

func (bicf *baseInterceptorsContainerFactory) createOneUnsignedTxInterceptor(topic string) (process.Interceptor, error) {
	if bicf.argInterceptorFactory == nil {
		return nil, process.ErrNilArgumentStruct
	}
	if check.IfNil(bicf.argInterceptorFactory.CoreComponents) {
		return nil, process.ErrNilCoreComponentsHolder
	}

	argProcessor := &processor.ArgTxInterceptorProcessor{
		ShardedDataCache: bicf.dataPool.UnsignedTransactions(),
		UserShardedPool:  disabled.NewShardedDataCacherNotifier(),
		TxValidator:      dataValidators.NewDisabledTxValidator(),
	}
	txProcessor, err := processor.NewTxInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	txFactory, err := interceptorFactory.NewInterceptedUnsignedTxDataFactory(bicf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	internalMarshaller := bicf.argInterceptorFactory.CoreComponents.InternalMarshalizer()
	interceptor, err := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:                topic,
			Marshalizer:          internalMarshaller,
			DataFactory:          txFactory,
			Processor:            txProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			CurrentPeerId:        bicf.mainMessenger.ID(),
			PreferredPeersHolder: bicf.preferredPeersHolder,
		},
	)
	if err != nil {
		return nil, err
	}

	return bicf.createTopicAndAssignHandler(topic, interceptor, true)
}

func (bicf *baseInterceptorsContainerFactory) createOneRewardTxInterceptor(topic string) (process.Interceptor, error) {
	if bicf.argInterceptorFactory == nil {
		return nil, process.ErrNilArgumentStruct
	}
	if check.IfNil(bicf.argInterceptorFactory.CoreComponents) {
		return nil, process.ErrNilCoreComponentsHolder
	}

	argProcessor := &processor.ArgTxInterceptorProcessor{
		ShardedDataCache: bicf.dataPool.RewardTransactions(),
		UserShardedPool:  disabled.NewShardedDataCacherNotifier(),
		TxValidator:      dataValidators.NewDisabledTxValidator(),
	}
	txProcessor, err := processor.NewTxInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	txFactory, err := interceptorFactory.NewInterceptedRewardTxDataFactory(bicf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	internalMarshaller := bicf.argInterceptorFactory.CoreComponents.InternalMarshalizer()
	interceptor, err := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:                topic,
			Marshalizer:          internalMarshaller,
			DataFactory:          txFactory,
			Processor:            txProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			CurrentPeerId:        bicf.mainMessenger.ID(),
			PreferredPeersHolder: bicf.preferredPeersHolder,
		},
	)
	if err != nil {
		return nil, err
	}

	return bicf.createTopicAndAssignHandler(topic, interceptor, true)
}

// ------- Hdr interceptor

func (bicf *baseInterceptorsContainerFactory) generateHeaderInterceptors() error {
	shardC := bicf.shardCoordinator

	hdrFactory, err := interceptorFactory.NewInterceptedShardHeaderDataFactory(bicf.argInterceptorFactory)
	if err != nil {
		return err
	}

	argProcessor := &processor.ArgHdrInterceptorProcessor{
		Headers:        bicf.dataPool.Headers(),
		BlockBlackList: bicf.blockBlackList,
	}
	hdrProcessor, err := processor.NewHdrInterceptorProcessor(argProcessor)
	if err != nil {
		return err
	}

	// compose header shard topic, for example: shardBlocks_0_META
	identifierHdr := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(core.MetachainShardId)

	// only one intrashard header topic
	interceptor, err := interceptors.NewSingleDataInterceptor(
		interceptors.ArgSingleDataInterceptor{
			Topic:                identifierHdr,
			DataFactory:          hdrFactory,
			Processor:            hdrProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			CurrentPeerId:        bicf.mainMessenger.ID(),
			PreferredPeersHolder: bicf.preferredPeersHolder,
		},
	)
	if err != nil {
		return err
	}

	_, err = bicf.createTopicAndAssignHandler(identifierHdr, interceptor, true)
	if err != nil {
		return err
	}

	return bicf.addInterceptorsToContainers([]string{identifierHdr}, []process.Interceptor{interceptor})
}

// ------- MiniBlocks interceptors

func (bicf *baseInterceptorsContainerFactory) generateMiniBlocksInterceptors() error {
	shardC := bicf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards+2)
	interceptorsSlice := make([]process.Interceptor, noOfShards+2)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(idx)

		interceptor, err := bicf.createOneMiniBlocksInterceptor(identifierMiniBlocks)
		if err != nil {
			return err
		}

		keys[int(idx)] = identifierMiniBlocks
		interceptorsSlice[int(idx)] = interceptor
	}

	identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.MetachainShardId)

	interceptor, err := bicf.createOneMiniBlocksInterceptor(identifierMiniBlocks)
	if err != nil {
		return err
	}

	keys[noOfShards] = identifierMiniBlocks
	interceptorsSlice[noOfShards] = interceptor

	identifierAllShardsMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.AllShardId)

	allShardsMiniBlocksInterceptor, err := bicf.createOneMiniBlocksInterceptor(identifierAllShardsMiniBlocks)
	if err != nil {
		return err
	}

	keys[noOfShards+1] = identifierAllShardsMiniBlocks
	interceptorsSlice[noOfShards+1] = allShardsMiniBlocksInterceptor

	return bicf.addInterceptorsToContainers(keys, interceptorsSlice)
}

func (bicf *baseInterceptorsContainerFactory) createOneMiniBlocksInterceptor(topic string) (process.Interceptor, error) {
	internalMarshaller := bicf.argInterceptorFactory.CoreComponents.InternalMarshalizer()
	hasher := bicf.argInterceptorFactory.CoreComponents.Hasher()
	argProcessor := &processor.ArgMiniblockInterceptorProcessor{
		MiniblockCache:   bicf.dataPool.MiniBlocks(),
		Marshalizer:      internalMarshaller,
		Hasher:           hasher,
		ShardCoordinator: bicf.shardCoordinator,
		WhiteListHandler: bicf.whiteListHandler,
	}
	miniblockProcessor, err := processor.NewMiniblockInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	miniblockFactory, err := interceptorFactory.NewInterceptedMiniblockDataFactory(bicf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	interceptor, err := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:                topic,
			Marshalizer:          internalMarshaller,
			DataFactory:          miniblockFactory,
			Processor:            miniblockProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			CurrentPeerId:        bicf.mainMessenger.ID(),
			PreferredPeersHolder: bicf.preferredPeersHolder,
		},
	)
	if err != nil {
		return nil, err
	}

	return bicf.createTopicAndAssignHandler(topic, interceptor, true)
}

// ------- MetachainHeader interceptors

func (bicf *baseInterceptorsContainerFactory) generateMetachainHeaderInterceptors() error {
	identifierHdr := factory.MetachainBlocksTopic

	hdrFactory, err := interceptorFactory.NewInterceptedMetaHeaderDataFactory(bicf.argInterceptorFactory)
	if err != nil {
		return err
	}

	argProcessor := &processor.ArgHdrInterceptorProcessor{
		Headers:        bicf.dataPool.Headers(),
		BlockBlackList: bicf.blockBlackList,
	}
	hdrProcessor, err := processor.NewHdrInterceptorProcessor(argProcessor)
	if err != nil {
		return err
	}

	// only one metachain header topic
	interceptor, err := interceptors.NewSingleDataInterceptor(
		interceptors.ArgSingleDataInterceptor{
			Topic:                identifierHdr,
			DataFactory:          hdrFactory,
			Processor:            hdrProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			CurrentPeerId:        bicf.mainMessenger.ID(),
			PreferredPeersHolder: bicf.preferredPeersHolder,
		},
	)
	if err != nil {
		return err
	}

	_, err = bicf.createTopicAndAssignHandler(identifierHdr, interceptor, true)
	if err != nil {
		return err
	}

	return bicf.addInterceptorsToContainers([]string{identifierHdr}, []process.Interceptor{interceptor})
}

func (bicf *baseInterceptorsContainerFactory) createOneTrieNodesInterceptor(topic string) (process.Interceptor, error) {
	trieNodesProcessor, err := processor.NewTrieNodesInterceptorProcessor(bicf.dataPool.TrieNodes())
	if err != nil {
		return nil, err
	}

	trieNodesFactory, err := interceptorFactory.NewInterceptedTrieNodeDataFactory(bicf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	internalMarshaller := bicf.argInterceptorFactory.CoreComponents.InternalMarshalizer()
	interceptor, err := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:                topic,
			Marshalizer:          internalMarshaller,
			DataFactory:          trieNodesFactory,
			Processor:            trieNodesProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			CurrentPeerId:        bicf.mainMessenger.ID(),
			PreferredPeersHolder: bicf.preferredPeersHolder,
		},
	)
	if err != nil {
		return nil, err
	}

	argChunkProcessor := processor.TrieNodesChunksProcessorArgs{
		Hasher:          bicf.hasher,
		ChunksCacher:    bicf.dataPool.TrieNodesChunks(),
		RequestInterval: chunksProcessorRequestInterval,
		RequestHandler:  bicf.requestHandler,
		Topic:           topic,
	}

	chunkProcessor, err := processor.NewTrieNodeChunksProcessor(argChunkProcessor)
	if err != nil {
		return nil, err
	}

	err = interceptor.SetChunkProcessor(chunkProcessor)
	if err != nil {
		return nil, err
	}

	return bicf.createTopicAndAssignHandler(topic, interceptor, true)
}

func (bicf *baseInterceptorsContainerFactory) generateUnsignedTxsInterceptors() error {
	shardC := bicf.shardCoordinator

	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards, noOfShards+1)
	interceptorsSlice := make([]process.Interceptor, noOfShards, noOfShards+1)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierScr := factory.UnsignedTransactionTopic + shardC.CommunicationIdentifier(idx)
		interceptor, err := bicf.createOneUnsignedTxInterceptor(identifierScr)
		if err != nil {
			return err
		}

		keys[int(idx)] = identifierScr
		interceptorsSlice[int(idx)] = interceptor
	}

	identifierScr := factory.UnsignedTransactionTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	interceptor, err := bicf.createOneUnsignedTxInterceptor(identifierScr)
	if err != nil {
		return err
	}

	keys = append(keys, identifierScr)
	interceptorsSlice = append(interceptorsSlice, interceptor)

	return bicf.addInterceptorsToContainers(keys, interceptorsSlice)
}

//------- PeerAuthentication interceptor

func (bicf *baseInterceptorsContainerFactory) generatePeerAuthenticationInterceptor() error {
	identifierPeerAuthentication := common.PeerAuthenticationTopic

	internalMarshaller := bicf.argInterceptorFactory.CoreComponents.InternalMarshalizer()
	argProcessor := processor.ArgPeerAuthenticationInterceptorProcessor{
		PeerAuthenticationCacher: bicf.dataPool.PeerAuthentications(),
		PeerShardMapper:          bicf.mainPeerShardMapper,
		Marshaller:               internalMarshaller,
		HardforkTrigger:          bicf.hardforkTrigger,
	}
	peerAuthenticationProcessor, err := processor.NewPeerAuthenticationInterceptorProcessor(argProcessor)
	if err != nil {
		return err
	}

	peerAuthenticationFactory, err := interceptorFactory.NewInterceptedPeerAuthenticationDataFactory(*bicf.argInterceptorFactory)
	if err != nil {
		return err
	}

	mdInterceptor, err := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:                identifierPeerAuthentication,
			Marshalizer:          internalMarshaller,
			DataFactory:          peerAuthenticationFactory,
			Processor:            peerAuthenticationProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			PreferredPeersHolder: bicf.preferredPeersHolder,
			CurrentPeerId:        bicf.mainMessenger.ID(),
		},
	)
	if err != nil {
		return err
	}

	err = createTopicAndAssignHandlerOnMessenger(identifierPeerAuthentication, mdInterceptor, true, bicf.mainMessenger)
	if err != nil {
		return err
	}

	return bicf.mainContainer.Add(identifierPeerAuthentication, mdInterceptor)
}

//------- Heartbeat interceptor

func (bicf *baseInterceptorsContainerFactory) generateHeartbeatInterceptor() error {
	shardC := bicf.shardCoordinator
	identifierHeartbeat := common.HeartbeatV2Topic + shardC.CommunicationIdentifier(shardC.SelfId())

	interceptor, err := bicf.createHeartbeatV2Interceptor(identifierHeartbeat, bicf.dataPool.Heartbeats(), bicf.mainPeerShardMapper)
	if err != nil {
		return err
	}

	return bicf.addInterceptorsToContainers([]string{identifierHeartbeat}, []process.Interceptor{interceptor})
}

func (bicf *baseInterceptorsContainerFactory) createHeartbeatV2Interceptor(
	identifier string,
	heartbeatCahcer storage.Cacher,
	peerShardMapper process.PeerShardMapper,
) (process.Interceptor, error) {
	argHeartbeatProcessor := processor.ArgHeartbeatInterceptorProcessor{
		HeartbeatCacher:  heartbeatCahcer,
		ShardCoordinator: bicf.shardCoordinator,
		PeerShardMapper:  peerShardMapper,
	}
	heartbeatProcessor, err := processor.NewHeartbeatInterceptorProcessor(argHeartbeatProcessor)
	if err != nil {
		return nil, err
	}

	heartbeatFactory, err := interceptorFactory.NewInterceptedHeartbeatDataFactory(*bicf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	interceptor, err := interceptors.NewSingleDataInterceptor(
		interceptors.ArgSingleDataInterceptor{
			Topic:                identifier,
			DataFactory:          heartbeatFactory,
			Processor:            heartbeatProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			PreferredPeersHolder: bicf.preferredPeersHolder,
			CurrentPeerId:        bicf.mainMessenger.ID(),
		},
	)
	if err != nil {
		return nil, err
	}

	return bicf.createTopicAndAssignHandler(identifier, interceptor, true)
}

// ------- PeerShard interceptor

func (bicf *baseInterceptorsContainerFactory) generatePeerShardInterceptor() error {
	identifier := common.ConnectionTopic

	interceptor, err := bicf.createPeerShardInterceptor(identifier, bicf.mainPeerShardMapper)
	if err != nil {
		return err
	}

	return bicf.addInterceptorsToContainers([]string{identifier}, []process.Interceptor{interceptor})
}

func (bicf *baseInterceptorsContainerFactory) createPeerShardInterceptor(
	identifier string,
	peerShardMapper process.PeerShardMapper,
) (process.Interceptor, error) {
	interceptedPeerShardFactory, err := interceptorFactory.NewInterceptedPeerShardFactory(*bicf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	argProcessor := processor.ArgPeerShardInterceptorProcessor{
		PeerShardMapper: peerShardMapper,
	}
	psiProcessor, err := processor.NewPeerShardInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	interceptor, err := interceptors.NewSingleDataInterceptor(
		interceptors.ArgSingleDataInterceptor{
			Topic:                identifier,
			DataFactory:          interceptedPeerShardFactory,
			Processor:            psiProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			CurrentPeerId:        bicf.mainMessenger.ID(),
			PreferredPeersHolder: bicf.preferredPeersHolder,
		},
	)
	if err != nil {
		return nil, err
	}

	return bicf.createTopicAndAssignHandler(identifier, interceptor, true)
}

func (bicf *baseInterceptorsContainerFactory) generateValidatorInfoInterceptor() error {
	identifier := common.ValidatorInfoTopic

	interceptedValidatorInfoFactory, err := interceptorFactory.NewInterceptedValidatorInfoDataFactory(*bicf.argInterceptorFactory)
	if err != nil {
		return err
	}

	internalMarshaller := bicf.argInterceptorFactory.CoreComponents.InternalMarshalizer()
	argProcessor := processor.ArgValidatorInfoInterceptorProcessor{
		ValidatorInfoPool: bicf.dataPool.ValidatorsInfo(),
	}

	validatorInfoProcessor, err := processor.NewValidatorInfoInterceptorProcessor(argProcessor)
	if err != nil {
		return err
	}

	mdInterceptor, err := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:                identifier,
			Marshalizer:          internalMarshaller,
			DataFactory:          interceptedValidatorInfoFactory,
			Processor:            validatorInfoProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			PreferredPeersHolder: bicf.preferredPeersHolder,
			CurrentPeerId:        bicf.mainMessenger.ID(),
		},
	)
	if err != nil {
		return err
	}

	interceptor, err := bicf.createTopicAndAssignHandler(identifier, mdInterceptor, true)
	if err != nil {
		return err
	}

	return bicf.addInterceptorsToContainers([]string{identifier}, []process.Interceptor{interceptor})
}

func (bicf *baseInterceptorsContainerFactory) addInterceptorsToContainers(keys []string, interceptors []process.Interceptor) error {
	err := bicf.mainContainer.AddMultiple(keys, interceptors)
	if err != nil {
		return err
	}

	if bicf.nodeOperationMode != common.FullArchiveMode {
		return nil
	}

	return bicf.fullArchiveContainer.AddMultiple(keys, interceptors)
}
