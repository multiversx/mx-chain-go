package dataThrottle

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

func (sdt *SendDataThrottle) Chans() []chan *p2p.SendableData {
	return sdt.chans
}

func (sdt *SendDataThrottle) Names() []string {
	return sdt.names
}

func (sdt *SendDataThrottle) NamesChans() map[string]chan *p2p.SendableData {
	return sdt.namesChans
}

func DefaultSendPipe() string {
	return defaultSendPipe
}
