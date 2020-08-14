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

var (
	log                   = logger.GetOrCreate("core/fullHistory")
	errInvalidBodyHandler = fmt.Errorf("cannot convert bodyHandler in body")
)

// HistoryRepositoryArguments is a structure that stores all components that are needed to a history processor
type HistoryRepositoryArguments struct {
	SelfShardID     uint32
	HistoryStorer   storage.Storer
	HashEpochStorer storage.Storer
	Marshalizer     marshal.Marshalizer
	Hasher          hashing.Hasher
}

// HistoryTransactionsData is a structure that stores information about history transactions
type HistoryTransactionsData struct {
	HeaderHash    []byte
	HeaderHandler data.HeaderHandler
	BodyHandler   data.BodyHandler
}

// HistoryTransactionWithEpoch is a structure for a history transaction that also contain epoch
type HistoryTransactionWithEpoch struct {
	Epoch uint32
	*TransactionsGroupMetadata
}

type historyProcessor struct {
	selfShardID     uint32
	historyStorer   storage.Storer
	marshalizer     marshal.Marshalizer
	hasher          hashing.Hasher
	hashEpochStorer hashEpochRepository
}

// NewHistoryRepository will create a new instance of HistoryRepository
func NewHistoryRepository(arguments HistoryRepositoryArguments) (*historyProcessor, error) {
	if check.IfNil(arguments.HistoryStorer) {
		return nil, core.ErrNilStore
	}
	if check.IfNil(arguments.Marshalizer) {
		return nil, core.ErrNilMarshalizer
	}
	if check.IfNil(arguments.Hasher) {
		return nil, core.ErrNilHasher
	}
	if check.IfNil(arguments.HashEpochStorer) {
		return nil, core.ErrNilStore
	}

	hashEpochStorer := newHashEpochStorer(arguments.HashEpochStorer, arguments.Marshalizer)

	return &historyProcessor{
		selfShardID:     arguments.SelfShardID,
		historyStorer:   arguments.HistoryStorer,
		marshalizer:     arguments.Marshalizer,
		hasher:          arguments.Hasher,
		hashEpochStorer: hashEpochStorer,
	}, nil
}

// PutTransactionsData will save in storage information about history transactions
func (hp *historyProcessor) PutTransactionsData(historyTxsData *HistoryTransactionsData) error {
	body, ok := historyTxsData.BodyHandler.(*block.Body)
	if !ok {
		return errInvalidBodyHandler
	}

	epoch := historyTxsData.HeaderHandler.GetEpoch()

	err := hp.hashEpochStorer.SaveEpoch(historyTxsData.HeaderHash, epoch)
	if err != nil {
		log.Warn("epochHashProcessor: cannot save header hash", "error", err.Error())
		return err
	}

	for _, mb := range body.MiniBlocks {
		err = hp.saveMiniblockData(historyTxsData, mb, epoch)
		if err != nil {
			continue
		}
	}

	return nil
}

func (hp *historyProcessor) saveMiniblockData(historyTxsData *HistoryTransactionsData, mb *block.MiniBlock, epoch uint32) error {
	mbHash, err := core.CalculateHash(hp.marshalizer, hp.hasher, mb)
	if err != nil {
		log.Warn("cannot calculate miniblock hash", "error", err.Error())
		return err
	}

	err = hp.hashEpochStorer.SaveEpoch(mbHash, epoch)
	if err != nil {
		log.Warn("epochHashProcessor: cannot save miniblock hash", "error", err.Error())
		return err
	}

	var status core.TransactionStatus
	if mb.ReceiverShardID == hp.selfShardID {
		status = core.TxStatusExecuted
	} else {
		status = core.TxStatusPartiallyExecuted
	}

	transactionsMetadata := buildTransactionsGroupMetadata(historyTxsData.HeaderHandler, historyTxsData.HeaderHash, mbHash, mb.ReceiverShardID, mb.SenderShardID, status)
	transactionsMetadataBytes, err := hp.marshalizer.Marshal(transactionsMetadata)
	if err != nil {
		log.Warn("cannot marshal history transaction", "error", err.Error())
		return err
	}

	for _, txHash := range mb.TxHashes {
		err = hp.saveTransactionMetadata(transactionsMetadataBytes, txHash, epoch)
		if err != nil {
			log.Warn("cannot save in storage",
				"hash", string(txHash),
				"error", err.Error())
		}
	}

	return nil
}

func (hp *historyProcessor) saveTransactionMetadata(historyTxBytes []byte, txHash []byte, epoch uint32) error {
	err := hp.hashEpochStorer.SaveEpoch(txHash, epoch)
	if err != nil {
		return fmt.Errorf("epochHashProcessor:cannot save transaction hash %w", err)
	}

	err = hp.historyStorer.Put(txHash, historyTxBytes)
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
func (hp *historyProcessor) GetTransactionsGroupMetadata(hash []byte) (*HistoryTransactionWithEpoch, error) {
	epoch, err := hp.hashEpochStorer.GetEpoch(hash)
	if err != nil {
		return nil, err
	}

	historyTxBytes, err := hp.historyStorer.GetFromEpoch(hash, epoch)
	if err != nil {
		return nil, err
	}

	historyTx := &TransactionsGroupMetadata{}
	err = hp.marshalizer.Unmarshal(historyTx, historyTxBytes)
	if err != nil {
		return nil, err
	}

	return &HistoryTransactionWithEpoch{
		Epoch:                     epoch,
		TransactionsGroupMetadata: historyTx,
	}, nil
}

// GetEpochForHash will return epoch for a given hash
func (hp *historyProcessor) GetEpochForHash(hash []byte) (uint32, error) {
	return hp.hashEpochStorer.GetEpoch(hash)
}

// IsEnabled will always returns true
func (hp *historyProcessor) IsEnabled() bool {
	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (hp *historyProcessor) IsInterfaceNil() bool {
	return hp == nil
}
