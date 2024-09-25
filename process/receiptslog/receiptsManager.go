package receiptslog

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/state"
)

// ArgsReceiptsManager is the structure that holds the components needed to a new receipts manager
type ArgsReceiptsManager struct {
	TrieHandler Interactor
}

// ArgsGenerateReceiptsAndSave is the DTO needed to provided input data to generate receipts
type ArgsGenerateReceiptsAndSave struct {
	// add data needed to create receipt structure
	// logs
}

type receiptsManager struct {
	trieInteractor Interactor
}

// NewReceiptsManager will create a new instance of receipts manager
func NewReceiptsManager(args ArgsReceiptsManager) (*receiptsManager, error) {
	if check.IfNil(args.TrieHandler) {
		return nil, ErrNilTrieInteractor
	}

	return &receiptsManager{
		trieInteractor: args.TrieHandler,
	}, nil
}

// GenerateReceiptsTrieAndSaveDataInStorage will generate the receipts trie based on the input data and return the receipt trie root hash
func (rm *receiptsManager) GenerateReceiptsTrieAndSaveDataInStorage(args ArgsGenerateReceiptsAndSave) ([]byte, error) {
	// TODO  generate the list of receipts

	err := rm.trieInteractor.CreateNewTrie()
	if err != nil {
		return nil, err
	}

	receipts := make([]state.Receipt, 0)
	for _, rec := range receipts {
		err = rm.trieInteractor.AddReceiptData(rec)
		if err != nil {
			return nil, err
		}
	}

	receiptsRootHash, err := rm.trieInteractor.Save()
	if err != nil {
		return nil, err
	}

	return receiptsRootHash, nil
}

// SyncReceiptsTrie will sync the receipts trie from network
func (rm *receiptsManager) SyncReceiptsTrie(receiptsRootHash []byte) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rm *receiptsManager) IsInterfaceNil() bool {
	return rm == nil
}
