package processorV2

import vmcommon "github.com/ElrondNetwork/elrond-vm-common"

type failureContext struct {
	sc            *scProcessor
	processFail   bool
	txHash        []byte
	errorMessage  string
	returnMessage []byte
	snapshot      int
	gasLocked     uint64
	logs          []*vmcommon.LogEntry
}

func NewFailureContext(sc *scProcessor) *failureContext {
	return &failureContext{
		sc:          sc,
		processFail: false,
	}
}

func (failCtx *failureContext) makeSnaphot() *failureContext {
	failCtx.snapshot = failCtx.sc.accounts.JournalLen()
	return failCtx
}

func (failCtx *failureContext) setTxHash(txHash []byte) *failureContext {
	failCtx.txHash = txHash
	return failCtx
}

func (failCtx *failureContext) setGasLocked(gasLocked uint64) *failureContext {
	failCtx.gasLocked = gasLocked
	return failCtx
}

func (failCtx *failureContext) setMessage(errorMessage string) *failureContext {
	failCtx.setMessages(errorMessage, []byte(errorMessage))
	return failCtx
}

func (failCtx *failureContext) setMessagesFromError(err error) *failureContext {
	failCtx.setMessage(err.Error())
	return failCtx
}

func (failCtx *failureContext) setMessages(errorMessage string, returnMessage []byte) *failureContext {
	failCtx.processFail = true
	failCtx.errorMessage = errorMessage
	failCtx.returnMessage = returnMessage
	return failCtx
}

func (failCtx *failureContext) setLogs(logs []*vmcommon.LogEntry) *failureContext {
	failCtx.logs = logs
	return failCtx
}
