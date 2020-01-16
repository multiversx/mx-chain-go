package genesis

import (
	"encoding/json"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

// ArgsNewStateExporter defines the arguments needed to create new state exporter
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

var log = logger.GetOrCreate("update/genesis/")

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

// ExportAll syncs and exports all the data from every shard for a certain epoch start block
func (se *stateExport) ExportAll(epoch uint32) error {
	err := se.stateSyncer.SyncAllState(epoch)
	if err != nil {
		return err
	}

	toExportTries, err := se.stateSyncer.GetAllTries()
	if err != nil {
		return err
	}

	metaBlock, err := se.stateSyncer.GetMetaBlock()
	if err != nil {
		return err
	}

	versionKey := CreateVersionKey(metaBlock)

	jsonData, err := json.Marshal(metaBlock)
	if err != nil {
		return err
	}

	err = se.writer.Write(MetaBlockFileName, versionKey, jsonData)
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

	se.writer.Finish()

	return nil
}

func (se *stateExport) exportTrie(key string, trie data.Trie) error {
	leaves, err := trie.GetAllLeaves()
	if err != nil {
		return err
	}

	accType, shId, err := GetTrieTypeAndShId(key)
	if err != nil {
		return err
	}

	if shId > se.shardCoordinator.NumberOfShards() && shId != sharding.MetachainShardId {
		return sharding.ErrInvalidShardId
	}

	rootHashKey := CreateRootHashKey(key)
	rootHash, err := trie.Root()
	if err != nil {
		return err
	}

	err = se.writer.Write(key, rootHashKey, rootHash)
	if err != nil {
		return err
	}

	for address, buff := range leaves {
		account, err := NewEmptyAccount(accType)
		if err != nil {
			log.Warn("error creating new account account", "address", address, "error", err)
			continue
		}
		err = se.marshalizer.Unmarshal(account, buff)
		if err != nil {
			log.Warn("error unmarshaling account", "address", address, "error", err)
			continue
		}

		jsonData, err := json.Marshal(account)
		if err != nil {
			log.Warn("error marshaling account", "address", address, "error", err)
			continue
		}

		keyToExport := CreateAccountKey(accType, shId, address)
		err = se.writer.Write(key, keyToExport, jsonData)
		if err != nil {
			return err
		}
	}

	return nil
}

func (se *stateExport) exportMBs(key string, mb *block.MiniBlock) error {
	marshaledData, err := json.Marshal(mb)
	if err != nil {
		return err
	}

	keyToSave := CreateMiniBlockKey(key)
	err = se.writer.Write(MiniBlocksFileName, keyToSave, marshaledData)
	if err != nil {
		return err
	}

	return nil
}

func (se *stateExport) exportTx(key string, tx data.TransactionHandler) error {
	marshaledData, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	keyToSave := CreateTransactionKey(key, tx)
	err = se.writer.Write(TransactionsFileName, keyToSave, marshaledData)
	if err != nil {
		return err
	}

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (se *stateExport) IsInterfaceNil() bool {
	return se == nil
}
