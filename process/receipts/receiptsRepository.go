package receipts

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("process/receipts")

type receiptsRepository struct {
	marshaller        marshal.Marshalizer
	hasher            hashing.Hasher
	store             dataRetriever.StorageService
	emptyReceiptsHash []byte
}

// NewReceiptsRepository creates a new receiptsRepository
func NewReceiptsRepository(args ArgsNewReceiptsRepository) (*receiptsRepository, error) {
	err := args.check()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errCannotCreateReceiptsRepository, err)
	}

	// We cache this special hash (as a struct member)
	emptyReceiptsHash, err := createEmptyReceiptsHash(args.Marshaller, args.Hasher)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errCannotCreateReceiptsRepository, err)
	}

	return &receiptsRepository{
		marshaller:        args.Marshaller,
		hasher:            args.Hasher,
		store:             args.Store,
		emptyReceiptsHash: emptyReceiptsHash,
	}, nil
}

func createEmptyReceiptsHash(marshaller marshal.Marshalizer, hasher hashing.Hasher) ([]byte, error) {
	emptyReceipts := &batch.Batch{Data: make([][]byte, 0)}
	return core.CalculateHash(marshaller, hasher, emptyReceipts)
}

// SaveReceipts saves the receipts in the storer (receipts unit) for a given block header
func (repository *receiptsRepository) SaveReceipts(holder *ReceiptsHolder, header data.HeaderHandler, headerHash []byte) error {
	storageKey := repository.decideStorageKey(header.GetReceiptsHash(), headerHash)

	log.Debug("receiptsRepository.SaveReceipts()", "headerNonce", header.GetNonce(), "storageKey", storageKey)

	receiptsBytes, err := marshalReceiptsHolder(holder, repository.marshaller)
	if err != nil {
		return err
	}

	// Nothing to be saved
	if len(receiptsBytes) == 0 {
		return nil
	}

	err = repository.getStorer().Put(storageKey, receiptsBytes)
	if err != nil {
		return fmt.Errorf("%w: %v", errCannotSaveReceipts, err)
	}

	return nil
}

func (repository *receiptsRepository) getStorer() storage.Storer {
	return repository.store.GetStorer(dataRetriever.ReceiptsUnit)
}

// LoadReceipts loads the receipts, given a block header
func (repository *receiptsRepository) LoadReceipts(header data.HeaderHandler, headerHash []byte) (*ReceiptsHolder, error) {
	storageKey := repository.decideStorageKey(header.GetReceiptsHash(), headerHash)

	batchBytes, err := repository.getStorer().GetFromEpoch(storageKey, header.GetEpoch())
	if err != nil {
		if storage.IsNotFoundInStorageErr(err) {
			return newReceiptsHolder(), nil
		}

		return nil, fmt.Errorf("%w: %v, storageKey = %s", errCannotLoadReceipts, err, hex.EncodeToString(storageKey))
	}

	holder, err := unmarshalReceiptsHolder(batchBytes, repository.marshaller)
	if err != nil {
		return nil, err
	}

	return holder, nil
}

func (repository *receiptsRepository) decideStorageKey(receiptsHash []byte, headerHash []byte) []byte {
	isEmptyReceiptsHash := bytes.Equal(receiptsHash, repository.emptyReceiptsHash)
	if isEmptyReceiptsHash {
		return headerHash
	}

	return receiptsHash
}

// IsInterfaceNil returns true if there is no value under the interface
func (repository *receiptsRepository) IsInterfaceNil() bool {
	return repository == nil
}
