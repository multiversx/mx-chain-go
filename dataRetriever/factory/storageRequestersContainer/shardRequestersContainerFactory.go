package storagerequesterscontainer

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/containers"
	storagerequesters "github.com/multiversx/mx-chain-go/dataRetriever/storageRequesters"
	"github.com/multiversx/mx-chain-go/process/factory"
)

var _ dataRetriever.RequestersContainerFactory = (*shardRequestersContainerFactory)(nil)

type shardRequestersContainerFactory struct {
	*baseRequestersContainerFactory
}

// NewShardRequestersContainerFactory creates a new container filled with topic requesters for shards
func NewShardRequestersContainerFactory(
	args FactoryArgs,
) (*shardRequestersContainerFactory, error) {
	container := containers.NewRequestersContainer()
	base := &baseRequestersContainerFactory{
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
		snapshotsEnabled:         args.GeneralConfig.StateTriesConfig.SnapshotsEnabled,
		enableEpochsHandler:      args.EnableEpochsHandler,
	}

	err := base.checkParams()
	if err != nil {
		return nil, err
	}

	return &shardRequestersContainerFactory{
		baseRequestersContainerFactory: base,
	}, nil
}

// Create returns a requester container that will hold all requesters in the system
func (srcf *shardRequestersContainerFactory) Create() (dataRetriever.RequestersContainer, error) {
	err := srcf.generateCommonRequesters()
	if err != nil {
		return nil, err
	}

	err = srcf.generateRewardRequester(
		factory.RewardsTransactionTopic,
		dataRetriever.RewardTransactionUnit,
	)
	if err != nil {
		return nil, err
	}

	err = srcf.generateHeaderRequesters()
	if err != nil {
		return nil, err
	}

	err = srcf.generateMetablockHeaderRequesters()
	if err != nil {
		return nil, err
	}

	err = srcf.generateTrieNodesRequesters()
	if err != nil {
		return nil, err
	}

	return srcf.container, nil
}

func (srcf *shardRequestersContainerFactory) generateHeaderRequesters() error {
	shardC := srcf.shardCoordinator

	// only one shard header topic, for example: shardBlocks_0_META
	identifierHdr := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(core.MetachainShardId)

	hdrStorer, err := srcf.store.GetStorer(dataRetriever.BlockHeaderUnit)
	if err != nil {
		return err
	}

	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardC.SelfId())
	hdrNonceStore, err := srcf.store.GetStorer(hdrNonceHashDataUnit)
	if err != nil {
		return err
	}

	arg := storagerequesters.ArgHeaderRequester{
		Messenger:                srcf.messenger,
		ResponseTopicName:        identifierHdr,
		NonceConverter:           srcf.uint64ByteSliceConverter,
		HdrStorage:               hdrStorer,
		HeadersNoncesStorage:     hdrNonceStore,
		ManualEpochStartNotifier: srcf.manualEpochStartNotifier,
		ChanGracefullyClose:      srcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	requester, err := storagerequesters.NewHeaderRequester(arg)
	if err != nil {
		return err
	}

	return srcf.container.Add(identifierHdr, requester)
}

func (srcf *shardRequestersContainerFactory) generateMetablockHeaderRequesters() error {
	// only one metachain header block topic
	// this is: metachainBlocks
	identifierHdr := factory.MetachainBlocksTopic
	hdrStorer, err := srcf.store.GetStorer(dataRetriever.MetaBlockUnit)
	if err != nil {
		return err
	}

	hdrNonceStore, err := srcf.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	if err != nil {
		return err
	}

	arg := storagerequesters.ArgHeaderRequester{
		Messenger:                srcf.messenger,
		ResponseTopicName:        identifierHdr,
		NonceConverter:           srcf.uint64ByteSliceConverter,
		HdrStorage:               hdrStorer,
		HeadersNoncesStorage:     hdrNonceStore,
		ManualEpochStartNotifier: srcf.manualEpochStartNotifier,
		ChanGracefullyClose:      srcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	requester, err := storagerequesters.NewHeaderRequester(arg)
	if err != nil {
		return err
	}

	return srcf.container.Add(identifierHdr, requester)
}

func (srcf *shardRequestersContainerFactory) generateTrieNodesRequesters() error {
	shardC := srcf.shardCoordinator

	keys := make([]string, 0)
	requestersSlice := make([]dataRetriever.Requester, 0)

	userAccountsStorer, err := srcf.store.GetStorer(dataRetriever.UserAccountsUnit)
	if err != nil {
		return err
	}

	identifierTrieNodes := factory.AccountTrieNodesTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	storageManager, userAccountsDataTrie, err := srcf.newImportDBTrieStorage(
		userAccountsStorer,
		dataRetriever.UserAccountsUnit,
		srcf.enableEpochsHandler,
	)
	if err != nil {
		return fmt.Errorf("%w while creating user accounts data trie storage getter", err)
	}
	arg := storagerequesters.ArgTrieRequester{
		Messenger:                srcf.messenger,
		ResponseTopicName:        identifierTrieNodes,
		Marshalizer:              srcf.marshalizer,
		TrieDataGetter:           userAccountsDataTrie,
		TrieStorageManager:       storageManager,
		ManualEpochStartNotifier: srcf.manualEpochStartNotifier,
		ChanGracefullyClose:      srcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	requester, err := storagerequesters.NewTrieNodeRequester(arg)
	if err != nil {
		return fmt.Errorf("%w while creating user accounts trie node requester", err)
	}

	requestersSlice = append(requestersSlice, requester)
	keys = append(keys, identifierTrieNodes)

	return srcf.container.AddMultiple(keys, requestersSlice)
}

func (srcf *shardRequestersContainerFactory) generateRewardRequester(
	topic string,
	unit dataRetriever.UnitType,
) error {

	shardC := srcf.shardCoordinator

	keys := make([]string, 0)
	requesterSlice := make([]dataRetriever.Requester, 0)

	identifierTx := topic + shardC.CommunicationIdentifier(core.MetachainShardId)
	requester, err := srcf.createTxRequester(identifierTx, unit)
	if err != nil {
		return err
	}

	requesterSlice = append(requesterSlice, requester)
	keys = append(keys, identifierTx)

	return srcf.container.AddMultiple(keys, requesterSlice)
}

// IsInterfaceNil returns true if there is no value under the interface
func (srcf *shardRequestersContainerFactory) IsInterfaceNil() bool {
	return srcf == nil
}
