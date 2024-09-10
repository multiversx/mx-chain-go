package receiptslog

import "github.com/multiversx/mx-chain-core-go/data/state"

// Interactor defines what a trie interactor should be able to do
type Interactor interface {
	CreateNewTrie() error
	AddReceiptData(receiptData state.Receipt) error
	Save() ([]byte, error)
	IsInterfaceNil() bool
}

// ReceiptsManagerHandler defines what a receipts manager should be able to do
type ReceiptsManagerHandler interface {
	GenerateReceiptsTrieAndSaveDataInStorage(args ArgsGenerateReceiptsAndSave) ([]byte, error)
	SyncReceiptsTrie(receiptsRootHash []byte) error
	IsInterfaceNil() bool
}
