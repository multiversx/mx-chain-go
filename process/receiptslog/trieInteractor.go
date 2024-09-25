package receiptslog

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/state"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/trie"
)

const maxTrieLevelInMemory = 5

// ArgsTrieInteractor is the structure that holds the components needed to a new  trie interactor
type ArgsTrieInteractor struct {
	ReceiptDataStorer   storage.Storer
	Marshaller          marshal.Marshalizer
	Hasher              hashing.Hasher
	EnableEpochsHandler common.EnableEpochsHandler
}

type trieInteractor struct {
	marshaller          marshal.Marshalizer
	hasher              hashing.Hasher
	enableEpochsHandler common.EnableEpochsHandler
	storage             storage.Storer

	localTrie common.Trie
}

// NewTrieInteractor will create a new instance of trie interactor
func NewTrieInteractor(args ArgsTrieInteractor) (*trieInteractor, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &trieInteractor{
		marshaller:          args.Marshaller,
		hasher:              args.Hasher,
		enableEpochsHandler: args.EnableEpochsHandler,
		storage:             args.ReceiptDataStorer,
	}, nil
}

// CreateNewTrie will create a new local trie(also will overwrite the old local trie)
func (ti *trieInteractor) CreateNewTrie() error {
	disabledStorageManager := &storageManager.StorageManagerStub{}

	localTrie, err := trie.NewTrie(disabledStorageManager, ti.marshaller, ti.hasher, ti.enableEpochsHandler, maxTrieLevelInMemory)
	if err != nil {
		return err
	}

	ti.localTrie = localTrie

	return nil
}

// AddReceiptData will add receipt data in local trie
func (ti *trieInteractor) AddReceiptData(receiptData state.Receipt) error {
	receiptDataBytes, err := ti.marshaller.Marshal(receiptData)
	if err != nil {
		return err
	}

	return ti.localTrie.Update(receiptData.TxHash, receiptDataBytes)
}

// Save will save all data from trie in storage and return the receipts root hash
func (ti *trieInteractor) Save() ([]byte, error) {
	dfsIterator, err := trie.NewDFSIterator(ti.localTrie)
	if err != nil {
		return nil, err
	}

	currentNodeData, errGet := dfsIterator.GetCurrentNodeInfo()
	if errGet != nil {
		return nil, errGet
	}

	serializedNodes := make([][]byte, 0)
	err = ti.saveNodeData(currentNodeData, serializedNodes)
	if err != nil {
		return nil, err
	}

	for dfsIterator.HasNext() {
		err = dfsIterator.Next()
		if err != nil {
			return nil, err
		}

		currentNodeData, errGet = dfsIterator.GetCurrentNodeInfo()
		if errGet != nil {
			return nil, errGet
		}

		err = ti.saveNodeData(currentNodeData, serializedNodes)
		if err != nil {
			return nil, err
		}

	}

	listOfSerializedNodesBytes, err := ti.marshaller.Marshal(&serializedNodes)
	if err != nil {
		return nil, err
	}

	receiptTrieRootHash, err := ti.localTrie.RootHash()
	if err != nil {
		return nil, err
	}

	err = ti.storage.Put(receiptTrieRootHash, listOfSerializedNodesBytes)
	if err != nil {
		return nil, err
	}

	return receiptTrieRootHash, nil
}

func (ti *trieInteractor) saveReceiptTxHashLeafKey(leafHash []byte, leafData []byte) error {
	receiptData := &state.Receipt{}
	err := ti.marshaller.Unmarshal(receiptData, leafData)
	if err != nil {
		return err
	}

	return ti.storage.Put(receiptData.TxHash, leafHash)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ti *trieInteractor) IsInterfaceNil() bool {
	return ti == nil
}

func checkArgs(args ArgsTrieInteractor) error {
	if check.IfNil(args.EnableEpochsHandler) {
		return process.ErrNilEnableEpochsHandler
	}
	if check.IfNil(args.Hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(args.Marshaller) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(args.ReceiptDataStorer) {
		return dataRetriever.ErrNilReceiptsStorage
	}

	return nil
}

func (ti *trieInteractor) saveNodeData(currentNodeData *trie.CurrentNodeInfo, serializedNodes [][]byte) error {
	if currentNodeData.Type != trie.LeafNodeType {
		serializedNodes = append(serializedNodes, currentNodeData.SerializedNode)
		return nil
	}

	err := ti.storage.Put(currentNodeData.Hash, currentNodeData.SerializedNode)
	if err != nil {
		return err
	}

	return ti.saveReceiptTxHashLeafKey(currentNodeData.Hash, currentNodeData.Value)
}
