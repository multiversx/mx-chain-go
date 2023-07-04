package processorV2

import vmcommon "github.com/multiversx/mx-chain-vm-common-go"

type failureContext struct {
	processFail   bool
	txHash        []byte
	errorMessage  string
	returnMessage []byte
	snapshot      int
	gasLocked     uint64
	logs          []*vmcommon.LogEntry
}

func NewFailureContext() *failureContext {
	return &failureContext{
		processFail: false,
	}
}

func (failCtx *failureContext) makeSnaphot(snapshot int) {
	failCtx.snapshot = snapshot
}

func (failCtx *failureContext) setTxHash(txHash []byte) {
	failCtx.txHash = txHash
}

func (failCtx *failureContext) setGasLocked(gasLocked uint64) {
	failCtx.gasLocked = gasLocked
}

func (failCtx *failureContext) setMessage(errorMessage string) {
	failCtx.setMessages(errorMessage, []byte(errorMessage))
}

func (failCtx *failureContext) setMessagesFromError(err error) {
	failCtx.setMessage(err.Error())
}

func (failCtx *failureContext) setMessages(errorMessage string, returnMessage []byte) {
	failCtx.processFail = true
	failCtx.errorMessage = errorMessage
	failCtx.returnMessage = returnMessage
}

func (failCtx *failureContext) setLogs(logs []*vmcommon.LogEntry) {
	failCtx.logs = logs
}
