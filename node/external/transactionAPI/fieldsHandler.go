package transactionAPI

import (
	"strings"
)

const (
	hashField              = "hash"
	nonceField             = "nonce"
	senderField            = "sender"
	receiverField          = "receiver"
	gasLimitField          = "gaslimit"
	gasPriceField          = "gasprice"
	rcvUsernameField       = "receiverusername"
	dataField              = "data"
	valueField             = "value"
	signatureField         = "signature"
	guardianField          = "guardian"
	guardianSignatureField = "guardiansignature"
	senderShardID          = "sendershard"
	receiverShardID        = "receivershard"
	wildCard               = "*"

	separator = ","
)

type fieldsHandler struct {
	HasNonce             bool
	HasSender            bool
	HasReceiver          bool
	HasGasLimit          bool
	HasGasPrice          bool
	HasRcvUsername       bool
	HasData              bool
	HasValue             bool
	HasSignature         bool
	HasSenderShardID     bool
	HasReceiverShardID   bool
	HasGuardian          bool
	HasGuardianSignature bool
}

func newFieldsHandler(parameters string) fieldsHandler {
	parameters = strings.ToLower(parameters)
	parametersMap := sliceToMap(strings.Split(parameters, separator))
	ph := fieldsHandler{
		HasNonce:             shouldConsiderField(parametersMap, nonceField),
		HasSender:            shouldConsiderField(parametersMap, senderField),
		HasReceiver:          shouldConsiderField(parametersMap, receiverField),
		HasGasLimit:          shouldConsiderField(parametersMap, gasLimitField),
		HasGasPrice:          shouldConsiderField(parametersMap, gasPriceField),
		HasRcvUsername:       shouldConsiderField(parametersMap, rcvUsernameField),
		HasData:              shouldConsiderField(parametersMap, dataField),
		HasValue:             shouldConsiderField(parametersMap, valueField),
		HasSignature:         shouldConsiderField(parametersMap, signatureField),
		HasSenderShardID:     shouldConsiderField(parametersMap, senderShardID),
		HasReceiverShardID:   shouldConsiderField(parametersMap, receiverShardID),
		HasGuardian:          shouldConsiderField(parametersMap, guardianField),
		HasGuardianSignature: shouldConsiderField(parametersMap, guardianSignatureField),
	}
	return ph
}

func shouldConsiderField(parametersMap map[string]struct{}, field string) bool {
	_, hasWildCard := parametersMap[wildCard]
	if hasWildCard {
		return true
	}

	_, has := parametersMap[field]
	return has
}

func sliceToMap(providedSlice []string) map[string]struct{} {
	result := make(map[string]struct{}, len(providedSlice))
	for _, entry := range providedSlice {
		result[entry] = struct{}{}
	}

	return result
}
