package disabled

import (
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/outport"
)

type disabledOutport struct{}

// NewDisabledOutport will create a new instance of disabledOutport
func NewDisabledOutport() *disabledOutport {
	return new(disabledOutport)
}

// SaveBlock does nothing
func (n *disabledOutport) SaveBlock(_ *outportcore.OutportBlockWithHeaderAndBody) error {
	return nil
}

// RevertIndexedBlock does nothing
func (n *disabledOutport) RevertIndexedBlock(_ *outportcore.HeaderDataWithBody) error {
	return nil
}

// SaveRoundsInfo does nothing
func (n *disabledOutport) SaveRoundsInfo(_ *outportcore.RoundsInfo) {
}

// SaveValidatorsPubKeys does nothing
func (n *disabledOutport) SaveValidatorsPubKeys(_ *outportcore.ValidatorsPubKeys) {
}

// SaveValidatorsRating does nothing
func (n *disabledOutport) SaveValidatorsRating(_ *outportcore.ValidatorsRating) {
}

// SaveAccounts does nothing
func (n *disabledOutport) SaveAccounts(_ *outportcore.Accounts) {
}

// FinalizedBlock does nothing
func (n *disabledOutport) FinalizedBlock(_ *outportcore.FinalizedBlock) {
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
