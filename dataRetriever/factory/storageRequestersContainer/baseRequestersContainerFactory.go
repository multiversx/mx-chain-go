package storagerequesterscontainer

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/disabled"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	disabledRequesters "github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers/requesters/disabled"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/storageRequesters"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	trieFactory "github.com/ElrondNetwork/elrond-go/trie/factory"
)

const defaultBeforeGracefulClose = time.Minute

type baseRequestersContainerFactory struct {
	container                dataRetriever.RequestersContainer
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

func (brcf *baseRequestersContainerFactory) checkParams() error {
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

func (brcf *baseRequestersContainerFactory) generateCommonRequesters() error {
	err := brcf.generateTxRequesters(
		factory.TransactionTopic,
		dataRetriever.TransactionUnit,
	)
	if err != nil {
		return err
	}

	err = brcf.generateTxRequesters(
		factory.UnsignedTransactionTopic,
		dataRetriever.UnsignedTransactionUnit,
	)
	if err != nil {
		return err
	}

	err = brcf.generateMiniBlocksRequesters()
	if err != nil {
		return err
	}

	err = brcf.generatePeerAuthenticationRequester()
	if err != nil {
		return err
	}

	return brcf.generateValidatorInfoRequester()
}

func (brcf *baseRequestersContainerFactory) generateTxRequesters(
	topic string,
	unit dataRetriever.UnitType,
) error {

	shardC := brcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards+1)
	requestersSlice := make([]dataRetriever.Requester, noOfShards+1)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := topic + shardC.CommunicationIdentifier(idx)
		requester, err := brcf.createTxRequester(identifierTx, unit)
		if err != nil {
			return err
		}

		requestersSlice[idx] = requester
		keys[idx] = identifierTx
	}

	identifierTx := topic + shardC.CommunicationIdentifier(core.MetachainShardId)
	requester, err := brcf.createTxRequester(identifierTx, unit)
	if err != nil {
		return err
	}

	requestersSlice[noOfShards] = requester
	keys[noOfShards] = identifierTx

	return brcf.container.AddMultiple(keys, requestersSlice)
}

func (brcf *baseRequestersContainerFactory) createTxRequester(
	responseTopic string,
	unit dataRetriever.UnitType,
) (dataRetriever.Requester, error) {

	txStorer, err := brcf.store.GetStorer(unit)
	if err != nil {
		return nil, err
	}

	arg := storagerequesters.ArgSliceRequester{
		Messenger:                brcf.messenger,
		ResponseTopicName:        responseTopic,
		Storage:                  txStorer,
		DataPacker:               brcf.dataPacker,
		Marshalizer:              brcf.marshalizer,
		ManualEpochStartNotifier: brcf.manualEpochStartNotifier,
		ChanGracefullyClose:      brcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	requester, err := storagerequesters.NewSliceRequester(arg)
	if err != nil {
		return nil, err
	}

	return requester, nil
}

func (brcf *baseRequestersContainerFactory) generateMiniBlocksRequesters() error {
	shardC := brcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards+2)
	requestersSlice := make([]dataRetriever.Requester, noOfShards+2)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(idx)
		requester, err := brcf.createMiniBlocksRequester(identifierMiniBlocks)
		if err != nil {
			return err
		}

		requestersSlice[idx] = requester
		keys[idx] = identifierMiniBlocks
	}

	identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	requester, err := brcf.createMiniBlocksRequester(identifierMiniBlocks)
	if err != nil {
		return err
	}

	requestersSlice[noOfShards] = requester
	keys[noOfShards] = identifierMiniBlocks

	identifierAllShardMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.AllShardId)
	allShardMiniblocksRequester, err := brcf.createMiniBlocksRequester(identifierAllShardMiniBlocks)
	if err != nil {
		return err
	}

	requestersSlice[noOfShards+1] = allShardMiniblocksRequester
	keys[noOfShards+1] = identifierAllShardMiniBlocks

	return brcf.container.AddMultiple(keys, requestersSlice)
}

func (brcf *baseRequestersContainerFactory) createMiniBlocksRequester(responseTopic string) (dataRetriever.Requester, error) {
	miniBlocksStorer, err := brcf.store.GetStorer(dataRetriever.MiniBlockUnit)
	if err != nil {
		return nil, err
	}

	arg := storagerequesters.ArgSliceRequester{
		Messenger:                brcf.messenger,
		ResponseTopicName:        responseTopic,
		Storage:                  miniBlocksStorer,
		DataPacker:               brcf.dataPacker,
		Marshalizer:              brcf.marshalizer,
		ManualEpochStartNotifier: brcf.manualEpochStartNotifier,
		ChanGracefullyClose:      brcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	mbRequester, err := storagerequesters.NewSliceRequester(arg)
	if err != nil {
		return nil, err
	}

	return mbRequester, nil
}

func (brcf *baseRequestersContainerFactory) newImportDBTrieStorage(
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

func (brcf *baseRequestersContainerFactory) generatePeerAuthenticationRequester() error {
	identifierPeerAuth := common.PeerAuthenticationTopic
	peerAuthRequester := disabledRequesters.NewDisabledRequester()

	return brcf.container.Add(identifierPeerAuth, peerAuthRequester)
}

func (brcf *baseRequestersContainerFactory) generateValidatorInfoRequester() error {
	validatorInfoStorer, err := brcf.store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	if err != nil {
		return err
	}

	identifierValidatorInfo := common.ValidatorInfoTopic
	arg := storagerequesters.ArgSliceRequester{
		Messenger:                brcf.messenger,
		ResponseTopicName:        identifierValidatorInfo,
		Storage:                  validatorInfoStorer,
		DataPacker:               brcf.dataPacker,
		Marshalizer:              brcf.marshalizer,
		ManualEpochStartNotifier: brcf.manualEpochStartNotifier,
		ChanGracefullyClose:      brcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	validatorInfoRequester, err := storagerequesters.NewSliceRequester(arg)
	if err != nil {
		return err
	}

	return brcf.container.Add(identifierValidatorInfo, validatorInfoRequester)
}
