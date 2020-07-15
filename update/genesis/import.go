package genesis

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	triesFactory "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/update"
)

var _ update.ImportHandler = (*stateImport)(nil)

const maxTrieLevelInMemory = uint(5)

// ArgsNewStateImport is the arguments structure to create a new state importer
type ArgsNewStateImport struct {
	Reader              update.MultiFileReader
	Hasher              hashing.Hasher
	Marshalizer         marshal.Marshalizer
	ShardID             uint32
	StorageConfig       config.StorageConfig
	TrieStorageManagers map[string]data.StorageManager
}

type stateImport struct {
	reader            update.MultiFileReader
	genesisHeaders    map[uint32]data.HeaderHandler
	transactions      map[string]data.TransactionHandler
	miniBlocks        map[string]*block.MiniBlock
	importedMetaBlock *block.MetaBlock
	tries             map[string]data.Trie
	accountDBsMap     map[uint32]state.AccountsAdapter
	validatorDB       state.AccountsAdapter

	hasher              hashing.Hasher
	marshalizer         marshal.Marshalizer
	shardID             uint32
	storageConfig       config.StorageConfig
	trieStorageManagers map[string]data.StorageManager
}

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
	if len(args.TrieStorageManagers) == 0 {
		return nil, update.ErrNilTrieStorageManagers
	}

	st := &stateImport{
		reader:              args.Reader,
		genesisHeaders:      make(map[uint32]data.HeaderHandler),
		transactions:        make(map[string]data.TransactionHandler),
		miniBlocks:          make(map[string]*block.MiniBlock),
		importedMetaBlock:   &block.MetaBlock{},
		tries:               make(map[string]data.Trie),
		hasher:              args.Hasher,
		marshalizer:         args.Marshalizer,
		accountDBsMap:       make(map[uint32]state.AccountsAdapter),
		trieStorageManagers: args.TrieStorageManagers,
		storageConfig:       args.StorageConfig,
		shardID:             args.ShardID,
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
			err = si.importState(fileName)
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
		return err
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

	objType, _, err := GetKeyTypeAndHash(key)
	if err != nil {
		return nil, err
	}

	object, err := NewObject(objType)
	if err != nil {
		return nil, err
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

func (si *stateImport) getTrie(shardID uint32, accType Type) (data.Trie, error) {
	trieString := core.ShardIdToString(shardID)
	if accType == ValidatorAccount {
		trieString = "validator"
	}

	trieForShard, ok := si.tries[trieString]
	if ok {
		return trieForShard, nil
	}

	trieStorageManager := si.trieStorageManagers[triesFactory.UserAccountTrie]
	if accType == ValidatorAccount {
		trieStorageManager = si.trieStorageManagers[triesFactory.PeerAccountTrie]
	}

	trieForShard, err := trie.NewTrie(trieStorageManager, si.marshalizer, si.hasher, maxTrieLevelInMemory)
	if err != nil {
		return nil, err
	}

	si.tries[trieString] = trieForShard

	return trieForShard, nil
}

func (si *stateImport) importDataTrie(fileName string, shID uint32) error {
	var originalRootHash, address, value []byte
	var key string
	var err error

	// read root hash - that is the first saved in the file
	key, originalRootHash, err = si.reader.ReadNextItem(fileName)
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

	dataTrie, err := trie.NewTrie(si.trieStorageManagers[triesFactory.UserAccountTrie], si.marshalizer, si.hasher, maxTrieLevelInMemory)
	if err != nil {
		return err
	}

	if len(originalRootHash) == 0 || bytes.Equal(originalRootHash, trie.EmptyTrieHash) {
		err = dataTrie.Commit()
		if err != nil {
			return err
		}
		si.tries[fileName] = dataTrie
		si.reader.CloseFile(fileName)
		return nil
	}

	for {
		key, value, err = si.reader.ReadNextItem(fileName)
		if err != nil {
			break
		}

		keyType, address, err = GetKeyTypeAndHash(key)
		if err != nil {
			break
		}
		if keyType != DataTrie {
			err = update.ErrKeyTypeMissMatch
			break
		}

		err = dataTrie.Update(address, value)
		if err != nil {
			break
		}
	}
	if err != update.ErrEndOfFile {
		return fmt.Errorf("%w fileName: %s", err, fileName)
	}

	err = dataTrie.Commit()
	if err != nil {
		return err
	}
	si.tries[fileName] = dataTrie
	si.reader.CloseFile(fileName)

	rootHash, err := dataTrie.Root()
	if err != nil {
		return err
	}

	if !bytes.Equal(rootHash, originalRootHash) {
		log.Warn("imported state rootHash does not match original ", "new", rootHash, "old", originalRootHash, "shardID", shID, "accType", DataTrie)
	}

	return nil
}

func (si *stateImport) getAccountsDB(accType Type, shardID uint32) (state.AccountsAdapter, data.Trie, error) {
	accountFactory, err := newAccountCreator(accType)
	if err != nil {
		return nil, nil, err
	}

	currentTrie, err := si.getTrie(shardID, accType)
	if err != nil {
		return nil, nil, err
	}

	if accType == ValidatorAccount {
		if check.IfNil(si.validatorDB) {
			accountsDB, err := state.NewAccountsDB(currentTrie, si.hasher, si.marshalizer, accountFactory)
			if err != nil {
				return nil, nil, err
			}
			si.validatorDB = accountsDB
		}
		return si.validatorDB, currentTrie, err
	}

	accountsDB, ok := si.accountDBsMap[shardID]
	if ok {
		return accountsDB, currentTrie, nil
	}

	accountsDB, err = state.NewAccountsDB(currentTrie, si.hasher, si.marshalizer, accountFactory)
	si.accountDBsMap[shardID] = accountsDB
	return accountsDB, currentTrie, err
}

func (si *stateImport) importState(fileName string) error {
	accType, shId, err := GetTrieTypeAndShId(fileName)
	if err != nil {
		return err
	}

	// no need to import validator account trie
	if accType == ValidatorAccount {
		return nil
	}

	if accType == DataTrie {
		return si.importDataTrie(fileName, shId)
	}

	accountsDB, mainTrie, err := si.getAccountsDB(accType, shId)
	if err != nil {
		return err
	}

	// read root hash - that is the first saved in the file
	key, rootHash, err := si.reader.ReadNextItem(fileName)
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

	if len(rootHash) == 0 || bytes.Equal(rootHash, trie.EmptyTrieHash) {
		return si.saveRootHash(accountsDB, accType, shId, rootHash)
	}

	var marshalledData []byte
	var address []byte
	for {
		key, marshalledData, err = si.reader.ReadNextItem(fileName)
		if err != nil {
			break
		}

		keyType, address, err = GetKeyTypeAndHash(key)
		if err != nil {
			break
		}
		if keyType != accType {
			err = update.ErrKeyTypeMissMatch
			break
		}

		err = mainTrie.Update(address, marshalledData)
		if err != nil {
			break
		}
	}

	if err != update.ErrEndOfFile {
		return fmt.Errorf("%w fileName: %s", err, fileName)
	}

	si.reader.CloseFile(fileName)
	return si.saveRootHash(accountsDB, accType, shId, rootHash)
}

func (si *stateImport) unMarshalAndSaveAccount(
	accType Type,
	address, buffer []byte,
	accountsDB state.AccountsAdapter,
	mainTrie data.Trie,
) error {
	account, err := NewEmptyAccount(accType, address)
	if err != nil {
		return err
	}

	err = si.marshalizer.Unmarshal(account, buffer)
	if err != nil {
		log.Trace("error unmarshaling account this is maybe a code error",
			"key", hex.EncodeToString(address),
			"error", err,
		)
		err = mainTrie.Update(address, buffer)
		return err
	}

	return accountsDB.SaveAccount(account)
}

func (si *stateImport) saveRootHash(
	accountsDB state.AccountsAdapter,
	accType Type,
	shardID uint32,
	originalRootHash []byte,
) error {
	rootHash, err := accountsDB.Commit()
	if err != nil {
		return err
	}

	if !bytes.Equal(rootHash, originalRootHash) {
		log.Warn("imported state rootHash does not match original ", "new", rootHash, "old", originalRootHash, "accType", accType, "shardID", shardID)
	}

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
