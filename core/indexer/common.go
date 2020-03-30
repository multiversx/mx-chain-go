package indexer

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

type commonProcessor struct {
	pubkeyConverter state.PubkeyConverter
}

func checkElasticSearchParams(arguments ElasticIndexerArgs) error {
	if check.IfNil(arguments.PubkeyConverter) {
		return ErrNilPubkeyConverter
	}
	if arguments.Url == "" {
		return core.ErrNilUrl
	}
	if arguments.UserName == "" {
		return ErrEmptyUserName
	}
	if arguments.Password == "" {
		return ErrEmptyPassword
	}
	if check.IfNil(arguments.ShardCoordinator) {
		return core.ErrNilCoordinator
	}
	if check.IfNil(arguments.Marshalizer) {
		return core.ErrNilMarshalizer
	}
	if check.IfNil(arguments.Hasher) {
		return core.ErrNilHasher
	}

	return nil
}

func (cm *commonProcessor) timestampMapping() io.Reader {
	return strings.NewReader(
		`{
				"settings": {"index": {"sort.field": "timestamp", "sort.order": "desc"}},
				"mappings": {"_doc": {"properties": {"timestamp": {"type": "date"}}}}
			}`,
	)
}

func (cm *commonProcessor) prepareGeneralInfo(tpsBenchmark statistics.TPSBenchmark) bytes.Buffer {
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

func (cm *commonProcessor) serializeShardInfo(shardInfo statistics.ShardStatistic) ([]byte, []byte) {
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

func (cm *commonProcessor) getTransactionByType(
	tx data.TransactionHandler,
	txHash []byte,
	mbHash []byte,
	blockHash []byte,
	mb *block.MiniBlock,
	header data.HeaderHandler,
	txStatus string,
) (*Transaction, error) {
	switch currentType := tx.(type) {
	case *transaction.Transaction:
		return cm.buildTransaction(currentType, txHash, mbHash, blockHash, mb, header, txStatus)
	case *smartContractResult.SmartContractResult:
		return cm.buildSmartContractResult(currentType, txHash, mbHash, blockHash, mb, header)
	case *rewardTx.RewardTx:
		return cm.buildRewardTransaction(currentType, txHash, mbHash, blockHash, mb, header)
	case *receipt.Receipt:
		return cm.buildReceiptTransaction(currentType, txHash, mbHash, blockHash, mb, header)
	default:
		return nil, fmt.Errorf("%w, type %v", ErrUnknownTransactionHandler, currentType)
	}
}

func (cm *commonProcessor) buildTransaction(
	tx *transaction.Transaction,
	txHash []byte,
	mbHash []byte,
	blockHash []byte,
	mb *block.MiniBlock,
	header data.HeaderHandler,
	txStatus string,
) (*Transaction, error) {
	indexedTx := &Transaction{
		Hash:          hex.EncodeToString(txHash),
		MBHash:        hex.EncodeToString(mbHash),
		BlockHash:     hex.EncodeToString(blockHash),
		Nonce:         tx.Nonce,
		Round:         header.GetRound(),
		Value:         tx.Value.String(),
		ReceiverShard: mb.ReceiverShardID,
		SenderShard:   mb.SenderShardID,
		GasPrice:      tx.GasPrice,
		GasLimit:      tx.GasLimit,
		Data:          tx.Data,
		Signature:     hex.EncodeToString(tx.Signature),
		Timestamp:     time.Duration(header.GetTimeStamp()),
		Status:        txStatus,
	}

	var err error
	indexedTx.Receiver, err = cm.pubkeyConverter.String(tx.RcvAddr)
	if err != nil {
		return nil, err
	}

	indexedTx.Sender, err = cm.pubkeyConverter.String(tx.SndAddr)
	if err != nil {
		return nil, err
	}

	return indexedTx, nil
}

func (cm *commonProcessor) buildSmartContractResult(
	scr *smartContractResult.SmartContractResult,
	txHash []byte,
	mbHash []byte,
	blockHash []byte,
	mb *block.MiniBlock,
	header data.HeaderHandler,
) (*Transaction, error) {
	indexedTx := &Transaction{
		Hash:          hex.EncodeToString(txHash),
		MBHash:        hex.EncodeToString(mbHash),
		BlockHash:     hex.EncodeToString(blockHash),
		Nonce:         scr.Nonce,
		Round:         header.GetRound(),
		Value:         scr.Value.String(),
		ReceiverShard: mb.ReceiverShardID,
		SenderShard:   mb.SenderShardID,
		GasPrice:      scr.GasPrice,
		GasLimit:      scr.GasPrice,
		Data:          scr.Data,
		Signature:     "",
		Timestamp:     time.Duration(header.GetTimeStamp()),
		Status:        "Success",
	}

	var err error
	indexedTx.Receiver, err = cm.pubkeyConverter.String(scr.RcvAddr)
	if err != nil {
		return nil, err
	}

	indexedTx.Sender, err = cm.pubkeyConverter.String(scr.SndAddr)
	if err != nil {
		return nil, err
	}

	return indexedTx, nil
}

func (cm *commonProcessor) buildRewardTransaction(
	rTx *rewardTx.RewardTx,
	txHash []byte,
	mbHash []byte,
	blockHash []byte,
	mb *block.MiniBlock,
	header data.HeaderHandler,
) (*Transaction, error) {
	indexedTx := &Transaction{
		Hash:          hex.EncodeToString(txHash),
		MBHash:        hex.EncodeToString(mbHash),
		BlockHash:     hex.EncodeToString(blockHash),
		Nonce:         0,
		Round:         rTx.Round,
		Value:         rTx.Value.String(),
		Sender:        metachainTpsDocID,
		ReceiverShard: mb.ReceiverShardID,
		SenderShard:   mb.SenderShardID,
		GasPrice:      0,
		GasLimit:      0,
		Data:          []byte(""),
		Signature:     "",
		Timestamp:     time.Duration(header.GetTimeStamp()),
		Status:        "Success",
	}

	var err error
	indexedTx.Receiver, err = cm.pubkeyConverter.String(rTx.RcvAddr)
	if err != nil {
		return nil, err
	}

	return indexedTx, nil
}

func (cm *commonProcessor) buildReceiptTransaction(
	rpt *receipt.Receipt,
	txHash []byte,
	mbHash []byte,
	blockHash []byte,
	mb *block.MiniBlock,
	header data.HeaderHandler,
) (*Transaction, error) {
	indexedTx := &Transaction{
		Hash:          hex.EncodeToString(txHash),
		MBHash:        hex.EncodeToString(mbHash),
		BlockHash:     hex.EncodeToString(blockHash),
		Nonce:         rpt.GetNonce(),
		Round:         header.GetRound(),
		Value:         rpt.Value.String(),
		ReceiverShard: mb.ReceiverShardID,
		SenderShard:   mb.SenderShardID,
		GasPrice:      0,
		GasLimit:      0,
		Data:          rpt.Data,
		Signature:     "",
		Timestamp:     time.Duration(header.GetTimeStamp()),
		Status:        "Success",
	}

	var err error
	indexedTx.Receiver, err = cm.pubkeyConverter.String(rpt.GetRcvAddr())
	if err != nil {
		return nil, err
	}

	indexedTx.Sender, err = cm.pubkeyConverter.String(rpt.GetSndAddr())
	if err != nil {
		return nil, err
	}

	return indexedTx, nil
}
