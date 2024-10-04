package processorV2

import (
	"github.com/multiversx/mx-chain-core-go/data"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

func appendOutcomeEmittedEvent(vmOutput *vmcommon.VMOutput, transaction data.TransactionHandler) {
	vmOutput.Logs = append(vmOutput.Logs, &vmcommon.LogEntry{
		Identifier: []byte(outcomeEmittedEvent),
		Address:    transaction.GetRcvAddr(),
		Topics: [][]byte{
			[]byte(vmOutput.ReturnCode.String()),
			[]byte(vmOutput.ReturnMessage),
		},
		Data: vmOutput.ReturnData,
	})
}
