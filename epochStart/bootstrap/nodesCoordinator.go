package bootstrap

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type nodesCoordinator struct {
	shuffler                sharding.NodesShuffler
	chance                  sharding.ChanceComputer
	numShards               map[uint32]uint32
	shardConsensusGroupSize uint32
	metaConsensusGroupSize  uint32

	nodesConfig map[uint32]*epochNodesConfig
}

type epochNodesConfig struct {
	nbShards            uint32
	shardID             uint32
	eligibleMap         map[uint32][]sharding.Validator
	waitingMap          map[uint32][]sharding.Validator
	expandedEligibleMap map[uint32][]sharding.Validator
	leavingList         []sharding.Validator
}

// ArgsNewStartInEpochNodesCoordinator -
type ArgsNewStartInEpochNodesCoordinator struct {
	Shuffler                sharding.NodesShuffler
	Chance                  sharding.ChanceComputer
	ShardConsensusGroupSize uint32
	MetaConsensusGroupSize  uint32
}

// NewStartInEpochNodesCoordinator creates an epoch start nodes coordinator
func NewStartInEpochNodesCoordinator(args ArgsNewStartInEpochNodesCoordinator) (*nodesCoordinator, error) {
	n := &nodesCoordinator{
		shuffler:                args.Shuffler,
		chance:                  args.Chance,
		shardConsensusGroupSize: args.ShardConsensusGroupSize,
		metaConsensusGroupSize:  args.MetaConsensusGroupSize,
		nodesConfig:             make(map[uint32]*epochNodesConfig),
		numShards:               make(map[uint32]uint32),
	}

	return n, nil
}

// ComputeNodesConfigForGenesis creates the actual node config for genesis
func (n *nodesCoordinator) ComputeNodesConfigForGenesis(nodesConfig *sharding.NodesSetup) (*sharding.EpochValidators, error) {
	eligibleNodesInfo, waitingNodesInfo := nodesConfig.InitialNodesInfo()

	eligibleValidators, err := sharding.NodesInfoToValidators(eligibleNodesInfo)
	if err != nil {
		return nil, err
	}

	waitingValidators, err := sharding.NodesInfoToValidators(waitingNodesInfo)
	if err != nil {
		return nil, err
	}

	err = n.setNodesPerShards(eligibleValidators, waitingValidators, nil, 0)
	epochValidators := epochNodesConfigToEpochValidators(n.nodesConfig[0])

	return epochValidators, nil
}

// ComputeNodesConfigFor computes the actual nodes config for the set epoch from the validator info
func (n *nodesCoordinator) ComputeNodesConfigFor(
	metaBlock *block.MetaBlock,
	validatorInfos []*state.ValidatorInfo,
) (*sharding.EpochValidators, error) {
	if check.IfNil(metaBlock) {
		return nil, epochStart.ErrNilHeaderHandler
	}
	if len(validatorInfos) == 0 {
		return nil, epochStart.ErrNilValidatorInfo
	}

	randomness := metaBlock.GetPrevRandSeed()
	newEpoch := metaBlock.GetEpoch()
	n.numShards[newEpoch] = uint32(len(metaBlock.EpochStart.LastFinalizedHeaders))

	sort.Slice(validatorInfos, func(i, j int) bool {
		return bytes.Compare(validatorInfos[i].PublicKey, validatorInfos[j].PublicKey) < 0
	})

	leaving, err := n.computeLeaving(validatorInfos)
	if err != nil {
		return nil, err
	}

	eligibleMap := make(map[uint32][]sharding.Validator)
	waitingMap := make(map[uint32][]sharding.Validator)
	newNodesMap := make([]sharding.Validator, 0)
	for i := uint32(0); i < n.numShards[newEpoch]; i++ {
		eligibleMap[i] = make([]sharding.Validator, 0)
		waitingMap[i] = make([]sharding.Validator, 0)
	}
	eligibleMap[core.MetachainShardId] = make([]sharding.Validator, 0)
	waitingMap[core.MetachainShardId] = make([]sharding.Validator, 0)

	mapValidatorInfo := make(map[string]*state.ValidatorInfo, len(validatorInfos))
	for _, validatorInfo := range validatorInfos {
		validator, err := sharding.NewValidator(validatorInfo.PublicKey, validatorInfo.RewardAddress)
		if err != nil {
			return nil, err
		}
		mapValidatorInfo[string(validatorInfo.PublicKey)] = validatorInfo

		switch validatorInfo.List {
		case string(core.WaitingList):
			waitingMap[validatorInfo.ShardId] = append(waitingMap[validatorInfo.ShardId], validator)
		case string(core.EligibleList):
			eligibleMap[validatorInfo.ShardId] = append(eligibleMap[validatorInfo.ShardId], validator)
		case string(core.NewList):
			newNodesMap = append(newNodesMap, validator)
		}
	}

	shufflerArgs := sharding.ArgsUpdateNodes{
		Eligible: eligibleMap,
		Waiting:  waitingMap,
		NewNodes: newNodesMap,
		Leaving:  leaving,
		Rand:     randomness,
		NbShards: n.numShards[newEpoch],
	}

	newEligibleMap, newWaitingMap, stillRemaining := n.shuffler.UpdateNodeLists(shufflerArgs)
	actualLeaving := sharding.ComputeActuallyLeaving(leaving, stillRemaining)
	err = n.setNodesPerShards(newEligibleMap, newWaitingMap, actualLeaving, newEpoch)
	if err != nil {
		log.Error("set nodes per shard failed", "error", err)
		return nil, err
	}

	err = n.expandSavedNodes(mapValidatorInfo, newEpoch)
	if err != nil {
		return nil, err
	}

	epochValidators := epochNodesConfigToEpochValidators(n.nodesConfig[newEpoch])

	return epochValidators, nil
}

func (n *nodesCoordinator) computeLeaving(allValidators []*state.ValidatorInfo) ([]sharding.Validator, error) {
	leavingValidators := make([]sharding.Validator, 0)
	minChances := n.chance.GetChance(0)
	for _, validator := range allValidators {

		chances := n.chance.GetChance(validator.TempRating)
		if chances < minChances {
			val, err := sharding.NewValidator(validator.PublicKey, validator.RewardAddress)
			if err != nil {
				return nil, err
			}
			leavingValidators = append(leavingValidators, val)
		}
	}

	return leavingValidators, nil
}

func (n *nodesCoordinator) setNodesPerShards(
	eligible map[uint32][]sharding.Validator,
	waiting map[uint32][]sharding.Validator,
	leaving []sharding.Validator,
	epoch uint32,
) error {
	nodesConfig, ok := n.nodesConfig[epoch]
	if !ok {
		nodesConfig = &epochNodesConfig{}
	}

	nodesList, ok := eligible[core.MetachainShardId]
	if !ok || uint32(len(nodesList)) < n.metaConsensusGroupSize {
		return fmt.Errorf("%w computed size %d needed size %d", epochStart.ErrSmallShardEligibleListSize, uint32(len(nodesList)), n.metaConsensusGroupSize)
	}

	for shardId := uint32(0); shardId < uint32(len(eligible)-1); shardId++ {
		nbNodesShard := uint32(len(eligible[shardId]))
		if nbNodesShard < n.shardConsensusGroupSize {
			return fmt.Errorf("%w computed size %d needed size %d", epochStart.ErrSmallShardEligibleListSize, nbNodesShard, n.shardConsensusGroupSize)
		}
	}

	// nbShards holds number of shards without meta
	nodesConfig.nbShards = uint32(len(eligible) - 1)
	nodesConfig.eligibleMap = eligible
	nodesConfig.waitingMap = waiting

	nodesConfig.leavingList = make([]sharding.Validator, 0, len(leaving))
	for _, validator := range leaving {
		nodesConfig.leavingList = append(nodesConfig.leavingList, validator)
	}

	n.nodesConfig[epoch] = nodesConfig
	return nil
}

// ComputeShardForSelfPublicKey -
func (n *nodesCoordinator) ComputeShardForSelfPublicKey(epoch uint32, pubKey []byte) uint32 {
	for shard, validators := range n.nodesConfig[epoch].eligibleMap {
		for _, v := range validators {
			if bytes.Equal(v.PubKey(), pubKey) {
				return shard
			}
		}
	}

	for shard, validators := range n.nodesConfig[epoch].waitingMap {
		for _, v := range validators {
			if bytes.Equal(v.PubKey(), pubKey) {
				return shard
			}
		}
	}

	return core.AllShardId
}

func (n *nodesCoordinator) expandSavedNodes(
	mapValidatorInfo map[string]*state.ValidatorInfo,
	epoch uint32,
) error {
	nodesConfig := n.nodesConfig[epoch]
	nodesConfig.expandedEligibleMap = make(map[uint32][]sharding.Validator)

	nrShards := len(nodesConfig.eligibleMap)
	var err error
	nodesConfig.expandedEligibleMap[core.MetachainShardId], err = n.expandEligibleList(nodesConfig.eligibleMap[core.MetachainShardId], mapValidatorInfo)
	if err != nil {
		return err
	}

	for shardId := uint32(0); shardId < uint32(nrShards-1); shardId++ {
		nodesConfig.expandedEligibleMap[shardId], err = n.expandEligibleList(nodesConfig.eligibleMap[shardId], mapValidatorInfo)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *nodesCoordinator) expandEligibleList(
	validators []sharding.Validator,
	mapValidatorInfo map[string]*state.ValidatorInfo,
) ([]sharding.Validator, error) {
	minChance := n.chance.GetChance(0)
	minSize := len(validators) * int(minChance)
	validatorList := make([]sharding.Validator, 0, minSize)

	for _, validatorInShard := range validators {
		pk := validatorInShard.PubKey()
		validatorInfo, ok := mapValidatorInfo[string(pk)]
		if !ok {
			return nil, epochStart.ErrNilValidatorInfo
		}

		chances := n.chance.GetChance(validatorInfo.TempRating)
		if chances < minChance {
			chances = minChance
		}

		for i := uint32(0); i < chances; i++ {
			validatorList = append(validatorList, validatorInShard)
		}
	}

	return validatorList, nil
}

func epochNodesConfigToEpochValidators(config *epochNodesConfig) *sharding.EpochValidators {
	result := &sharding.EpochValidators{
		EligibleValidators: make(map[string][]*sharding.SerializableValidator, len(config.eligibleMap)),
		WaitingValidators:  make(map[string][]*sharding.SerializableValidator, len(config.waitingMap)),
		LeavingValidators:  make([]*sharding.SerializableValidator, 0),
	}

	for k, v := range config.eligibleMap {
		result.EligibleValidators[fmt.Sprint(k)] = sharding.ValidatorArrayToSerializableValidatorArray(v)
	}

	for k, v := range config.waitingMap {
		result.WaitingValidators[fmt.Sprint(k)] = sharding.ValidatorArrayToSerializableValidatorArray(v)
	}

	for _, v := range config.leavingList {
		result.LeavingValidators = append(result.LeavingValidators, &sharding.SerializableValidator{
			PubKey:  v.PubKey(),
			Address: v.Address(),
		})
	}

	return result
}

// IsInterfaceNil returns true if underlying object is nil
func (n *nodesCoordinator) IsInterfaceNil() bool {
	return n == nil
}
