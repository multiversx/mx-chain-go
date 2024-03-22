package systemSmartContracts

import "github.com/multiversx/mx-chain-core-go/core/check"

type sovereignESDT struct {
	baseTokenIDCreator  TokenIdentifierCreatorHandler
	prefixWithSeparator []byte
}

func NewSovereignTokenIDCreator(tokenIDCreator TokenIdentifierCreatorHandler, prefix string) (*sovereignESDT, error) {
	if check.IfNil(tokenIDCreator) {
		return nil, nil
	}

	return &sovereignESDT{
		baseTokenIDCreator:  tokenIDCreator,
		prefixWithSeparator: append([]byte(prefix), []byte(tickerSeparator)...),
	}, nil
}

func (se *sovereignESDT) CreateNewTokenIdentifier(caller []byte, ticker []byte) ([]byte, error) {
	tokenID, err := se.baseTokenIDCreator.CreateNewTokenIdentifier(caller, ticker)
	if err != nil {
		return nil, err
	}

	return append(se.prefixWithSeparator, tokenID...), nil
}

func (se *sovereignESDT) IsInterfaceNil() bool {
	return se == nil
}
