package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type StringNewer struct {
	Data string
}

func (sn *StringNewer) ID() string {
	return sn.Data
}

func (sn *StringNewer) New() p2p.Newer {
	return &StringNewer{}
}
