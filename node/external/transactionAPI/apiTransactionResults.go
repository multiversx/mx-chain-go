package transactionAPI

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dblookupext"
	"github.com/ElrondNetwork/elrond-go/node/filters"
	datafield "github.com/ElrondNetwork/elrond-vm-common/parsers/dataField"
)

type apiTransactionResultsProcessor struct {
	txUnmarshaller         *txUnmarshaller
	addressPubKeyConverter core.PubkeyConverter
	historyRepository      dblookupext.HistoryRepository
	storageService         dataRetriever.StorageService
	marshalizer            marshal.Marshalizer
	dataFieldParser        DataFieldParser
	selfShardID            uint32
}

func newAPITransactionResultProcessor(
	addressPubKeyConverter core.PubkeyConverter,
	historyRepository dblookupext.HistoryRepository,
	storageService dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	txUnmarshaller *txUnmarshaller,
	selfShardID uint32,
	dataFieldParser DataFieldParser,
) *apiTransactionResultsProcessor {
	return &apiTransactionResultsProcessor{
		txUnmarshaller:         txUnmarshaller,
		addressPubKeyConverter: addressPubKeyConverter,
		historyRepository:      historyRepository,
		storageService:         storageService,
		marshalizer:            marshalizer,
		selfShardID:            selfShardID,
		dataFieldParser:        dataFieldParser,
	}
}

func (arp *apiTransactionResultsProcessor) putResultsInTransaction(hash []byte, tx *transaction.ApiTransactionResult, epoch uint32) {
	arp.putLogsInTransaction(hash, tx, epoch)

	resultsHashes, err := arp.historyRepository.GetResultsHashesByTxHash(hash, epoch)
	if err != nil {
		return
	}

	if len(resultsHashes.ReceiptsHash) > 0 {
		arp.putReceiptInTransaction(tx, resultsHashes.ReceiptsHash, epoch)
		return
	}

	arp.putSmartContractResultsInTransaction(tx, resultsHashes.ScResultsHashesAndEpoch)
}

func (arp *apiTransactionResultsProcessor) putReceiptInTransaction(tx *transaction.ApiTransactionResult, receiptHash []byte, epoch uint32) {
	rec, err := arp.getReceiptFromStorage(receiptHash, epoch)
	if err != nil {
		log.Warn("nodeTransactionEvents.putReceiptInTransaction() cannot get receipt from storage",
			"hash", receiptHash,
			"error", err.Error())
		return
	}

	tx.Receipt = rec
}

func (arp *apiTransactionResultsProcessor) getReceiptFromStorage(hash []byte, epoch uint32) (*transaction.ApiReceipt, error) {
	receiptsStorer := arp.storageService.GetStorer(dataRetriever.UnsignedTransactionUnit)
	receiptBytes, err := receiptsStorer.GetFromEpoch(hash, epoch)
	if err != nil {
		return nil, err
	}

	return arp.txUnmarshaller.unmarshalReceipt(receiptBytes)
}

func (arp *apiTransactionResultsProcessor) putSmartContractResultsInTransaction(
	tx *transaction.ApiTransactionResult,
	scrHashesEpoch []*dblookupext.ScResultsHashesAndEpoch,
) {
	for _, scrHashesE := range scrHashesEpoch {
		arp.putSmartContractResultsInTransactionByHashesAndEpoch(tx, scrHashesE.ScResultsHashes, scrHashesE.Epoch)
	}

	statusFilters := filters.NewStatusFilters(arp.selfShardID)
	statusFilters.SetStatusIfIsFailedESDTTransfer(tx)
}

func (arp *apiTransactionResultsProcessor) putSmartContractResultsInTransactionByHashesAndEpoch(tx *transaction.ApiTransactionResult, scrsHashes [][]byte, epoch uint32) {
	for _, scrHash := range scrsHashes {
		scr, err := arp.getScrFromStorage(scrHash, epoch)
		if err != nil {
			log.Warn("putSmartContractResultsInTransaction cannot get result from storage",
				"hash", scrHash,
				"error", err.Error())
			continue
		}

		scrAPI := arp.adaptSmartContractResult(scrHash, scr)
		arp.putLogsInSCR(scrHash, epoch, scrAPI)

		tx.SmartContractResults = append(tx.SmartContractResults, scrAPI)
	}
}

func (arp *apiTransactionResultsProcessor) putLogsInTransaction(hash []byte, tx *transaction.ApiTransactionResult, epoch uint32) {
	logsAndEvents, err := arp.getLogsAndEvents(hash, epoch)
	if err != nil || logsAndEvents == nil {
		return
	}

	logsAPI := arp.prepareLogsAndEvents(logsAndEvents)
	tx.Logs = logsAPI
}

func (arp *apiTransactionResultsProcessor) putLogsInSCR(scrHash []byte, epoch uint32, scr *transaction.ApiSmartContractResult) {
	logsAndEvents, err := arp.getLogsAndEvents(scrHash, epoch)
	if err != nil {
		return
	}

	logsAPI := arp.prepareLogsAndEvents(logsAndEvents)
	scr.Logs = logsAPI
}

func (arp *apiTransactionResultsProcessor) getScrFromStorage(hash []byte, epoch uint32) (*smartContractResult.SmartContractResult, error) {
	unsignedTxsStorer := arp.storageService.GetStorer(dataRetriever.UnsignedTransactionUnit)
	scrBytes, err := unsignedTxsStorer.GetFromEpoch(hash, epoch)
	if err != nil {
		return nil, err
	}

	scr := &smartContractResult.SmartContractResult{}
	err = arp.marshalizer.Unmarshal(scr, scrBytes)
	if err != nil {
		return nil, err
	}

	return scr, nil
}

func (arp *apiTransactionResultsProcessor) adaptSmartContractResult(scrHash []byte, scr *smartContractResult.SmartContractResult) *transaction.ApiSmartContractResult {
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

	if len(scr.SndAddr) == arp.addressPubKeyConverter.Len() {
		apiSCR.SndAddr = arp.addressPubKeyConverter.Encode(scr.SndAddr)
	}

	if len(scr.RcvAddr) == arp.addressPubKeyConverter.Len() {
		apiSCR.RcvAddr = arp.addressPubKeyConverter.Encode(scr.RcvAddr)
	}

	if len(scr.RelayerAddr) == arp.addressPubKeyConverter.Len() {
		apiSCR.RelayerAddr = arp.addressPubKeyConverter.Encode(scr.RelayerAddr)
	}

	if len(scr.OriginalSender) == arp.addressPubKeyConverter.Len() {
		apiSCR.OriginalSender = arp.addressPubKeyConverter.Encode(scr.OriginalSender)
	}

	res := arp.dataFieldParser.Parse(scr.Data, scr.GetSndAddr(), scr.GetRcvAddr())
	apiSCR.Operation = res.Operation
	apiSCR.Function = res.Function
	apiSCR.ESDTValues = res.ESDTValues
	apiSCR.Tokens = res.Tokens
	apiSCR.Receivers = datafield.EncodeBytesSlice(arp.addressPubKeyConverter.Encode, res.Receivers)
	apiSCR.ReceiversShardIDs = res.ReceiversShardID
	apiSCR.IsRelayed = res.IsRelayed

	return apiSCR
}

func (arp *apiTransactionResultsProcessor) prepareLogsAndEvents(logsAndEvents *transaction.Log) *transaction.ApiLogs {
	addrEncoded := arp.addressPubKeyConverter.Encode(logsAndEvents.Address)

	logsAPI := &transaction.ApiLogs{
		Address: addrEncoded,
		Events:  make([]*transaction.Events, 0, len(logsAndEvents.Events)),
	}

	for _, event := range logsAndEvents.Events {
		logsAPI.Events = append(logsAPI.Events, &transaction.Events{
			Address:    arp.addressPubKeyConverter.Encode(event.Address),
			Identifier: string(event.Identifier),
			Topics:     event.Topics,
			Data:       event.Data,
		})
	}

	return logsAPI
}

func (arp *apiTransactionResultsProcessor) getLogsAndEvents(hash []byte, epoch uint32) (*transaction.Log, error) {
	logsAndEventsStorer := arp.storageService.GetStorer(dataRetriever.TxLogsUnit)
	logsAndEventsBytes, err := logsAndEventsStorer.GetFromEpoch(hash, epoch)
	if err != nil {
		return nil, err
	}

	txLog := &transaction.Log{}
	err = arp.marshalizer.Unmarshal(txLog, logsAndEventsBytes)
	if err != nil {
		return nil, err
	}

	return txLog, nil
}
