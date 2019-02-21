package loadBalancer

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

func (oplb *OutgoingPipeLoadBalancer) Chans() []chan *p2p.SendableData {
	return oplb.chans
}

func (oplb *OutgoingPipeLoadBalancer) Names() []string {
	return oplb.names
}

func (oplb *OutgoingPipeLoadBalancer) NamesChans() map[string]chan *p2p.SendableData {
	return oplb.namesChans
}

func DefaultSendPipe() string {
	return defaultSendPipe
}
