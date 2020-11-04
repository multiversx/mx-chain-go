package metachain

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.EpochStartRewardsCreator = (*rewardsCreator)(nil)

var zero = big.NewInt(0)

type ArgsNewRewardsCreator struct {
	BaseRewardsCreatorArgs
}

type rewardsCreator struct {
	*baseRewardsCreator
	mapRewardsPerBlockPerValidator map[uint32]*big.Int
}

type rewardInfoData struct {
	accumulatedFees *big.Int
	address         string
	protocolRewards *big.Int
}

// NewEpochStartRewardsCreator creates a new rewards creator object
func NewEpochStartRewardsCreator(args ArgsNewRewardsCreator) (*rewardsCreator, error) {
	brc, err := NewBaseRewardsCreator(args.BaseRewardsCreatorArgs)
	if err != nil {
		return nil, err
	}

	rc := &rewardsCreator{
		baseRewardsCreator: brc,
	}

	return rc, nil
}

// CreateRewardsMiniBlocks creates the rewards miniblocks according to economics data and validator info
func (rc *rewardsCreator) CreateRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error) {
	if check.IfNil(metaBlock) {
		return nil, epochStart.ErrNilHeaderHandler
	}

	rc.clean()
	rc.flagDelegationSystemSCEnabled.Toggle(metaBlock.GetEpoch() >= rc.delegationSystemSCEnableEpoch)

	miniBlocks := make(block.MiniBlockSlice, rc.shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i <= rc.shardCoordinator.NumberOfShards(); i++ {
		miniBlocks[i] = &block.MiniBlock{}
		miniBlocks[i].SenderShardID = core.MetachainShardId
		miniBlocks[i].ReceiverShardID = i
		miniBlocks[i].Type = block.RewardsBlock
		miniBlocks[i].TxHashes = make([][]byte, 0)
	}
	miniBlocks[rc.shardCoordinator.NumberOfShards()].ReceiverShardID = core.MetachainShardId

	protocolSustainabilityRwdTx, protocolSustainabilityShardId, err := rc.createProtocolSustainabilityRewardTransaction(metaBlock)
	if err != nil {
		return nil, err
	}

	rc.fillRewardsPerBlockPerNode(&metaBlock.EpochStart.Economics)
	err = rc.addValidatorRewardsToMiniBlocks(validatorsInfo, metaBlock, miniBlocks, protocolSustainabilityRwdTx)
	if err != nil {
		return nil, err
	}

	totalWithoutDevelopers := big.NewInt(0).Sub(metaBlock.EpochStart.Economics.TotalToDistribute, metaBlock.DevFeesInEpoch)
	difference := big.NewInt(0).Sub(totalWithoutDevelopers, rc.accumulatedRewards)
	log.Debug("arithmetic difference in end of epoch rewards economics", "value", difference)
	protocolSustainabilityRwdTx.Value.Add(protocolSustainabilityRwdTx.Value, difference)
	if protocolSustainabilityRwdTx.Value.Cmp(big.NewInt(0)) < 0 {
		log.Error("negative rewards protocol sustainability")
		protocolSustainabilityRwdTx.Value.SetUint64(0)
	}
	rc.protocolSustainability.Set(protocolSustainabilityRwdTx.Value)

	protocolSustainabilityRwdHash, errHash := core.CalculateHash(rc.marshalizer, rc.hasher, protocolSustainabilityRwdTx)
	if errHash != nil {
		return nil, errHash
	}

	rc.currTxs.AddTx(protocolSustainabilityRwdHash, protocolSustainabilityRwdTx)
	miniBlocks[protocolSustainabilityShardId].TxHashes = append(miniBlocks[protocolSustainabilityShardId].TxHashes, protocolSustainabilityRwdHash)

	for shId := uint32(0); shId <= rc.shardCoordinator.NumberOfShards(); shId++ {
		sort.Slice(miniBlocks[shId].TxHashes, func(i, j int) bool {
			return bytes.Compare(miniBlocks[shId].TxHashes[i], miniBlocks[shId].TxHashes[j]) < 0
		})
	}

	finalMiniBlocks := make(block.MiniBlockSlice, 0)
	for i := uint32(0); i <= rc.shardCoordinator.NumberOfShards(); i++ {
		if len(miniBlocks[i].TxHashes) > 0 {
			finalMiniBlocks = append(finalMiniBlocks, miniBlocks[i])
		}
	}

	return finalMiniBlocks, nil
}

func (rc *rewardsCreator) fillRewardsPerBlockPerNode(economicsData *block.Economics) {
	rc.mapRewardsPerBlockPerValidator = make(map[uint32]*big.Int)
	for i := uint32(0); i < rc.shardCoordinator.NumberOfShards(); i++ {
		consensusSize := big.NewInt(int64(rc.nodesConfigProvider.ConsensusGroupSize(i)))
		rc.mapRewardsPerBlockPerValidator[i] = big.NewInt(0).Div(economicsData.RewardsPerBlock, consensusSize)
		log.Debug("rewardsPerBlockPerValidator", "shardID", i, "value", rc.mapRewardsPerBlockPerValidator[i].String())
	}

	consensusSize := big.NewInt(int64(rc.nodesConfigProvider.ConsensusGroupSize(core.MetachainShardId)))
	rc.mapRewardsPerBlockPerValidator[core.MetachainShardId] = big.NewInt(0).Div(economicsData.RewardsPerBlock, consensusSize)
	log.Debug("rewardsPerBlockPerValidator", "shardID", core.MetachainShardId, "value", rc.mapRewardsPerBlockPerValidator[core.MetachainShardId].String())
}

func (rc *rewardsCreator) addValidatorRewardsToMiniBlocks(
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	metaBlock *block.MetaBlock,
	miniBlocks block.MiniBlockSlice,
	protocolSustainabilityRwdTx *rewardTx.RewardTx,
) error {
	rwdAddrValidatorInfo := rc.computeValidatorInfoPerRewardAddress(validatorsInfo, protocolSustainabilityRwdTx)
	for _, rwdInfo := range rwdAddrValidatorInfo {
		rwdTx, rwdTxHash, err := rc.createRewardFromRwdInfo(rwdInfo, metaBlock)
		if err != nil {
			return err
		}
		if rwdTx.Value.Cmp(zero) <= 0 {
			continue
		}

		rc.accumulatedRewards.Add(rc.accumulatedRewards, rwdTx.Value)
		mbId := rc.shardCoordinator.ComputeId([]byte(rwdInfo.address))
		if mbId == core.MetachainShardId {
			mbId = rc.shardCoordinator.NumberOfShards()

			if !rc.flagDelegationSystemSCEnabled.IsSet() || !rc.isSystemDelegationSC(rwdTx.RcvAddr) {
				protocolSustainabilityRwdTx.Value.Add(protocolSustainabilityRwdTx.Value, rwdTx.GetValue())
				continue
			}
		}

		if rwdTx.Value.Cmp(big.NewInt(0)) < 0 {
			log.Error("negative rewards", "rcv", rwdTx.RcvAddr)
			continue
		}
		rc.currTxs.AddTx(rwdTxHash, rwdTx)
		miniBlocks[mbId].TxHashes = append(miniBlocks[mbId].TxHashes, rwdTxHash)
	}

	return nil
}

func (rc *rewardsCreator) computeValidatorInfoPerRewardAddress(
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	protocolSustainabilityRwd *rewardTx.RewardTx,
) map[string]*rewardInfoData {

	rwdAddrValidatorInfo := make(map[string]*rewardInfoData)

	for _, shardValidatorsInfo := range validatorsInfo {
		for _, validatorInfo := range shardValidatorsInfo {
			rewardsPerBlockPerNodeForShard := rc.mapRewardsPerBlockPerValidator[validatorInfo.ShardId]
			protocolRewardValue := big.NewInt(0).Mul(rewardsPerBlockPerNodeForShard, big.NewInt(0).SetUint64(uint64(validatorInfo.NumSelectedInSuccessBlocks)))

			if validatorInfo.LeaderSuccess == 0 && validatorInfo.ValidatorFailure == 0 {
				protocolSustainabilityRwd.Value.Add(protocolSustainabilityRwd.Value, protocolRewardValue)
				continue
			}

			rwdInfo, ok := rwdAddrValidatorInfo[string(validatorInfo.RewardAddress)]
			if !ok {
				rwdInfo = &rewardInfoData{
					accumulatedFees: big.NewInt(0),
					protocolRewards: big.NewInt(0),
					address:         string(validatorInfo.RewardAddress),
				}
				rwdAddrValidatorInfo[string(validatorInfo.RewardAddress)] = rwdInfo
			}

			rwdInfo.accumulatedFees.Add(rwdInfo.accumulatedFees, validatorInfo.AccumulatedFees)
			rwdInfo.protocolRewards.Add(rwdInfo.protocolRewards, protocolRewardValue)
		}
	}

	return rwdAddrValidatorInfo
}

// VerifyRewardsMiniBlocks verifies if received rewards miniblocks are correct
func (rc *rewardsCreator) VerifyRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) error {
	if check.IfNil(metaBlock) {
		return epochStart.ErrNilHeaderHandler
	}

	createdMiniBlocks, err := rc.CreateRewardsMiniBlocks(metaBlock, validatorsInfo)
	if err != nil {
		return err
	}

	return rc.verifyCreatedRewardMiniblocksWithMetaBlock(metaBlock, createdMiniBlocks)
}
