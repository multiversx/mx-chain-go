package node

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go/core/dblookupext"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

func (n *Node) putEventsInTransaction(hash []byte, tx *transaction.ApiTransactionResult, epoch uint32) {
	eventsHashes, err := n.historyRepository.GetEventsHashesByTxHash(hash, epoch)
	if err != nil {
		return
	}

	if len(eventsHashes.ReceiptsHash) > 0 {
		// if transaction hash a receipt after receipt is added we can return
		n.putReceiptInTransaction(tx, eventsHashes.ReceiptsHash, epoch)
		return
	}

	n.putSmartContractResultsInTransaction(tx, eventsHashes.ScrHashesEpoch)
}

func (n *Node) putReceiptInTransaction(tx *transaction.ApiTransactionResult, recHash []byte, epoch uint32) {
	rec, err := n.getReceiptFromStorage(recHash, epoch)
	if err != nil {
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
	scrHashesEpoch []*dblookupext.ScrHashesAndEpoch,
) {
	for _, scrHashesE := range scrHashesEpoch {
		for _, scrHash := range scrHashesE.SmartContractResultsHashes {
			scr, err := n.getScrFromStorage(scrHash, scrHashesE.Epoch)
			if err != nil {
				continue
			}
			tx.ScResults = append(tx.ScResults, n.adaptSmartContractResult(scr))
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

func (n *Node) adaptSmartContractResult(scr *smartContractResult.SmartContractResult) *transaction.SmartContractResultApi {
	apiSCR := &transaction.SmartContractResultApi{
		Nonce:          scr.Nonce,
		Value:          scr.Value,
		RcvAddr:        n.addressPubkeyConverter.Encode(scr.RcvAddr),
		SndAddr:        n.addressPubkeyConverter.Encode(scr.SndAddr),
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

	if len(scr.RelayerAddr) != 0 {
		apiSCR.RelayerAddr = n.addressPubkeyConverter.Encode(scr.RelayerAddr)
	}

	if len(scr.OriginalSender) != 0 {
		apiSCR.OriginalSender = n.addressPubkeyConverter.Encode(scr.OriginalSender)
	}

	return apiSCR
}
