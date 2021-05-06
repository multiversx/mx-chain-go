package node

import (
	"bytes"
	"encoding/hex"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/dblookupext"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

func (n *Node) putResultsInTransaction(hash []byte, tx *transaction.ApiTransactionResult, epoch uint32) {
	resultsHashes, err := n.historyRepository.GetResultsHashesByTxHash(hash, epoch)
	if err != nil {
		return
	}

	if len(resultsHashes.ReceiptsHash) > 0 {
		n.putReceiptInTransaction(tx, resultsHashes.ReceiptsHash, epoch)
		return
	}

	n.putSmartContractResultsInTransaction(tx, resultsHashes.ScResultsHashesAndEpoch)
}

func (n *Node) putReceiptInTransaction(tx *transaction.ApiTransactionResult, recHash []byte, epoch uint32) {
	rec, err := n.getReceiptFromStorage(recHash, epoch)
	if err != nil {
		log.Warn("nodeTransactionEvents.putReceiptInTransaction() cannot get receipt from storage",
			"hash", hex.EncodeToString(recHash))
		return
	}

	tx.Receipt = n.adaptReceipt(rec)
}

func (n *Node) getReceiptFromStorage(hash []byte, epoch uint32) (*receipt.Receipt, error) {
	receiptsStorer := n.store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	receiptBytes, err := receiptsStorer.GetFromEpoch(hash, epoch)
	if err != nil {
		return nil, err
	}

	rec := &receipt.Receipt{}
	err = n.internalMarshalizer.Unmarshal(rec, receiptBytes)
	if err != nil {
		return nil, err
	}

	return rec, nil
}

func (n *Node) adaptReceipt(rcpt *receipt.Receipt) *transaction.ReceiptApi {
	return &transaction.ReceiptApi{
		Value:   rcpt.Value,
		SndAddr: n.addressPubkeyConverter.Encode(rcpt.SndAddr),
		Data:    string(rcpt.Data),
		TxHash:  hex.EncodeToString(rcpt.TxHash),
	}
}

func (n *Node) putSmartContractResultsInTransaction(
	tx *transaction.ApiTransactionResult,
	scrHashesEpoch []*dblookupext.ScResultsHashesAndEpoch,
) {
	for _, scrHashesE := range scrHashesEpoch {
		for _, scrHash := range scrHashesE.ScResultsHashes {
			scr, err := n.getScrFromStorage(scrHash, scrHashesE.Epoch)
			if err != nil {
				log.Warn("putSmartContractResultsInTransaction cannot get result from storage",
					"hash", hex.EncodeToString(scrHash),
					"error", err.Error())
				continue
			}

			tx.SmartContractResults = append(tx.SmartContractResults, n.adaptSmartContractResult(scrHash, scr))
		}
	}

	n.setStatusIfIsESDTTransferFail(tx)
}

func (n *Node) setStatusIfIsESDTTransferFail(tx *transaction.ApiTransactionResult) {
	if len(tx.SmartContractResults) < 1 {
		return
	}

	// check if cross shard destination me
	if !(tx.SourceShard != tx.DestinationShard && n.shardCoordinator.SelfId() == tx.DestinationShard) {
		return
	}

	// check if is an ESDT transfer
	if !strings.HasPrefix(string(tx.Data), core.BuiltInFunctionESDTTransfer) {
		return
	}

	for _, scr := range tx.SmartContractResults {
		if bytes.HasPrefix([]byte(scr.Data), tx.Data) && scr.Nonce == tx.Nonce {
			tx.Status = transaction.TxStatusFail
			return
		}
	}
}

func (n *Node) getScrFromStorage(hash []byte, epoch uint32) (*smartContractResult.SmartContractResult, error) {
	unsignedTxsStorer := n.store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	scrBytes, err := unsignedTxsStorer.GetFromEpoch(hash, epoch)
	if err != nil {
		return nil, err
	}

	scr := &smartContractResult.SmartContractResult{}
	err = n.internalMarshalizer.Unmarshal(scr, scrBytes)
	if err != nil {
		return nil, err
	}

	return scr, nil
}

func (n *Node) adaptSmartContractResult(scrHash []byte, scr *smartContractResult.SmartContractResult) *transaction.ApiSmartContractResult {
	apiSCR := &transaction.ApiSmartContractResult{
		Hash:           hex.EncodeToString(scrHash),
		Nonce:          scr.Nonce,
		Value:          scr.Value,
		RelayedValue:   scr.RelayedValue,
		Code:           string(scr.Code),
		Data:           string(scr.Data),
		PrevTxHash:     hex.EncodeToString(scr.PrevTxHash),
		OriginalTxHash: hex.EncodeToString(scr.OriginalTxHash),
		GasLimit:       scr.GasLimit,
		GasPrice:       scr.GasPrice,
		CallType:       scr.CallType,
		CodeMetadata:   string(scr.CodeMetadata),
		ReturnMessage:  string(scr.ReturnMessage),
	}

	if len(scr.SndAddr) != 0 {
		apiSCR.SndAddr = n.addressPubkeyConverter.Encode(scr.SndAddr)
	}

	if len(scr.RcvAddr) != 0 {
		apiSCR.RcvAddr = n.addressPubkeyConverter.Encode(scr.RcvAddr)
	}

	if len(scr.RelayerAddr) != 0 {
		apiSCR.RelayerAddr = n.addressPubkeyConverter.Encode(scr.RelayerAddr)
	}

	if len(scr.OriginalSender) != 0 {
		apiSCR.OriginalSender = n.addressPubkeyConverter.Encode(scr.OriginalSender)
	}

	return apiSCR
}
