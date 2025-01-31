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
	relayerField           = "relayer"
	relayerSignatureField  = "relayersignature"
	wildCard               = "*"

	separator = ","
)

type fieldsHandler struct {
	fieldsMap map[string]struct{}
}

func newFieldsHandler(parameters string) fieldsHandler {
	if len(parameters) == 0 {
		return fieldsHandler{
			fieldsMap: map[string]struct{}{
				hashField: {}, // hash should always be returned
			},
		}
	}

	parameters = strings.ToLower(parameters)
	fieldsMap := sliceToMap(strings.Split(parameters, separator))
	fieldsMap[hashField] = struct{}{} // hashField should always be returned

	return fieldsHandler{
		fieldsMap: fieldsMap,
	}
}

// IsFieldSet returns true if the provided field is set
func (handler *fieldsHandler) IsFieldSet(field string) bool {
	_, hasWildCard := handler.fieldsMap[wildCard]
	if hasWildCard {
		return true
	}

	_, has := handler.fieldsMap[strings.ToLower(field)]
	return has
}

func sliceToMap(providedSlice []string) map[string]struct{} {
	result := make(map[string]struct{}, len(providedSlice))
	for _, entry := range providedSlice {
		result[entry] = struct{}{}
	}

	return result
}
