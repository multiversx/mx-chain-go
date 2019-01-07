package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type StringCreatorMock struct {
	Data string
}

func (sn *StringCreatorMock) ID() string {
	return sn.Data
}

func (sn *StringCreatorMock) Create() p2p.Creator {
	return &StringCreatorMock{}
}
