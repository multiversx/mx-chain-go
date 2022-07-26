package transactionsfee

import (
	"bytes"
	"encoding/hex"
	"strings"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const (
	GasRefundForRelayerMessage = "gas refund for relayer"
)

func isSCRForSenderWithRefund(dbScResult *smartContractResult.SmartContractResult, txHash []byte, tx data.TransactionHandlerWithGasUsedAndFee) bool {
	isForSender := bytes.Equal(dbScResult.RcvAddr, tx.GetSndAddr())
	isRightNonce := dbScResult.Nonce == tx.GetNonce()+1
	isFromCurrentTx := bytes.Equal(dbScResult.PrevTxHash, txHash)
	isScrDataOk := isDataOk(dbScResult.Data)

	return isFromCurrentTx && isForSender && isRightNonce && isScrDataOk
}

func isRefundForRelayed(dbScResult *smartContractResult.SmartContractResult, tx data.TransactionHandlerWithGasUsedAndFee) bool {
	isForRelayed := string(dbScResult.ReturnMessage) == GasRefundForRelayerMessage
	isForSender := bytes.Equal(dbScResult.RcvAddr, tx.GetSndAddr())
	differentHash := !bytes.Equal(dbScResult.OriginalTxHash, dbScResult.PrevTxHash)

	return isForRelayed && isForSender && differentHash
}

func isDataOk(data []byte) bool {
	dataFieldStr := "@" + hex.EncodeToString([]byte(vmcommon.Ok.String()))

	return strings.HasPrefix(string(data), dataFieldStr)
}
