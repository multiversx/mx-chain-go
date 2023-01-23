package receipts

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("process/receipts")

type receiptsRepository struct {
	marshaller        marshal.Marshalizer
	hasher            hashing.Hasher
	storer            storage.Storer
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

	storer, err := args.Store.GetStorer(dataRetriever.ReceiptsUnit)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errCannotCreateReceiptsRepository, err)
	}

	return &receiptsRepository{
		marshaller:        args.Marshaller,
		hasher:            args.Hasher,
		storer:            storer,
		emptyReceiptsHash: emptyReceiptsHash,
	}, nil
}

func createEmptyReceiptsHash(marshaller marshal.Marshalizer, hasher hashing.Hasher) ([]byte, error) {
	emptyReceipts := &batch.Batch{Data: make([][]byte, 0)}
	return core.CalculateHash(marshaller, hasher, emptyReceipts)
}

// SaveReceipts saves the receipts in the storer (receipts unit) for a given block header
func (repository *receiptsRepository) SaveReceipts(holder common.ReceiptsHolder, header data.HeaderHandler, headerHash []byte) error {
	if check.IfNil(holder) {
		return errNilReceiptsHolder
	}
	if check.IfNil(header) {
		return errNilBlockHeader
	}
	if len(headerHash) == 0 {
		return errEmptyBlockHash
	}

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

	err = repository.storer.Put(storageKey, receiptsBytes)
	if err != nil {
		return fmt.Errorf("%w: %v", errCannotSaveReceipts, err)
	}

	return nil
}

// LoadReceipts loads the receipts, given a block header
func (repository *receiptsRepository) LoadReceipts(header data.HeaderHandler, headerHash []byte) (common.ReceiptsHolder, error) {
	storageKey := repository.decideStorageKey(header.GetReceiptsHash(), headerHash)

	batchBytes, err := repository.storer.GetFromEpoch(storageKey, header.GetEpoch())
	if err != nil {
		if storage.IsNotFoundInStorageErr(err) {
			return createEmptyReceiptsHolder(), nil
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
