package genesis

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	commonDisabled "github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/disabled"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/update"
)

var _ update.ImportHandler = (*stateImport)(nil)

const maxTrieLevelInMemory = uint(5)

// ArgsNewStateImport is the arguments structure to create a new state importer
type ArgsNewStateImport struct {
	Hasher              hashing.Hasher
	Marshalizer         marshal.Marshalizer
	ShardID             uint32
	StorageConfig       config.StorageConfig
	TrieStorageManagers map[string]common.StorageManager
	HardforkStorer      update.HardforkStorer
	AddressConverter    core.PubkeyConverter
	EnableEpochsHandler common.EnableEpochsHandler
}

type stateImport struct {
	genesisHeaders               map[uint32]data.HeaderHandler
	transactions                 map[string]data.TransactionHandler
	miniBlocks                   map[string]*block.MiniBlock
	importedEpochStartMetaBlock  data.MetaHeaderHandler
	importedUnFinishedMetaBlocks map[string]data.MetaHeaderHandler
	tries                        map[string]common.Trie
	accountDBsMap                map[uint32]state.AccountsDBImporter
	validatorDB                  state.AccountsDBImporter
	hardforkStorer               update.HardforkStorer

	hasher              hashing.Hasher
	marshalizer         marshal.Marshalizer
	shardID             uint32
	storageConfig       config.StorageConfig
	trieStorageManagers map[string]common.StorageManager
	addressConverter    core.PubkeyConverter
	enableEpochsHandler common.EnableEpochsHandler
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
	if check.IfNil(args.AddressConverter) {
		return nil, update.ErrNilAddressConverter
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, errors.ErrNilEnableEpochsHandler
	}

	st := &stateImport{
		genesisHeaders:               make(map[uint32]data.HeaderHandler),
		transactions:                 make(map[string]data.TransactionHandler),
		miniBlocks:                   make(map[string]*block.MiniBlock),
		importedEpochStartMetaBlock:  &block.MetaBlock{},
		importedUnFinishedMetaBlocks: make(map[string]data.MetaHeaderHandler),
		tries:                        make(map[string]common.Trie),
		hasher:                       args.Hasher,
		marshalizer:                  args.Marshalizer,
		accountDBsMap:                make(map[uint32]state.AccountsDBImporter),
		trieStorageManagers:          args.TrieStorageManagers,
		storageConfig:                args.StorageConfig,
		shardID:                      args.ShardID,
		hardforkStorer:               args.HardforkStorer,
		addressConverter:             args.AddressConverter,
		enableEpochsHandler:          args.EnableEpochsHandler,
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

func newAccountCreator(
	accType Type,
	hasher hashing.Hasher,
	marshaller marshal.Marshalizer,
	handler common.EnableEpochsHandler,
) (state.AccountFactory, error) {
	switch accType {
	case UserAccount:
		args := factory.ArgsAccountCreator{
			Hasher:              hasher,
			Marshaller:          marshaller,
			EnableEpochsHandler: handler,
		}
		return factory.NewAccountCreator(args)
	case ValidatorAccount:
		return factory.NewPeerAccountCreator(), nil
	}
	return nil, update.ErrUnknownType
}

func (si *stateImport) getTrie(shardID uint32, accType Type) (common.Trie, error) {
	trieString := core.ShardIdToString(shardID)
	if accType == ValidatorAccount {
		trieString = "validator"
	}

	trieForShard, ok := si.tries[trieString]
	if ok {
		return trieForShard, nil
	}

	trieStorageManager := si.trieStorageManagers[dataRetriever.UserAccountsUnit.String()]
	if accType == ValidatorAccount {
		trieStorageManager = si.trieStorageManagers[dataRetriever.PeerAccountsUnit.String()]
	}

	trieForShard, err := trie.NewTrie(trieStorageManager, si.marshalizer, si.hasher, si.enableEpochsHandler, maxTrieLevelInMemory)
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

	dataTrie, err := trie.NewTrie(si.trieStorageManagers[dataRetriever.UserAccountsUnit.String()], si.marshalizer, si.hasher, si.enableEpochsHandler, maxTrieLevelInMemory)
	if err != nil {
		return err
	}

	if common.IsEmptyTrie(originalRootHash) {
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
		// TODO this will not work for a partially migrated trie
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

func (si *stateImport) getAccountsDB(accType Type, shardID uint32, accountFactory state.AccountFactory) (state.AccountsDBImporter, common.Trie, error) {
	currentTrie, err := si.getTrie(shardID, accType)
	if err != nil {
		return nil, nil, err
	}

	if accType == ValidatorAccount {
		if check.IfNil(si.validatorDB) {
			argsAccountDB := state.ArgsAccountsDB{
				Trie:                  currentTrie,
				Hasher:                si.hasher,
				Marshaller:            si.marshalizer,
				AccountFactory:        accountFactory,
				StoragePruningManager: disabled.NewDisabledStoragePruningManager(),
				ProcessingMode:        common.Normal,
				ProcessStatusHandler:  commonDisabled.NewProcessStatusHandler(),
				AppStatusHandler:      commonDisabled.NewAppStatusHandler(),
				AddressConverter:      si.addressConverter,
				EnableEpochsHandler:   si.enableEpochsHandler,
			}
			accountsDB, errCreate := state.NewAccountsDB(argsAccountDB)
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

	argsAccountDB := state.ArgsAccountsDB{
		Trie:                  currentTrie,
		Hasher:                si.hasher,
		Marshaller:            si.marshalizer,
		AccountFactory:        accountFactory,
		StoragePruningManager: disabled.NewDisabledStoragePruningManager(),
		ProcessingMode:        common.Normal,
		ProcessStatusHandler:  commonDisabled.NewProcessStatusHandler(),
		AppStatusHandler:      commonDisabled.NewAppStatusHandler(),
		AddressConverter:      si.addressConverter,
		EnableEpochsHandler:   si.enableEpochsHandler,
	}
	accountsDB, err = state.NewAccountsDB(argsAccountDB)
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

	accountFactory, err := newAccountCreator(accType, si.hasher, si.marshalizer, si.enableEpochsHandler)
	if err != nil {
		return err
	}

	accountsDB, mainTrie, err := si.getAccountsDB(accType, shId, accountFactory)
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

	if common.IsEmptyTrie(rootHash) {
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

		err = si.unMarshalAndSaveAccount(address, accountFactory, marshaledData, accountsDB, mainTrie)
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
	address []byte,
	accountCreator state.AccountFactory,
	buffer []byte,
	accountsDB state.AccountsDBImporter,
	mainTrie common.Trie,
) error {
	account, err := accountCreator.CreateAccount(address)
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
func (si *stateImport) GetHardForkMetaBlock() data.MetaHeaderHandler {
	return si.importedEpochStartMetaBlock
}

// GetUnFinishedMetaBlocks returns all imported unFinished metablocks
func (si *stateImport) GetUnFinishedMetaBlocks() map[string]data.MetaHeaderHandler {
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

// Close tries to close state import objects
func (si *stateImport) Close() error {
	return si.hardforkStorer.Close()
}

// IsInterfaceNil returns true if underlying object is nil
func (si *stateImport) IsInterfaceNil() bool {
	return si == nil
}
