package transactions

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type txDBBuilder struct {
	esdtProc               *esdtTransactionProcessor
	addressPubkeyConverter core.PubkeyConverter
	shardCoordinator       sharding.Coordinator
	txFeeCalculator        process.TransactionFeeCalculator
}

func newTransactionDBBuilder(
	addressPubkeyConverter core.PubkeyConverter,
	shardCoordinator sharding.Coordinator,
	txFeeCalculator process.TransactionFeeCalculator,
) *txDBBuilder {
	esdtProc := newEsdtTransactionHandler()

	return &txDBBuilder{
		esdtProc:               esdtProc,
		addressPubkeyConverter: addressPubkeyConverter,
		shardCoordinator:       shardCoordinator,
		txFeeCalculator:        txFeeCalculator,
	}
}

func (tbb *txDBBuilder) buildTransaction(
	tx *transaction.Transaction,
	txHash []byte,
	mbHash []byte,
	mb *block.MiniBlock,
	header data.HeaderHandler,
	txStatus string,
) *types.Transaction {
	var tokenIdentifier, esdtValue string
	if isESDTTx := tbb.esdtProc.isESDTTx(tx); isESDTTx {
		tokenIdentifier, esdtValue = tbb.esdtProc.getTokenIdentifierAndValue(tx)
	}

	gasUsed := tbb.txFeeCalculator.ComputeGasLimit(tx)
	fee := tbb.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, gasUsed)

	return &types.Transaction{
		Hash:                hex.EncodeToString(txHash),
		MBHash:              hex.EncodeToString(mbHash),
		Nonce:               tx.Nonce,
		Round:               header.GetRound(),
		Value:               tx.Value.String(),
		Receiver:            tbb.addressPubkeyConverter.Encode(tx.RcvAddr),
		Sender:              tbb.addressPubkeyConverter.Encode(tx.SndAddr),
		ReceiverShard:       mb.ReceiverShardID,
		SenderShard:         mb.SenderShardID,
		GasPrice:            tx.GasPrice,
		GasLimit:            tx.GasLimit,
		Data:                tx.Data,
		Signature:           hex.EncodeToString(tx.Signature),
		Timestamp:           time.Duration(header.GetTimeStamp()),
		Status:              txStatus,
		EsdtTokenIdentifier: tokenIdentifier,
		EsdtValue:           esdtValue,
		GasUsed:             gasUsed,
		Fee:                 fee.String(),
		ReceiverUserName:    tx.RcvUserName,
		SenderUserName:      tx.SndUserName,
		RcvAddrBytes:        tx.RcvAddr,
	}
}

func (tbb *txDBBuilder) buildRewardTransaction(
	rTx *rewardTx.RewardTx,
	txHash []byte,
	mbHash []byte,
	mb *block.MiniBlock,
	header data.HeaderHandler,
	txStatus string,
) *types.Transaction {
	return &types.Transaction{
		Hash:          hex.EncodeToString(txHash),
		MBHash:        hex.EncodeToString(mbHash),
		Nonce:         0,
		Round:         rTx.Round,
		Value:         rTx.Value.String(),
		Receiver:      tbb.addressPubkeyConverter.Encode(rTx.RcvAddr),
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

func (tbb *txDBBuilder) convertScResultInDatabaseScr(
	scHash string,
	sc *smartContractResult.SmartContractResult,
	header data.HeaderHandler,
) *types.ScResult {
	relayerAddr := ""
	if len(sc.RelayerAddr) > 0 {
		relayerAddr = tbb.addressPubkeyConverter.Encode(sc.RelayerAddr)
	}

	var tokenIdentifier, esdtValue string

	isESDTTx := tbb.esdtProc.isESDTTx(sc)
	if isESDTTx {
		tokenIdentifier, esdtValue = tbb.esdtProc.getTokenIdentifierAndValue(sc)
	}

	return &types.ScResult{
		Hash:                hex.EncodeToString([]byte(scHash)),
		Nonce:               sc.Nonce,
		GasLimit:            sc.GasLimit,
		GasPrice:            sc.GasPrice,
		Value:               sc.Value.String(),
		Sender:              tbb.addressPubkeyConverter.Encode(sc.SndAddr),
		Receiver:            tbb.addressPubkeyConverter.Encode(sc.RcvAddr),
		RelayerAddr:         relayerAddr,
		RelayedValue:        sc.RelayedValue.String(),
		Code:                string(sc.Code),
		Data:                sc.Data,
		PreTxHash:           hex.EncodeToString(sc.PrevTxHash),
		OriginalTxHash:      hex.EncodeToString(sc.OriginalTxHash),
		CallType:            strconv.Itoa(int(sc.CallType)),
		CodeMetadata:        sc.CodeMetadata,
		ReturnMessage:       string(sc.ReturnMessage),
		EsdtTokenIdentifier: tokenIdentifier,
		EsdtValue:           esdtValue,
		Timestamp:           time.Duration(header.GetTimeStamp()),
	}
}

func (tbb *txDBBuilder) convertReceiptInDatabaseReceipt(
	recHash string,
	rec *receipt.Receipt,
	header data.HeaderHandler,
) *types.Receipt {
	return &types.Receipt{
		Hash:      hex.EncodeToString([]byte(recHash)),
		Value:     rec.Value.String(),
		Sender:    tbb.addressPubkeyConverter.Encode(rec.SndAddr),
		Data:      string(rec.Data),
		TxHash:    hex.EncodeToString(rec.TxHash),
		Timestamp: time.Duration(header.GetTimeStamp()),
	}
}

func (tbb *txDBBuilder) addScrsReceiverToAlteredAccounts(
	alteredAddress map[string]*types.AlteredAccount,
	scrs []*types.ScResult,
) {
	for _, scr := range scrs {
		receiverAddr, _ := tbb.addressPubkeyConverter.Decode(scr.Receiver)
		shardID := tbb.shardCoordinator.ComputeId(receiverAddr)
		if shardID != tbb.shardCoordinator.SelfId() {
			continue
		}

		egldBalanceNotChanged := scr.Value == "" || scr.Value == "0"
		esdtBalanceNotChanged := scr.EsdtValue == "" || scr.EsdtValue == "0"
		if egldBalanceNotChanged && esdtBalanceNotChanged {
			// the smart contract results that don't alter the balance of the receiver address should be ignored
			continue
		}
		encodedReceiverAddress := scr.Receiver
		alteredAddress[encodedReceiverAddress] = &types.AlteredAccount{
			IsESDTOperation: scr.EsdtTokenIdentifier != "" && scr.EsdtValue != "",
			TokenIdentifier: scr.EsdtTokenIdentifier,
		}
	}
}
