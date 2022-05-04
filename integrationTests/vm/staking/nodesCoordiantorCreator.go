package staking

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/factory"
	integrationMocks "github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/testscommon/stakingcommon"
)

const (
	shuffleBetweenShards = false
	adaptivity           = false
	hysteresis           = float32(0.2)
	initialRating        = 5
)

func createNodesCoordinator(
	eligibleMap map[uint32][]nodesCoordinator.Validator,
	waitingMap map[uint32][]nodesCoordinator.Validator,
	numOfMetaNodes uint32,
	numOfShards uint32,
	numOfEligibleNodesPerShard uint32,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	coreComponents factory.CoreComponentsHolder,
	bootStorer storage.Storer,
	nodesCoordinatorRegistryFactory nodesCoordinator.NodesCoordinatorRegistryFactory,
	maxNodesConfig []config.MaxNodesChangeConfig,
) nodesCoordinator.NodesCoordinator {
	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           numOfEligibleNodesPerShard,
		NodesMeta:            numOfMetaNodes,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: maxNodesConfig,
		EnableEpochs: config.EnableEpochs{
			StakingV4EnableEpoch:                     stakingV4EnableEpoch,
			StakingV4DistributeAuctionToWaitingEpoch: stakingV4DistributeAuctionToWaitingEpoch,
		},
	}
	nodeShuffler, _ := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)

	cache, _ := lrucache.NewCache(10000)
	argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
		ShardConsensusGroupSize:         shardConsensusGroupSize,
		MetaConsensusGroupSize:          metaConsensusGroupSize,
		Marshalizer:                     coreComponents.InternalMarshalizer(),
		Hasher:                          coreComponents.Hasher(),
		ShardIDAsObserver:               core.MetachainShardId,
		NbShards:                        numOfShards,
		EligibleNodes:                   eligibleMap,
		WaitingNodes:                    waitingMap,
		SelfPublicKey:                   eligibleMap[core.MetachainShardId][0].PubKey(),
		ConsensusGroupCache:             cache,
		ShuffledOutHandler:              &integrationMocks.ShuffledOutHandlerStub{},
		ChanStopNode:                    coreComponents.ChanStopNodeProcess(),
		IsFullArchive:                   false,
		Shuffler:                        nodeShuffler,
		BootStorer:                      bootStorer,
		EpochStartNotifier:              coreComponents.EpochStartNotifierWithConfirm(),
		StakingV4EnableEpoch:            stakingV4EnableEpoch,
		NodesCoordinatorRegistryFactory: nodesCoordinatorRegistryFactory,
		NodeTypeProvider:                coreComponents.NodeTypeProvider(),
	}

	baseNodesCoordinator, _ := nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
	nodesCoord, _ := nodesCoordinator.NewIndexHashedNodesCoordinatorWithRater(baseNodesCoordinator, coreComponents.Rater())

	return nodesCoord
}

func createGenesisNodes(
	numOfMetaNodes uint32,
	numOfShards uint32,
	numOfNodesPerShard uint32,
	numOfWaitingNodesPerShard uint32,
	marshaller marshal.Marshalizer,
	stateComponents factory.StateComponentsHandler,
) (map[uint32][]nodesCoordinator.Validator, map[uint32][]nodesCoordinator.Validator) {
	addressStartIdx := uint32(0)
	eligibleGenesisNodes := generateGenesisNodeInfoMap(numOfMetaNodes, numOfShards, numOfNodesPerShard, addressStartIdx)
	eligibleValidators, _ := nodesCoordinator.NodesInfoToValidators(eligibleGenesisNodes)

	addressStartIdx = numOfMetaNodes + numOfShards*numOfNodesPerShard
	waitingGenesisNodes := generateGenesisNodeInfoMap(numOfWaitingNodesPerShard, numOfShards, numOfWaitingNodesPerShard, addressStartIdx)
	waitingValidators, _ := nodesCoordinator.NodesInfoToValidators(waitingGenesisNodes)

	registerValidators(eligibleValidators, stateComponents, marshaller, common.EligibleList)
	registerValidators(waitingValidators, stateComponents, marshaller, common.WaitingList)

	return eligibleValidators, waitingValidators
}

func createGenesisNodesWithCustomConfig(
	owners map[string]*OwnerStats,
	marshaller marshal.Marshalizer,
	stateComponents factory.StateComponentsHandler,
) (map[uint32][]nodesCoordinator.Validator, map[uint32][]nodesCoordinator.Validator) {
	eligibleGenesis := make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
	waitingGenesis := make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)

	for owner, ownerStats := range owners {
		for shardID, ownerEligibleKeys := range ownerStats.EligibleBlsKeys {
			for _, eligibleKey := range ownerEligibleKeys {
				validator := integrationMocks.NewNodeInfo(eligibleKey, eligibleKey, shardID, initialRating)
				eligibleGenesis[shardID] = append(eligibleGenesis[shardID], validator)

				pubKey := validator.PubKeyBytes()

				peerAccount, _ := state.NewPeerAccount(pubKey)
				peerAccount.SetTempRating(initialRating)
				peerAccount.ShardId = shardID
				peerAccount.BLSPublicKey = pubKey
				peerAccount.List = string(common.EligibleList)
				_ = stateComponents.PeerAccounts().SaveAccount(peerAccount)

				stakingcommon.RegisterValidatorKeys(
					stateComponents.AccountsAdapter(),
					[]byte(owner),
					[]byte(owner),
					[][]byte{pubKey},
					ownerStats.TotalStake,
					marshaller,
				)

			}
		}

		for shardID, ownerWaitingKeys := range ownerStats.WaitingBlsKeys {
			for _, waitingKey := range ownerWaitingKeys {
				validator := integrationMocks.NewNodeInfo(waitingKey, waitingKey, shardID, initialRating)
				waitingGenesis[shardID] = append(waitingGenesis[shardID], validator)

				pubKey := validator.PubKeyBytes()

				peerAccount, _ := state.NewPeerAccount(pubKey)
				peerAccount.SetTempRating(initialRating)
				peerAccount.ShardId = shardID
				peerAccount.BLSPublicKey = pubKey
				peerAccount.List = string(common.WaitingList)
				_ = stateComponents.PeerAccounts().SaveAccount(peerAccount)

				stakingcommon.RegisterValidatorKeys(
					stateComponents.AccountsAdapter(),
					[]byte(owner),
					[]byte(owner),
					[][]byte{pubKey},
					ownerStats.TotalStake,
					marshaller,
				)

			}
		}
	}

	eligible, _ := nodesCoordinator.NodesInfoToValidators(eligibleGenesis)
	waiting, _ := nodesCoordinator.NodesInfoToValidators(waitingGenesis)

	//registerValidators(eligible, stateComponents, marshaller, common.EligibleList)
	//registerValidators(waiting, stateComponents, marshaller, common.WaitingList)

	return eligible, waiting
}

func generateGenesisNodeInfoMap(
	numOfMetaNodes uint32,
	numOfShards uint32,
	numOfNodesPerShard uint32,
	addressStartIdx uint32,
) map[uint32][]nodesCoordinator.GenesisNodeInfoHandler {
	validatorsMap := make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
	id := addressStartIdx
	for shardId := uint32(0); shardId < numOfShards; shardId++ {
		for n := uint32(0); n < numOfNodesPerShard; n++ {
			addr := generateAddress(id)
			validator := integrationMocks.NewNodeInfo(addr, addr, shardId, initialRating)
			validatorsMap[shardId] = append(validatorsMap[shardId], validator)
			id++
		}
	}

	for n := uint32(0); n < numOfMetaNodes; n++ {
		addr := generateAddress(id)
		validator := integrationMocks.NewNodeInfo(addr, addr, core.MetachainShardId, initialRating)
		validatorsMap[core.MetachainShardId] = append(validatorsMap[core.MetachainShardId], validator)
		id++
	}

	return validatorsMap
}

func registerValidators(
	validators map[uint32][]nodesCoordinator.Validator,
	stateComponents factory.StateComponentsHolder,
	marshaller marshal.Marshalizer,
	list common.PeerType,
) {
	for shardID, validatorsInShard := range validators {
		for _, val := range validatorsInShard {
			pubKey := val.PubKey()

			peerAccount, _ := state.NewPeerAccount(pubKey)
			peerAccount.SetTempRating(initialRating)
			peerAccount.ShardId = shardID
			peerAccount.BLSPublicKey = pubKey
			peerAccount.List = string(list)
			_ = stateComponents.PeerAccounts().SaveAccount(peerAccount)

			stakingcommon.RegisterValidatorKeys(
				stateComponents.AccountsAdapter(),
				pubKey,
				pubKey,
				[][]byte{pubKey},
				big.NewInt(2*nodePrice),
				marshaller,
			)
		}
	}
}
