package genesis

import (
	"encoding/json"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
)

type ArgsNewStateExporter struct {
	TrieSyncHandlers update.TrieSyncContainer
}

type stateExport struct {
	accountsContainer update.AccountsHandlerContainer
	stateStore        storage.Storer
	stateSyncer       update.StateSyncer
	trieSyncHandlers  update.TrieSyncContainer
	shardCoordinator  sharding.Coordinator
	marshalizer       marshal.Marshalizer
	exportMarshalizer marshal.Marshalizer
}

// json marshalizer unmarshal all the trie data and write somewhere the new data

func NewStateExporter() (*stateExport, error) {
	return nil, nil
}

func (se *stateExport) ExportAll(epoch uint32) error {
	err := se.stateSyncer.SyncAllState(epoch)
	if err != nil {
		return err
	}

	toExportTries, err := se.stateSyncer.GetAllTries()
	if err != nil {
		return err
	}

	for key, trie := range toExportTries {
		err = se.exportTrie(key, trie)
		if err != nil {
			return err
		}
	}

	toExportMBs, err := se.stateSyncer.GetAllMiniBlocks()
	for key, mb := range toExportMBs {
		err := se.exportMBs(key, mb)
		if err != nil {
			return err
		}
	}

	toExportTransactions, err := se.stateSyncer.GetAllTransactions()
	for key, tx := range toExportTransactions {
		err := se.exportTx(key, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func newEmptyAccount(accountType factory.Type) state.AccountHandler {
	switch accountType {
	case factory.UserAccount:
		return &state.Account{}
	case factory.ShardStatistics:
		return &state.MetaAccount{}
	case factory.ValidatorAccount:
		return &state.PeerAccount{}
	}

	return &state.Account{}
}

func (se *stateExport) exportTrie(key string, trie data.Trie) error {
	leaves, err := trie.GetAllLeaves()
	if err != nil {
		return err
	}

	accType, shId, err := update.GetTrieTypeAndShId(key)
	if err != nil {
		return err
	}

	if shId > se.shardCoordinator.NumberOfShards() && shId != sharding.MetachainShardId {
		return sharding.ErrInvalidShardId
	}

	for address, buff := range leaves {
		account := newEmptyAccount(accType)
		err := se.marshalizer.Unmarshal(account, buff)
		if err != nil {
			log.Warn("error unmarshalling account", "address", address, "error", err)
			continue
		}

		jsonData, err := json.Marshal(account)
		if err != nil {
			log.Warn("error marshalling account", "address", address, "error", err)
			continue
		}

		keyToExport := update.CreateAccountKey(accType, shId, address)
		err = se.stateStore.Put([]byte(keyToExport), jsonData)
		if err != nil {
			return err
		}
	}

	return nil
}

func (se *stateExport) exportMBs(key string, mb *block.MiniBlock) error {
	marshalledData, err := json.Marshal(mb)
	if err != nil {
		return err
	}

	keyToSave := update.CreateMiniBlockKey(key)
	err = se.stateStore.Put([]byte(keyToSave), marshalledData)
	if err != nil {
		return err
	}

	return nil
}

func (se *stateExport) exportTx(key string, tx data.TransactionHandler) error {
	marshalledData, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	keyToSave := update.CreateTransactionKey(key, tx)
	err = se.stateStore.Put([]byte(keyToSave), marshalledData)
	if err != nil {
		return err
	}

	return nil
}

func (se *stateExport) IsInterfaceNil() bool {
	return se == nil
}
