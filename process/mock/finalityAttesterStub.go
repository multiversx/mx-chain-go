package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type ValidityAttesterStub struct {
	CheckBlockBasicValidityCalled func(headerHandler data.HeaderHandler) error
}

func (vas *ValidityAttesterStub) CheckBlockBasicValidity(headerHandler data.HeaderHandler) error {
	if vas.CheckBlockBasicValidityCalled != nil {
		return vas.CheckBlockBasicValidityCalled(headerHandler)
	}

	return nil
}

func (vas *ValidityAttesterStub) IsInterfaceNil() bool {
	return vas == nil
}
