package loadBalancer

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

func (oplb *OutgoingChannelLoadBalancer) Chans() []chan *p2p.SendableData {
	return oplb.chans
}

func (oplb *OutgoingChannelLoadBalancer) Names() []string {
	return oplb.names
}

func (oplb *OutgoingChannelLoadBalancer) NamesChans() map[string]chan *p2p.SendableData {
	return oplb.namesChans
}

func DefaultSendChannel() string {
	return defaultSendChannel
}
