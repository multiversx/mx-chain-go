//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. miniblockMetadata.proto

package fullHistory

import (
	"fmt"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("core/fullHistory")

// HistoryRepositoryArguments is a structure that stores all components that are needed to a history processor
type HistoryRepositoryArguments struct {
	SelfShardID                 uint32
	MiniblocksMetadataStorer    storage.Storer
	MiniblockHashByTxHashStorer storage.Storer
	EpochByHashStorer           storage.Storer
	Marshalizer                 marshal.Marshalizer
	Hasher                      hashing.Hasher
}

// MiniblockMetadataWithEpoch holds miniblock metadata and epoch
type MiniblockMetadataWithEpoch struct {
	Epoch uint32
	*MiniblockMetadata
}

type historyProcessor struct {
	selfShardID                uint32
	miniblocksMetadataStorer   storage.Storer
	miniblockHashByTxHashIndex interface{}
	epochByHashIndex           *epochByHashIndex
	marshalizer                marshal.Marshalizer
	hasher                     hashing.Hasher
}

// NewHistoryRepository will create a new instance of HistoryRepository
func NewHistoryRepository(arguments HistoryRepositoryArguments) (*historyProcessor, error) {
	if check.IfNil(arguments.MiniblocksMetadataStorer) {
		return nil, core.ErrNilStore
	}
	if check.IfNil(arguments.MiniblockHashByTxHashStorer) {
		return nil, core.ErrNilStore
	}
	if check.IfNil(arguments.EpochByHashStorer) {
		return nil, core.ErrNilStore
	}
	if check.IfNil(arguments.Marshalizer) {
		return nil, core.ErrNilMarshalizer
	}
	if check.IfNil(arguments.Hasher) {
		return nil, core.ErrNilHasher
	}

	hashToEpochIndex := newHashToEpochIndex(arguments.EpochByHashStorer, arguments.Marshalizer)

	return &historyProcessor{
		selfShardID:              arguments.SelfShardID,
		miniblocksMetadataStorer: arguments.MiniblocksMetadataStorer,
		marshalizer:              arguments.Marshalizer,
		hasher:                   arguments.Hasher,
		epochByHashIndex:         hashToEpochIndex,
	}, nil
}

// RecordBlock records a block
func (hp *historyProcessor) RecordBlock(blockHeaderHash []byte, blockHeader data.HeaderHandler, blockBody data.BodyHandler) error {
	body, ok := blockBody.(*block.Body)
	if !ok {
		return errCannotCastToBlockBody
	}

	epoch := blockHeader.GetEpoch()

	err := hp.epochByHashIndex.saveEpochByHash(blockHeaderHash, epoch)
	if err != nil {
		return newErrCannotSaveEpochByHash("block header", blockHeaderHash, err)
	}

	for _, miniblock := range body.MiniBlocks {
		if miniblock.Type == block.PeerBlock {
			continue
		}

		err = hp.saveMiniblockMetadata(blockHeaderHash, blockHeader, miniblock, epoch)
		if err != nil {
			continue
		}
	}

	return nil
}

func (hp *historyProcessor) saveMiniblockMetadata(blockHeaderHash []byte, blockHeader data.HeaderHandler, miniblock *block.MiniBlock, epoch uint32) error {
	miniblockHash, err := core.CalculateHash(hp.marshalizer, hp.hasher, miniblock)
	if err != nil {
		return err
	}

	err = hp.epochByHashIndex.saveEpochByHash(miniblockHash, epoch)
	if err != nil {
		return newErrCannotSaveEpochByHash("miniblock", miniblockHash, err)
	}

	miniblockMetadata := &MiniblockMetadata{
		HeaderHash:  blockHeaderHash,
		MbHash:      miniblockHash,
		Round:       blockHeader.GetRound(),
		HeaderNonce: blockHeader.GetNonce(),
		SndShardID:  miniblock.GetSenderShardID(),
		RcvShardID:  miniblock.GetReceiverShardID(),
		Status:      []byte(hp.getMiniblockStatus(miniblock)),
	}

	miniblockMetadataBytes, err := hp.marshalizer.Marshal(miniblockMetadata)
	if err != nil {
		return err
	}

	for _, txHash := range miniblock.TxHashes {
		// Question for review: so, it means that, for 1000 txs in a miniblock, we save the same thing 1000 times, right?
		err = hp.saveTransactionMetadata(miniblockMetadataBytes, txHash, epoch)
		if err != nil {
			log.Warn("cannot save tx in storage", "hash", string(txHash), "error", err.Error())
		}
	}

	return nil
}

func (hp *historyProcessor) getMiniblockStatus(miniblock *block.MiniBlock) core.TransactionStatus {
	if miniblock.Type == block.InvalidBlock {
		return core.TxStatusInvalid
	}
	if miniblock.ReceiverShardID == hp.selfShardID {
		return core.TxStatusExecuted
	}

	return core.TxStatusPartiallyExecuted
}

func (hp *historyProcessor) saveTransactionMetadata(historyTxBytes []byte, txHash []byte, epoch uint32) error {
	err := hp.epochByHashIndex.saveEpochByHash(txHash, epoch)
	if err != nil {
		return newErrCannotSaveEpochByHash("tx", txHash, err)
	}

	err = hp.miniblocksMetadataStorer.Put(txHash, historyTxBytes)
	if err != nil {
		return fmt.Errorf("cannot save in storage history transaction %w", err)
	}

	return nil
}

// GetMiniblockMetadataByTxHash will return a history transaction for the given hash from storage
func (hp *historyProcessor) GetMiniblockMetadataByTxHash(hash []byte) (*MiniblockMetadataWithEpoch, error) {
	epoch, err := hp.epochByHashIndex.getEpochByHash(hash)
	if err != nil {
		return nil, err
	}

	metadataBytes, err := hp.miniblocksMetadataStorer.GetFromEpoch(hash, epoch)
	if err != nil {
		return nil, err
	}

	metadata := &MiniblockMetadata{}
	err = hp.marshalizer.Unmarshal(metadata, metadataBytes)
	if err != nil {
		return nil, err
	}

	return &MiniblockMetadataWithEpoch{
		Epoch:             epoch,
		MiniblockMetadata: metadata,
	}, nil
}

// GetEpochByHash will return epoch for a given hash
func (hp *historyProcessor) GetEpochByHash(hash []byte) (uint32, error) {
	return hp.epochByHashIndex.getEpochByHash(hash)
}

// IsEnabled will always returns true
func (hp *historyProcessor) IsEnabled() bool {
	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (hp *historyProcessor) IsInterfaceNil() bool {
	return hp == nil
}
