package receiptslog

import (
	"context"
	"github.com/multiversx/mx-chain-core-go/data/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/trie"
)

// ArgsReceiptsManager is the structure that holds the components needed to a new receipts manager
type ArgsReceiptsManager struct {
	TrieHandler        Interactor
	ReceiptsDataSyncer ReceiptsDataSyncer
}

// ArgsGenerateReceiptsAndSave is the DTO needed to provided input data to generate receipts
type ArgsGenerateReceiptsAndSave struct {
	// add data needed to create receipt structure
	// logs
}

type receiptsManager struct {
	trieInteractor     Interactor
	receiptsDataSyncer ReceiptsDataSyncer
}

// NewReceiptsManager will create a new instance of receipts manager
func NewReceiptsManager(args ArgsReceiptsManager) (*receiptsManager, error) {
	if args.TrieHandler == nil {
		return nil, ErrNilTrieInteractor
	}
	if args.ReceiptsDataSyncer == nil {
		return nil, ErrNilReceiptsDataSyncer
	}

	return &receiptsManager{
		trieInteractor:     args.TrieHandler,
		receiptsDataSyncer: args.ReceiptsDataSyncer,
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

func (rm *receiptsManager) SyncReceiptsTrie(receiptsRootHash []byte) error {
	receiptTrieBranchNodesBytes, err := rm.syncBranchNodesData(receiptsRootHash)
	if err != nil {
		return err
	}

	nodesMap, err := rm.trieInteractor.GetBranchNodesMap(receiptTrieBranchNodesBytes)
	if err != nil {
		return err
	}

	leafNodesHashes, err := trie.GetLeafHashesAndPutNodesInRamStorage(nodesMap, nil)
	if err != nil {
		return err
	}

	err = rm.syncLeafNodesAndPutInStorer(leafNodesHashes, nil)
	if err != nil {
		return err
	}

	return nil
}

func (rm *receiptsManager) syncBranchNodesData(receiptsRootHash []byte) ([]byte, error) {
	err := rm.receiptsDataSyncer.SyncReceiptsDataFor([][]byte{receiptsRootHash}, context.Background())
	if err != nil {
		return nil, err
	}
	receiptsDataMap, err := rm.receiptsDataSyncer.GetReceiptsData()
	if err != nil {
		return nil, err
	}
	rm.receiptsDataSyncer.ClearFields()

	receiptTrieBranchNodesBytes := receiptsDataMap[string(receiptsRootHash)]

	return receiptTrieBranchNodesBytes, nil
}

func (rm *receiptsManager) syncLeafNodesAndPutInStorer(hashes [][]byte, db storage.Storer) error {
	err := rm.receiptsDataSyncer.SyncReceiptsDataFor(hashes, context.Background())
	if err != nil {
		return err
	}
	leafNodesMap, err := rm.receiptsDataSyncer.GetReceiptsData()
	if err != nil {
		return err
	}
	rm.receiptsDataSyncer.ClearFields()

	for leafHash, leafBytes := range leafNodesMap {
		err = db.Put([]byte(leafHash), leafBytes)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rm *receiptsManager) IsInterfaceNil() bool {
	return rm == nil
}
