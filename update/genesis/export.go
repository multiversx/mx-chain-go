package genesis

import (
	"encoding/json"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

type ArgsNewStateExporter struct {
	ShardCoordinator sharding.Coordinator
	StateSyncer      update.StateSyncer
	Marshalizer      marshal.Marshalizer
	Writer           update.MultiFileWriter
}

type stateExport struct {
	writer           update.MultiFileWriter
	stateSyncer      update.StateSyncer
	shardCoordinator sharding.Coordinator
	marshalizer      marshal.Marshalizer
}

// NewStateExporter exports all the data at a specific moment to a set of files
func NewStateExporter(args ArgsNewStateExporter) (*stateExport, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, data.ErrNilShardCoordinator
	}
	if check.IfNil(args.StateSyncer) {
		return nil, update.ErrNilStateSyncer
	}
	if check.IfNil(args.Marshalizer) {
		return nil, data.ErrNilMarshalizer
	}
	if check.IfNil(args.Writer) {
		return nil, epochStart.ErrNilStorage
	}

	se := &stateExport{
		writer:           args.Writer,
		stateSyncer:      args.StateSyncer,
		shardCoordinator: args.ShardCoordinator,
		marshalizer:      args.Marshalizer,
	}

	return se, nil
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

	versionKey := update.CreateVersionKey(se.stateSyncer.GetMetaBlock())
	jsonData, err := json.Marshal(se.stateSyncer.GetMetaBlock())
	if err != nil {
		return err
	}

	err = se.writer.Write(update.MetaBlockFileName, versionKey, jsonData)
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
	if err != nil {
		return err
	}

	for key, mb := range toExportMBs {
		err := se.exportMBs(key, mb)
		if err != nil {
			return err
		}
	}

	toExportTransactions, err := se.stateSyncer.GetAllTransactions()
	if err != nil {
		return err
	}

	for key, tx := range toExportTransactions {
		err := se.exportTx(key, tx)
		if err != nil {
			return err
		}
	}

	return nil
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

	rootHashKey := update.CreateRootHashKey(key)
	rootHash, err := trie.Root()
	if err != nil {
		return err
	}

	err = se.writer.Write(key, rootHashKey, rootHash)
	if err != nil {
		return err
	}

	for address, buff := range leaves {
		account, err := update.NewEmptyAccount(accType)
		if err != nil {
			log.Warn("error creating new account account", "address", address, "error", err)
			continue
		}
		err = se.marshalizer.Unmarshal(account, buff)
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
		err = se.writer.Write(key, keyToExport, jsonData)
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
	err = se.writer.Write(update.MiniBlocksFileName, keyToSave, marshalledData)
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
	err = se.writer.Write(update.TransactionsFileName, keyToSave, marshalledData)
	if err != nil {
		return err
	}

	return nil
}

func (se *stateExport) IsInterfaceNil() bool {
	return se == nil
}
