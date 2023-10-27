package storagerequesterscontainer

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/containers"
	"github.com/multiversx/mx-chain-go/dataRetriever/storageRequesters"
	"github.com/multiversx/mx-chain-go/process/factory"
)

var _ dataRetriever.RequestersContainerFactory = (*metaRequestersContainerFactory)(nil)

type metaRequestersContainerFactory struct {
	*baseRequestersContainerFactory
}

// NewMetaRequestersContainerFactory creates a new container filled with topic requesters for metachain
func NewMetaRequestersContainerFactory(
	args FactoryArgs,
) (*metaRequestersContainerFactory, error) {
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

	return &metaRequestersContainerFactory{
		baseRequestersContainerFactory: base,
	}, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (mrcf *metaRequestersContainerFactory) Create() (dataRetriever.RequestersContainer, error) {
	err := mrcf.generateCommonRequesters()
	if err != nil {
		return nil, err
	}

	err = mrcf.generateShardHeaderRequesters()
	if err != nil {
		return nil, err
	}

	err = mrcf.generateMetaChainHeaderRequesters()
	if err != nil {
		return nil, err
	}

	err = mrcf.generateRewardsRequesters(
		factory.RewardsTransactionTopic,
		dataRetriever.RewardTransactionUnit,
	)
	if err != nil {
		return nil, err
	}

	err = mrcf.generateTrieNodesRequesters()
	if err != nil {
		return nil, err
	}

	return mrcf.container, nil
}

func (mrcf *metaRequestersContainerFactory) generateShardHeaderRequesters() error {
	shardC := mrcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards)
	requestersSlice := make([]dataRetriever.Requester, noOfShards)

	// wire up to topics: shardBlocks_0_META, shardBlocks_1_META ...
	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierHeader := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(idx)
		requester, err := mrcf.createShardHeaderRequester(identifierHeader, idx)
		if err != nil {
			return err
		}

		requestersSlice[idx] = requester
		keys[idx] = identifierHeader
	}

	return mrcf.container.AddMultiple(keys, requestersSlice)
}

func (mrcf *metaRequestersContainerFactory) createShardHeaderRequester(
	responseTopicName string,
	shardID uint32,
) (dataRetriever.Requester, error) {
	hdrStorer, err := mrcf.store.GetStorer(dataRetriever.BlockHeaderUnit)
	if err != nil {
		return nil, err
	}

	// TODO change this data unit creation method through a factory or func
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardID)
	hdrNonceStore, err := mrcf.store.GetStorer(hdrNonceHashDataUnit)
	if err != nil {
		return nil, err
	}

	arg := storagerequesters.ArgHeaderRequester{
		Messenger:                mrcf.messenger,
		ResponseTopicName:        responseTopicName,
		NonceConverter:           mrcf.uint64ByteSliceConverter,
		HdrStorage:               hdrStorer,
		HeadersNoncesStorage:     hdrNonceStore,
		ManualEpochStartNotifier: mrcf.manualEpochStartNotifier,
		ChanGracefullyClose:      mrcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	requester, err := storagerequesters.NewHeaderRequester(arg)
	if err != nil {
		return nil, err
	}

	return requester, nil
}

func (mrcf *metaRequestersContainerFactory) generateMetaChainHeaderRequesters() error {
	identifierHeader := factory.MetachainBlocksTopic
	requester, err := mrcf.createMetaChainHeaderRequester()
	if err != nil {
		return err
	}

	return mrcf.container.Add(identifierHeader, requester)
}

func (mrcf *metaRequestersContainerFactory) createMetaChainHeaderRequester() (dataRetriever.Requester, error) {
	hdrStorer, err := mrcf.store.GetStorer(dataRetriever.MetaBlockUnit)
	if err != nil {
		return nil, err
	}

	hdrNonceStore, err := mrcf.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	if err != nil {
		return nil, err
	}

	arg := storagerequesters.ArgHeaderRequester{
		Messenger:                mrcf.messenger,
		ResponseTopicName:        factory.MetachainBlocksTopic,
		NonceConverter:           mrcf.uint64ByteSliceConverter,
		HdrStorage:               hdrStorer,
		HeadersNoncesStorage:     hdrNonceStore,
		ManualEpochStartNotifier: mrcf.manualEpochStartNotifier,
		ChanGracefullyClose:      mrcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	requester, err := storagerequesters.NewHeaderRequester(arg)
	if err != nil {
		return nil, err
	}

	return requester, nil
}

func (mrcf *metaRequestersContainerFactory) generateTrieNodesRequesters() error {
	keys := make([]string, 0)
	requestersSlice := make([]dataRetriever.Requester, 0)

	userAccountsStorer, err := mrcf.store.GetStorer(dataRetriever.UserAccountsUnit)
	if err != nil {
		return err
	}

	identifierTrieNodes := factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	storageManager, userAccountsDataTrie, err := mrcf.newImportDBTrieStorage(
		userAccountsStorer,
		dataRetriever.UserAccountsUnit,
		mrcf.enableEpochsHandler,
	)
	if err != nil {
		return fmt.Errorf("%w while creating user accounts data trie storage getter", err)
	}
	arg := storagerequesters.ArgTrieRequester{
		Messenger:                mrcf.messenger,
		ResponseTopicName:        identifierTrieNodes,
		Marshalizer:              mrcf.marshalizer,
		TrieDataGetter:           userAccountsDataTrie,
		TrieStorageManager:       storageManager,
		ManualEpochStartNotifier: mrcf.manualEpochStartNotifier,
		ChanGracefullyClose:      mrcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	requester, err := storagerequesters.NewTrieNodeRequester(arg)
	if err != nil {
		return fmt.Errorf("%w while creating user accounts trie node requester", err)
	}

	requestersSlice = append(requestersSlice, requester)
	keys = append(keys, identifierTrieNodes)

	peerAccountsStorer, err := mrcf.store.GetStorer(dataRetriever.PeerAccountsUnit)
	if err != nil {
		return err
	}

	identifierTrieNodes = factory.ValidatorTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	storageManager, peerAccountsDataTrie, err := mrcf.newImportDBTrieStorage(
		peerAccountsStorer,
		dataRetriever.PeerAccountsUnit,
		mrcf.enableEpochsHandler,
	)
	if err != nil {
		return fmt.Errorf("%w while creating peer accounts data trie storage getter", err)
	}
	arg = storagerequesters.ArgTrieRequester{
		Messenger:                mrcf.messenger,
		ResponseTopicName:        identifierTrieNodes,
		Marshalizer:              mrcf.marshalizer,
		TrieDataGetter:           peerAccountsDataTrie,
		TrieStorageManager:       storageManager,
		ManualEpochStartNotifier: mrcf.manualEpochStartNotifier,
		ChanGracefullyClose:      mrcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}

	requester, err = storagerequesters.NewTrieNodeRequester(arg)
	if err != nil {
		return fmt.Errorf("%w while creating peer accounts trie node requester", err)
	}

	requestersSlice = append(requestersSlice, requester)
	keys = append(keys, identifierTrieNodes)

	return mrcf.container.AddMultiple(keys, requestersSlice)
}

func (mrcf *metaRequestersContainerFactory) generateRewardsRequesters(
	topic string,
	unit dataRetriever.UnitType,
) error {

	shardC := mrcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards)
	requestersSlice := make([]dataRetriever.Requester, noOfShards)

	// wire up to topics: shardBlocks_0_META, shardBlocks_1_META ...
	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := topic + shardC.CommunicationIdentifier(idx)
		requester, err := mrcf.createTxRequester(identifierTx, unit)
		if err != nil {
			return err
		}

		requestersSlice[idx] = requester
		keys[idx] = identifierTx
	}

	return mrcf.container.AddMultiple(keys, requestersSlice)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mrcf *metaRequestersContainerFactory) IsInterfaceNil() bool {
	return mrcf == nil
}
