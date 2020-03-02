package metachain

import (
	"bytes"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"sort"
)

// ArgsNewValidatorInfoCreator defines the arguments structure needed to create a new validatorInfo creator
type ArgsNewValidatorInfoCreator struct {
	ShardCoordinator sharding.Coordinator
	MiniBlockStorage storage.Storer
	Hasher           hashing.Hasher
	Marshalizer      marshal.Marshalizer
}

type validatorInfoCreator struct {
	shardCoordinator sharding.Coordinator
	miniBlockStorage storage.Storer

	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
}

// NewEpochStartRewardsCreator creates a new rewards creator object
func NewValidatorInfoCreator(args ArgsNewValidatorInfoCreator) (*validatorInfoCreator, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, epochStart.ErrNilShardCoordinator
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

	vic := &validatorInfoCreator{
		shardCoordinator: args.ShardCoordinator,
		hasher:           args.Hasher,
		marshalizer:      args.Marshalizer,
		miniBlockStorage: args.MiniBlockStorage,
	}

	return vic, nil
}

// CreateValidatorInfoMiniBlocks creates the validatorInfo miniblocks according to the validatorInfo
func (r *validatorInfoCreator) CreateValidatorInfoMiniBlocks(validatorInfos map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error) {
	miniblocks := make([]*block.MiniBlock, 0)

	for shardId, validators := range validatorInfos {
		if len(validators) == 0 {
			continue
		}

		miniBlock := &block.MiniBlock{}
		miniBlock.SenderShardID = r.shardCoordinator.SelfId()
		miniBlock.ReceiverShardID = shardId
		miniBlock.TxHashes = make([][]byte, len(validators))
		miniBlock.Type = block.PeerBlock

		validatorCopy := make([]*state.ValidatorInfo, len(validators))
		copy(validatorCopy, validators)
		sort.Slice(validatorCopy, func(a, b int) bool {
			return bytes.Compare(validatorCopy[a].PublicKey, validatorCopy[b].PublicKey) < 0
		})

		for index, validator := range validatorCopy {
			marshalizedValidator, err := r.marshalizer.Marshal(validator)
			if err != nil {
				return nil, err
			}
			miniBlock.TxHashes[index] = marshalizedValidator
		}

		miniblocks = append(miniblocks, miniBlock)
	}

	return miniblocks, nil
}

// VerifyRewardsMiniBlocks verifies if received rewards miniblocks are correct
func (r *validatorInfoCreator) VerifyValidatorInfoMiniBlocks(metaBlock *block.MetaBlock, validatorInfos map[uint32][]*state.ValidatorInfo) error {
	createdMiniBlocks, err := r.CreateValidatorInfoMiniBlocks(validatorInfos)
	if err != nil {
		return err
	}

	numReceivedRewardsMBs := 0
	for _, miniBlockHdr := range metaBlock.MiniBlockHeaders {
		if miniBlockHdr.Type != block.PeerBlock {
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
			return epochStart.ErrValidatorMiniBlockHashDoesNotMatch
		}
	}

	if len(createdMiniBlocks) != numReceivedRewardsMBs {
		return epochStart.ErrValidatorMiniBlockHashDoesNotMatch
	}

	return nil
}

// SaveValidatorInfoBlockToStorage saves created data to storage
func (r *validatorInfoCreator) SaveValidatorInfoBlocksToStorage(metaBlock *block.MetaBlock, body block.Body) {
	for _, miniBlock := range body {
		if miniBlock.Type != block.PeerBlock {
			continue
		}

		for _, mbHeader := range metaBlock.MiniBlockHeaders {
			if mbHeader.Type != block.PeerBlock {
				continue
			}

			marshaledData, err := r.marshalizer.Marshal(miniBlock)
			if err != nil {
				continue
			}

			_ = r.miniBlockStorage.Put(mbHeader.Hash, marshaledData)
		}
	}
}

// DeleteValidatorInfoBlockFromStorage deletes data from storage
func (r *validatorInfoCreator) DeleteValidatorInfoBlocksFromStorage(metaBlock *block.MetaBlock) {
	for _, mbHeader := range metaBlock.MiniBlockHeaders {
		_ = r.miniBlockStorage.Remove(mbHeader.Hash)
	}
}

// IsInterfaceNil return true if underlying object is nil
func (r *validatorInfoCreator) IsInterfaceNil() bool {
	return r == nil
}
