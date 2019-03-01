package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type ChannelLoadBalancerStub struct {
	AddChannelCalled                    func(pipe string) error
	RemoveChannelCalled                 func(pipe string) error
	GetChannelOrDefaultCalled           func(pipe string) chan *p2p.SendableData
	CollectOneElementFromChannelsCalled func() *p2p.SendableData
}

func (plbs *ChannelLoadBalancerStub) AddChannel(pipe string) error {
	return plbs.AddChannelCalled(pipe)
}

func (plbs *ChannelLoadBalancerStub) RemoveChannel(pipe string) error {
	return plbs.RemoveChannelCalled(pipe)
}

func (plbs *ChannelLoadBalancerStub) GetChannelOrDefault(pipe string) chan *p2p.SendableData {
	return plbs.GetChannelOrDefaultCalled(pipe)
}

func (plbs *ChannelLoadBalancerStub) CollectOneElementFromChannels() *p2p.SendableData {
	return plbs.CollectOneElementFromChannelsCalled()
}
