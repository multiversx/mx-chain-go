package receiptslog

import (
	"context"
	"github.com/multiversx/mx-chain-core-go/data/state"
)

// Interactor defines what a trie interactor should be able to do
type Interactor interface {
	CreateNewTrie() error
	AddReceiptData(receiptData state.Receipt) error
	Save() ([]byte, error)
	GetSerializedNode(nodeHash []byte) ([]byte, error)
	GetBranchNodesMap(branchNodesSerialized []byte) (map[string][]byte, error)
	IsInterfaceNil() bool
}

// ReceiptsManagerHandler defines what a receipts manager should be able to do
type ReceiptsManagerHandler interface {
	GenerateReceiptsTrieAndSaveDataInStorage(args ArgsGenerateReceiptsAndSave) ([]byte, error)
	SyncReceiptsTrie(receiptsRootHash []byte) error
	IsInterfaceNil() bool
}

// ReceiptsDataSyncer defines what a receipts data syncer should be able to do
type ReceiptsDataSyncer interface {
	SyncReceiptsDataFor(hashes [][]byte, ctx context.Context) error
	GetReceiptsData() (map[string][]byte, error)
	ClearFields()
	IsInterfaceNil() bool
}
