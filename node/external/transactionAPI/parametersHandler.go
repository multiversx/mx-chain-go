package transactionAPI

import "strings"

const (
	hash     = "hash"
	nonce    = "nonce"
	sender   = "sender"
	receiver = "receiver"
	gaslimit = "gaslimit"
	gasprice = "gasprice"
)

type parametersHandler struct {
	HasHash     bool
	HasNonce    bool
	HasSender   bool
	HasReceiver bool
	HasGasLimit bool
	HasGasPrice bool
}

func newParametersHandler(parameters string) parametersHandler {
	ph := parametersHandler{
		HasHash:     strings.Contains(parameters, hash),
		HasNonce:    strings.Contains(parameters, nonce),
		HasSender:   strings.Contains(parameters, sender),
		HasReceiver: strings.Contains(parameters, receiver),
		HasGasLimit: strings.Contains(parameters, gaslimit),
		HasGasPrice: strings.Contains(parameters, gasprice),
	}
	return ph
}
