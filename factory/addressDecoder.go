package factory

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/errors"
)

func decodeAddresses(pkConverter core.PubkeyConverter, stringAddresses []string) ([][]byte, error) {
	if check.IfNil(pkConverter) {
		return nil, errors.ErrNilPubKeyConverter
	}
	decodedAddresses := make([][]byte, len(stringAddresses))
	for i, stringAddress := range stringAddresses {
		decodedAddress, errDecode := pkConverter.Decode(stringAddress)
		if errDecode != nil {
			return nil, errDecode
		}
		decodedAddresses[i] = decodedAddress
	}
	return decodedAddresses, nil
}
