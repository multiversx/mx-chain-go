package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type ChannelLoadBalancerStub struct {
	AddChannelCalled                    func(pipe string) error
	RemoveChannelCalled                 func(pipe string) error
	GetChannelOrDefaultCalled           func(pipe string) chan *p2p.SendableData
	CollectOneElementFromChannelsCalled func() *p2p.SendableData
}

func (clbs *ChannelLoadBalancerStub) AddChannel(pipe string) error {
	return clbs.AddChannelCalled(pipe)
}

func (clbs *ChannelLoadBalancerStub) RemoveChannel(pipe string) error {
	return clbs.RemoveChannelCalled(pipe)
}

func (clbs *ChannelLoadBalancerStub) GetChannelOrDefault(pipe string) chan *p2p.SendableData {
	return clbs.GetChannelOrDefaultCalled(pipe)
}

func (clbs *ChannelLoadBalancerStub) CollectOneElementFromChannels() *p2p.SendableData {
	return clbs.CollectOneElementFromChannelsCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (clbs *ChannelLoadBalancerStub) IsInterfaceNil() bool {
	return clbs == nil
}
