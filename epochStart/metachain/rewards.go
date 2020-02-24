package metachain

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgsNewRewardsCreator defines the arguments structure needed to create a new rewards creator
type ArgsNewRewardsCreator struct {
	ShardCoordinator sharding.Coordinator
	AddrConverter    state.AddressConverter
	RewardsStorage   storage.Storer
	MiniBlockStorage storage.Storer
	Hasher           hashing.Hasher
	Marshalizer      marshal.Marshalizer
}

type rewardsCreator struct {
	currTxs          dataRetriever.TransactionCacher
	shardCoordinator sharding.Coordinator
	addrConverter    state.AddressConverter
	rewardsStorage   storage.Storer
	miniBlockStorage storage.Storer

	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
}

type rewardInfoData struct {
	LeaderSuccess              uint32
	LeaderFailure              uint32
	ValidatorSuccess           uint32
	ValidatorFailure           uint32
	NumSelectedInSuccessBlocks uint32
	AccumulatedFees            *big.Int
}

// NewEpochStartRewardsCreator creates a new rewards creator object
func NewEpochStartRewardsCreator(args ArgsNewRewardsCreator) (*rewardsCreator, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, epochStart.ErrNilShardCoordinator
	}
	if check.IfNil(args.AddrConverter) {
		return nil, epochStart.ErrNilAddressConverter
	}
	if check.IfNil(args.RewardsStorage) {
		return nil, epochStart.ErrNilStorage
	}
	if check.IfNil(args.Marshalizer) {
		return nil, epochStart.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, epochStart.ErrNilHasher
	}
	if check.IfNil(args.MiniBlockStorage) {
		return nil, epochStart.ErrNilStorage
	}

	currTxsCache, err := dataPool.NewCurrentBlockPool()
	if err != nil {
		return nil, err
	}

	r := &rewardsCreator{
		currTxs:          currTxsCache,
		shardCoordinator: args.ShardCoordinator,
		addrConverter:    args.AddrConverter,
		rewardsStorage:   args.RewardsStorage,
		hasher:           args.Hasher,
		marshalizer:      args.Marshalizer,
		miniBlockStorage: args.MiniBlockStorage,
	}

	return r, nil
}

// CreateBlockStarted announces block creation started and cleans inside data
func (r *rewardsCreator) clean() {
	r.currTxs.Clean()
}

// CreateRewardsMiniBlocks creates the rewards miniblocks according to economics data and validator info
func (r *rewardsCreator) CreateRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorInfos map[uint32][]*state.ValidatorInfoData) (data.BodyHandler, error) {
	r.clean()

	miniBlocks := make(block.Body, r.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < r.shardCoordinator.NumberOfShards(); i++ {
		miniBlocks[i] = &block.MiniBlock{}
		miniBlocks[i].SenderShardID = sharding.MetachainShardId
		miniBlocks[i].ReceiverShardID = i
		miniBlocks[i].Type = block.RewardsBlock
		miniBlocks[i].TxHashes = make([][]byte, 0)
	}

	rwdAddrValidatorInfo := r.computeValidatorInfoPerRewardAddress(validatorInfos)

	for address, rwdInfo := range rwdAddrValidatorInfo {
		addrContainer, err := r.addrConverter.CreateAddressFromPublicKeyBytes([]byte(address))
		if err != nil {
			return nil, err
		}

		rwdTx, rwdTxHash, err := r.createRewardFromRwdInfo([]byte(address), rwdInfo, &metaBlock.EpochStart.Economics, metaBlock)
		if err != nil {
			return nil, err
		}

		r.currTxs.AddTx(rwdTxHash, rwdTx)

		shardId := r.shardCoordinator.ComputeId(addrContainer)
		miniBlocks[shardId].TxHashes = append(miniBlocks[shardId].TxHashes, rwdTxHash)
	}

	for shId := uint32(0); shId < r.shardCoordinator.NumberOfShards(); shId++ {
		sort.Slice(miniBlocks[shId].TxHashes, func(i, j int) bool {
			return bytes.Compare(miniBlocks[shId].TxHashes[i], miniBlocks[shId].TxHashes[j]) < 0
		})
	}

	finalMiniBlocks := make(block.Body, 0)
	for i := uint32(0); i < r.shardCoordinator.NumberOfShards(); i++ {
		if len(miniBlocks[i].TxHashes) > 0 {
			finalMiniBlocks = append(finalMiniBlocks, miniBlocks[i])
		}
	}

	return finalMiniBlocks, nil
}

func (r *rewardsCreator) computeValidatorInfoPerRewardAddress(
	validatorInfos map[uint32][]*state.ValidatorInfoData,
) map[string]*rewardInfoData {

	rwdAddrValidatorInfo := make(map[string]*rewardInfoData)

	for _, shardValidatorInfos := range validatorInfos {
		for _, validatorInfo := range shardValidatorInfos {
			rwdInfo, ok := rwdAddrValidatorInfo[string(validatorInfo.RewardAddress)]
			if !ok {
				rwdInfo = &rewardInfoData{
					AccumulatedFees: big.NewInt(0),
				}
				rwdAddrValidatorInfo[string(validatorInfo.RewardAddress)] = rwdInfo
			}

			rwdInfo.LeaderSuccess += validatorInfo.LeaderSuccess
			rwdInfo.LeaderFailure += validatorInfo.LeaderFailure
			rwdInfo.ValidatorFailure += validatorInfo.ValidatorFailure
			rwdInfo.ValidatorSuccess += validatorInfo.ValidatorSuccess
			rwdInfo.NumSelectedInSuccessBlocks += validatorInfo.NumSelectedInSuccessBlocks

			rwdInfo.AccumulatedFees.Add(rwdInfo.AccumulatedFees, validatorInfo.AccumulatedFees)
		}
	}

	return rwdAddrValidatorInfo
}

func (r *rewardsCreator) createRewardFromRwdInfo(
	address []byte,
	rwdInfo *rewardInfoData,
	economicsData *block.Economics,
	metaBlock *block.MetaBlock,
) (*rewardTx.RewardTx, []byte, error) {
	rwdTx := &rewardTx.RewardTx{
		Round:   metaBlock.GetRound(),
		Value:   big.NewInt(0).Set(rwdInfo.AccumulatedFees),
		RcvAddr: address,
		Epoch:   metaBlock.Epoch,
	}

	protocolRewardValue := big.NewInt(0).Mul(economicsData.RewardsPerBlockPerNode, big.NewInt(0).SetUint64(uint64(rwdInfo.NumSelectedInSuccessBlocks)))
	rwdTx.Value.Add(rwdTx.Value, protocolRewardValue)

	rwdTxHash, err := core.CalculateHash(r.marshalizer, r.hasher, rwdTx)
	if err != nil {
		return nil, nil, err
	}

	return rwdTx, rwdTxHash, nil
}

// VerifyRewardsMiniBlocks verifies if received rewards miniblocks are correct
func (r *rewardsCreator) VerifyRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorInfos map[uint32][]*state.ValidatorInfoData) error {
	createdBody, err := r.CreateRewardsMiniBlocks(metaBlock, validatorInfos)
	if err != nil {
		return err
	}

	createdMiniBlocks := createdBody.(block.Body)
	numReceivedRewardsMBs := 0
	for _, miniBlockHdr := range metaBlock.MiniBlockHeaders {
		if miniBlockHdr.Type != block.RewardsBlock {
			continue
		}

		numReceivedRewardsMBs++
		createdMiniBlock := createdMiniBlocks[miniBlockHdr.ReceiverShardID]
		createdMBHash, err := core.CalculateHash(r.marshalizer, r.hasher, createdMiniBlock)
		if err != nil {
			return err
		}

		if !bytes.Equal(createdMBHash, miniBlockHdr.Hash) {
			// TODO: add display debug prints of miniblocks contents
			return epochStart.ErrRewardMiniBlockHashDoesNotMatch
		}
	}

	if len(createdMiniBlocks) != numReceivedRewardsMBs {
		return epochStart.ErrRewardMiniBlocksNumDoesNotMatch
	}

	return nil
}

// CreateMarshalizedData creates the marshalized data to be sent to shards
func (r *rewardsCreator) CreateMarshalizedData(body block.Body) map[string][][]byte {
	txs := make(map[string][][]byte)

	for _, miniBlock := range body {
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		broadCastTopic := createBroadcastTopic(r.shardCoordinator, miniBlock.ReceiverShardID)
		if _, ok := txs[broadCastTopic]; !ok {
			txs[broadCastTopic] = make([][]byte, 0, len(miniBlock.TxHashes))
		}

		for _, txHash := range miniBlock.TxHashes {
			rwdTx, err := r.currTxs.GetTx(txHash)
			if err != nil {
				continue
			}

			marshaledData, err := r.marshalizer.Marshal(rwdTx)
			if err != nil {
				continue
			}

			txs[broadCastTopic] = append(txs[broadCastTopic], marshaledData)
		}
	}

	return txs
}

// SaveTxBlockToStorage saves created data to storage
func (r *rewardsCreator) SaveTxBlockToStorage(metaBlock *block.MetaBlock, body block.Body) {
	for _, miniBlock := range body {
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		for _, txHash := range miniBlock.TxHashes {
			rwdTx, err := r.currTxs.GetTx(txHash)
			if err != nil {
				continue
			}

			marshaledData, err := r.marshalizer.Marshal(rwdTx)
			if err != nil {
				continue
			}

			_ = r.rewardsStorage.Put(txHash, marshaledData)
		}

		for _, mbHeader := range metaBlock.MiniBlockHeaders {
			if mbHeader.Type != block.RewardsBlock {
				continue
			}
			if mbHeader.ReceiverShardID != miniBlock.ReceiverShardID {
				continue
			}

			marshaledData, err := r.marshalizer.Marshal(miniBlock)
			if err != nil {
				continue
			}

			_ = r.miniBlockStorage.Put(mbHeader.Hash, marshaledData)
		}
	}
	r.clean()
}

// DeleteTxsFromStorage deletes data from storage
func (r *rewardsCreator) DeleteTxsFromStorage(metaBlock *block.MetaBlock, body block.Body) {
	for _, miniBlock := range body {
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		for _, txHash := range miniBlock.TxHashes {
			_ = r.rewardsStorage.Remove(txHash)
		}
	}

	for _, mbHeader := range metaBlock.MiniBlockHeaders {
		_ = r.miniBlockStorage.Remove(mbHeader.Hash)
	}
}

// IsInterfaceNil return true if underlying object is nil
func (r *rewardsCreator) IsInterfaceNil() bool {
	return r == nil
}

func createBroadcastTopic(shardC sharding.Coordinator, destShId uint32) string {
	transactionTopic := factory.RewardsTransactionTopic +
		shardC.CommunicationIdentifier(destShId)
	return transactionTopic
}
