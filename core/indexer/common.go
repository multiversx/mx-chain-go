package indexer

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type objectsMap = map[string]interface{}

type commonProcessor struct {
	shardCoordinator         sharding.Coordinator
	addressPubkeyConverter   core.PubkeyConverter
	validatorPubkeyConverter core.PubkeyConverter
	txFeeCalculator          process.TransactionFeeCalculator
}

func prepareGeneralInfo(tpsBenchmark statistics.TPSBenchmark) bytes.Buffer {
	var buff bytes.Buffer

	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, metachainTpsDocID, tpsIndex, "\n"))
	generalInfo := TPS{
		LiveTPS:               tpsBenchmark.LiveTPS(),
		PeakTPS:               tpsBenchmark.PeakTPS(),
		NrOfShards:            tpsBenchmark.NrOfShards(),
		BlockNumber:           tpsBenchmark.BlockNumber(),
		RoundNumber:           tpsBenchmark.RoundNumber(),
		RoundTime:             tpsBenchmark.RoundTime(),
		AverageBlockTxCount:   tpsBenchmark.AverageBlockTxCount(),
		LastBlockTxCount:      tpsBenchmark.LastBlockTxCount(),
		TotalProcessedTxCount: tpsBenchmark.TotalProcessedTxCount(),
	}

	serializedInfo, err := json.Marshal(generalInfo)
	if err != nil {
		log.Debug("indexer: could not serialize tps info, will skip indexing tps this round")
		return buff
	}
	// append a newline foreach element in the bulk we create
	serializedInfo = append(serializedInfo, "\n"...)

	buff.Grow(len(meta) + len(serializedInfo))
	_, err = buff.Write(meta)
	if err != nil {
		log.Warn("elastic search: update TPS write meta", "error", err.Error())
	}
	_, err = buff.Write(serializedInfo)
	if err != nil {
		log.Warn("elastic search: update TPS write serialized info", "error", err.Error())
	}

	return buff
}

func serializeShardInfo(shardInfo statistics.ShardStatistic) ([]byte, []byte) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s%d", "_type" : "%s" } }%s`,
		shardTpsDocIDPrefix, shardInfo.ShardID(), tpsIndex, "\n"))

	bigTxCount := big.NewInt(int64(shardInfo.AverageBlockTxCount()))
	shardTPS := TPS{
		ShardID:               shardInfo.ShardID(),
		LiveTPS:               shardInfo.LiveTPS(),
		PeakTPS:               shardInfo.PeakTPS(),
		AverageTPS:            shardInfo.AverageTPS(),
		AverageBlockTxCount:   bigTxCount,
		CurrentBlockNonce:     shardInfo.CurrentBlockNonce(),
		LastBlockTxCount:      shardInfo.LastBlockTxCount(),
		TotalProcessedTxCount: shardInfo.TotalProcessedTxCount(),
	}

	serializedInfo, err := json.Marshal(shardTPS)
	if err != nil {
		log.Debug("indexer: could not serialize tps info, will skip indexing tps this shard")
		return nil, nil
	}
	// append a newline foreach element in the bulk we create
	serializedInfo = append(serializedInfo, "\n"...)

	return serializedInfo, meta
}

func (cm *commonProcessor) buildTransaction(
	tx *transaction.Transaction,
	txHash []byte,
	mbHash []byte,
	mb *block.MiniBlock,
	header data.HeaderHandler,
	txStatus string,
) *Transaction {
	gasUsed := cm.txFeeCalculator.ComputeGasLimit(tx)
	fee := cm.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, gasUsed)

	return &Transaction{
		Hash:          hex.EncodeToString(txHash),
		MBHash:        hex.EncodeToString(mbHash),
		Nonce:         tx.Nonce,
		Round:         header.GetRound(),
		Value:         tx.Value.String(),
		Receiver:      cm.addressPubkeyConverter.Encode(tx.RcvAddr),
		Sender:        cm.addressPubkeyConverter.Encode(tx.SndAddr),
		ReceiverShard: cm.shardCoordinator.ComputeId(tx.RcvAddr),
		SenderShard:   mb.SenderShardID,
		GasPrice:      tx.GasPrice,
		GasLimit:      tx.GasLimit,
		Data:          tx.Data,
		Signature:     hex.EncodeToString(tx.Signature),
		Timestamp:     time.Duration(header.GetTimeStamp()),
		Status:        txStatus,
		GasUsed:       gasUsed,
		Fee:           fee.String(),
		rcvAddrBytes:  tx.RcvAddr,
	}
}

func (cm *commonProcessor) buildRewardTransaction(
	rTx *rewardTx.RewardTx,
	txHash []byte,
	mbHash []byte,
	mb *block.MiniBlock,
	header data.HeaderHandler,
	txStatus string,
) *Transaction {
	return &Transaction{
		Hash:          hex.EncodeToString(txHash),
		MBHash:        hex.EncodeToString(mbHash),
		Nonce:         0,
		Round:         rTx.Round,
		Value:         rTx.Value.String(),
		Receiver:      cm.addressPubkeyConverter.Encode(rTx.RcvAddr),
		Sender:        fmt.Sprintf("%d", core.MetachainShardId),
		ReceiverShard: mb.ReceiverShardID,
		SenderShard:   mb.SenderShardID,
		GasPrice:      0,
		GasLimit:      0,
		Data:          make([]byte, 0),
		Signature:     "",
		Timestamp:     time.Duration(header.GetTimeStamp()),
		Status:        txStatus,
	}
}

func (cm *commonProcessor) convertScResultInDatabaseScr(scHash string, sc *smartContractResult.SmartContractResult) ScResult {
	relayerAddr := ""
	if len(sc.RelayerAddr) > 0 {
		relayerAddr = cm.addressPubkeyConverter.Encode(sc.RelayerAddr)
	}

	return ScResult{
		Hash:           hex.EncodeToString([]byte(scHash)),
		Nonce:          sc.Nonce,
		GasLimit:       sc.GasLimit,
		GasPrice:       sc.GasPrice,
		Value:          sc.Value.String(),
		Sender:         cm.addressPubkeyConverter.Encode(sc.SndAddr),
		Receiver:       cm.addressPubkeyConverter.Encode(sc.RcvAddr),
		RelayerAddr:    relayerAddr,
		RelayedValue:   sc.RelayedValue.String(),
		Code:           string(sc.Code),
		Data:           sc.Data,
		PreTxHash:      hex.EncodeToString(sc.PrevTxHash),
		OriginalTxHash: hex.EncodeToString(sc.OriginalTxHash),
		CallType:       strconv.Itoa(int(sc.CallType)),
		CodeMetadata:   sc.CodeMetadata,
		ReturnMessage:  string(sc.ReturnMessage),
	}
}

func serializeBulkMiniBlocks(
	hdrShardID uint32,
	bulkMbs []*Miniblock,
	getAlreadyIndexedItems func(hashes []string, index string) (map[string]bool, error),
) (bytes.Buffer, map[string]bool) {
	var err error
	var buff bytes.Buffer

	mbsHashes := make([]string, len(bulkMbs))
	for idx := range bulkMbs {
		mbsHashes[idx] = bulkMbs[idx].Hash
	}

	existsInDb, err := getAlreadyIndexedItems(mbsHashes, miniblocksIndex)
	if err != nil {
		log.Warn("indexer get indexed items miniblocks",
			"error", err.Error())
		return buff, make(map[string]bool)
	}

	for _, mb := range bulkMbs {
		var meta, serializedData []byte
		if !existsInDb[mb.Hash] {
			//insert miniblock in database
			meta = []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, mb.Hash, "_doc", "\n"))
			serializedData, err = json.Marshal(mb)
			if err != nil {
				log.Debug("indexer: marshal",
					"error", "could not serialize miniblock, will skip indexing",
					"mb hash", mb.Hash)
				continue
			}
		} else {
			// update miniblock
			meta = []byte(fmt.Sprintf(`{ "update" : { "_id" : "%s" } }%s`, mb.Hash, "\n"))
			if hdrShardID == mb.SenderShardID {
				// update sender block hash
				serializedData = []byte(fmt.Sprintf(`{ "doc" : { "senderBlockHash" : "%s" } }`, mb.SenderBlockHash))
			} else {
				// update receiver block hash
				serializedData = []byte(fmt.Sprintf(`{ "doc" : { "receiverBlockHash" : "%s" } }`, mb.ReceiverBlockHash))
			}
		}

		buff = prepareBufferMiniblocks(buff, meta, serializedData)
	}

	return buff, existsInDb
}

func prepareBufferMiniblocks(buff bytes.Buffer, meta, serializedData []byte) bytes.Buffer {
	// append a newline for each element
	serializedData = append(serializedData, "\n"...)
	buff.Grow(len(meta) + len(serializedData))
	_, err := buff.Write(meta)
	if err != nil {
		log.Warn("elastic search: serialize bulk miniblocks, write meta", "error", err.Error())
	}
	_, err = buff.Write(serializedData)
	if err != nil {
		log.Warn("elastic search: serialize bulk miniblocks, write serialized miniblock", "error", err.Error())
	}

	return buff
}

func serializeTransactions(
	transactions []*Transaction,
	selfShardID uint32,
	_ func(hashes []string, index string) (map[string]bool, error),
	mbsHashInDB map[string]bool,
) ([]bytes.Buffer, error) {
	var err error

	var buff bytes.Buffer
	buffSlice := make([]bytes.Buffer, 0)
	for _, tx := range transactions {
		isMBOfTxInDB := mbsHashInDB[tx.MBHash]
		meta, serializedData, errPrepareTx := prepareSerializedDataForATransaction(tx, selfShardID, isMBOfTxInDB)
		if errPrepareTx != nil {
			log.Warn("error preparing transaction for indexing", "tx hash", tx.Hash, "error", err)
			return nil, errPrepareTx
		}

		// append a newline for each element
		serializedData = append(serializedData, "\n"...)

		buffLenWithCurrentTx := buff.Len() + len(meta) + len(serializedData)
		if buffLenWithCurrentTx > bulkSizeThreshold && buff.Len() != 0 {
			buffSlice = append(buffSlice, buff)
			buff = bytes.Buffer{}
		}

		buff.Grow(len(meta) + len(serializedData))
		_, err = buff.Write(meta)
		if err != nil {
			log.Warn("elastic search: serialize bulk tx, write meta", "error", err.Error())
			return nil, err

		}
		_, err = buff.Write(serializedData)
		if err != nil {
			log.Warn("elastic search: serialize bulk tx, write serialized tx", "error", err.Error())
			return nil, err
		}
	}

	// check if the last buffer contains data
	if buff.Len() != 0 {
		buffSlice = append(buffSlice, buff)
	}

	return buffSlice, nil
}

func serializeAccounts(accounts map[string]*AccountInfo) ([]bytes.Buffer, error) {
	var err error

	var buff bytes.Buffer
	buffSlice := make([]bytes.Buffer, 0)
	for address, acc := range accounts {
		meta, serializedData, errPrepareAcc := prepareSerializedAccountInfo(address, acc)
		if len(meta) == 0 {
			log.Warn("cannot prepare serializes account info", "error", errPrepareAcc)
			return nil, err
		}

		// append a newline for each element
		serializedData = append(serializedData, "\n"...)

		buffLenWithCurrentAcc := buff.Len() + len(meta) + len(serializedData)
		if buffLenWithCurrentAcc > bulkSizeThreshold && buff.Len() != 0 {
			buffSlice = append(buffSlice, buff)
			buff = bytes.Buffer{}
		}

		buff.Grow(len(meta) + len(serializedData))
		_, err = buff.Write(meta)
		if err != nil {
			log.Warn("elastic search: serialize bulk accounts, write meta", "error", err.Error())
			return nil, err
		}
		_, err = buff.Write(serializedData)
		if err != nil {
			log.Warn("elastic search: serialize bulk accounts, write serialized account", "error", err.Error())
			return nil, err
		}
	}

	// check if the last buffer contains data
	if buff.Len() != 0 {
		buffSlice = append(buffSlice, buff)
	}

	return buffSlice, nil
}

func serializeAccountsHistory(accounts map[string]*AccountBalanceHistory) ([]bytes.Buffer, error) {
	var err error

	var buff bytes.Buffer
	buffSlice := make([]bytes.Buffer, 0)
	for address, acc := range accounts {
		meta, serializedData, errPrepareAcc := prepareSerializedAccountBalanceHistory(address, acc)
		if errPrepareAcc != nil {
			log.Warn("cannot prepare serializes account balance history", "error", err)
			return nil, err
		}

		// append a newline for each element
		serializedData = append(serializedData, "\n"...)

		buffLenWithCurrentAccountHistory := buff.Len() + len(meta) + len(serializedData)
		if buffLenWithCurrentAccountHistory > bulkSizeThreshold && buff.Len() != 0 {
			buffSlice = append(buffSlice, buff)
			buff = bytes.Buffer{}
		}

		buff.Grow(len(meta) + len(serializedData))
		_, err = buff.Write(meta)
		if err != nil {
			log.Warn("elastic search: serialize bulk accounts history, write meta", "error", err.Error())
			return nil, err
		}
		_, err = buff.Write(serializedData)
		if err != nil {
			log.Warn("elastic search: serialize bulk accounts history, write serialized account history", "error", err.Error())
			return nil, err
		}
	}

	// check if the last buffer contains data
	if buff.Len() != 0 {
		buffSlice = append(buffSlice, buff)
	}

	return buffSlice, nil
}

func prepareSerializedDataForATransaction(
	tx *Transaction,
	selfShardID uint32,
	_ bool,
) ([]byte, []byte, error) {
	metaData := []byte(fmt.Sprintf(`{"update":{"_id":"%s", "_type": "_doc"}}%s`, tx.Hash, "\n"))

	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize transaction, will skip indexing",
			"tx hash", tx.Hash)
		return nil, nil, err
	}

	if isIntraShardOrInvalid(tx, selfShardID) {
		// if transaction is intra-shard, use basic insert as data can be re-written at forks
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, tx.Hash, "_doc", "\n"))
		log.Trace("indexer tx is intra shard or invalid tx", "meta", string(meta), "marshaledTx", string(marshaledTx))

		return meta, marshaledTx, nil
	}

	if !isCrossShardDstMe(tx, selfShardID) {
		// if transaction is cross-shard and current shard ID is source, use upsert without updating anything
		serializedData :=
			[]byte(fmt.Sprintf(`{"script":{"source":"return"},"upsert":%s}`,
				string(marshaledTx)))
		log.Trace("indexer tx is on sender shard", "metaData", string(metaData), "serializedData", string(serializedData))

		return metaData, serializedData, nil
	}

	// if transaction is cross-shard and current shard ID is destination, use upsert with updating fields
	marshaledLog, err := json.Marshal(tx.Log)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize transaction log, will skip indexing",
			"tx hash", tx.Hash)
		return nil, nil, err
	}
	scResults, err := json.Marshal(tx.SmartContractResults)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize smart contract results, will skip indexing",
			"tx hash", tx.Hash)
		return nil, nil, err
	}

	marshaledTimestamp, err := json.Marshal(tx.Timestamp)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize timestamp, will skip indexing",
			"tx hash", tx.Hash)
		return nil, nil, err
	}

	serializedData := []byte(fmt.Sprintf(`{"script":{"source":"`+
		`ctx._source.status = params.status;`+
		`ctx._source.miniBlockHash = params.miniBlockHash;`+
		`ctx._source.log = params.log;`+
		`ctx._source.scResults = params.scResults;`+
		`ctx._source.timestamp = params.timestamp;`+
		`ctx._source.gasUsed = params.gasUsed;`+
		`ctx._source.fee = params.fee;`+
		`","lang": "painless","params":`+
		`{"status": "%s", "miniBlockHash": "%s", "log": %s, "scResults": %s, "timestamp": %s, "gasUsed": %d, "fee": "%s"}},"upsert":%s}`,
		tx.Status, tx.MBHash, string(marshaledLog), string(scResults), string(marshaledTimestamp), tx.GasUsed, tx.Fee, string(marshaledTx)))

	log.Trace("indexer tx is on destination shard", "metaData", string(metaData), "serializedData", string(serializedData))

	return metaData, serializedData, nil
}

func isRelayedTx(tx *Transaction) bool {
	return strings.HasPrefix(string(tx.Data), "relayedTx") && len(tx.SmartContractResults) > 0
}

func prepareSerializedAccountInfo(address string, account *AccountInfo) ([]byte, []byte, error) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, address, "\n"))
	serializedData, err := json.Marshal(account)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize account, will skip indexing",
			"address", address)
		return nil, nil, err
	}

	return meta, serializedData, nil
}

func prepareSerializedAccountBalanceHistory(address string, account *AccountBalanceHistory) ([]byte, []byte, error) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, address, "\n"))
	serializedData, err := json.Marshal(account)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize account history entry, will skip indexing",
			"address", address)
		return nil, nil, err
	}

	return meta, serializedData, nil
}

func isCrossShardDstMe(tx *Transaction, selfShardID uint32) bool {
	return tx.SenderShard != tx.ReceiverShard && tx.ReceiverShard == selfShardID
}

func isIntraShardOrInvalid(tx *Transaction, selfShardID uint32) bool {
	return (tx.SenderShard == tx.ReceiverShard && tx.ReceiverShard == selfShardID) || tx.Status == transaction.TxStatusInvalid.String()
}

func getDecodedResponseMultiGet(response objectsMap) map[string]bool {
	founded := make(map[string]bool)
	interfaceSlice, ok := response["docs"].([]interface{})
	if !ok {
		return founded
	}

	for _, element := range interfaceSlice {
		obj := element.(objectsMap)
		_, ok = obj["error"]
		if ok {
			continue
		}
		founded[obj["_id"].(string)] = obj["found"].(bool)
	}

	return founded
}

// GetElasticTemplatesAndPolicies will return elastic templates and policies
func GetElasticTemplatesAndPolicies(path string, useKibana bool) (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
	indexTemplates := make(map[string]*bytes.Buffer)
	indexPolicies := make(map[string]*bytes.Buffer)
	var err error

	indexes := []string{"opendistro", txIndex, blockIndex, miniblocksIndex, tpsIndex, ratingIndex, roundIndex, validatorsIndex, accountsIndex, accountsHistoryIndex}
	for _, index := range indexes {
		indexTemplates[index], err = getTemplateByIndex(path, index)
		if err != nil {
			return nil, nil, err
		}
	}

	if useKibana {
		indexesPolicies := []string{txPolicy, blockPolicy, miniblocksPolicy, ratingPolicy, roundPolicy, validatorsPolicy, accountsHistoryPolicy}
		for _, indexPolicy := range indexesPolicies {
			indexPolicies[indexPolicy], err = getPolicyByIndex(path, indexPolicy)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return indexTemplates, indexPolicies, nil
}

func getTemplateByIndex(path string, index string) (*bytes.Buffer, error) {
	indexTemplate := &bytes.Buffer{}

	fileName := fmt.Sprintf("%s.json", index)
	filePath := filepath.Join(path, fileName)
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("getTemplateByIndex: %w, path %s, error %s", ErrReadTemplatesFile, filePath, err.Error())
	}

	indexTemplate.Grow(len(fileBytes))
	_, err = indexTemplate.Write(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("getTemplateByIndex: %w, path %s, error %s", ErrWriteToBuffer, filePath, err.Error())
	}

	return indexTemplate, nil
}

func getPolicyByIndex(path string, index string) (*bytes.Buffer, error) {
	indexPolicy := &bytes.Buffer{}

	fileName := fmt.Sprintf("%s.json", index)
	filePath := filepath.Join(path, fileName)
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("getPolicyByIndex: %w, path %s, error %s", ErrReadPolicyFile, filePath, err.Error())
	}

	indexPolicy.Grow(len(fileBytes))
	_, err = indexPolicy.Write(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("getPolicyByIndex: %w, path %s, error %s", ErrWriteToBuffer, filePath, err.Error())
	}

	return indexPolicy, nil
}

func stringValueToBigInt(strValue string) *big.Int {
	value, ok := big.NewInt(0).SetString(strValue, 10)
	if !ok {
		return big.NewInt(0)
	}

	return value
}
