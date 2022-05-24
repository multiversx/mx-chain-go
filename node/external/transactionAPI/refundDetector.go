package transactionAPI

import (
	"bytes"
	"encoding/hex"

	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type refundDetectorInput struct {
	Value         string
	Data          []byte
	ReturnMessage string
	GasLimit      uint64
}

type refundDetector struct {
}

func newRefundDetector() *refundDetector {
	return &refundDetector{}
}

// Also see: https://github.com/ElrondNetwork/elastic-indexer-go/blob/master/process/transactions/scrsDataToTransactions.go
func (detector *refundDetector) isRefund(result refundDetectorInput) bool {
	hasValue := result.Value != "0" && result.Value != ""
	hasNoGasLimit := result.GasLimit == 0
	hasReturnCodeOK := detector.isReturnCodeOK(result.Data)
	isRefundForRelayTxSender := result.ReturnMessage == gasRefundForRelayerMessage
	isSuccessful := hasReturnCodeOK || isRefundForRelayTxSender

	return hasValue && hasNoGasLimit && isSuccessful
}

// Also see: https://github.com/ElrondNetwork/elastic-indexer-go/blob/master/process/transactions/checkers.go
func (detector *refundDetector) isReturnCodeOK(resultData []byte) bool {
	okCode := vmcommon.Ok.String()
	okMarker := []byte("@" + hex.EncodeToString([]byte(okCode)))
	okMarkerBackwardsCompatible := []byte("@" + okCode)

	containsOk := bytes.Contains(resultData, okMarker)
	containsOkBackwardsCompatible := bytes.Contains(resultData, okMarkerBackwardsCompatible)

	return containsOk || containsOkBackwardsCompatible
}
