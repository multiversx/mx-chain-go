package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/consensus/broadcast/delayed"
)

// DelayBlockBroadcasterStub -
type DelayBlockBroadcasterStub struct {
	SetLeaderDataCalled         func(data *delayed.DelayedBroadcastData) error
	SetValidatorDataCalled      func(data *delayed.DelayedBroadcastData) error
	SetHeaderForValidatorCalled func(vData *delayed.ValidatorHeaderBroadcastData) error
	SetBroadcastHandlersCalled  func(
		mbBroadcast func(mbData map[uint32][]byte) error,
		txBroadcast func(txData map[string][][]byte) error,
		headerBroadcast func(header data.HeaderHandler) error,
	) error
	CloseCalled func()
}

// SetLeaderData -
func (stub *DelayBlockBroadcasterStub) SetLeaderData(data *delayed.DelayedBroadcastData) error {
	if stub.SetLeaderDataCalled != nil {
		return stub.SetLeaderDataCalled(data)
	}

	return nil
}

// SetValidatorData -
func (stub *DelayBlockBroadcasterStub) SetValidatorData(data *delayed.DelayedBroadcastData) error {
	if stub.SetValidatorDataCalled != nil {
		return stub.SetValidatorDataCalled(data)
	}

	return nil
}

// SetHeaderForValidator -
func (stub *DelayBlockBroadcasterStub) SetHeaderForValidator(vData *delayed.ValidatorHeaderBroadcastData) error {
	if stub.SetHeaderForValidatorCalled != nil {
		return stub.SetHeaderForValidatorCalled(vData)
	}

	return nil
}

// SetBroadcastHandlers -
func (stub *DelayBlockBroadcasterStub) SetBroadcastHandlers(
	mbBroadcast func(mbData map[uint32][]byte) error,
	txBroadcast func(txData map[string][][]byte) error,
	headerBroadcast func(header data.HeaderHandler) error,
) error {
	if stub.SetBroadcastHandlersCalled != nil {
		return stub.SetBroadcastHandlersCalled(mbBroadcast, txBroadcast, headerBroadcast)
	}

	return nil
}

// Close -
func (stub *DelayBlockBroadcasterStub) Close() {
	if stub.CloseCalled != nil {
		stub.CloseCalled()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *DelayBlockBroadcasterStub) IsInterfaceNil() bool {
	return stub == nil
}
