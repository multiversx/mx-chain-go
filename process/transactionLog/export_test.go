package transactionLog

import (
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

func NewPlainLogProcessor(args ArgTxLogProcessor) *txLogProcessor {
	return &txLogProcessor{
		marshalizer: args.Marshalizer,
		storer: args.Storer,
	}
}

func (tlp *txLogProcessor) ComputeEvent(vmLog *vmcommon.LogEntry) *transaction.Event {
	return tlp.computeEvent(vmLog)
}
