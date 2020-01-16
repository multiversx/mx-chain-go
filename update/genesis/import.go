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
		case MetaBlockFileName:
			err = si.importMetaBlock()
		case MiniBlocksFileName:
			err = si.importMiniBlocks()
		case TransactionsFileName:
			err = si.importTransactions()
		default:
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
	var object interface{}
	for {
		object, err = si.readNextElement(TransactionsFileName)
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

func (si *stateImport) importState(fileName string) error {
	accType, _, err := GetTrieTypeAndShId(fileName)
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

	keyType, _, err := GetKeyTypeAndHash(key)
	if err != nil {
		return err
	}

	if keyType != RootHash {
		return core.ErrWrongTypeAssertion
	}

	oldRootHash := value
	log.Debug("old root hash", "value", oldRootHash)

	for {
		key, value, err := si.reader.ReadNextItem(fileName)
		if err != nil {
			break
		}

		_, address, err := GetKeyTypeAndHash(key)
		if err != nil {
			break
		}

		account, err := NewEmptyAccount(accType)
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

func (si *stateImport) ProcessTransactions() error {
	return nil
}

func (si *stateImport) CreateGenesisBlocks() error {
	return nil
}

func (si *stateImport) GetAllGenesisBlocks() map[uint32]data.HeaderHandler {
	return si.genesisHeaders
}

func (si *stateImport) IsInterfaceNil() bool {
	return si == nil
}
