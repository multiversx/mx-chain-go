package metachain

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/state"
)

var _ process.RewardsCreator = (*rewardsCreator)(nil)

var zero = big.NewInt(0)

// ArgsNewRewardsCreator defines the arguments structure needed to create a new rewards creator
type ArgsNewRewardsCreator struct {
	BaseRewardsCreatorArgs
}

type rewardsCreator struct {
	*baseRewardsCreator
}

type rewardInfoData struct {
	accumulatedFees     *big.Int
	address             string
	rewardsFromProtocol *big.Int
}

// NewRewardsCreator creates a new rewards creator object
func NewRewardsCreator(args ArgsNewRewardsCreator) (*rewardsCreator, error) {
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
func (rc *rewardsCreator) CreateRewardsMiniBlocks(
	metaBlock data.MetaHeaderHandler,
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	computedEconomics *block.Economics,
) (block.MiniBlockSlice, error) {
	if check.IfNil(metaBlock) {
		return nil, epochStart.ErrNilHeaderHandler
	}
	rc.mutRewardsData.Lock()
	defer rc.mutRewardsData.Unlock()

	rc.clean()
	rc.flagDelegationSystemSCEnabled.SetValue(metaBlock.GetEpoch() >= rc.delegationSystemSCEnableEpoch)

	economicsData := metaBlock.GetEpochStartHandler().GetEconomicsHandler()
	log.Debug("rewardsCreator.CreateRewardsMiniBlocks",
		"totalToDistribute", economicsData.GetTotalToDistribute(),
		"rewardsForProtocolSustainability", economicsData.GetRewardsForProtocolSustainability(),
		"rewardsPerBlock", economicsData.GetRewardsPerBlock(),
		"devFeesInEpoch", metaBlock.GetDevFeesInEpoch(),
	)

	miniBlocks := rc.initializeRewardsMiniBlocks()

	protSustRwdTx, protSustShardId, err := rc.createProtocolSustainabilityRewardTransaction(metaBlock, computedEconomics)
	if err != nil {
		return nil, err
	}

	rc.fillBaseRewardsPerBlockPerNode(economicsData.GetRewardsPerBlock())
	err = rc.addValidatorRewardsToMiniBlocks(validatorsInfo, metaBlock, miniBlocks, protSustRwdTx)
	if err != nil {
		return nil, err
	}

	totalWithoutDevelopers := big.NewInt(0).Sub(economicsData.GetTotalToDistribute(), metaBlock.GetDevFeesInEpoch())
	difference := big.NewInt(0).Sub(totalWithoutDevelopers, rc.accumulatedRewards)
	log.Debug("arithmetic difference in end of epoch rewards economics", "epoch", metaBlock.GetEpoch(), "value", difference)
	rc.adjustProtocolSustainabilityRewards(protSustRwdTx, difference)
	err = rc.addProtocolRewardToMiniBlocks(protSustRwdTx, miniBlocks, protSustShardId)
	if err != nil {
		return nil, err
	}

	return rc.finalizeMiniBlocks(miniBlocks), nil
}

func (rc *rewardsCreator) adjustProtocolSustainabilityRewards(protocolSustainabilityRwdTx *rewardTx.RewardTx, dustRewards *big.Int) {
	if protocolSustainabilityRwdTx.Value.Cmp(big.NewInt(0)) < 0 {
		log.Error("negative rewards protocol sustainability")
		protocolSustainabilityRwdTx.Value.SetUint64(0)
	}

	if dustRewards.Cmp(big.NewInt(0)) < 0 {
		log.Debug("will adjust protocol rewards with negative value", "dustRewards", dustRewards.String())
	}

	protocolSustainabilityRwdTx.Value.Add(protocolSustainabilityRwdTx.Value, dustRewards)

	log.Debug("baseRewardsCreator.adjustProtocolSustainabilityRewards - rewardsCreator",
		"epoch", protocolSustainabilityRwdTx.GetEpoch(),
		"destination", protocolSustainabilityRwdTx.GetRcvAddr(),
		"value", protocolSustainabilityRwdTx.GetValue().String())

	rc.protocolSustainabilityValue.Set(protocolSustainabilityRwdTx.Value)
}

func (rc *rewardsCreator) addValidatorRewardsToMiniBlocks(
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	metaBlock data.HeaderHandler,
	miniBlocks block.MiniBlockSlice,
	protocolSustainabilityRwdTx *rewardTx.RewardTx,
) error {
	rwdAddrValidatorInfo := rc.computeValidatorInfoPerRewardAddress(validatorsInfo, protocolSustainabilityRwdTx, metaBlock.GetEpoch())
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
				log.Debug("rewardsCreator.addValidatorRewardsToMiniBlocks - not supported metaChain address",
					"move to protocol vault", rwdTx.GetValue())
				protocolSustainabilityRwdTx.Value.Add(protocolSustainabilityRwdTx.Value, rwdTx.GetValue())
				continue
			}
		}

		if rwdTx.Value.Cmp(big.NewInt(0)) < 0 {
			log.Error("negative rewards", "rcv", rwdTx.RcvAddr)
			continue
		}
		rc.currTxs.AddTx(rwdTxHash, rwdTx)

		log.Debug("rewardsCreator.addValidatorRewardsToMiniBlocks",
			"epoch", rwdTx.GetEpoch(),
			"destination", rwdTx.GetRcvAddr(),
			"value", rwdTx.GetValue().String())

		miniBlocks[mbId].TxHashes = append(miniBlocks[mbId].TxHashes, rwdTxHash)
	}

	return nil
}

func (rc *rewardsCreator) computeValidatorInfoPerRewardAddress(
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	protocolSustainabilityRwd *rewardTx.RewardTx,
	epoch uint32,
) map[string]*rewardInfoData {

	rwdAddrValidatorInfo := make(map[string]*rewardInfoData)

	for _, shardValidatorsInfo := range validatorsInfo {
		for _, validatorInfo := range shardValidatorsInfo {
			rewardsPerBlockPerNodeForShard := rc.mapBaseRewardsPerBlockPerValidator[validatorInfo.ShardId]
			protocolRewardValue := big.NewInt(0).Mul(rewardsPerBlockPerNodeForShard, big.NewInt(0).SetUint64(uint64(validatorInfo.NumSelectedInSuccessBlocks)))

			isFix1Enabled := rc.isRewardsFix1Enabled(epoch)
			if isFix1Enabled && validatorInfo.LeaderSuccess == 0 && validatorInfo.ValidatorSuccess == 0 {
				protocolSustainabilityRwd.Value.Add(protocolSustainabilityRwd.Value, protocolRewardValue)
				continue
			}
			if !isFix1Enabled && validatorInfo.LeaderSuccess == 0 && validatorInfo.ValidatorFailure == 0 {
				protocolSustainabilityRwd.Value.Add(protocolSustainabilityRwd.Value, protocolRewardValue)
				continue
			}

			rwdInfo, ok := rwdAddrValidatorInfo[string(validatorInfo.RewardAddress)]
			if !ok {
				rwdInfo = &rewardInfoData{
					accumulatedFees:     big.NewInt(0),
					rewardsFromProtocol: big.NewInt(0),
					address:             string(validatorInfo.RewardAddress),
				}
				rwdAddrValidatorInfo[string(validatorInfo.RewardAddress)] = rwdInfo
			}

			rwdInfo.accumulatedFees.Add(rwdInfo.accumulatedFees, validatorInfo.AccumulatedFees)
			rwdInfo.rewardsFromProtocol.Add(rwdInfo.rewardsFromProtocol, protocolRewardValue)
		}
	}

	return rwdAddrValidatorInfo
}

// VerifyRewardsMiniBlocks verifies if received rewards miniblocks are correct
func (rc *rewardsCreator) VerifyRewardsMiniBlocks(
	metaBlock data.MetaHeaderHandler,
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	computedEconomics *block.Economics,
) error {
	if check.IfNil(metaBlock) {
		return epochStart.ErrNilHeaderHandler
	}

	createdMiniBlocks, err := rc.CreateRewardsMiniBlocks(metaBlock, validatorsInfo, computedEconomics)
	if err != nil {
		return err
	}

	return rc.verifyCreatedRewardMiniBlocksWithMetaBlock(metaBlock, createdMiniBlocks)
}

// IsInterfaceNil return true if underlying object is nil
func (rc *rewardsCreator) IsInterfaceNil() bool {
	return rc == nil
}

func (rc *rewardsCreator) isRewardsFix1Enabled(epoch uint32) bool {
	return epoch > rc.rewardsFix1EnableEpoch
}
