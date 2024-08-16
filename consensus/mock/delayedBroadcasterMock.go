package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/consensus/broadcast"
)

type DelayedBroadcasterMock struct {
	SetLeaderDataCalled         func(data *broadcast.DelayedBroadcastData) error
	SetValidatorDataCalled      func(data *broadcast.DelayedBroadcastData) error
	SetHeaderForValidatorCalled func(vData *broadcast.ValidatorHeaderBroadcastData) error
	SetBroadcastHandlersCalled  func(mbBroadcast func(mbData map[uint32][]byte, pkBytes []byte) error, txBroadcast func(txData map[string][][]byte, pkBytes []byte) error, headerBroadcast func(header data.HeaderHandler, pkBytes []byte) error) error
	CloseCalled                 func()
}

func (d DelayedBroadcasterMock) SetLeaderData(data *broadcast.DelayedBroadcastData) error {
	if d.SetLeaderDataCalled != nil {
		return d.SetLeaderDataCalled(data)
	}
	return nil
}

// SetValidatorData -
func (d DelayedBroadcasterMock) SetValidatorData(data *broadcast.DelayedBroadcastData) error {
	if d.SetValidatorDataCalled != nil {
		return d.SetValidatorDataCalled(data)
	}
	return nil
}

func (d DelayedBroadcasterMock) SetHeaderForValidator(vData *broadcast.ValidatorHeaderBroadcastData) error {
	if d.SetHeaderForValidatorCalled != nil {
		return d.SetHeaderForValidatorCalled(vData)
	}
	return nil
}

func (d DelayedBroadcasterMock) SetBroadcastHandlers(mbBroadcast func(mbData map[uint32][]byte, pkBytes []byte) error, txBroadcast func(txData map[string][][]byte, pkBytes []byte) error, headerBroadcast func(header data.HeaderHandler, pkBytes []byte) error) error {
	if d.SetBroadcastHandlersCalled != nil {
		return d.SetBroadcastHandlersCalled(mbBroadcast, txBroadcast, headerBroadcast)
	}
	return nil
}

func (d DelayedBroadcasterMock) Close() {
	if d.CloseCalled != nil {
		d.CloseCalled()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (dbb *DelayedBroadcasterMock) IsInterfaceNil() bool {
	return dbb == nil
}
