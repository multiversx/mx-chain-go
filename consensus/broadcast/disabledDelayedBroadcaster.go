package broadcast

import "github.com/ElrondNetwork/elrond-go-core/data"

type DisabledDelayedBroadcaster struct{}

func (ddb *DisabledDelayedBroadcaster) SetLeaderData(data *delayedBroadcastData) error {
	return nil
}

func (ddb *DisabledDelayedBroadcaster) SetValidatorData(data *delayedBroadcastData) error {
	return nil
}

func (ddb *DisabledDelayedBroadcaster) SetHeaderForValidator(vData *validatorHeaderBroadcastData) error {
	return nil
}

func (ddb *DisabledDelayedBroadcaster) SetBroadcastHandlers(
	mbBroadcast func(mbData map[uint32][]byte) error,
	txBroadcast func(txData map[string][][]byte) error,
	headerBroadcast func(header data.HeaderHandler) error,
) error {
	return nil
}

func (ddb *DisabledDelayedBroadcaster) Close() {
}
