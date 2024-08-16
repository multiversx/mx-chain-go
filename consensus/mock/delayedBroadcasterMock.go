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

// SetLeaderData -
func (mock *DelayedBroadcasterMock) SetLeaderData(data *broadcast.DelayedBroadcastData) error {
	if mock.SetLeaderDataCalled != nil {
		return mock.SetLeaderDataCalled(data)
	}
	return nil
}

// SetValidatorData -
func (mock *DelayedBroadcasterMock) SetValidatorData(data *broadcast.DelayedBroadcastData) error {
	if mock.SetValidatorDataCalled != nil {
		return mock.SetValidatorDataCalled(data)
	}
	return nil
}

// SetHeaderForValidator -
func (mock *DelayedBroadcasterMock) SetHeaderForValidator(vData *broadcast.ValidatorHeaderBroadcastData) error {
	if mock.SetHeaderForValidatorCalled != nil {
		return mock.SetHeaderForValidatorCalled(vData)
	}
	return nil
}

// SetBroadcastHandlers -
func (mock *DelayedBroadcasterMock) SetBroadcastHandlers(mbBroadcast func(mbData map[uint32][]byte, pkBytes []byte) error, txBroadcast func(txData map[string][][]byte, pkBytes []byte) error, headerBroadcast func(header data.HeaderHandler, pkBytes []byte) error) error {
	if mock.SetBroadcastHandlersCalled != nil {
		return mock.SetBroadcastHandlersCalled(mbBroadcast, txBroadcast, headerBroadcast)
	}
	return nil
}

// Close -
func (mock *DelayedBroadcasterMock) Close() {
	if mock.CloseCalled != nil {
		mock.CloseCalled()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (mock *DelayedBroadcasterMock) IsInterfaceNil() bool {
	return mock == nil
}
