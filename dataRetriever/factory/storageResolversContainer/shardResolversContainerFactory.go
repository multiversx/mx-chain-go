package storageResolversContainers

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/storageResolvers"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

var _ dataRetriever.ResolversContainerFactory = (*shardResolversContainerFactory)(nil)

type shardResolversContainerFactory struct {
	*baseResolversContainerFactory
}

// NewShardResolversContainerFactory creates a new container filled with topic resolvers for shards
func NewShardResolversContainerFactory(
	args FactoryArgs,
) (*shardResolversContainerFactory, error) {
	container := containers.NewResolversContainer()
	base := &baseResolversContainerFactory{
		container:                container,
		shardCoordinator:         args.ShardCoordinator,
		messenger:                args.Messenger,
		store:                    args.Store,
		marshalizer:              args.Marshalizer,
		hasher:                   args.Hasher,
		uint64ByteSliceConverter: args.Uint64ByteSliceConverter,
		dataPacker:               args.DataPacker,
		manualEpochStartNotifier: args.ManualEpochStartNotifier,
		chanGracefullyClose:      args.ChanGracefullyClose,
		generalConfig:            args.GeneralConfig,
		shardIDForTries:          args.ShardIDForTries,
		chainID:                  args.ChainID,
		workingDir:               args.WorkingDirectory,
	}

	err := base.checkParams()
	if err != nil {
		return nil, err
	}

	return &shardResolversContainerFactory{
		baseResolversContainerFactory: base,
	}, nil
}

// Create returns a resolver container that will hold all resolvers in the system
func (srcf *shardResolversContainerFactory) Create() (dataRetriever.ResolversContainer, error) {
	err := srcf.generateTxResolvers(
		factory.TransactionTopic,
		dataRetriever.TransactionUnit,
	)
	if err != nil {
		return nil, err
	}

	err = srcf.generateTxResolvers(
		factory.UnsignedTransactionTopic,
		dataRetriever.UnsignedTransactionUnit,
	)
	if err != nil {
		return nil, err
	}

	err = srcf.generateRewardResolver(
		factory.RewardsTransactionTopic,
		dataRetriever.RewardTransactionUnit,
	)
	if err != nil {
		return nil, err
	}

	err = srcf.generateHeaderResolvers()
	if err != nil {
		return nil, err
	}

	err = srcf.generateMiniBlocksResolvers()
	if err != nil {
		return nil, err
	}

	err = srcf.generateMetablockHeaderResolvers()
	if err != nil {
		return nil, err
	}

	err = srcf.generateTrieNodesResolvers()
	if err != nil {
		return nil, err
	}

	return srcf.container, nil
}

//------- Hdr resolver

func (srcf *shardResolversContainerFactory) generateHeaderResolvers() error {
	shardC := srcf.shardCoordinator

	//only one shard header topic, for example: shardBlocks_0_META
	identifierHdr := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(core.MetachainShardId)

	hdrStorer := srcf.store.GetStorer(dataRetriever.BlockHeaderUnit)

	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardC.SelfId())
	hdrNonceStore := srcf.store.GetStorer(hdrNonceHashDataUnit)
	arg := storageResolvers.ArgHeaderResolver{
		Messenger:                srcf.messenger,
		ResponseTopicName:        identifierHdr,
		NonceConverter:           srcf.uint64ByteSliceConverter,
		HdrStorage:               hdrStorer,
		HeadersNoncesStorage:     hdrNonceStore,
		ManualEpochStartNotifier: srcf.manualEpochStartNotifier,
		ChanGracefullyClose:      srcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	resolver, err := storageResolvers.NewHeaderResolver(arg)
	if err != nil {
		return err
	}

	return srcf.container.Add(identifierHdr, resolver)
}

//------- MetaBlockHeaderResolvers

func (srcf *shardResolversContainerFactory) generateMetablockHeaderResolvers() error {
	//only one metachain header block topic
	//this is: metachainBlocks
	identifierHdr := factory.MetachainBlocksTopic
	hdrStorer := srcf.store.GetStorer(dataRetriever.MetaBlockUnit)

	hdrNonceStore := srcf.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	arg := storageResolvers.ArgHeaderResolver{
		Messenger:                srcf.messenger,
		ResponseTopicName:        identifierHdr,
		NonceConverter:           srcf.uint64ByteSliceConverter,
		HdrStorage:               hdrStorer,
		HeadersNoncesStorage:     hdrNonceStore,
		ManualEpochStartNotifier: srcf.manualEpochStartNotifier,
		ChanGracefullyClose:      srcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	resolver, err := storageResolvers.NewHeaderResolver(arg)
	if err != nil {
		return err
	}

	return srcf.container.Add(identifierHdr, resolver)
}

func (srcf *shardResolversContainerFactory) generateTrieNodesResolvers() error {
	shardC := srcf.shardCoordinator

	keys := make([]string, 0)
	resolversSlice := make([]dataRetriever.Resolver, 0)

	identifierTrieNodes := factory.AccountTrieNodesTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	storageManager, userAccountsDataTrie, err := srcf.newImportDBTrieStorage(srcf.generalConfig.AccountsTrieStorage)
	if err != nil {
		return fmt.Errorf("%w while creating user accounts data trie storage getter", err)
	}
	arg := storageResolvers.ArgTrieResolver{
		Messenger:                srcf.messenger,
		ResponseTopicName:        identifierTrieNodes,
		Marshalizer:              srcf.marshalizer,
		TrieDataGetter:           userAccountsDataTrie,
		TrieStorageManager:       storageManager,
		ManualEpochStartNotifier: srcf.manualEpochStartNotifier,
		ChanGracefullyClose:      srcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	resolver, err := storageResolvers.NewTrieNodeResolver(arg)
	if err != nil {
		return fmt.Errorf("%w while creating user accounts trie node resolver", err)
	}

	resolversSlice = append(resolversSlice, resolver)
	keys = append(keys, identifierTrieNodes)

	return srcf.container.AddMultiple(keys, resolversSlice)
}

func (srcf *shardResolversContainerFactory) generateRewardResolver(
	topic string,
	unit dataRetriever.UnitType,
) error {

	shardC := srcf.shardCoordinator

	keys := make([]string, 0)
	resolverSlice := make([]dataRetriever.Resolver, 0)

	identifierTx := topic + shardC.CommunicationIdentifier(core.MetachainShardId)
	resolver, err := srcf.createTxResolver(identifierTx, unit)
	if err != nil {
		return err
	}

	resolverSlice = append(resolverSlice, resolver)
	keys = append(keys, identifierTx)

	return srcf.container.AddMultiple(keys, resolverSlice)
}

// IsInterfaceNil returns true if there is no value under the interface
func (srcf *shardResolversContainerFactory) IsInterfaceNil() bool {
	return srcf == nil
}
