package transactionAPI

import (
	"bytes"
	"strings"

	"github.com/ElrondNetwork/elrond-go-core/core"
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
func (detector *refundDetector) isRefund(input refundDetectorInput) bool {
	hasValue := input.Value != "0" && input.Value != ""
	hasReturnCodeOK := detector.isReturnCodeOK(input.Data)
	isRefundForRelayTxSender := strings.Contains(input.ReturnMessage, core.GasRefundForRelayerMessage)
	isSuccessful := hasReturnCodeOK || isRefundForRelayTxSender

	return hasValue && isSuccessful
}

// Also see: https://github.com/ElrondNetwork/elastic-indexer-go/blob/master/process/transactions/checkers.go
func (detector *refundDetector) isReturnCodeOK(resultData []byte) bool {
	containsOk := bytes.Contains(resultData, []byte(okReturnCodeMarker))
	containsOkBackwardsCompatible := bytes.Contains(resultData, []byte(okReturnCodeMarkerBackwardsCompatible))

	return containsOk || containsOkBackwardsCompatible
}
