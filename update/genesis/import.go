package genesis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	trieFactory "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/update"
)

// ArgsNewStateImport is the arguments structure to create a new state importer
type ArgsNewStateImport struct {
	Reader              update.MultiFileReader
	Hasher              hashing.Hasher
	Marshalizer         marshal.Marshalizer
	TriesContainer      state.TriesHolder
	TrieStorageManagers map[string]data.StorageManager
	ShardID             uint32
	StorageConfig       config.StorageConfig
}

type stateImport struct {
	reader             update.MultiFileReader
	genesisHeaders     map[uint32]data.HeaderHandler
	transactions       map[string]data.TransactionHandler
	miniBlocks         map[string]*block.MiniBlock
	importedMetaBlock  *block.MetaBlock
	tries              map[uint32]data.Trie
	accountDBsMap      map[uint32]state.AccountsAdapter
	validatorDB        state.AccountsAdapter
	newRootHashes      map[uint32][]byte
	validatorsRootHash []byte

	hasher              hashing.Hasher
	marshalizer         marshal.Marshalizer
	storage             dataRetriever.StorageService
	triesContainer      state.TriesHolder
	trieStorageManagers map[string]data.StorageManager
	shardID             uint32
	storageConfig       config.StorageConfig
}

// TODO: think about the state of validators - epochs.
// NewStateImport creates an importer which reads all the files for a new start
func NewStateImport(args ArgsNewStateImport) (*stateImport, error) {
	if check.IfNil(args.Reader) {
		return nil, update.ErrNilMultiFileReader
	}
	if check.IfNil(args.Hasher) {
		return nil, update.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, update.ErrNilMarshalizer
	}
	if check.IfNil(args.TriesContainer) {
		return nil, update.ErrNilStorage
	}

	st := &stateImport{
		reader:            args.Reader,
		genesisHeaders:    make(map[uint32]data.HeaderHandler),
		transactions:      make(map[string]data.TransactionHandler),
		miniBlocks:        make(map[string]*block.MiniBlock),
		importedMetaBlock: &block.MetaBlock{},
		tries:             make(map[uint32]data.Trie),
		newRootHashes:     make(map[uint32][]byte),
		hasher:            args.Hasher,
		marshalizer:       args.Marshalizer,
		accountDBsMap:     make(map[uint32]state.AccountsAdapter),
	}

	return st, nil
}

// ImportAll imports all the relevant files for the new genesis
func (si *stateImport) ImportAll() error {
	files := si.reader.GetFileNames()
	if len(files) == 0 {
		return update.ErrNoFileToImport
	}

	var err error
	for _, fileName := range files {
		switch fileName {
		case MetaBlockFileName:
			err = si.importMetaBlock()
		case MiniBlocksFileName:
			err = si.importMiniBlocks()
		case TransactionsFileName:
			err = si.importTransactions()
		default:
			splitString := strings.Split(fileName, atSep)
			canImportState := len(splitString) > 1 && splitString[0] == TrieFileName
			if !canImportState {
				continue
			}
			err = si.importState(splitString[0], splitString[1])
		}
		if err != nil {
			return err
		}
	}

	si.reader.Finish()

	return nil
}

func (si *stateImport) importMetaBlock() error {
	object, err := si.readNextElement(MetaBlockFileName)
	if err != nil {
		return nil
	}

	metaBlock, ok := object.(*block.MetaBlock)
	if !ok {
		return update.ErrWrongTypeAssertion
	}

	si.importedMetaBlock = metaBlock

	return nil
}

func (si *stateImport) importTransactions() error {
	var err error
	var object interface{}
	for {
		object, err = si.readNextElement(TransactionsFileName)
		if err != nil {
			break
		}

		tx, ok := object.(data.TransactionHandler)
		if !ok {
			err = fmt.Errorf("%w wanted a transaction handler", update.ErrWrongTypeAssertion)
			break
		}

		hash, err := core.CalculateHash(si.marshalizer, si.hasher, tx)
		if err != nil {
			break
		}

		si.transactions[string(hash)] = tx
	}

	if err != update.ErrEndOfFile {
		return fmt.Errorf("%w fileName %s", err, TransactionsFileName)
	}

	return nil
}

func (si *stateImport) readNextElement(fileName string) (interface{}, error) {
	key, value, err := si.reader.ReadNextItem(fileName)
	if err != nil {
		return nil, err
	}

	objType, readHash, err := GetKeyTypeAndHash(key)
	if err != nil {
		return nil, err
	}

	object, err := NewObject(objType)
	if err != nil {
		return nil, err
	}

	hash := si.hasher.Compute(string(value))
	if !bytes.Equal(readHash, hash) {
		return nil, update.ErrHashMissmatch
	}

	err = json.Unmarshal(value, object)
	if err != nil {
		return nil, err
	}

	return object, nil
}

func (si *stateImport) importMiniBlocks() error {
	var err error
	var object interface{}
	for {
		object, err = si.readNextElement(MiniBlocksFileName)
		if err != nil {
			break
		}

		miniBlock, ok := object.(*block.MiniBlock)
		if !ok {
			err = fmt.Errorf("%w wanted a miniblock", update.ErrWrongTypeAssertion)
			break
		}

		hash, err := core.CalculateHash(si.marshalizer, si.hasher, miniBlock)
		if err != nil {
			break
		}

		si.miniBlocks[string(hash)] = miniBlock
	}

	if err != update.ErrEndOfFile {
		return fmt.Errorf("%w fileName %s", err, MiniBlocksFileName)
	}

	return nil
}

func newAccountCreator(accType Type) (state.AccountFactory, error) {
	switch accType {
	case UserAccount:
		return factory.NewAccountCreator(), nil
	case ValidatorAccount:
		return factory.NewPeerAccountCreator(), nil
	}
	return nil, update.ErrUnknownType
}

func getTrieContainerID(accType Type) []byte {
	switch accType {
	case UserAccount:
		return []byte(trieFactory.UserAccountTrie)
	case ValidatorAccount:
		return []byte(trieFactory.PeerAccountTrie)
	}

	return nil
}

func (si *stateImport) getTrie(shardID uint32, accType Type) (data.Trie, error) {
	if shardID == si.shardID {
		accountsDBType := getTrieContainerID(accType)
		trieForShard := si.triesContainer.Get(accountsDBType)
		return trieForShard, nil
	}

	trieForShard, ok := si.tries[shardID]
	if ok {
		return trieForShard, nil
	}

	dbConfig := storageFactory.GetDBFromConfig(si.storageConfig.DB)
	dbConfig.FilePath = path.Join(si.storageConfig.DB.FilePath, core.GetShardIdString(shardID))
	storageForShard, err := storageUnit.NewStorageUnitFromConf(
		storageFactory.GetCacherFromConfig(si.storageConfig.Cache),
		dbConfig,
		storageFactory.GetBloomFromConfig(si.storageConfig.Bloom),
	)
	if err != nil {
		return nil, err
	}

	trieStorage, err := trie.NewTrieStorageManagerWithoutPruning(storageForShard)
	if err != nil {
		return nil, err
	}

	trieForShard, err = trie.NewTrie(trieStorage, si.marshalizer, si.hasher)
	if err != nil {
		return nil, err
	}

	si.tries[shardID] = trieForShard

	return trieForShard, nil
}

func (si *stateImport) importState(fileName string, trieKey string) error {
	accType, shId, err := GetTrieTypeAndShId(trieKey)
	if err != nil {
		return err
	}

	accountFactory, err := newAccountCreator(accType)
	if err != nil {
		return err
	}

	currentTrie, err := si.getTrie(shId, accType)
	if err != nil {
		return err
	}

	accountsDB, err := state.NewAccountsDB(currentTrie, si.hasher, si.marshalizer, accountFactory)
	if err != nil {
		return err
	}

	// read root hash - that is the first saved in the file
	key, value, err := si.reader.ReadNextItem(fileName)
	if err != nil {
		return err
	}

	keyType, _, err := GetKeyTypeAndHash(key)
	if err != nil {
		return err
	}

	if keyType != RootHash {
		return fmt.Errorf("%w wanted a roothash", update.ErrWrongTypeAssertion)
	}

	oldRootHash := value
	log.Debug("old root hash", "value", oldRootHash)

	var address []byte
	var account state.AccountHandler
	for {
		key, value, err = si.reader.ReadNextItem(fileName)
		if err != nil {
			break
		}

		_, address, err = GetKeyTypeAndHash(key)
		if err != nil {
			break
		}

		account, err = NewEmptyAccount(accType)
		if err != nil {
			break
		}

		err = json.Unmarshal(value, account)
		if err != nil {
			break
		}

		if !bytes.Equal(account.AddressContainer().Bytes(), address) {
			return update.ErrHashMissmatch
		}

		err = accountsDB.SaveAccount(account)
		if err != nil {
			break
		}
	}

	if err != update.ErrEndOfFile {
		return fmt.Errorf("%w fileName: %s", err, fileName)
	}

	return si.saveRootHash(accountsDB, accType, shId)
}

func (si *stateImport) saveRootHash(accountsDB state.AccountsAdapter, accType Type, shardID uint32) error {
	rootHash, err := accountsDB.Commit()
	if err != nil {
		return err
	}

	if si.shardID == shardID {
		accountsDB.SnapshotState(rootHash)
	}

	if accType == ValidatorAccount {
		si.validatorsRootHash = rootHash
		si.validatorDB = accountsDB
		return nil
	}

	si.accountDBsMap[shardID] = accountsDB
	si.newRootHashes[shardID] = rootHash

	return nil
}

// GetAccountsDBForShard returns the accounts DB for a specific shard
func (si *stateImport) GetAccountsDBForShard(shardID uint32) state.AccountsAdapter {
	return si.accountDBsMap[shardID]
}

// GetTransactions returns all pending imported transactions
func (si *stateImport) GetTransactions() map[string]data.TransactionHandler {
	return si.transactions
}

// GetHardForkMetaBlock returns the hardFork metablock
func (si *stateImport) GetHardForkMetaBlock() *block.MetaBlock {
	return si.importedMetaBlock
}

// GetMiniBlocks returns all imported pending miniblocks
func (si *stateImport) GetMiniBlocks() map[string]*block.MiniBlock {
	return si.miniBlocks
}

// GetValidatorAccountsDB returns the imported validator accounts DB
func (si *stateImport) GetValidatorAccountsDB() state.AccountsAdapter {
	return si.validatorDB
}

// IsInterfaceNil returns true if underlying object is nil
func (si *stateImport) IsInterfaceNil() bool {
	return si == nil
}
