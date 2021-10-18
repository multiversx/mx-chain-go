package disabled

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go/outport"
)

type disabledOutport struct{}

// NewDisabledOutport will create a new instance of disabledOutport
func NewDisabledOutport() *disabledOutport {
	return new(disabledOutport)
}

// SaveBlock does nothing
func (n *disabledOutport) SaveBlock(_ *indexer.ArgsSaveBlockData) {
}

// RevertIndexedBlock does nothing
func (n *disabledOutport) RevertIndexedBlock(_ data.HeaderHandler, _ data.BodyHandler) {
}

// SaveRoundsInfo does nothing
func (n *disabledOutport) SaveRoundsInfo(_ []*indexer.RoundInfo) {
}

// SaveValidatorsPubKeys does nothing
func (n *disabledOutport) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) {
}

// SaveValidatorsRating does nothing
func (n *disabledOutport) SaveValidatorsRating(_ string, _ []*indexer.ValidatorRatingInfo) {
}

// SaveAccounts does nothing
func (n *disabledOutport) SaveAccounts(_ uint64, _ []data.UserAccountHandler) {
}

// FinalizedBlock does nothing
func (n *disabledOutport) FinalizedBlock(_ []byte) {
}

// Close does nothing
func (n *disabledOutport) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (n *disabledOutport) IsInterfaceNil() bool {
	return n == nil
}

// SubscribeDriver does nothing
func (n *disabledOutport) SubscribeDriver(_ outport.Driver) error {
	return nil
}

// HasDrivers does nothing
func (n *disabledOutport) HasDrivers() bool {
	return false
}
