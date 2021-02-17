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
	Hasher              hashing.Hasher
	Marshalizer         marshal.Marshalizer
	ShardID             uint32
	StorageConfig       config.StorageConfig
	TrieStorageManagers map[string]data.StorageManager
	HardforkStorer      update.HardforkStorer
}

type stateImport struct {
	genesisHeaders               map[uint32]data.HeaderHandler
	transactions                 map[string]data.TransactionHandler
	miniBlocks                   map[string]*block.MiniBlock
	importedEpochStartMetaBlock  *block.MetaBlock
	importedUnFinishedMetaBlocks map[string]*block.MetaBlock
	tries                        map[string]data.Trie
	accountDBsMap                map[uint32]state.AccountsDBImporter
	validatorDB                  state.AccountsDBImporter
	hardforkStorer               update.HardforkStorer

	hasher              hashing.Hasher
	marshalizer         marshal.Marshalizer
	shardID             uint32
	storageConfig       config.StorageConfig
	trieStorageManagers map[string]data.StorageManager
}

// NewStateImport creates an importer which reads all the files for a new start
func NewStateImport(args ArgsNewStateImport) (*stateImport, error) {
	if check.IfNil(args.Hasher) {
		return nil, update.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, update.ErrNilMarshalizer
	}
	if len(args.TrieStorageManagers) == 0 {
		return nil, update.ErrNilTrieStorageManagers
	}
	if check.IfNil(args.HardforkStorer) {
		return nil, update.ErrNilHardforkStorer
	}

	st := &stateImport{
		genesisHeaders:               make(map[uint32]data.HeaderHandler),
		transactions:                 make(map[string]data.TransactionHandler),
		miniBlocks:                   make(map[string]*block.MiniBlock),
		importedEpochStartMetaBlock:  &block.MetaBlock{},
		importedUnFinishedMetaBlocks: make(map[string]*block.MetaBlock),
		tries:                        make(map[string]data.Trie),
		hasher:                       args.Hasher,
		marshalizer:                  args.Marshalizer,
		accountDBsMap:                make(map[uint32]state.AccountsDBImporter),
		trieStorageManagers:          args.TrieStorageManagers,
		storageConfig:                args.StorageConfig,
		shardID:                      args.ShardID,
		hardforkStorer:               args.HardforkStorer,
	}

	return st, nil
}

// ImportAll imports all the relevant files for the new genesis
func (si *stateImport) ImportAll() error {
	var errFound error

	si.hardforkStorer.RangeKeys(func(identifier string, keys [][]byte) bool {
		var err error
		switch identifier {
		case EpochStartMetaBlockIdentifier:
			err = si.importEpochStartMetaBlock(identifier, keys)
		case UnFinishedMetaBlocksIdentifier:
			err = si.importUnFinishedMetaBlocks(identifier, keys)
		case MiniBlocksIdentifier:
			err = si.importMiniBlocks(identifier, keys)
		case TransactionsIdentifier:
			err = si.importTransactions(identifier, keys)
		default:
			splitString := strings.Split(identifier, atSep)
			canImportState := len(splitString) > 1 && splitString[0] == TrieIdentifier
			if !canImportState {
				return true
			}
			err = si.importState(identifier, keys)
		}
		if err != nil {
			errFound = err
			return false
		}

		return true
	})

	err := si.hardforkStorer.Close()
	if errFound != nil {
		return errFound
	}

	return err
}

func (si *stateImport) importEpochStartMetaBlock(identifier string, keys [][]byte) error {
	if len(keys) != 1 {
		return update.ErrExpectedOneStartOfEpochMetaBlock
	}
	object, err := si.createElement(identifier, string(keys[0]))
	if err != nil {
		return err
	}

	metaBlock, ok := object.(*block.MetaBlock)
	if !ok {
		return update.ErrWrongTypeAssertion
	}

	si.importedEpochStartMetaBlock = metaBlock

	return nil
}

func (si *stateImport) importUnFinishedMetaBlocks(identifier string, keys [][]byte) error {
	var err error
	var object interface{}
	for _, key := range keys {
		object, err = si.createElement(identifier, string(key))
		if err != nil {
			break
		}

		metaBlock, ok := object.(*block.MetaBlock)
		if !ok {
			return update.ErrWrongTypeAssertion
		}

		var hash []byte
		hash, err = core.CalculateHash(si.marshalizer, si.hasher, metaBlock)
		if err != nil {
			break
		}

		si.importedUnFinishedMetaBlocks[string(hash)] = metaBlock
	}

	if err != nil {
		return fmt.Errorf("%w identifier %s", err, UnFinishedMetaBlocksIdentifier)
	}

	return nil
}

func (si *stateImport) importTransactions(identifier string, keys [][]byte) error {
	var err error
	var object interface{}
	for _, key := range keys {
		object, err = si.createElement(identifier, string(key))
		if err != nil {
			break
		}

		tx, ok := object.(data.TransactionHandler)
		if !ok {
			err = fmt.Errorf("%w: wanted a transaction handler", update.ErrWrongTypeAssertion)
			break
		}

		var hash []byte
		hash, err = core.CalculateHash(si.marshalizer, si.hasher, tx)
		if err != nil {
			break
		}

		si.transactions[string(hash)] = tx
	}

	if err != nil {
		return fmt.Errorf("%w identifier %s", err, TransactionsIdentifier)
	}

	return nil
}

func (si *stateImport) createElement(identifier string, key string) (interface{}, error) {
	objType, _, err := GetKeyTypeAndHash(key)
	if err != nil {
		return nil, err
	}

	object, err := NewObject(objType)
	if err != nil {
		return nil, err
	}

	value, err := si.hardforkStorer.Get(identifier, []byte(key))
	if err != nil {
		return nil, fmt.Errorf("%w, key not found for %s, error: %s",
			update.ErrImportingData, hex.EncodeToString([]byte(key)), err.Error())
	}

	err = json.Unmarshal(value, object)
	if err != nil {
		return nil, err
	}

	return object, nil
}

func (si *stateImport) importMiniBlocks(identifier string, keys [][]byte) error {
	var err error
	var object interface{}
	for _, key := range keys {
		object, err = si.createElement(identifier, string(key))
		if err != nil {
			break
		}

		miniBlock, ok := object.(*block.MiniBlock)
		if !ok {
			err = fmt.Errorf("%w: wanted a miniblock", update.ErrWrongTypeAssertion)
			break
		}

		var hash []byte
		hash, err = core.CalculateHash(si.marshalizer, si.hasher, miniBlock)
		if err != nil {
			break
		}

		si.miniBlocks[string(hash)] = miniBlock
	}

	if err != nil {
		return fmt.Errorf("%w identifier %s", err, MiniBlocksIdentifier)
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

func (si *stateImport) importDataTrie(identifier string, shID uint32, keys [][]byte) error {
	var originalRootHash, address, value []byte
	var err error

	if len(keys) == 0 {
		return fmt.Errorf("%w missing original root hash", update.ErrImportingData)
	}

	originalRootHash, err = si.hardforkStorer.Get(identifier, keys[0])
	if err != nil {
		return err
	}

	keyType, _, err := GetKeyTypeAndHash(string(keys[0]))
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
		si.tries[identifier] = dataTrie

		return nil
	}

	for i := 1; i < len(keys); i++ {
		key := keys[i]
		value, err = si.hardforkStorer.Get(identifier, key)
		if err != nil {
			break
		}

		keyType, address, err = GetKeyTypeAndHash(string(key))
		if err != nil {
			break
		}
		if keyType != DataTrie {
			err = update.ErrKeyTypeMismatch
			break
		}

		err = dataTrie.Update(address, value)
		if err != nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("%w identifier: %s", err, identifier)
	}

	err = dataTrie.Commit()
	if err != nil {
		return err
	}
	si.tries[identifier] = dataTrie

	rootHash, err := dataTrie.RootHash()
	if err != nil {
		return err
	}

	if !bytes.Equal(rootHash, originalRootHash) {
		log.Warn("imported state rootHash does not match original ", "new", rootHash, "old", originalRootHash, "shardID", shID, "accType", DataTrie)
	}

	return nil
}

func (si *stateImport) getAccountsDB(accType Type, shardID uint32) (state.AccountsDBImporter, data.Trie, error) {
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
			accountsDB, errCreate := state.NewAccountsDB(currentTrie, si.hasher, si.marshalizer, accountFactory)
			if errCreate != nil {
				return nil, nil, errCreate
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

func (si *stateImport) importState(identifier string, keys [][]byte) error {
	accType, shId, err := GetTrieTypeAndShId(identifier)
	if err != nil {
		return err
	}

	// no need to import validator account trie
	if accType == ValidatorAccount {
		return nil
	}

	if accType == DataTrie {
		return si.importDataTrie(identifier, shId, keys)
	}

	accountsDB, mainTrie, err := si.getAccountsDB(accType, shId)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return fmt.Errorf("%w missing root hash", update.ErrImportingData)
	}

	// read root hash - that is the first saved in the file
	rootHash, err := si.hardforkStorer.Get(identifier, keys[0])
	if err != nil {
		return err
	}

	keyType, _, err := GetKeyTypeAndHash(string(keys[0]))
	if err != nil {
		return err
	}

	if keyType != RootHash {
		return fmt.Errorf("%w wanted a roothash, got %v", update.ErrWrongTypeAssertion, keyType)
	}

	log.Debug("importing state", "shard ID", shId, "root hash", rootHash)

	if len(rootHash) == 0 || bytes.Equal(rootHash, trie.EmptyTrieHash) {
		return si.saveRootHash(accountsDB, accType, shId, rootHash)
	}

	var marshaledData []byte
	var address []byte
	for i := 1; i < len(keys); i++ {
		key := keys[i]
		marshaledData, err = si.hardforkStorer.Get(identifier, key)
		if err != nil {
			break
		}

		keyType, address, err = GetKeyTypeAndHash(string(key))
		if err != nil {
			break
		}
		if keyType != accType {
			err = update.ErrKeyTypeMismatch
			break
		}

		err = si.unMarshalAndSaveAccount(accType, address, marshaledData, accountsDB, mainTrie)
		if err != nil {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("%w identifier: %s", err, identifier)
	}

	return si.saveRootHash(accountsDB, accType, shId, rootHash)
}

func (si *stateImport) unMarshalAndSaveAccount(
	accType Type,
	address, buffer []byte,
	accountsDB state.AccountsDBImporter,
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

	return accountsDB.ImportAccount(account)
}

func (si *stateImport) saveRootHash(
	accountsDB state.AccountsDBImporter,
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

	log.Debug("committed trie", "shard ID", shardID, "root hash", rootHash)

	return nil
}

// GetAccountsDBForShard returns the accounts DB for a specific shard
func (si *stateImport) GetAccountsDBForShard(shardID uint32) state.AccountsAdapter {
	adb, ok := si.accountDBsMap[shardID]
	if !ok {
		return nil
	}

	accountsAdapter, ok := adb.(state.AccountsAdapter)
	if !ok {
		return nil
	}

	return accountsAdapter
}

// GetTransactions returns all pending imported transactions
func (si *stateImport) GetTransactions() map[string]data.TransactionHandler {
	return si.transactions
}

// GetHardForkMetaBlock returns the hardFork metablock
func (si *stateImport) GetHardForkMetaBlock() *block.MetaBlock {
	return si.importedEpochStartMetaBlock
}

// GetUnFinishedMetaBlocks returns all imported unFinished metablocks
func (si *stateImport) GetUnFinishedMetaBlocks() map[string]*block.MetaBlock {
	return si.importedUnFinishedMetaBlocks
}

// GetMiniBlocks returns all imported pending miniblocks
func (si *stateImport) GetMiniBlocks() map[string]*block.MiniBlock {
	return si.miniBlocks
}

// GetValidatorAccountsDB returns the imported validator accounts DB
func (si *stateImport) GetValidatorAccountsDB() state.AccountsAdapter {
	accountsAdapter, ok := si.validatorDB.(state.AccountsAdapter)
	if !ok {
		return nil
	}

	return accountsAdapter
}

// IsInterfaceNil returns true if underlying object is nil
func (si *stateImport) IsInterfaceNil() bool {
	return si == nil
}
