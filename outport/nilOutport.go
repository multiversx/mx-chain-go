package outport

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/outport/drivers"
	"github.com/ElrondNetwork/elrond-go/outport/types"
)

type nilOutport struct{}

// NewNilOutport --
func NewNilOutport() OutportHandler {
	return new(nilOutport)
}

// SaveBlock -
func (n *nilOutport) SaveBlock(_ types.ArgsSaveBlocks) {
}

// RevertBlock -
func (n *nilOutport) RevertBlock(_ data.HeaderHandler, _ data.BodyHandler) {
}

// SaveRoundsInfo -
func (n *nilOutport) SaveRoundsInfo(_ []types.RoundInfo) {
}

// UpdateTPS -
func (n *nilOutport) UpdateTPS(_ statistics.TPSBenchmark) {
}

// SaveValidatorsPubKeys -
func (n *nilOutport) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) {
}

// SaveValidatorsRating -
func (n *nilOutport) SaveValidatorsRating(_ string, _ []types.ValidatorRatingInfo) {
}

// SaveAccounts -
func (n *nilOutport) SaveAccounts(_ []state.UserAccountHandler) {
}

// Close -
func (n *nilOutport) Close() error {
	return nil
}

// IsInterfaceNil -
func (n *nilOutport) IsInterfaceNil() bool {
	return n == nil
}

// SubscribeDriver -
func (n *nilOutport) SubscribeDriver(_ drivers.Driver) error {
	return nil
}

// HasDrivers -
func (n *nilOutport) HasDrivers() bool {
	return false
}
