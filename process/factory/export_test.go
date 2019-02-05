package factory

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

func (p *processorsCreator) SetMessenger(messenger p2p.Messenger) {
	p.messenger = messenger
}
