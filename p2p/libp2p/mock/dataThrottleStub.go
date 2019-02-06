package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type DataThrottleStub struct {
	AddPipeCalled             func(pipe string) error
	RemovePipeCalled          func(pipe string) error
	GetChannelOrDefaultCalled func(pipe string) chan *p2p.SendableData
	CollectFromPipesCalled    func() []*p2p.SendableData
}

func (dts *DataThrottleStub) AddPipe(pipe string) error {
	return dts.AddPipeCalled(pipe)
}

func (dts *DataThrottleStub) RemovePipe(pipe string) error {
	return dts.RemovePipeCalled(pipe)
}

func (dts *DataThrottleStub) GetChannelOrDefault(pipe string) chan *p2p.SendableData {
	return dts.GetChannelOrDefaultCalled(pipe)
}

func (dts *DataThrottleStub) CollectFromPipes() []*p2p.SendableData {
	return dts.CollectFromPipesCalled()
}
