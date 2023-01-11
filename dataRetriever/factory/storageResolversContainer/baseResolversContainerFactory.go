package storageResolversContainers

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	disabledResolvers "github.com/multiversx/mx-chain-go/dataRetriever/resolvers/disabled"
	"github.com/multiversx/mx-chain-go/dataRetriever/storageResolvers"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	trieFactory "github.com/multiversx/mx-chain-go/trie/factory"
)

const defaultBeforeGracefulClose = time.Minute

type baseResolversContainerFactory struct {
	container                dataRetriever.ResolversContainer
	shardCoordinator         sharding.Coordinator
	messenger                dataRetriever.TopicMessageHandler
	store                    dataRetriever.StorageService
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	dataPacker               dataRetriever.DataPacker
	manualEpochStartNotifier dataRetriever.ManualEpochStartNotifier
	chanGracefullyClose      chan endProcess.ArgEndProcess
	generalConfig            config.Config
	shardIDForTries          uint32
	chainID                  string
	workingDir               string
}

func (brcf *baseResolversContainerFactory) checkParams() error {
	if check.IfNil(brcf.shardCoordinator) {
		return dataRetriever.ErrNilShardCoordinator
	}
	if check.IfNil(brcf.messenger) {
		return dataRetriever.ErrNilMessenger
	}
	if check.IfNil(brcf.store) {
		return dataRetriever.ErrNilStore
	}
	if check.IfNil(brcf.marshalizer) {
		return dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(brcf.uint64ByteSliceConverter) {
		return dataRetriever.ErrNilUint64ByteSliceConverter
	}
	if check.IfNil(brcf.dataPacker) {
		return dataRetriever.ErrNilDataPacker
	}
	if check.IfNil(brcf.manualEpochStartNotifier) {
		return dataRetriever.ErrNilManualEpochStartNotifier
	}
	if brcf.chanGracefullyClose == nil {
		return dataRetriever.ErrNilGracefullyCloseChannel
	}
	if check.IfNil(brcf.hasher) {
		return dataRetriever.ErrNilHasher
	}

	return nil
}

func (brcf *baseResolversContainerFactory) generateTxResolvers(
	topic string,
	unit dataRetriever.UnitType,
) error {

	shardC := brcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards+1)
	resolverSlice := make([]dataRetriever.Resolver, noOfShards+1)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := topic + shardC.CommunicationIdentifier(idx)
		resolver, err := brcf.createTxResolver(identifierTx, unit)
		if err != nil {
			return err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierTx
	}

	identifierTx := topic + shardC.CommunicationIdentifier(core.MetachainShardId)
	resolver, err := brcf.createTxResolver(identifierTx, unit)
	if err != nil {
		return err
	}

	resolverSlice[noOfShards] = resolver
	keys[noOfShards] = identifierTx

	return brcf.container.AddMultiple(keys, resolverSlice)
}

func (brcf *baseResolversContainerFactory) createTxResolver(
	responseTopic string,
	unit dataRetriever.UnitType,
) (dataRetriever.Resolver, error) {

	txStorer, err := brcf.store.GetStorer(unit)
	if err != nil {
		return nil, err
	}

	arg := storageResolvers.ArgSliceResolver{
		Messenger:                brcf.messenger,
		ResponseTopicName:        responseTopic,
		Storage:                  txStorer,
		DataPacker:               brcf.dataPacker,
		Marshalizer:              brcf.marshalizer,
		ManualEpochStartNotifier: brcf.manualEpochStartNotifier,
		ChanGracefullyClose:      brcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	resolver, err := storageResolvers.NewSliceResolver(arg)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

func (brcf *baseResolversContainerFactory) generateMiniBlocksResolvers() error {
	shardC := brcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards+2)
	resolverSlice := make([]dataRetriever.Resolver, noOfShards+2)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(idx)
		resolver, err := brcf.createMiniBlocksResolver(identifierMiniBlocks)
		if err != nil {
			return err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierMiniBlocks
	}

	identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	resolver, err := brcf.createMiniBlocksResolver(identifierMiniBlocks)
	if err != nil {
		return err
	}

	resolverSlice[noOfShards] = resolver
	keys[noOfShards] = identifierMiniBlocks

	identifierAllShardMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.AllShardId)
	allShardMiniblocksResolver, err := brcf.createMiniBlocksResolver(identifierAllShardMiniBlocks)
	if err != nil {
		return err
	}

	resolverSlice[noOfShards+1] = allShardMiniblocksResolver
	keys[noOfShards+1] = identifierAllShardMiniBlocks

	return brcf.container.AddMultiple(keys, resolverSlice)
}

func (brcf *baseResolversContainerFactory) createMiniBlocksResolver(responseTopic string) (dataRetriever.Resolver, error) {
	miniBlocksStorer, err := brcf.store.GetStorer(dataRetriever.MiniBlockUnit)
	if err != nil {
		return nil, err
	}

	arg := storageResolvers.ArgSliceResolver{
		Messenger:                brcf.messenger,
		ResponseTopicName:        responseTopic,
		Storage:                  miniBlocksStorer,
		DataPacker:               brcf.dataPacker,
		Marshalizer:              brcf.marshalizer,
		ManualEpochStartNotifier: brcf.manualEpochStartNotifier,
		ChanGracefullyClose:      brcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	mbResolver, err := storageResolvers.NewSliceResolver(arg)
	if err != nil {
		return nil, err
	}

	return mbResolver, nil
}

func (brcf *baseResolversContainerFactory) newImportDBTrieStorage(
	mainStorer storage.Storer,
	checkpointsStorer storage.Storer,
) (common.StorageManager, dataRetriever.TrieDataGetter, error) {
	pathManager, err := storageFactory.CreatePathManager(
		storageFactory.ArgCreatePathManager{
			WorkingDir: brcf.workingDir,
			ChainID:    brcf.chainID,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	trieFactoryArgs := trieFactory.TrieFactoryArgs{
		Marshalizer:              brcf.marshalizer,
		Hasher:                   brcf.hasher,
		PathManager:              pathManager,
		TrieStorageManagerConfig: brcf.generalConfig.TrieStorageManagerConfig,
	}
	trieFactoryInstance, err := trieFactory.NewTrieFactory(trieFactoryArgs)
	if err != nil {
		return nil, nil, err
	}

	args := trieFactory.TrieCreateArgs{
		MainStorer:         mainStorer,
		CheckpointsStorer:  checkpointsStorer,
		PruningEnabled:     brcf.generalConfig.StateTriesConfig.AccountsStatePruningEnabled,
		CheckpointsEnabled: brcf.generalConfig.StateTriesConfig.CheckpointsEnabled,
		MaxTrieLevelInMem:  brcf.generalConfig.StateTriesConfig.MaxStateTrieLevelInMemory,
		SnapshotsEnabled:   brcf.generalConfig.StateTriesConfig.SnapshotsEnabled,
		IdleProvider:       disabled.NewProcessStatusHandler(),
	}
	return trieFactoryInstance.Create(args)
}

func (brcf *baseResolversContainerFactory) generatePeerAuthenticationResolver() error {
	identifierPeerAuth := common.PeerAuthenticationTopic
	peerAuthResolver := disabledResolvers.NewDisabledPeerAuthenticatorResolver()

	return brcf.container.Add(identifierPeerAuth, peerAuthResolver)
}

func (brcf *baseResolversContainerFactory) generateValidatorInfoResolver() error {
	validatorInfoStorer, err := brcf.store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	if err != nil {
		return err
	}

	identifierValidatorInfo := common.ValidatorInfoTopic
	arg := storageResolvers.ArgSliceResolver{
		Messenger:                brcf.messenger,
		ResponseTopicName:        identifierValidatorInfo,
		Storage:                  validatorInfoStorer,
		DataPacker:               brcf.dataPacker,
		Marshalizer:              brcf.marshalizer,
		ManualEpochStartNotifier: brcf.manualEpochStartNotifier,
		ChanGracefullyClose:      brcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	validatorInfoResolver, err := storageResolvers.NewSliceResolver(arg)
	if err != nil {
		return err
	}

	return brcf.container.Add(identifierValidatorInfo, validatorInfoResolver)
}
