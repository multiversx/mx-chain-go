package transactionAPI

import (
	"bytes"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
)

// RefundDetectorInput will contain the needed input
type RefundDetectorInput struct {
	Value         string
	Data          []byte
	ReturnMessage string
	GasLimit      uint64
}

type refundDetector struct {
}

// NewRefundDetector will create a new instance of *refundDetector
func NewRefundDetector() *refundDetector {
	return &refundDetector{}
}

// IsRefund will verify if the provided input is a refund
// Also see: https://github.com/multiversx/mx-chain-es-indexer-go/blob/master/process/transactions/scrsDataToTransactions.go
func (detector *refundDetector) IsRefund(input RefundDetectorInput) bool {
	hasValue := input.Value != "0" && input.Value != ""
	hasReturnCodeOK := detector.isReturnCodeOK(input.Data)
	isRefundForRelayTxSender := strings.Contains(input.ReturnMessage, core.GasRefundForRelayerMessage)
	isSuccessful := hasReturnCodeOK || isRefundForRelayTxSender

	return hasValue && isSuccessful
}

// Also see: https://github.com/multiversx/mx-chain-es-indexer-go/blob/master/process/transactions/checkers.go
func (detector *refundDetector) isReturnCodeOK(resultData []byte) bool {
	containsOk := bytes.Contains(resultData, []byte(okReturnCodeMarker))
	containsOkBackwardsCompatible := bytes.Contains(resultData, []byte(okReturnCodeMarkerBackwardsCompatible))

	return containsOk || containsOkBackwardsCompatible
}
