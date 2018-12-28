package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type StringCreator struct {
	Data string
}

func (sn *StringCreator) ID() string {
	return sn.Data
}

func (sn *StringCreator) Create() p2p.Creator {
	return &StringCreator{}
}
