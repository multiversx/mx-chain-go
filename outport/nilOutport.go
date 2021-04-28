package outport

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

type nilOutport struct{}

// NewNilOutport wil create a new instance of nilOurpot
func NewNilOutport() OutportHandler {
	return new(nilOutport)
}

// SaveBlock does nothing
func (n *nilOutport) SaveBlock(_ *indexer.ArgsSaveBlockData) {
}

// RevertIndexedBlock does nothing
func (n *nilOutport) RevertIndexedBlock(_ data.HeaderHandler, _ data.BodyHandler) {
}

// SaveRoundsInfo does nothing
func (n *nilOutport) SaveRoundsInfo(_ []*indexer.RoundInfo) {
}

// UpdateTPS does nothing
func (n *nilOutport) UpdateTPS(_ statistics.TPSBenchmark) {
}

// SaveValidatorsPubKeys does nothing
func (n *nilOutport) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) {
}

// SaveValidatorsRating does nothing
func (n *nilOutport) SaveValidatorsRating(_ string, _ []*indexer.ValidatorRatingInfo) {
}

// SaveAccounts does nothing
func (n *nilOutport) SaveAccounts(_ uint64, _ []state.UserAccountHandler) {
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
