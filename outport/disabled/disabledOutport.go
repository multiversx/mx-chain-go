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

// SaveBlock returns nil
func (n *disabledOutport) SaveBlock(_ *indexer.ArgsSaveBlockData) error {
	return nil
}

// RevertIndexedBlock returns nil
func (n *disabledOutport) RevertIndexedBlock(_ data.HeaderHandler, _ data.BodyHandler) error {
	return nil
}

// SaveRoundsInfo returns nil
func (n *disabledOutport) SaveRoundsInfo(_ []*indexer.RoundInfo) error {
	return nil
}

// SaveValidatorsPubKeys returns nil
func (n *disabledOutport) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) error {
	return nil
}

// SaveValidatorsRating returns nil
func (n *disabledOutport) SaveValidatorsRating(_ string, _ []*indexer.ValidatorRatingInfo) error {
	return nil
}

// SaveAccounts returns nil
func (n *disabledOutport) SaveAccounts(_ uint64, _ []data.UserAccountHandler) error {
	return nil
}

// FinalizedBlock returns nil
func (n *disabledOutport) FinalizedBlock(_ []byte) error {
	return nil
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
