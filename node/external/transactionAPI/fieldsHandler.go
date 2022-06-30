package transactionAPI

import "strings"

const (
	hashField     = "hash"
	nonceField    = "nonce"
	senderField   = "sender"
	receiverField = "receiver"
	gasLimitField = "gaslimit"
	gasPriceField = "gasprice"
)

type fieldsHandler struct {
	HasNonce    bool
	HasSender   bool
	HasReceiver bool
	HasGasLimit bool
	HasGasPrice bool
}

func newFieldsHandler(parameters string) fieldsHandler {
	ph := fieldsHandler{
		HasNonce:    strings.Contains(parameters, nonceField),
		HasSender:   strings.Contains(parameters, senderField),
		HasReceiver: strings.Contains(parameters, receiverField),
		HasGasLimit: strings.Contains(parameters, gasLimitField),
		HasGasPrice: strings.Contains(parameters, gasPriceField),
	}
	return ph
}
