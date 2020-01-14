package genesis

import (
	"bytes"
	"encoding/json"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/update"
)

type ArgsNewStateImport struct {
}

type stateImport struct {
	reader            update.MultiFileReader
	genesisHeaders    map[uint32]data.HeaderHandler
	transactions      map[string]data.TransactionHandler
	miniBlocks        map[string]*block.MiniBlock
	importedMetaBlock *block.MetaBlock
	tries             map[string]data.Trie

	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
}

func NewStateImport(args ArgsNewStateImport) (*stateImport, error) {
	return nil, nil
}

func (si *stateImport) ImportAll() error {
	files := si.reader.GetFileNames()
	if len(files) == 0 {
		return update.ErrNoFileToImport
	}

	var err error
	for _, fileName := range files {
		switch fileName {
		case update.MetaBlockFileName:
			err = si.importMetaBlock()
		case update.MiniBlocksFileName:
			err = si.importMiniBlocks()
		case update.TransactionsFileName:
			err = si.importTransactions()
		default:
			err = si.importTrie(fileName)
		}
		if err != nil {
			return err
		}
	}

	err = si.processTransactions()
	if err != nil {
		return err
	}

	return nil
}

func (si *stateImport) importMetaBlock() error {

	object, err := si.readNextElement(update.MetaBlockFileName)
	if err != nil {
		return nil
	}

	metaBlock, ok := object.(*block.MetaBlock)
	if !ok {
		return core.ErrWrongTypeAssertion
	}

	si.importedMetaBlock = metaBlock

	return nil
}

func (si *stateImport) importTransactions() error {
	var err error
	for {
		object, err := si.readNextElement(update.TransactionsFileName)
		if err != nil {
			break
		}

		tx, ok := object.(data.TransactionHandler)
		if !ok {
			err = core.ErrWrongTypeAssertion
			break
		}

		hash, err := core.CalculateHash(si.marshalizer, si.hasher, tx)
		if err != nil {
			break
		}

		si.transactions[string(hash)] = tx
	}

	if err != update.ErrEndOfFile {
		return err
	}

	return nil
}

func (si *stateImport) readNextElement(fileName string) (interface{}, error) {
	key, value, err := si.reader.ReadNextItem(fileName)
	if err != nil {
		return nil, err
	}

	objType, readHash, err := update.GetKeyTypeAndHash(key)
	if err != nil {
		return nil, err
	}

	object, err := update.NewObject(objType)
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
	for {
		object, err := si.readNextElement(update.MiniBlocksFileName)
		if err != nil {
			break
		}

		miniBlock, ok := object.(*block.MiniBlock)
		if !ok {
			err = core.ErrWrongTypeAssertion
			break
		}

		hash, err := core.CalculateHash(si.marshalizer, si.hasher, miniBlock)
		if err != nil {
			break
		}

		si.miniBlocks[string(hash)] = miniBlock
	}

	if err != update.ErrEndOfFile {
		return err
	}

	return nil
}

func (si *stateImport) importTrie(fileName string) error {
	accType, _, err := update.GetTrieTypeAndShId(fileName)
	if err != nil {
		return err
	}

	accountFactory, err := factory.NewAccountFactoryCreator(accType)
	if err != nil {
		return err
	}

	accountsDB, err := state.NewAccountsDB(si.tries[fileName], si.hasher, si.marshalizer, accountFactory)
	if err != nil {
		return err
	}

	// read root hash - that is the first saved in the file
	key, value, err := si.reader.ReadNextItem(fileName)
	if err != nil {
		return err
	}

	keyType, _, err := update.GetKeyTypeAndHash(key)
	if err != nil {
		return err
	}

	if keyType != update.RootHash {
		return core.ErrWrongTypeAssertion
	}

	oldRootHash := value
	log.Debug("old root hash", "value", oldRootHash)

	for {
		key, value, err := si.reader.ReadNextItem(fileName)
		if err != nil {
			break
		}

		_, address, err := update.GetKeyTypeAndHash(key)
		if err != nil {
			break
		}

		account, err := update.NewEmptyAccount(accType)
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
		return err
	}

	return nil
}

func (si *stateImport) processTransactions() error {
	return nil
}

func (si *stateImport) GetAllGenesisBlocks() map[uint32]data.HeaderHandler {
	return si.genesisHeaders
}

func (si *stateImport) ImportDataFor(shId uint32) error {

	return nil
}

func (si *stateImport) IsInterfaceNil() bool {
	return si == nil
}
