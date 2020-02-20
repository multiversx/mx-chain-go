package metachain

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type rewardsCreator struct {
	currTxs          dataRetriever.TransactionCacher
	shardCoordinator sharding.Coordinator
	addrConverter    state.AddressConverter
}

type rewardInfoData struct {
	LeaderSuccess    uint32
	LeaderFailure    uint32
	ValidatorSuccess uint32
	ValidatorFailure uint32
	AccumulatedFees  *big.Int
}

func NewEpochStartRewardsCreator() (*rewardsCreator, error) {
	return &rewardsCreator{}, nil
}

func (r *rewardsCreator) CreateBlockStarted() {
	r.currTxs.Clean()
}

func (r *rewardsCreator) CreateRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorInfos map[uint32][]*state.ValidatorInfoData) (data.BodyHandler, error) {
	miniBlocks := make(block.Body, r.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < r.shardCoordinator.NumberOfShards(); i++ {
		miniBlocks[i].SenderShardID = sharding.MetachainShardId
		miniBlocks[i].ReceiverShardID = i
		miniBlocks[i].Type = block.RewardsBlock
		miniBlocks[i].TxHashes = make([][]byte, 0)
	}

	rwdAddrValidatorInfo := make(map[string]*rewardInfoData)
	for _, shardValidatorInfos := range validatorInfos {
		for _, validatorInfo := range shardValidatorInfos {
			rwdInfo, ok := rwdAddrValidatorInfo[string(validatorInfo.RewardAddress)]
			if !ok {
				rwdAddrValidatorInfo[string(validatorInfo.RewardAddress)] = &rewardInfoData{
					AccumulatedFees: big.NewInt(0),
				}
				rwdInfo = rwdAddrValidatorInfo[string(validatorInfo.RewardAddress)]
			}

			rwdInfo.LeaderSuccess += validatorInfo.LeaderSuccess
			rwdInfo.LeaderFailure += validatorInfo.LeaderFailure
			rwdInfo.ValidatorFailure += validatorInfo.ValidatorFailure
			rwdInfo.ValidatorSuccess += validatorInfo.ValidatorSuccess
			rwdInfo.AccumulatedFees.Add(rwdInfo.AccumulatedFees, validatorInfo.AccumulatedFees)
		}
	}

	for address, rwdInfo := range rwdAddrValidatorInfo {
		addrContainer, err := r.addrConverter.CreateAddressFromPublicKeyBytes([]byte(address))
		if err != nil {
			return nil, err
		}

		rwdTx, rwdTxHash := r.createRewardFromRwdInfo(rwdInfo, &metaBlock.EpochStart.Economics)
		r.currTxs.AddTx(rwdTxHash, rwdTx)

		shardId := r.shardCoordinator.ComputeId(addrContainer)
		miniBlocks[shardId].TxHashes = append(miniBlocks[shardId].TxHashes, rwdTxHash)
	}

	for shId := uint32(0); shId < r.shardCoordinator.NumberOfShards(); shId++ {
		sort.Slice(miniBlocks[shId].TxHashes, func(i, j int) bool {
			return bytes.Compare(miniBlocks[shId].TxHashes[i], miniBlocks[shId].TxHashes[j]) < 0
		})
	}

	return miniBlocks, nil
}

func (r *rewardsCreator) createRewardFromRwdInfo(rwdInfo *rewardInfoData, economicsData *block.Economics) (*rewardTx.RewardTx, []byte) {
	return nil, nil
}

func (r *rewardsCreator) VerifyRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorInfos map[uint32][]*state.ValidatorInfoData) error {
	return nil
}

func (r *rewardsCreator) CreateMarshalizedData() map[string][][]byte {
	return nil
}

func (r *rewardsCreator) SaveTxBlockToStorage(body block.Body) error {
	return nil
}

func (r *rewardsCreator) DeleteTxsFromStorage(body block.Body) error {
	return nil
}

func (r *rewardsCreator) IsInterfaceNil() bool {
	return r == nil
}
