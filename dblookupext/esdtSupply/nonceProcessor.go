package esdtSupply

import (
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const (
	processedBlockKey = "processed-block"
)

type nonceProcessor struct {
	marshalizer marshal.Marshalizer
	storer      storage.Storer
}

func newNonceProcessor(marshalizer marshal.Marshalizer, storer storage.Storer) *nonceProcessor {
	return &nonceProcessor{
		marshalizer: marshalizer,
		storer:      storer,
	}
}

func (np *nonceProcessor) shouldProcessLog(blockNonce uint64, isRevert bool) (bool, error) {
	nonceFromStorage, err := np.getLatestProcessedBlockNonceFromStorage()
	if err != nil {
		return false, err
	}

	if isRevert {
		return blockNonce == nonceFromStorage, nil
	}

	return blockNonce > nonceFromStorage, nil
}

func (np *nonceProcessor) getLatestProcessedBlockNonceFromStorage() (uint64, error) {
	processedBlockBytes, err := np.storer.Get([]byte(processedBlockKey))
	if err != nil {
		log.Warn("logsProcessor.getLatestProcessedBlockNonceFromStorage nothing in storage")
		return 0, nil
	}

	processedBlock := &ProcessedBlockNonce{}
	err = np.marshalizer.Unmarshal(processedBlock, processedBlockBytes)
	if err != nil {
		return 0, err
	}

	return processedBlock.Nonce, nil
}

func (np *nonceProcessor) saveNonceInStorage(nonce uint64) error {
	processedBlock := &ProcessedBlockNonce{
		Nonce: nonce,
	}

	processedBlockBytes, err := np.marshalizer.Marshal(processedBlock)
	if err != nil {
		return err
	}

	return np.storer.Put([]byte(processedBlockKey), processedBlockBytes)
}
