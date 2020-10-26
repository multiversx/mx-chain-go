package indexer

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

type objectsMap = map[string]interface{}

type commonProcessor struct {
	addressPubkeyConverter   core.PubkeyConverter
	validatorPubkeyConverter core.PubkeyConverter
	minGasLimit              uint64
	gasPerDataByte           uint64
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
	gasUsed := cm.minGasLimit + uint64(len(tx.Data))*cm.gasPerDataByte

	return &Transaction{
		Hash:          hex.EncodeToString(txHash),
		MBHash:        hex.EncodeToString(mbHash),
		Nonce:         tx.Nonce,
		Round:         header.GetRound(),
		Value:         tx.Value.String(),
		Receiver:      cm.addressPubkeyConverter.Encode(tx.RcvAddr),
		Sender:        cm.addressPubkeyConverter.Encode(tx.SndAddr),
		ReceiverShard: mb.ReceiverShardID,
		SenderShard:   mb.SenderShardID,
		GasPrice:      tx.GasPrice,
		GasLimit:      tx.GasLimit,
		Data:          tx.Data,
		Signature:     hex.EncodeToString(tx.Signature),
		Timestamp:     time.Duration(header.GetTimeStamp()),
		Status:        txStatus,
		GasUsed:       gasUsed,
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
) []bytes.Buffer {
	var err error

	var buff bytes.Buffer
	buffSlice := make([]bytes.Buffer, 0)
	for _, tx := range transactions {
		isMBOfTxInDB := mbsHashInDB[tx.MBHash]
		meta, serializedData := prepareSerializedDataForATransaction(tx, selfShardID, isMBOfTxInDB)
		if len(meta) == 0 {
			continue
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
		}
		_, err = buff.Write(serializedData)
		if err != nil {
			log.Warn("elastic search: serialize bulk tx, write serialized tx", "error", err.Error())
		}
	}

	// check if the last buffer contains data
	if buff.Len() != 0 {
		buffSlice = append(buffSlice, buff)
	}

	return buffSlice
}

func serializeAccounts(accounts map[string]*AccountInfo) []bytes.Buffer {
	var err error

	var buff bytes.Buffer
	buffSlice := make([]bytes.Buffer, 0)
	for address, acc := range accounts {
		meta, serializedData := prepareSerializedAccountInfo(address, acc)
		if len(meta) == 0 {
			continue
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
		}
		_, err = buff.Write(serializedData)
		if err != nil {
			log.Warn("elastic search: serialize bulk accounts, write serialized account", "error", err.Error())
		}
	}

	// check if the last buffer contains data
	if buff.Len() != 0 {
		buffSlice = append(buffSlice, buff)
	}

	return buffSlice
}

func serializeAccountsHistory(accounts map[string]*AccountBalanceHistory) []bytes.Buffer {
	var err error

	var buff bytes.Buffer
	buffSlice := make([]bytes.Buffer, 0)
	for address, acc := range accounts {
		meta, serializedData := prepareSerializedAccountBalanceHistory(address, acc)
		if len(meta) == 0 {
			continue
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
		}
		_, err = buff.Write(serializedData)
		if err != nil {
			log.Warn("elastic search: serialize bulk accounts history, write serialized account history", "error", err.Error())
		}
	}

	// check if the last buffer contains data
	if buff.Len() != 0 {
		buffSlice = append(buffSlice, buff)
	}

	return buffSlice
}

func prepareSerializedDataForATransaction(
	tx *Transaction,
	selfShardID uint32,
	_ bool,
) (meta []byte, serializedData []byte) {
	meta = []byte(fmt.Sprintf(`{"update":{"_id":"%s", "_type": "_doc"}}%s`, tx.Hash, "\n"))

	marshalledTx, err := json.Marshal(tx)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize transaction, will skip indexing",
			"tx hash", tx.Hash)
		return
	}
	if !isCrossShardDstMe(tx, selfShardID) || tx.Status == transaction.TxStatusInvalid.String() {
		serializedData =
			[]byte(fmt.Sprintf(`{"script":{"source":""},"upsert":%s}`,
				string(marshalledTx)))
		log.Warn("log meta - insert", "data", string(meta))
		log.Warn("log srld data - insert", "data", string(serializedData))
		return
	}

	marshalizedLog, err := json.Marshal(tx.Log)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize transaction log, will skip indexing",
			"tx hash", tx.Hash)
		return nil, nil
	}
	scResults, err := json.Marshal(tx.SmartContractResults)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize smart contract results, will skip indexing",
			"tx hash", tx.Hash)
		return nil, nil
	}

	marshalizedTimestamp, err := json.Marshal(tx.Timestamp)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize timestamp, will skip indexing",
			"tx hash", tx.Hash)
		return nil, nil
	}

	if tx.GasUsed == tx.GasLimit {
		// do not update gasUsed because it is the same with gasUsed when transaction was saved first time in database
		serializedData =
			[]byte(fmt.Sprintf(`{"script":{"source":"`+
				`ctx._source.status = params.status;`+
				`ctx._source.log = params.log;`+
				`ctx._source.scResults = params.scResults;`+
				`ctx._source.timestamp = params.timestamp;`+
				`","lang": "painless","params":`+
				`{"status": "%s", "log": %s, "scResults": %s, "timestamp": %s}},"upsert":%s}`,
				tx.Status, string(marshalizedLog), string(scResults), string(marshalizedTimestamp), string(marshalledTx)))
	} else {
		// update gasUsed because was changed (is a smart contract operation)
		serializedData =
			[]byte(fmt.Sprintf(`{"script":{"source":"`+
				`ctx._source.status = params.status;`+
				`ctx._source.log = params.log;`+
				`ctx._source.scResults = params.scResults;`+
				`ctx._source.timestamp = params.timestamp;`+
				`ctx._source.gasUsed = params.gasUsed;`+
				`","lang": "painless","params":`+
				`{"status": "%s", "log": %s, "scResults": %s, "timestamp": %s, "gasUsed": %d}},"upsert":%s}`,
				tx.Status, string(marshalizedLog), string(scResults), string(marshalizedTimestamp), tx.GasUsed, string(marshalledTx)))
	}

	log.Warn("log meta - update", "data", string(meta))
	log.Warn("log srld data - update", "data", string(serializedData))
	//log.Debug("index tx request", "request", string(serializedData))

	return
}

func prepareSerializedAccountInfo(address string, account *AccountInfo) (meta []byte, serializedData []byte) {
	var err error
	meta = []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, address, "\n"))
	serializedData, err = json.Marshal(account)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize account, will skip indexing",
			"address", address)
		return
	}

	return
}

func prepareSerializedAccountBalanceHistory(address string, account *AccountBalanceHistory) (meta []byte, serializedData []byte) {
	var err error
	meta = []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, address, "\n"))
	serializedData, err = json.Marshal(account)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize account history entry, will skip indexing",
			"address", address)
		return
	}

	return
}

func isCrossShardDstMe(tx *Transaction, selfShardID uint32) bool {
	return tx.SenderShard != tx.ReceiverShard && tx.ReceiverShard == selfShardID
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
func GetElasticTemplatesAndPolicies() (map[string]*bytes.Buffer, map[string]*bytes.Buffer) {
	indexTemplates := make(map[string]*bytes.Buffer)
	indexPolicies := make(map[string]*bytes.Buffer)

	indexes := []string{"opendistro", txIndex, blockIndex, miniblocksIndex, tpsIndex, ratingIndex, roundIndex, validatorsIndex, accountsIndex, accountsHistoryIndex}
	for _, index := range indexes {
		indexTemplates[index] = getTemplateByIndex(index)
	}

	indexesPolicies := []string{txPolicy, blockPolicy, miniblocksPolicy, tpsPolicy, ratingPolicy, roundPolicy, validatorsPolicy, accountsPolicy, accountsHistoryPolicy}
	for _, indexPolicy := range indexesPolicies {
		indexPolicies[indexPolicy] = getPolicyByIndex(indexPolicy)
	}

	return indexTemplates, indexPolicies
}

func getTemplateByIndex(index string) *bytes.Buffer {
	indexTemplate := &bytes.Buffer{}

	filePath := fmt.Sprintf("./config/elasticIndexTemplates/%s.json", index)
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Error("cannot read bytes from elastic template file", "path", filePath, "err", err)
		return &bytes.Buffer{}
	}

	indexTemplate.Grow(len(fileBytes))
	_, err = indexTemplate.Write(fileBytes)
	if err != nil {
		log.Error("getTemplateByIndex: cannot write bytes to buffer", "err", err)
		return &bytes.Buffer{}
	}

	return indexTemplate
}

func getPolicyByIndex(index string) *bytes.Buffer {
	indexPolicy := &bytes.Buffer{}

	filePath := fmt.Sprintf("./config/elasticIndexTemplates/%s.json", index)
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Error("cannot read bytes from elastic policy file", "path", filePath, "err", err)
		return &bytes.Buffer{}
	}

	indexPolicy.Grow(len(fileBytes))
	_, err = indexPolicy.Write(fileBytes)
	if err != nil {
		log.Error("getPolicyByIndex: cannot write bytes to buffer", "err", err)
		return &bytes.Buffer{}
	}

	return indexPolicy
}
