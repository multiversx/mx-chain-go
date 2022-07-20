package blockAPI

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

func (bap *baseAPIBlockProcessor) getReceiptsBatch(receiptsHash []byte, headerHash []byte, epoch uint32, options api.BlockQueryOptions) (*batch.Batch, error) {
	receiptsStorageKey := bap.getStorageKeyForReceipts(receiptsHash, headerHash)
	receiptsBatch := &batch.Batch{}

	batchBytes, err := bap.getFromStorerWithEpoch(dataRetriever.ReceiptsUnit, receiptsStorageKey, epoch)
	if err != nil {
		if storage.IsNotFoundInStorageErr(err) {
			return receiptsBatch, nil
		}

		return nil, fmt.Errorf("%w (receipts): %v, storageKey = %s", errCannotLoadMiniblocks, err, hex.EncodeToString(receiptsStorageKey))
	}

	if len(batchBytes) == 0 {
		return receiptsBatch, nil
	}

	err = bap.marshalizer.Unmarshal(receiptsBatch, batchBytes)
	if err != nil {
		return nil, fmt.Errorf("%w (receipts): %v, storageKey = %s", errCannotUnmarshalMiniblocks, err, hex.EncodeToString(receiptsStorageKey))
	}

	return receiptsBatch, nil
}

func (bap *baseAPIBlockProcessor) getStorageKeyForReceipts(receiptsHash []byte, headerHash []byte) []byte {
	isEmptyReceiptsHash := bytes.Equal(receiptsHash, bap.emptyReceiptsHash)
	if isEmptyReceiptsHash {
		return headerHash
	}

	return receiptsHash
}
