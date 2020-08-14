//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. transactionsGroupMetadata.proto

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

// TransactionsGroupMetadataWithEpoch is a structure for a history transaction that also contain epoch
type TransactionsGroupMetadataWithEpoch struct {
	Epoch uint32
	*TransactionsGroupMetadata
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
		err = hp.saveMiniblockData(blockHeaderHash, blockHeader, miniblock, epoch)
		if err != nil {
			continue
		}
	}

	return nil
}

func (hp *historyProcessor) saveMiniblockData(blockHeaderHash []byte, blockHeader data.HeaderHandler, miniblock *block.MiniBlock, epoch uint32) error {
	miniblockHash, err := core.CalculateHash(hp.marshalizer, hp.hasher, miniblock)
	if err != nil {
		return err
	}

	err = hp.epochByHashIndex.saveEpochByHash(miniblockHash, epoch)
	if err != nil {
		return newErrCannotSaveEpochByHash("miniblock", miniblockHash, err)
	}

	var status core.TransactionStatus
	if miniblock.ReceiverShardID == hp.selfShardID {
		status = core.TxStatusExecuted
	} else {
		status = core.TxStatusPartiallyExecuted
	}

	transactionsGroupMetadata := buildTransactionsGroupMetadata(
		blockHeader,
		blockHeaderHash,
		miniblockHash,
		miniblock.ReceiverShardID,
		miniblock.SenderShardID,
		status,
	)
	transactionsGroupMetadataBytes, err := hp.marshalizer.Marshal(transactionsGroupMetadata)
	if err != nil {
		return err
	}

	for _, txHash := range miniblock.TxHashes {
		// Question for review: so, it means that, for 1000 txs in a miniblock, we save the same thing 1000 times, right?
		err = hp.saveTransactionMetadata(transactionsGroupMetadataBytes, txHash, epoch)
		if err != nil {
			log.Warn("cannot save tx in storage", "hash", string(txHash), "error", err.Error())
		}
	}

	return nil
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

func buildTransactionsGroupMetadata(
	headerHandler data.HeaderHandler,
	headerHash []byte,
	mbHash []byte,
	rcvShardID uint32,
	sndShardID uint32,
	status core.TransactionStatus,
) *TransactionsGroupMetadata {
	return &TransactionsGroupMetadata{
		HeaderHash:  headerHash,
		MbHash:      mbHash,
		Round:       headerHandler.GetRound(),
		HeaderNonce: headerHandler.GetNonce(),
		RcvShardID:  rcvShardID,
		SndShardID:  sndShardID,
		Status:      []byte(status),
	}
}

// GetTransactionsGroupMetadata will return a history transaction for the given hash from storage
func (hp *historyProcessor) GetTransactionsGroupMetadata(hash []byte) (*TransactionsGroupMetadataWithEpoch, error) {
	epoch, err := hp.epochByHashIndex.getEpochByHash(hash)
	if err != nil {
		return nil, err
	}

	historyTxBytes, err := hp.miniblocksMetadataStorer.GetFromEpoch(hash, epoch)
	if err != nil {
		return nil, err
	}

	historyTx := &TransactionsGroupMetadata{}
	err = hp.marshalizer.Unmarshal(historyTx, historyTxBytes)
	if err != nil {
		return nil, err
	}

	return &TransactionsGroupMetadataWithEpoch{
		Epoch:                     epoch,
		TransactionsGroupMetadata: historyTx,
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
