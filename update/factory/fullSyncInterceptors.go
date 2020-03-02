package factory

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/throttler"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/dataValidators"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	interceptorFactory "github.com/ElrondNetwork/elrond-go/process/interceptors/factory"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

var _ process.InterceptorsContainerFactory = (*fullSyncInterceptorsContainerFactory)(nil)

// fullSyncInterceptorsContainerFactory will handle the creation the interceptors container for shards
type fullSyncInterceptorsContainerFactory struct {
	container              process.InterceptorsContainer
	shardCoordinator       sharding.Coordinator
	accounts               state.AccountsAdapter
	marshalizer            marshal.Marshalizer
	hasher                 hashing.Hasher
	store                  dataRetriever.StorageService
	dataPool               dataRetriever.PoolsHolder
	messenger              process.TopicHandler
	multiSigner            crypto.MultiSigner
	nodesCoordinator       sharding.NodesCoordinator
	blackList              process.BlackListHandler
	argInterceptorFactory  *interceptorFactory.ArgInterceptedDataFactory
	globalThrottler        process.InterceptorThrottler
	maxTxNonceDeltaAllowed int
	keyGen                 crypto.KeyGenerator
	singleSigner           crypto.SingleSigner
	addrConverter          state.AddressConverter
	whiteListHandler       dataRetriever.WhiteListHandler
}

// ArgsNewFullSyncInterceptorsContainerFactory holds the arguments needed for fullSyncInterceptorsContainerFactory
type ArgsNewFullSyncInterceptorsContainerFactory struct {
	Accounts               state.AccountsAdapter
	ShardCoordinator       sharding.Coordinator
	NodesCoordinator       sharding.NodesCoordinator
	Messenger              process.TopicHandler
	Store                  dataRetriever.StorageService
	Marshalizer            marshal.Marshalizer
	Hasher                 hashing.Hasher
	KeyGen                 crypto.KeyGenerator
	BlockSignKeyGen        crypto.KeyGenerator
	SingleSigner           crypto.SingleSigner
	BlockSingleSigner      crypto.SingleSigner
	MultiSigner            crypto.MultiSigner
	DataPool               dataRetriever.PoolsHolder
	AddrConverter          state.AddressConverter
	MaxTxNonceDeltaAllowed int
	TxFeeHandler           process.FeeHandler
	BlackList              process.BlackListHandler
	HeaderSigVerifier      process.InterceptedHeaderSigVerifier
	ChainID                []byte
	SizeCheckDelta         uint32
	ValidityAttester       process.ValidityAttester
	EpochStartTrigger      process.EpochStartTriggerHandler
	WhiteListHandler       dataRetriever.WhiteListHandler
	InterceptorsContainer  process.InterceptorsContainer
}

// NewFullSyncInterceptorsContainerFactory is responsible for creating a new interceptors factory object
func NewFullSyncInterceptorsContainerFactory(
	args ArgsNewFullSyncInterceptorsContainerFactory,
) (*fullSyncInterceptorsContainerFactory, error) {
	if args.SizeCheckDelta > 0 {
		args.Marshalizer = marshal.NewSizeCheckUnmarshalizer(args.Marshalizer, args.SizeCheckDelta)
	}
	err := checkBaseParams(
		args.ShardCoordinator,
		args.Accounts,
		args.Marshalizer,
		args.Hasher,
		args.Store,
		args.DataPool,
		args.Messenger,
		args.MultiSigner,
		args.NodesCoordinator,
		args.BlackList,
	)
	if err != nil {
		return nil, err
	}

	if check.IfNil(args.KeyGen) {
		return nil, process.ErrNilKeyGen
	}
	if check.IfNil(args.SingleSigner) {
		return nil, process.ErrNilSingleSigner
	}
	if check.IfNil(args.AddrConverter) {
		return nil, process.ErrNilAddressConverter
	}
	if check.IfNil(args.TxFeeHandler) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.BlockSignKeyGen) {
		return nil, process.ErrNilKeyGen
	}
	if check.IfNil(args.BlockSingleSigner) {
		return nil, process.ErrNilSingleSigner
	}
	if check.IfNil(args.HeaderSigVerifier) {
		return nil, process.ErrNilHeaderSigVerifier
	}
	if len(args.ChainID) == 0 {
		return nil, process.ErrInvalidChainID
	}
	if check.IfNil(args.ValidityAttester) {
		return nil, process.ErrNilValidityAttester
	}
	if check.IfNil(args.EpochStartTrigger) {
		return nil, process.ErrNilEpochStartTrigger
	}
	if check.IfNil(args.InterceptorsContainer) {
		return nil, update.ErrNilInterceptorsContainer
	}
	if check.IfNil(args.WhiteListHandler) {
		return nil, update.ErrNilWhiteListHandler
	}

	argInterceptorFactory := &interceptorFactory.ArgInterceptedDataFactory{
		Marshalizer:       args.Marshalizer,
		Hasher:            args.Hasher,
		ShardCoordinator:  args.ShardCoordinator,
		MultiSigVerifier:  args.MultiSigner,
		NodesCoordinator:  args.NodesCoordinator,
		KeyGen:            args.KeyGen,
		BlockKeyGen:       args.BlockSignKeyGen,
		Signer:            args.SingleSigner,
		BlockSigner:       args.BlockSingleSigner,
		AddrConv:          args.AddrConverter,
		FeeHandler:        args.TxFeeHandler,
		HeaderSigVerifier: args.HeaderSigVerifier,
		ChainID:           args.ChainID,
		ValidityAttester:  args.ValidityAttester,
		EpochStartTrigger: args.EpochStartTrigger,
	}

	icf := &fullSyncInterceptorsContainerFactory{
		container:              args.InterceptorsContainer,
		accounts:               args.Accounts,
		shardCoordinator:       args.ShardCoordinator,
		messenger:              args.Messenger,
		store:                  args.Store,
		marshalizer:            args.Marshalizer,
		hasher:                 args.Hasher,
		multiSigner:            args.MultiSigner,
		dataPool:               args.DataPool,
		nodesCoordinator:       args.NodesCoordinator,
		argInterceptorFactory:  argInterceptorFactory,
		blackList:              args.BlackList,
		maxTxNonceDeltaAllowed: args.MaxTxNonceDeltaAllowed,
		keyGen:                 args.KeyGen,
		singleSigner:           args.SingleSigner,
		addrConverter:          args.AddrConverter,
		whiteListHandler:       args.WhiteListHandler,
	}

	icf.globalThrottler, err = throttler.NewNumGoRoutineThrottler(numGoRoutines)
	if err != nil {
		return nil, err
	}

	return icf, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (sicf *fullSyncInterceptorsContainerFactory) Create() (process.InterceptorsContainer, error) {
	err := sicf.generateTxInterceptors()
	if err != nil {
		return nil, err
	}

	err = sicf.generateUnsignedTxsInterceptorsForShard()
	if err != nil {
		return nil, err
	}

	err = sicf.generateRewardTxInterceptor()
	if err != nil {
		return nil, err
	}

	err = sicf.generateHeaderInterceptors()
	if err != nil {
		return nil, err
	}

	err = sicf.generateMiniBlocksInterceptors()
	if err != nil {
		return nil, err
	}

	err = sicf.generateMetachainHeaderInterceptors()
	if err != nil {
		return nil, err
	}

	err = sicf.generateTrieNodesInterceptors()
	if err != nil {
		return nil, err
	}

	err = sicf.generateMetablockInterceptors()
	if err != nil {
		return nil, err
	}

	err = sicf.generateShardHeaderInterceptors()
	if err != nil {
		return nil, err
	}

	err = sicf.generateTxInterceptors()
	if err != nil {
		return nil, err
	}

	err = sicf.generateUnsignedTxsInterceptors()
	if err != nil {
		return nil, err
	}

	err = sicf.generateRewardTxInterceptors()
	if err != nil {
		return nil, err
	}

	err = sicf.generateMiniBlocksInterceptors()
	if err != nil {
		return nil, err
	}

	err = sicf.generateTrieNodesInterceptors()
	if err != nil {
		return nil, err
	}

	return sicf.container, nil
}

//------- Unsigned transactions interceptors

func (sicf *fullSyncInterceptorsContainerFactory) generateUnsignedTxsInterceptorsForShard() error {
	err := sicf.generateUnsignedTxsInterceptors()
	if err != nil {
		return err
	}

	shardC := sicf.shardCoordinator
	identifierTx := factory.UnsignedTransactionTopic + shardC.CommunicationIdentifier(core.MetachainShardId)

	interceptor, err := sicf.createOneUnsignedTxInterceptor(identifierTx)
	if err != nil {
		return err
	}

	return sicf.container.Add(identifierTx, interceptor)
}

func (sicf *fullSyncInterceptorsContainerFactory) generateTrieNodesInterceptors() error {
	shardC := sicf.shardCoordinator

	keys := make([]string, 0)
	interceptorsSlice := make([]process.Interceptor, 0)

	identifierTrieNodes := factory.AccountTrieNodesTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	interceptor, err := sicf.createOneTrieNodesInterceptor(identifierTrieNodes)
	if err != nil {
		return err
	}

	keys = append(keys, identifierTrieNodes)
	interceptorsSlice = append(interceptorsSlice, interceptor)

	identifierTrieNodes = factory.ValidatorTrieNodesTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	interceptor, err = sicf.createOneTrieNodesInterceptor(identifierTrieNodes)
	if err != nil {
		return err
	}

	keys = append(keys, identifierTrieNodes)
	interceptorsSlice = append(interceptorsSlice, interceptor)

	return sicf.container.AddMultiple(keys, interceptorsSlice)
}

//------- Reward transactions interceptors

func (sicf *fullSyncInterceptorsContainerFactory) generateRewardTxInterceptor() error {
	shardC := sicf.shardCoordinator

	keys := make([]string, 0)
	interceptorSlice := make([]process.Interceptor, 0)

	identifierTx := factory.RewardsTransactionTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	interceptor, err := sicf.createOneRewardTxInterceptor(identifierTx)
	if err != nil {
		return err
	}

	keys = append(keys, identifierTx)
	interceptorSlice = append(interceptorSlice, interceptor)

	return sicf.container.AddMultiple(keys, interceptorSlice)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sicf *fullSyncInterceptorsContainerFactory) IsInterfaceNil() bool {
	return sicf == nil
}

const numGoRoutines = 2000

func checkBaseParams(
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	store dataRetriever.StorageService,
	dataPool dataRetriever.PoolsHolder,
	messenger process.TopicHandler,
	multiSigner crypto.MultiSigner,
	nodesCoordinator sharding.NodesCoordinator,
	blackList process.BlackListHandler,
) error {
	if check.IfNil(shardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(messenger) {
		return process.ErrNilMessenger
	}
	if check.IfNil(store) {
		return process.ErrNilStore
	}
	if check.IfNil(marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(multiSigner) {
		return process.ErrNilMultiSigVerifier
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
		return process.ErrNilBlackListHandler
	}

	return nil
}

func (bicf *fullSyncInterceptorsContainerFactory) createTopicAndAssignHandler(
	topic string,
	interceptor process.Interceptor,
	createChannel bool,
) (process.Interceptor, error) {

	err := bicf.messenger.CreateTopic(topic, createChannel)
	if err != nil {
		return nil, err
	}

	return interceptor, bicf.messenger.RegisterMessageProcessor(topic, interceptor)
}

//------- Tx interceptors

func (bicf *fullSyncInterceptorsContainerFactory) generateTxInterceptors() error {
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

	//tx interceptor for metachain topic
	identifierTx := factory.TransactionTopic + shardC.CommunicationIdentifier(core.MetachainShardId)

	interceptor, err := bicf.createOneTxInterceptor(identifierTx)
	if err != nil {
		return err
	}

	keys = append(keys, identifierTx)
	interceptorSlice = append(interceptorSlice, interceptor)

	return bicf.container.AddMultiple(keys, interceptorSlice)
}

func (bicf *fullSyncInterceptorsContainerFactory) createOneTxInterceptor(topic string) (process.Interceptor, error) {
	txValidator, err := dataValidators.NewTxValidator(bicf.accounts, bicf.shardCoordinator, bicf.maxTxNonceDeltaAllowed)
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

	interceptor, err := interceptors.NewMultiDataInterceptor(
		bicf.marshalizer,
		txFactory,
		txProcessor,
		bicf.globalThrottler,
	)
	if err != nil {
		return nil, err
	}

	return bicf.createTopicAndAssignHandler(topic, interceptor, true)
}

func (bicf *fullSyncInterceptorsContainerFactory) createOneUnsignedTxInterceptor(topic string) (process.Interceptor, error) {
	//TODO replace the nil tx validator with white list validator
	txValidator, err := mock.NewNilTxValidator()
	if err != nil {
		return nil, err
	}

	argProcessor := &processor.ArgTxInterceptorProcessor{
		ShardedDataCache: bicf.dataPool.UnsignedTransactions(),
		TxValidator:      txValidator,
	}
	txProcessor, err := processor.NewTxInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	txFactory, err := interceptorFactory.NewInterceptedUnsignedTxDataFactory(bicf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	interceptor, err := interceptors.NewMultiDataInterceptor(
		bicf.marshalizer,
		txFactory,
		txProcessor,
		bicf.globalThrottler,
	)
	if err != nil {
		return nil, err
	}

	return bicf.createTopicAndAssignHandler(topic, interceptor, true)
}

func (bicf *fullSyncInterceptorsContainerFactory) createOneRewardTxInterceptor(topic string) (process.Interceptor, error) {
	//TODO replace the nil tx validator with white list validator
	txValidator, err := mock.NewNilTxValidator()
	if err != nil {
		return nil, err
	}

	argProcessor := &processor.ArgTxInterceptorProcessor{
		ShardedDataCache: bicf.dataPool.RewardTransactions(),
		TxValidator:      txValidator,
	}
	txProcessor, err := processor.NewTxInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	txFactory, err := interceptorFactory.NewInterceptedRewardTxDataFactory(bicf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	interceptor, err := interceptors.NewMultiDataInterceptor(
		bicf.marshalizer,
		txFactory,
		txProcessor,
		bicf.globalThrottler,
	)
	if err != nil {
		return nil, err
	}

	return bicf.createTopicAndAssignHandler(topic, interceptor, true)
}

//------- Hdr interceptor

func (bicf *fullSyncInterceptorsContainerFactory) generateHeaderInterceptors() error {
	shardC := bicf.shardCoordinator
	//TODO implement other HeaderHandlerProcessValidator that will check the header's nonce
	// against blockchain's latest nonce - k finality
	hdrValidator, err := dataValidators.NewNilHeaderValidator()
	if err != nil {
		return err
	}

	hdrFactory, err := interceptorFactory.NewInterceptedShardHeaderDataFactory(bicf.argInterceptorFactory)
	if err != nil {
		return err
	}

	argProcessor := &processor.ArgHdrInterceptorProcessor{
		Headers:      bicf.dataPool.Headers(),
		HdrValidator: hdrValidator,
		BlackList:    bicf.blackList,
	}
	hdrProcessor, err := processor.NewHdrInterceptorProcessor(argProcessor)
	if err != nil {
		return err
	}

	//only one intrashard header topic
	interceptor, err := interceptors.NewSingleDataInterceptor(
		hdrFactory,
		hdrProcessor,
		bicf.globalThrottler,
	)
	if err != nil {
		return err
	}

	// compose header shard topic, for example: shardBlocks_0_META
	identifierHdr := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	_, err = bicf.createTopicAndAssignHandler(identifierHdr, interceptor, true)
	if err != nil {
		return err
	}

	return bicf.container.Add(identifierHdr, interceptor)
}

//------- MiniBlocks interceptors

func (bicf *fullSyncInterceptorsContainerFactory) generateMiniBlocksInterceptors() error {
	shardC := bicf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards+1)
	interceptorsSlice := make([]process.Interceptor, noOfShards+1)

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

	return bicf.container.AddMultiple(keys, interceptorsSlice)
}

func (bicf *fullSyncInterceptorsContainerFactory) createOneMiniBlocksInterceptor(topic string) (process.Interceptor, error) {
	argProcessor := &processor.ArgTxBodyInterceptorProcessor{
		MiniblockCache:   bicf.dataPool.MiniBlocks(),
		Marshalizer:      bicf.marshalizer,
		Hasher:           bicf.hasher,
		ShardCoordinator: bicf.shardCoordinator,
	}
	txBlockBodyProcessor, err := processor.NewTxBodyInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	txFactory, err := interceptorFactory.NewInterceptedTxBlockBodyDataFactory(bicf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	interceptor, err := interceptors.NewSingleDataInterceptor(
		txFactory,
		txBlockBodyProcessor,
		bicf.globalThrottler,
	)
	if err != nil {
		return nil, err
	}

	return bicf.createTopicAndAssignHandler(topic, interceptor, true)
}

//------- MetachainHeader interceptors

func (bicf *fullSyncInterceptorsContainerFactory) generateMetachainHeaderInterceptors() error {
	identifierHdr := factory.MetachainBlocksTopic
	//TODO implement other HeaderHandlerProcessValidator that will check the header's nonce
	// against blockchain's latest nonce - k finality
	hdrValidator, err := dataValidators.NewNilHeaderValidator()
	if err != nil {
		return err
	}

	hdrFactory, err := interceptorFactory.NewInterceptedMetaHeaderDataFactory(bicf.argInterceptorFactory)
	if err != nil {
		return err
	}

	argProcessor := &processor.ArgHdrInterceptorProcessor{
		Headers:      bicf.dataPool.Headers(),
		HdrValidator: hdrValidator,
		BlackList:    bicf.blackList,
	}
	hdrProcessor, err := processor.NewHdrInterceptorProcessor(argProcessor)
	if err != nil {
		return err
	}

	//only one metachain header topic
	interceptor, err := interceptors.NewSingleDataInterceptor(
		hdrFactory,
		hdrProcessor,
		bicf.globalThrottler,
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

func (bicf *fullSyncInterceptorsContainerFactory) createOneTrieNodesInterceptor(topic string) (process.Interceptor, error) {
	trieNodesProcessor, err := processor.NewTrieNodesInterceptorProcessor(bicf.dataPool.TrieNodes())
	if err != nil {
		return nil, err
	}

	trieNodesFactory, err := interceptorFactory.NewInterceptedTrieNodeDataFactory(bicf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	interceptor, err := interceptors.NewMultiDataInterceptor(
		bicf.marshalizer,
		trieNodesFactory,
		trieNodesProcessor,
		bicf.globalThrottler,
	)
	if err != nil {
		return nil, err
	}

	return bicf.createTopicAndAssignHandler(topic, interceptor, true)
}

func (bicf *fullSyncInterceptorsContainerFactory) generateUnsignedTxsInterceptors() error {
	shardC := bicf.shardCoordinator

	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards)
	interceptorsSlice := make([]process.Interceptor, noOfShards)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierScr := factory.UnsignedTransactionTopic + shardC.CommunicationIdentifier(idx)
		interceptor, err := bicf.createOneUnsignedTxInterceptor(identifierScr)
		if err != nil {
			return err
		}

		keys[int(idx)] = identifierScr
		interceptorsSlice[int(idx)] = interceptor
	}

	return bicf.container.AddMultiple(keys, interceptorsSlice)
}

func (bicf *fullSyncInterceptorsContainerFactory) setWhiteListHandlerToInterceptors() error {
	var err error

	bicf.container.Iterate(func(key string, interceptor process.Interceptor) bool {
		errFound := interceptor.SetIsDataForCurrentShardVerifier(bicf.whiteListHandler)
		if errFound != nil {
			err = errFound
			return false
		}
		return true
	})

	return err
}
