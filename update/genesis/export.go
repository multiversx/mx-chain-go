package genesis

import (
	"encoding/hex"
	"encoding/json"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

var _ update.ExportHandler = (*stateExport)(nil)

// ArgsNewStateExporter defines the arguments needed to create new state exporter
type ArgsNewStateExporter struct {
	ShardCoordinator sharding.Coordinator
	StateSyncer      update.StateSyncer
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	Writer           update.MultiFileWriter
}

type stateExport struct {
	writer           update.MultiFileWriter
	stateSyncer      update.StateSyncer
	shardCoordinator sharding.Coordinator
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
}

var log = logger.GetOrCreate("update/genesis")

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
	if check.IfNil(args.Hasher) {
		return nil, update.ErrNilHasher
	}

	se := &stateExport{
		writer:           args.Writer,
		stateSyncer:      args.StateSyncer,
		shardCoordinator: args.ShardCoordinator,
		marshalizer:      args.Marshalizer,
		hasher:           args.Hasher,
	}

	return se, nil
}

// ExportAll syncs and exports all the data from every shard for a certain epoch start block
func (se *stateExport) ExportAll(epoch uint32) error {
	err := se.stateSyncer.SyncAllState(epoch)
	if err != nil {
		return err
	}

	defer se.writer.Finish()

	err = se.exportMeta()
	if err != nil {
		return err
	}

	err = se.exportAllTries()
	if err != nil {
		return err
	}

	err = se.exportAllMiniBlocks()
	if err != nil {
		return err
	}

	err = se.exportAllTransactions()
	if err != nil {
		return err
	}

	return nil
}

func (se *stateExport) exportAllTransactions() error {
	toExportTransactions, err := se.stateSyncer.GetAllTransactions()
	if err != nil {
		return err
	}

	log.Debug("Exported transactions", "len", len(toExportTransactions))
	for key, tx := range toExportTransactions {
		err := se.exportTx(key, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (se *stateExport) exportAllMiniBlocks() error {
	toExportMBs, err := se.stateSyncer.GetAllMiniBlocks()
	if err != nil {
		return err
	}

	log.Debug("Exported miniBlocks", "len", len(toExportMBs))
	for key, mb := range toExportMBs {
		err := se.exportMBs(key, mb)
		if err != nil {
			return err
		}
	}

	return nil
}

func (se *stateExport) exportAllTries() error {
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

	return nil
}

func (se *stateExport) exportMeta() error {
	metaBlock, err := se.stateSyncer.GetEpochStartMetaBlock()
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(metaBlock)
	if err != nil {
		return err
	}

	metaHash := se.hasher.Compute(string(jsonData))
	versionKey := CreateVersionKey(metaBlock, metaHash)

	err = se.writer.Write(MetaBlockFileName, versionKey, jsonData)
	if err != nil {
		return err
	}

	log.Debug("Exported metaBlock", "rootHash", metaBlock.RootHash)

	return nil
}

func (se *stateExport) exportTrie(key string, trie data.Trie) error {
	fileName := TrieFileName + atSep + key

	leaves, err := trie.GetAllLeaves()
	if err != nil {
		return err
	}

	accType, shId, err := GetTrieTypeAndShId(fileName)
	if err != nil {
		return err
	}

	if shId > se.shardCoordinator.NumberOfShards() && shId != core.MetachainShardId {
		return sharding.ErrInvalidShardId
	}

	rootHashKey := CreateRootHashKey(key)
	rootHash, err := trie.Root()
	if err != nil {
		return err
	}

	err = se.writer.Write(fileName, rootHashKey, rootHash)
	if err != nil {
		return err
	}

	if accType == DataTrie {
		return se.exportDataTries(leaves, accType, shId, fileName)
	}

	return se.exportAccountLeafs(leaves, accType, shId, fileName)
}

func (se *stateExport) exportDataTries(leafs map[string][]byte, accType Type, shId uint32, fileName string) error {
	for address, buff := range leafs {
		keyToExport := CreateAccountKey(accType, shId, address)
		err := se.writer.Write(fileName, keyToExport, buff)
		if err != nil {
			return err
		}
	}

	se.writer.CloseFile(fileName)
	return nil
}

func (se *stateExport) exportAccountLeafs(leafs map[string][]byte, accType Type, shId uint32, fileName string) error {
	for address, buff := range leafs {
		keyToExport := CreateAccountKey(accType, shId, address)
		account, err := NewEmptyAccount(accType, []byte(address))
		if err != nil {
			log.Warn("error creating new account account", "address", address, "error", err)
			continue
		}
		err = se.marshalizer.Unmarshal(account, buff)
		if err != nil {
			log.Trace("error unmarshaling account this is maybe a code error",
				"address", hex.EncodeToString([]byte(address)),
				"error", err,
			)

			err = se.writer.Write(fileName, keyToExport, buff)
			if err != nil {
				return err
			}

			continue
		}

		jsonData, err := json.Marshal(account)
		if err != nil {
			log.Warn("error marshaling account", "address", address, "error", err)
			continue
		}

		err = se.writer.Write(fileName, keyToExport, jsonData)
		if err != nil {
			return err
		}
	}

	se.writer.CloseFile(fileName)
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
