package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type PipeLoadBalancerStub struct {
	AddPipeCalled             func(pipe string) error
	RemovePipeCalled          func(pipe string) error
	GetChannelOrDefaultCalled func(pipe string) chan *p2p.SendableData
	CollectFromPipesCalled    func() []*p2p.SendableData
}

func (plbs *PipeLoadBalancerStub) AddPipe(pipe string) error {
	return plbs.AddPipeCalled(pipe)
}

func (plbs *PipeLoadBalancerStub) RemovePipe(pipe string) error {
	return plbs.RemovePipeCalled(pipe)
}

func (plbs *PipeLoadBalancerStub) GetChannelOrDefault(pipe string) chan *p2p.SendableData {
	return plbs.GetChannelOrDefaultCalled(pipe)
}

func (plbs *PipeLoadBalancerStub) CollectFromPipes() []*p2p.SendableData {
	return plbs.CollectFromPipesCalled()
}
