package node

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go/core/dblookupext"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node/filters"
)

func (n *Node) putResultsInTransaction(hash []byte, tx *transaction.ApiTransactionResult, epoch uint32) {
	n.putLogsInTransaction(hash, tx, epoch)

	resultsHashes, err := n.processComponents.HistoryRepository().GetResultsHashesByTxHash(hash, epoch)
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
	receiptsStorer := n.dataComponents.StorageService().GetStorer(dataRetriever.UnsignedTransactionUnit)
	receiptBytes, err := receiptsStorer.GetFromEpoch(hash, epoch)
	if err != nil {
		return nil, err
	}

	rec := &receipt.Receipt{}
	err = n.coreComponents.InternalMarshalizer().Unmarshal(rec, receiptBytes)
	if err != nil {
		return nil, err
	}

	return rec, nil
}

func (n *Node) adaptReceipt(rcpt *receipt.Receipt) *transaction.ReceiptApi {
	return &transaction.ReceiptApi{
		Value:   rcpt.Value,
		SndAddr: n.coreComponents.AddressPubKeyConverter().Encode(rcpt.SndAddr),
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

			scrAPI := n.adaptSmartContractResult(scrHash, scr)
			n.putLogsInSCR(scrHash, scrHashesE.Epoch, scrAPI)

			tx.SmartContractResults = append(tx.SmartContractResults, scrAPI)
		}
	}

	statusFilters := filters.NewStatusFilters(n.processComponents.ShardCoordinator().SelfId())
	statusFilters.SetStatusIfIsFailedESDTTransfer(tx)
}

func (n *Node) putLogsInTransaction(hash []byte, tx *transaction.ApiTransactionResult, epoch uint32) {
	logsAndEvents, err := n.getLogsAndEvents(hash, epoch)
	if err != nil {
		return
	}

	logsAPI := n.prepareLogsAndEvents(logsAndEvents)
	tx.Logs = logsAPI
}

func (n *Node) putLogsInSCR(scrHash []byte, epoch uint32, scr *transaction.ApiSmartContractResult) {
	logsAndEvents, err := n.getLogsAndEvents(scrHash, epoch)
	if err != nil {
		return
	}

	logsAPI := n.prepareLogsAndEvents(logsAndEvents)
	scr.Logs = logsAPI
}

func (n *Node) getScrFromStorage(hash []byte, epoch uint32) (*smartContractResult.SmartContractResult, error) {
	unsignedTxsStorer := n.dataComponents.StorageService().GetStorer(dataRetriever.UnsignedTransactionUnit)
	scrBytes, err := unsignedTxsStorer.GetFromEpoch(hash, epoch)
	if err != nil {
		return nil, err
	}

	scr := &smartContractResult.SmartContractResult{}
	err = n.coreComponents.InternalMarshalizer().Unmarshal(scr, scrBytes)
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

	addressPubKeyConverter := n.coreComponents.AddressPubKeyConverter()
	if len(scr.SndAddr) != 0 {
		apiSCR.SndAddr = addressPubKeyConverter.Encode(scr.SndAddr)
	}

	if len(scr.RcvAddr) != 0 {
		apiSCR.RcvAddr = addressPubKeyConverter.Encode(scr.RcvAddr)
	}

	if len(scr.RelayerAddr) != 0 {
		apiSCR.RelayerAddr = addressPubKeyConverter.Encode(scr.RelayerAddr)
	}

	if len(scr.OriginalSender) != 0 {
		apiSCR.OriginalSender = addressPubKeyConverter.Encode(scr.OriginalSender)
	}

	return apiSCR
}

func (n *Node) prepareLogsAndEvents(logsAndEvents *transaction.Log) *transaction.LogsAPI {
	addressPubKeyConverter := n.coreComponents.AddressPubKeyConverter()
	addrEncoded := addressPubKeyConverter.Encode(logsAndEvents.Address)

	logsAPI := &transaction.LogsAPI{
		Address: addrEncoded,
		Events:  make([]*transaction.Events, 0, len(logsAndEvents.Events)),
	}

	for _, event := range logsAndEvents.Events {
		logsAPI.Events = append(logsAPI.Events, &transaction.Events{
			Address:    addressPubKeyConverter.Encode(event.Address),
			Identifier: string(event.Identifier),
			Topics:     event.Topics,
			Data:       event.Data,
		})
	}

	return logsAPI
}

func (n *Node) getLogsAndEvents(hash []byte, epoch uint32) (*transaction.Log, error) {
	logsAndEventsStorer := n.dataComponents.StorageService().GetStorer(dataRetriever.TxLogsUnit)
	logsAndEventsBytes, err := logsAndEventsStorer.GetFromEpoch(hash, epoch)
	if err != nil {
		return nil, err
	}

	txLog := &transaction.Log{}
	err = n.coreComponents.InternalMarshalizer().Unmarshal(txLog, logsAndEventsBytes)
	if err != nil {
		return nil, err
	}

	return txLog, nil
}
