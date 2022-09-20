package factory

import (
	"errors"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/closing"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/cache"
)

// CreateShardCoordinator is the shard coordinator factory
func CreateShardCoordinator(
	nodesConfig sharding.GenesisNodesSetupHandler,
	pubKey crypto.PublicKey,
	prefsConfig config.PreferencesConfig,
	log logger.Logger,
) (sharding.Coordinator, core.NodeType, error) {

	selfShardId, err := getShardIdFromNodePubKey(pubKey, nodesConfig)
	nodeType := core.NodeTypeValidator
	if err == sharding.ErrPublicKeyNotFoundInGenesis {
		nodeType = core.NodeTypeObserver
		log.Info("starting as observer node")

		selfShardId, err = common.ProcessDestinationShardAsObserver(prefsConfig.DestinationShardAsObserver)
		if err != nil {
			return nil, "", err
		}
		var pubKeyBytes []byte
		if selfShardId == common.DisabledShardIDAsObserver {
			pubKeyBytes, err = pubKey.ToByteArray()
			if err != nil {
				return nil, core.NodeTypeObserver, fmt.Errorf("%w while assigning random shard ID for observer", err)
			}

			selfShardId = common.AssignShardForPubKeyWhenNotSpecified(pubKeyBytes, nodesConfig.NumberOfShards())
		}
	}
	if err != nil {
		return nil, "", err
	}

	var shardName string
	if selfShardId == core.MetachainShardId {
		shardName = common.MetachainShardName
	} else {
		shardName = fmt.Sprintf("%d", selfShardId)
	}
	log.Info("shard info", "started in shard", shardName)

	shardCoordinator, err := sharding.NewMultiShardCoordinator(nodesConfig.NumberOfShards(), selfShardId)
	if err != nil {
		return nil, "", err
	}

	return shardCoordinator, nodeType, nil
}

func getShardIdFromNodePubKey(pubKey crypto.PublicKey, nodesConfig sharding.GenesisNodesSetupHandler) (uint32, error) {
	if pubKey == nil {
		return 0, errors.New("nil public key")
	}

	publicKey, err := pubKey.ToByteArray()
	if err != nil {
		return 0, err
	}

	selfShardId, err := nodesConfig.GetShardIDForPubKey(publicKey)
	if err != nil {
		return 0, err
	}

	return selfShardId, err
}

// CreateNodesCoordinator is the nodes coordinator factory
func CreateNodesCoordinator(
	nodeShufflerOut ShuffleOutCloser,
	nodesConfig sharding.GenesisNodesSetupHandler,
	prefsConfig config.PreferencesConfig,
	epochStartNotifier epochStart.RegistrationHandler,
	pubKey crypto.PublicKey,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	ratingAndListIndexHandler nodesCoordinator.ChanceComputer,
	bootStorer storage.Storer,
	nodeShuffler nodesCoordinator.NodesShuffler,
	currentShardID uint32,
	bootstrapParameters BootstrapParamsHolder,
	startEpoch uint32,
	chanNodeStop chan endProcess.ArgEndProcess,
	nodeTypeProvider core.NodeTypeProviderHandler,
	enableEpochsHandler common.EnableEpochsHandler,
	validatorInfoCacher epochStart.ValidatorInfoCacher,
) (nodesCoordinator.NodesCoordinator, error) {
	if chanNodeStop == nil {
		return nil, nodesCoordinator.ErrNilNodeStopChannel
	}
	shardIDAsObserver, err := common.ProcessDestinationShardAsObserver(prefsConfig.DestinationShardAsObserver)
	if err != nil {
		return nil, err
	}
	var pubKeyBytes []byte
	if shardIDAsObserver == common.DisabledShardIDAsObserver {
		pubKeyBytes, err = pubKey.ToByteArray()
		if err != nil {
			return nil, fmt.Errorf("%w while assigning random shard ID for observer", err)
		}

		shardIDAsObserver = common.AssignShardForPubKeyWhenNotSpecified(pubKeyBytes, nodesConfig.NumberOfShards())
	}

	nbShards := nodesConfig.NumberOfShards()
	shardConsensusGroupSize := int(nodesConfig.GetShardConsensusGroupSize())
	metaConsensusGroupSize := int(nodesConfig.GetMetaConsensusGroupSize())
	eligibleNodesInfo, waitingNodesInfo := nodesConfig.InitialNodesInfo()

	eligibleValidators, errEligibleValidators := nodesCoordinator.NodesInfoToValidators(eligibleNodesInfo)
	if errEligibleValidators != nil {
		return nil, errEligibleValidators
	}

	waitingValidators, errWaitingValidators := nodesCoordinator.NodesInfoToValidators(waitingNodesInfo)
	if errWaitingValidators != nil {
		return nil, errWaitingValidators
	}

	currentEpoch := startEpoch
	if bootstrapParameters.NodesConfig() != nil {
		nodeRegistry := bootstrapParameters.NodesConfig()
		currentEpoch = bootstrapParameters.Epoch()
		epochsConfig, ok := nodeRegistry.EpochsConfig[fmt.Sprintf("%d", currentEpoch)]
		if ok {
			eligibles := epochsConfig.EligibleValidators
			eligibleValidators, err = nodesCoordinator.SerializableValidatorsToValidators(eligibles)
			if err != nil {
				return nil, err
			}

			waitings := epochsConfig.WaitingValidators
			waitingValidators, err = nodesCoordinator.SerializableValidatorsToValidators(waitings)
			if err != nil {
				return nil, err
			}
		}
	}

	pubKeyBytes, err = pubKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	consensusGroupCache, err := cache.NewLRUCache(25000)
	if err != nil {
		return nil, err
	}

	shuffledOutHandler, err := sharding.NewShuffledOutTrigger(pubKeyBytes, currentShardID, nodeShufflerOut.EndOfProcessingHandler)
	if err != nil {
		return nil, err
	}

	argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
		ShardConsensusGroupSize: shardConsensusGroupSize,
		MetaConsensusGroupSize:  metaConsensusGroupSize,
		Marshalizer:             marshalizer,
		Hasher:                  hasher,
		Shuffler:                nodeShuffler,
		EpochStartNotifier:      epochStartNotifier,
		BootStorer:              bootStorer,
		ShardIDAsObserver:       shardIDAsObserver,
		NbShards:                nbShards,
		EligibleNodes:           eligibleValidators,
		WaitingNodes:            waitingValidators,
		SelfPublicKey:           pubKeyBytes,
		ConsensusGroupCache:     consensusGroupCache,
		ShuffledOutHandler:      shuffledOutHandler,
		Epoch:                   currentEpoch,
		StartEpoch:              startEpoch,
		ChanStopNode:            chanNodeStop,
		NodeTypeProvider:        nodeTypeProvider,
		IsFullArchive:           prefsConfig.FullArchive,
		EnableEpochsHandler:     enableEpochsHandler,
		ValidatorInfoCacher:     validatorInfoCacher,
	}

	baseNodesCoordinator, err := nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
	if err != nil {
		return nil, err
	}

	nodesCoord, err := nodesCoordinator.NewIndexHashedNodesCoordinatorWithRater(baseNodesCoordinator, ratingAndListIndexHandler)
	if err != nil {
		return nil, err
	}

	return nodesCoord, nil
}

// CreateNodesShuffleOut is the nodes shuffler closer factory
func CreateNodesShuffleOut(
	nodesConfig sharding.GenesisNodesSetupHandler,
	epochConfig config.EpochStartConfig,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
) (ShuffleOutCloser, error) {

	maxThresholdEpochDuration := epochConfig.MaxShuffledOutRestartThreshold
	if !(maxThresholdEpochDuration >= 0.0 && maxThresholdEpochDuration <= 1.0) {
		return nil, fmt.Errorf("invalid max threshold for shuffled out handler")
	}
	minThresholdEpochDuration := epochConfig.MinShuffledOutRestartThreshold
	if !(minThresholdEpochDuration >= 0.0 && minThresholdEpochDuration <= 1.0) {
		return nil, fmt.Errorf("invalid min threshold for shuffled out handler")
	}

	epochDuration := int64(nodesConfig.GetRoundDuration()) * epochConfig.RoundsPerEpoch
	minDurationBeforeStopProcess := int64(minThresholdEpochDuration * float64(epochDuration))
	maxDurationBeforeStopProcess := int64(maxThresholdEpochDuration * float64(epochDuration))

	minDurationInterval := time.Millisecond * time.Duration(minDurationBeforeStopProcess)
	maxDurationInterval := time.Millisecond * time.Duration(maxDurationBeforeStopProcess)

	log.Debug("closing.NewShuffleOutCloser",
		"minDurationInterval", minDurationInterval,
		"maxDurationInterval", maxDurationInterval,
	)

	nodeShufflerOut, err := closing.NewShuffleOutCloser(
		minDurationInterval,
		maxDurationInterval,
		chanStopNodeProcess,
		log,
	)
	if err != nil {
		return nil, err
	}

	return nodeShufflerOut, nil
}
