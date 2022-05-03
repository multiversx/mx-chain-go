package interceptorscontainer

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/dataValidators"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	interceptorFactory "github.com/ElrondNetwork/elrond-go/process/interceptors/factory"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
)

const numGoRoutines = 100
const chunksProcessorRequestInterval = time.Millisecond * 400
const minTimespanDurationInSec = int64(1)

type baseInterceptorsContainerFactory struct {
	container              process.InterceptorsContainer
	shardCoordinator       sharding.Coordinator
	accounts               state.AccountsAdapter
	store                  dataRetriever.StorageService
	dataPool               dataRetriever.PoolsHolder
	messenger              process.TopicHandler
	nodesCoordinator       nodesCoordinator.NodesCoordinator
	blockBlackList         process.TimeCacher
	argInterceptorFactory  *interceptorFactory.ArgInterceptedDataFactory
	globalThrottler        process.InterceptorThrottler
	maxTxNonceDeltaAllowed int
	antifloodHandler       process.P2PAntifloodHandler
	whiteListHandler       process.WhiteListHandler
	whiteListerVerifiedTxs process.WhiteListHandler
	preferredPeersHolder   process.PreferredPeersHolderHandler
	hasher                 hashing.Hasher
	requestHandler         process.RequestHandler
	peerShardMapper        process.PeerShardMapper
	hardforkTrigger        heartbeat.HardforkTrigger
}

func checkBaseParams(
	coreComponents process.CoreComponentsHolder,
	cryptoComponents process.CryptoComponentsHolder,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
	store dataRetriever.StorageService,
	dataPool dataRetriever.PoolsHolder,
	messenger process.TopicHandler,
	nodesCoordinator nodesCoordinator.NodesCoordinator,
	blackList process.TimeCacher,
	antifloodHandler process.P2PAntifloodHandler,
	whiteListHandler process.WhiteListHandler,
	whiteListerVerifiedTxs process.WhiteListHandler,
	preferredPeersHolder process.PreferredPeersHolderHandler,
	requestHandler process.RequestHandler,
	peerShardMapper process.PeerShardMapper,
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
	if check.IfNil(messenger) {
		return process.ErrNilMessenger
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
	if check.IfNil(cryptoComponents.MultiSigner()) {
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
	if check.IfNil(peerShardMapper) {
		return process.ErrNilPeerShardMapper
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

	err := bicf.messenger.CreateTopic(topic, createChannel)
	if err != nil {
		return nil, err
	}

	return interceptor, bicf.messenger.RegisterMessageProcessor(topic, common.DefaultInterceptorsIdentifier, interceptor)
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

	return bicf.container.AddMultiple(keys, interceptorSlice)
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
		bicf.maxTxNonceDeltaAllowed,
	)
	if err != nil {
		return nil, err
	}

	argProcessor := &processor.ArgTxInterceptorProcessor{
		ShardedDataCache: bicf.dataPool.Transactions(),
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

	internalMarshalizer := bicf.argInterceptorFactory.CoreComponents.InternalMarshalizer()
	interceptor, err := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:                topic,
			Marshalizer:          internalMarshalizer,
			DataFactory:          txFactory,
			Processor:            txProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			CurrentPeerId:        bicf.messenger.ID(),
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

	internalMarshalizer := bicf.argInterceptorFactory.CoreComponents.InternalMarshalizer()
	interceptor, err := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:                topic,
			Marshalizer:          internalMarshalizer,
			DataFactory:          txFactory,
			Processor:            txProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			CurrentPeerId:        bicf.messenger.ID(),
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

	internalMarshalizer := bicf.argInterceptorFactory.CoreComponents.InternalMarshalizer()
	interceptor, err := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:                topic,
			Marshalizer:          internalMarshalizer,
			DataFactory:          txFactory,
			Processor:            txProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			CurrentPeerId:        bicf.messenger.ID(),
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
			CurrentPeerId:        bicf.messenger.ID(),
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

	return bicf.container.Add(identifierHdr, interceptor)
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

	allShardsMiniBlocksInterceptorinterceptor, err := bicf.createOneMiniBlocksInterceptor(identifierAllShardsMiniBlocks)
	if err != nil {
		return err
	}

	keys[noOfShards+1] = identifierAllShardsMiniBlocks
	interceptorsSlice[noOfShards+1] = allShardsMiniBlocksInterceptorinterceptor

	return bicf.container.AddMultiple(keys, interceptorsSlice)
}

func (bicf *baseInterceptorsContainerFactory) createOneMiniBlocksInterceptor(topic string) (process.Interceptor, error) {
	internalMarshalizer := bicf.argInterceptorFactory.CoreComponents.InternalMarshalizer()
	hasher := bicf.argInterceptorFactory.CoreComponents.Hasher()
	argProcessor := &processor.ArgMiniblockInterceptorProcessor{
		MiniblockCache:   bicf.dataPool.MiniBlocks(),
		Marshalizer:      internalMarshalizer,
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
			Marshalizer:          internalMarshalizer,
			DataFactory:          miniblockFactory,
			Processor:            miniblockProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			CurrentPeerId:        bicf.messenger.ID(),
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
			CurrentPeerId:        bicf.messenger.ID(),
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

	return bicf.container.Add(identifierHdr, interceptor)
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

	internalMarshalizer := bicf.argInterceptorFactory.CoreComponents.InternalMarshalizer()
	interceptor, err := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:                topic,
			Marshalizer:          internalMarshalizer,
			DataFactory:          trieNodesFactory,
			Processor:            trieNodesProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			CurrentPeerId:        bicf.messenger.ID(),
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

	return bicf.container.AddMultiple(keys, interceptorsSlice)
}

//------- PeerAuthentication interceptor

func (bicf *baseInterceptorsContainerFactory) generatePeerAuthenticationInterceptor() error {
	identifierPeerAuthentication := common.PeerAuthenticationTopic

	internalMarshaller := bicf.argInterceptorFactory.CoreComponents.InternalMarshalizer()
	argProcessor := processor.ArgPeerAuthenticationInterceptorProcessor{
		PeerAuthenticationCacher: bicf.dataPool.PeerAuthentications(),
		PeerShardMapper:          bicf.peerShardMapper,
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
			CurrentPeerId:        bicf.messenger.ID(),
		},
	)
	if err != nil {
		return err
	}

	interceptor, err := bicf.createTopicAndAssignHandler(identifierPeerAuthentication, mdInterceptor, true)
	if err != nil {
		return err
	}

	return bicf.container.Add(identifierPeerAuthentication, interceptor)
}

//------- Heartbeat interceptor

func (bicf *baseInterceptorsContainerFactory) generateHeartbeatInterceptor() error {
	shardC := bicf.shardCoordinator
	identifierHeartbeat := common.HeartbeatV2Topic + shardC.CommunicationIdentifier(shardC.SelfId())

	argHeartbeatProcessor := processor.ArgHeartbeatInterceptorProcessor{
		HeartbeatCacher:  bicf.dataPool.Heartbeats(),
		ShardCoordinator: shardC,
		PeerShardMapper:  bicf.peerShardMapper,
	}
	heartbeatProcessor, err := processor.NewHeartbeatInterceptorProcessor(argHeartbeatProcessor)
	if err != nil {
		return err
	}

	heartbeatFactory, err := interceptorFactory.NewInterceptedHeartbeatDataFactory(*bicf.argInterceptorFactory)
	if err != nil {
		return err
	}

	internalMarshalizer := bicf.argInterceptorFactory.CoreComponents.InternalMarshalizer()
	mdInterceptor, err := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:                identifierHeartbeat,
			Marshalizer:          internalMarshalizer,
			DataFactory:          heartbeatFactory,
			Processor:            heartbeatProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			PreferredPeersHolder: bicf.preferredPeersHolder,
			CurrentPeerId:        bicf.messenger.ID(),
		},
	)
	if err != nil {
		return err
	}

	interceptor, err := bicf.createTopicAndAssignHandler(identifierHeartbeat, mdInterceptor, true)
	if err != nil {
		return err
	}

	return bicf.container.Add(identifierHeartbeat, interceptor)
}

// ------- ValidatorInfo interceptor

func (bicf *baseInterceptorsContainerFactory) generateValidatorInfoInterceptor() error {
	identifier := common.ConnectionTopic

	interceptedValidatorInfoFactory, err := interceptorFactory.NewInterceptedValidatorInfoFactory(*bicf.argInterceptorFactory)
	if err != nil {
		return err
	}

	argProcessor := processor.ArgValidatorInfoInterceptorProcessor{
		PeerShardMapper: bicf.peerShardMapper,
	}
	hdrProcessor, err := processor.NewValidatorInfoInterceptorProcessor(argProcessor)
	if err != nil {
		return err
	}

	interceptor, err := interceptors.NewSingleDataInterceptor(
		interceptors.ArgSingleDataInterceptor{
			Topic:                identifier,
			DataFactory:          interceptedValidatorInfoFactory,
			Processor:            hdrProcessor,
			Throttler:            bicf.globalThrottler,
			AntifloodHandler:     bicf.antifloodHandler,
			WhiteListRequest:     bicf.whiteListHandler,
			CurrentPeerId:        bicf.messenger.ID(),
			PreferredPeersHolder: bicf.preferredPeersHolder,
		},
	)
	if err != nil {
		return err
	}

	_, err = bicf.createTopicAndAssignHandler(identifier, interceptor, true)
	if err != nil {
		return err
	}

	return bicf.container.Add(identifier, interceptor)
}
