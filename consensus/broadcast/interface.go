package broadcast

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// DelayedBroadcaster exposes functionality for handling the consensus members broadcasting of delay data
type DelayedBroadcaster interface {
	SetLeaderData(data *DelayedBroadcastData) error
	SetValidatorData(data *DelayedBroadcastData) error
	SetHeaderForValidator(vData *ValidatorHeaderBroadcastData) error
	SetBroadcastHandlers(
		mbBroadcast func(mbData map[uint32][]byte, pkBytes []byte) error,
		txBroadcast func(txData map[string][][]byte, pkBytes []byte) error,
		headerBroadcast func(header data.HeaderHandler, pkBytes []byte) error,
	) error
	Close()
	IsInterfaceNil() bool
}
