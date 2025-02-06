package transactionsfee

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

func (tep *transactionsFeeProcessor) isESDTOperationWithSCCall(tx data.TransactionHandler) bool {
	res := tep.dataFieldParser.Parse(tx.GetData(), tx.GetSndAddr(), tx.GetRcvAddr(), tep.shardCoordinator.NumberOfShards())

	isESDTTransferOperation := res.Operation == core.BuiltInFunctionESDTTransfer ||
		res.Operation == core.BuiltInFunctionESDTNFTTransfer || res.Operation == core.BuiltInFunctionMultiESDTNFTTransfer

	isReceiverSC := core.IsSmartContractAddress(tx.GetRcvAddr())
	hasFunction := res.Function != ""
	if !hasFunction {
		return false
	}

	if !bytes.Equal(tx.GetSndAddr(), tx.GetRcvAddr()) {
		return isESDTTransferOperation && isReceiverSC && hasFunction
	}

	if len(res.Receivers) == 0 {
		return false
	}

	isReceiverSC = core.IsSmartContractAddress(res.Receivers[0])

	return isESDTTransferOperation && isReceiverSC && hasFunction
}

func isSCRForSenderWithRefund(scr *smartContractResult.SmartContractResult, txHashHex string, tx data.TransactionHandler) bool {
	isForSender := bytes.Equal(scr.RcvAddr, tx.GetSndAddr())
	isRightNonce := scr.Nonce == tx.GetNonce()+1
	isFromCurrentTx := hex.EncodeToString(scr.PrevTxHash) == txHashHex
	isScrDataOk := isDataOk(scr.Data)

	return isFromCurrentTx && isForSender && isRightNonce && isScrDataOk
}

func isRefundForRelayed(dbScResult *smartContractResult.SmartContractResult, tx data.TransactionHandler) bool {
	isRelayedV3 := common.IsRelayedTxV3(tx)
	isForRelayed := string(dbScResult.ReturnMessage) == core.GasRefundForRelayerMessage
	isForSender := bytes.Equal(dbScResult.RcvAddr, tx.GetSndAddr())
	isForRelayerV3 := isForRelayerOfV3(dbScResult, tx)
	differentHash := !bytes.Equal(dbScResult.OriginalTxHash, dbScResult.PrevTxHash)

	isRefundForRelayedV1V2 := isForRelayed && isForSender && differentHash && !isRelayedV3
	isRefundForRelayedV3 := isForRelayed && isForRelayerV3 && isRelayedV3

	return isRefundForRelayedV1V2 || isRefundForRelayedV3
}

func isForRelayerOfV3(scr *smartContractResult.SmartContractResult, tx data.TransactionHandler) bool {
	relayedTx, isRelayedV3 := tx.(data.RelayedTransactionHandler)
	if !isRelayedV3 {
		return false
	}

	return bytes.Equal(relayedTx.GetRelayerAddr(), scr.RcvAddr)
}

func isDataOk(data []byte) bool {
	okReturnDataNewVersion := []byte("@" + hex.EncodeToString([]byte(vmcommon.Ok.String())))
	okReturnDataOldVersion := []byte("@" + vmcommon.Ok.String()) // backwards compatible

	return bytes.Contains(data, okReturnDataNewVersion) || bytes.Contains(data, okReturnDataOldVersion)
}

func isSCRWithRefundNoTx(scr *smartContractResult.SmartContractResult) bool {
	hasRefund := scr.Value.Cmp(big.NewInt(0)) != 0
	isSuccessful := isDataOk(scr.Data)
	isRefundForRelayTxSender := string(scr.ReturnMessage) == core.GasRefundForRelayerMessage

	ok := isSuccessful || isRefundForRelayTxSender
	differentHash := !bytes.Equal(scr.OriginalTxHash, scr.PrevTxHash)

	return ok && differentHash && hasRefund
}

func isRelayedTx(tx *transactionWithResults) bool {
	txData := string(tx.GetTxHandler().GetData())
	isRelayed := strings.HasPrefix(txData, core.RelayedTransaction) || strings.HasPrefix(txData, core.RelayedTransactionV2)
	return isRelayed && len(tx.scrs) > 0
}
