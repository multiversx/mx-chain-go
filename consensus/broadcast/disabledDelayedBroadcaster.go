package broadcast

import "github.com/ElrondNetwork/elrond-go-core/data"

type disabledDelayedBroadcaster struct{}

// NewDisabledDelayedBroadcaster will create a new instance of disabledDelayedBroadcaster
func NewDisabledDelayedBroadcaster() *disabledDelayedBroadcaster {
	return new(disabledDelayedBroadcaster)
}

// SetLeaderData returns nil
func (ddb *disabledDelayedBroadcaster) SetLeaderData(data *delayedBroadcastData) error {
	return nil
}

// SetValidatorData returns nil
func (ddb *disabledDelayedBroadcaster) SetValidatorData(data *delayedBroadcastData) error {
	return nil
}

// SetHeaderForValidator returns nil
func (ddb *disabledDelayedBroadcaster) SetHeaderForValidator(vData *validatorHeaderBroadcastData) error {
	return nil
}

// SetBroadcastHandlers returns nil
func (ddb *disabledDelayedBroadcaster) SetBroadcastHandlers(
	mbBroadcast func(mbData map[uint32][]byte) error,
	txBroadcast func(txData map[string][][]byte) error,
	headerBroadcast func(header data.HeaderHandler) error,
) error {
	return nil
}

// Close does nothing
func (ddb *disabledDelayedBroadcaster) Close() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (ddb *disabledDelayedBroadcaster) IsInterfaceNil() bool {
	return ddb == nil
}
