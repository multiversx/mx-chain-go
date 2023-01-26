package transactionAPI

import (
	"strings"
)

const (
	hashField        = "hash"
	nonceField       = "nonce"
	senderField      = "sender"
	receiverField    = "receiver"
	gasLimitField    = "gaslimit"
	gasPriceField    = "gasprice"
	rcvUsernameField = "receiverusername"
	dataField        = "data"
	valueField       = "value"
)

type fieldsHandler struct {
	HasNonce       bool
	HasSender      bool
	HasReceiver    bool
	HasGasLimit    bool
	HasGasPrice    bool
	HasRcvUsername bool
	HasData        bool
	HasValue       bool
}

func newFieldsHandler(parameters string) fieldsHandler {
	parameters = strings.ToLower(parameters)
	ph := fieldsHandler{
		HasNonce:       strings.Contains(parameters, nonceField),
		HasSender:      strings.Contains(parameters, senderField),
		HasReceiver:    strings.Contains(parameters, receiverField),
		HasGasLimit:    strings.Contains(parameters, gasLimitField),
		HasGasPrice:    strings.Contains(parameters, gasPriceField),
		HasRcvUsername: strings.Contains(parameters, rcvUsernameField),
		HasData:        strings.Contains(parameters, dataField),
		HasValue:       strings.Contains(parameters, valueField),
	}
	return ph
}
