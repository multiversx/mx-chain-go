package receiptslog

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/state"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/trie"
)

// ArgsReceiptsManager is the structure that holds the components needed to a new receipts manager
type ArgsReceiptsManager struct {
	TrieHandler        Interactor
	ReceiptsDataSyncer ReceiptsDataSyncer
	Marshaller         marshal.Marshalizer
	Hasher             hashing.Hasher
}

// ArgsGenerateReceiptsAndSave is the DTO needed to provided input data to generate receipts
type ArgsGenerateReceiptsAndSave struct {
	// add data needed to create receipt structure
	// logs
}

type receiptsManager struct {
	trieInteractor     Interactor
	receiptsDataSyncer ReceiptsDataSyncer
	marshaller         marshal.Marshalizer
	hasher             hashing.Hasher
}

// NewReceiptsManager will create a new instance of receipts manager
func NewReceiptsManager(args ArgsReceiptsManager) (*receiptsManager, error) {
	if check.IfNil(args.TrieHandler) {
		return nil, ErrNilTrieInteractor
	}
	if check.IfNil(args.ReceiptsDataSyncer) {
		return nil, ErrNilReceiptsDataSyncer
	}

	return &receiptsManager{
		trieInteractor:     args.TrieHandler,
		receiptsDataSyncer: args.ReceiptsDataSyncer,
		marshaller:         args.Marshaller,
		hasher:             args.Hasher,
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
	nodesMap, err := rm.syncBranchNodesData(receiptsRootHash)
	if err != nil {
		return err
	}

	memoryDB := database.NewMemDB()
	leafNodesHashes, err := trie.GetLeafHashesAndPutNodesInRamStorage(nodesMap, memoryDB, rm.hasher, rm.marshaller)
	if err != nil {
		return err
	}

	err = rm.syncLeafNodesAndPutInStorer(leafNodesHashes, memoryDB)
	if err != nil {
		return err
	}

	newTrie, err := rm.trieInteractor.RecreateTrieFromDB(receiptsRootHash, memoryDB)
	if err != nil {
		return err
	}

	newTrieRootHash, err := rm.trieInteractor.SaveNewTrie(newTrie)
	if err != nil {
		return err
	}

	if !bytes.Equal(newTrieRootHash, receiptsRootHash) {
		return fmt.Errorf("%v , expected=%s, actual=%s", ErrReceiptTrieRootHashDoesNotMatch, hex.EncodeToString(receiptsRootHash), hex.EncodeToString(newTrieRootHash))
	}

	return nil
}

func (rm *receiptsManager) syncBranchNodesData(receiptsRootHash []byte) (map[string][]byte, error) {
	receiptsDataMap, err := rm.syncData([][]byte{receiptsRootHash})
	if err != nil {
		return nil, err
	}

	receiptTrieBranchNodesBytes := receiptsDataMap[string(receiptsRootHash)]

	serializedNodes := state.NewSerializedNodesMap()
	err = rm.marshaller.Unmarshal(serializedNodes, receiptTrieBranchNodesBytes)
	if err != nil {
		return nil, err
	}

	return serializedNodes.SerializedNodes, nil
}

func (rm *receiptsManager) syncData(hashes [][]byte) (map[string][]byte, error) {
	err := rm.receiptsDataSyncer.SyncReceiptsDataFor(hashes, context.Background())
	if err != nil {
		return nil, err
	}

	defer rm.receiptsDataSyncer.ClearFields()

	return rm.receiptsDataSyncer.GetReceiptsData()
}

func (rm *receiptsManager) syncLeafNodesAndPutInStorer(hashes [][]byte, db storage.Persister) error {
	leafNodesMap, err := rm.syncData(hashes)
	if err != nil {
		return err
	}

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
