package outport

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/outport/types"
)

type nilOutport struct{}

// NewNilOutport wil create a new instance of nilOurpot
func NewNilOutport() OutportHandler {
	return new(nilOutport)
}

// SaveBlock does nothing
func (n *nilOutport) SaveBlock(_ types.ArgsSaveBlocks) {
}

// RevertBlock does nothing
func (n *nilOutport) RevertBlock(_ data.HeaderHandler, _ data.BodyHandler) {
}

// SaveRoundsInfo does nothing
func (n *nilOutport) SaveRoundsInfo(_ []types.RoundInfo) {
}

// UpdateTPS does nothing
func (n *nilOutport) UpdateTPS(_ statistics.TPSBenchmark) {
}

// SaveValidatorsPubKeys does nothing
func (n *nilOutport) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) {
}

// SaveValidatorsRating does nothing
func (n *nilOutport) SaveValidatorsRating(_ string, _ []types.ValidatorRatingInfo) {
}

// SaveAccounts does nothing
func (n *nilOutport) SaveAccounts(_ []state.UserAccountHandler) {
}

// Close does nothing
func (n *nilOutport) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (n *nilOutport) IsInterfaceNil() bool {
	return n == nil
}

// SubscribeDriver does nothing
func (n *nilOutport) SubscribeDriver(_ Driver) error {
	return nil
}

// HasDrivers does nothing
func (n *nilOutport) HasDrivers() bool {
	return false
}
