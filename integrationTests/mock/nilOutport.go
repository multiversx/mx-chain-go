package mock

import (
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/outport"
)

type nilOutport struct{}

// NewNilOutport -
func NewNilOutport() *nilOutport {
	return new(nilOutport)
}

// SaveBlock -
func (n *nilOutport) SaveBlock(_ *outportcore.OutportBlock) {
}

// RevertIndexedBlock -
func (n *nilOutport) RevertIndexedBlock(_ *outportcore.BlockData) {
}

// SaveRoundsInfo -
func (n *nilOutport) SaveRoundsInfo(_ *outportcore.RoundsInfo) {
}

// SaveValidatorsPubKeys -
func (n *nilOutport) SaveValidatorsPubKeys(_ *outportcore.ValidatorsPubKeys) {
}

// SaveValidatorsRating -
func (n *nilOutport) SaveValidatorsRating(_ *outportcore.ValidatorsRating) {
}

// SaveAccounts -
func (n *nilOutport) SaveAccounts(_ *outportcore.Accounts) {
}

// FinalizedBlock -
func (n *nilOutport) FinalizedBlock(_ *outportcore.FinalizedBlock) {
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
func (n *nilOutport) SubscribeDriver(_ outport.Driver) error {
	return nil
}

// HasDrivers -
func (n *nilOutport) HasDrivers() bool {
	return false
}
