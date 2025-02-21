package broadcast

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/broadcast/shared"
)

// DelayedBroadcaster exposes functionality for handling the consensus members broadcasting of delay data
type DelayedBroadcaster interface {
	SetLeaderData(data *shared.DelayedBroadcastData) error
	SetValidatorData(data *shared.DelayedBroadcastData) error
	SetHeaderForValidator(vData *shared.ValidatorHeaderBroadcastData) error
	SetBroadcastHandlers(
		mbBroadcast func(mbData map[uint32][]byte, pkBytes []byte) error,
		txBroadcast func(txData map[string][][]byte, pkBytes []byte) error,
		headerBroadcast func(header data.HeaderHandler, pkBytes []byte) error,
		consensusMessageBroadcast func(message *consensus.Message) error,
	) error
	Close()
	IsInterfaceNil() bool
}
