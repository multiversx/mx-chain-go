package transactionsfee

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

func isSCRForSenderWithRefund(scr *smartContractResult.SmartContractResult, txHash []byte, tx data.TransactionHandlerWithGasUsedAndFee) bool {
	isForSender := bytes.Equal(scr.RcvAddr, tx.GetSndAddr())
	isRightNonce := scr.Nonce == tx.GetNonce()+1
	isFromCurrentTx := bytes.Equal(scr.PrevTxHash, txHash)
	isScrDataOk := isDataOk(scr.Data)

	return isFromCurrentTx && isForSender && isRightNonce && isScrDataOk
}

func isRefundForRelayed(dbScResult *smartContractResult.SmartContractResult, tx data.TransactionHandlerWithGasUsedAndFee) bool {
	isForRelayed := string(dbScResult.ReturnMessage) == core.GasRefundForRelayerMessage
	isForSender := bytes.Equal(dbScResult.RcvAddr, tx.GetSndAddr())
	differentHash := !bytes.Equal(dbScResult.OriginalTxHash, dbScResult.PrevTxHash)

	return isForRelayed && isForSender && differentHash
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
