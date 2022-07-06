package transactionAPI

import (
	"encoding/hex"
	"errors"
	"fmt"

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
	refundDetector         *refundDetector
	logsFacade             LogsFacade
}

func newAPITransactionResultProcessor(
	addressPubKeyConverter core.PubkeyConverter,
	historyRepository dblookupext.HistoryRepository,
	storageService dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	txUnmarshaller *txUnmarshaller,
	logsFacade LogsFacade,
	selfShardID uint32,
	dataFieldParser DataFieldParser,
) *apiTransactionResultsProcessor {
	refundDetector := newRefundDetector()

	return &apiTransactionResultsProcessor{
		txUnmarshaller:         txUnmarshaller,
		addressPubKeyConverter: addressPubKeyConverter,
		historyRepository:      historyRepository,
		storageService:         storageService,
		marshalizer:            marshalizer,
		selfShardID:            selfShardID,
		refundDetector:         refundDetector,
		logsFacade:             logsFacade,
		dataFieldParser:        dataFieldParser,
	}
}

func (arp *apiTransactionResultsProcessor) putResultsInTransaction(hash []byte, tx *transaction.ApiTransactionResult, epoch uint32) error {
	// TODO: Note that the following call produces an effect even if the function "putResultsInTransaction" results in an error.
	// TODO: Refactor this package to use less functions with side-effects.
	arp.loadLogsIntoTransaction(hash, tx, epoch)

	resultsHashes, err := arp.historyRepository.GetResultsHashesByTxHash(hash, epoch)
	if err != nil {
		// It's perfectly normal to have transactions without SCRs.
		if errors.Is(err, dblookupext.ErrNotFoundInStorage) {
			return nil
		}
		return err
	}

	if len(resultsHashes.ReceiptsHash) > 0 {
		return arp.putReceiptInTransaction(tx, resultsHashes.ReceiptsHash, epoch)
	}

	return arp.putSmartContractResultsInTransaction(tx, resultsHashes.ScResultsHashesAndEpoch)
}

func (arp *apiTransactionResultsProcessor) putReceiptInTransaction(tx *transaction.ApiTransactionResult, receiptHash []byte, epoch uint32) error {
	rec, err := arp.getReceiptFromStorage(receiptHash, epoch)
	if err != nil {
		return fmt.Errorf("%w: %v, hash = %s", errCannotLoadReceipts, err, hex.EncodeToString(receiptHash))
	}

	tx.Receipt = rec
	return nil
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
) error {
	for _, scrHashesE := range scrHashesEpoch {
		err := arp.putSmartContractResultsInTransactionByHashesAndEpoch(tx, scrHashesE.ScResultsHashes, scrHashesE.Epoch)
		if err != nil {
			return err
		}
	}

	statusFilters := filters.NewStatusFilters(arp.selfShardID)
	statusFilters.SetStatusIfIsFailedESDTTransfer(tx)
	return nil
}

func (arp *apiTransactionResultsProcessor) putSmartContractResultsInTransactionByHashesAndEpoch(tx *transaction.ApiTransactionResult, scrsHashes [][]byte, epoch uint32) error {
	for _, scrHash := range scrsHashes {
		scr, err := arp.getScrFromStorage(scrHash, epoch)
		if err != nil {
			return fmt.Errorf("%w: %v, hash = %s", errCannotLoadContractResults, err, hex.EncodeToString(scrHash))
		}

		scrAPI := arp.adaptSmartContractResult(scrHash, scr)
		arp.loadLogsIntoContractResults(scrHash, epoch, scrAPI)

		tx.SmartContractResults = append(tx.SmartContractResults, scrAPI)
	}

	return nil
}

func (arp *apiTransactionResultsProcessor) loadLogsIntoTransaction(hash []byte, tx *transaction.ApiTransactionResult, epoch uint32) {
	var err error

	tx.Logs, err = arp.logsFacade.GetLog(hash, epoch)
	if err != nil {
		log.Trace("loadLogsIntoTransaction()", "hash", hash, "epoch", epoch, "err", err)
	}
}

func (arp *apiTransactionResultsProcessor) loadLogsIntoContractResults(scrHash []byte, epoch uint32, scr *transaction.ApiSmartContractResult) {
	var err error

	scr.Logs, err = arp.logsFacade.GetLog(scrHash, epoch)
	if err != nil {
		log.Trace("loadLogsIntoContractResults()", "hash", scrHash, "epoch", epoch, "err", err)
	}
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
	isRefund := arp.refundDetector.isRefund(refundDetectorInput{
		Value:         scr.Value.String(),
		Data:          scr.Data,
		ReturnMessage: string(scr.ReturnMessage),
		GasLimit:      scr.GasLimit,
	})

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
		IsRefund:       isRefund,
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
