package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go/outport"
)

type nilOutport struct{}

// NewNilOutport -
func NewNilOutport() *nilOutport {
	return new(nilOutport)
}

// SaveBlock -
func (n *nilOutport) SaveBlock(_ *indexer.ArgsSaveBlockData) error {
	return nil
}

// RevertIndexedBlock -
func (n *nilOutport) RevertIndexedBlock(_ data.HeaderHandler, _ data.BodyHandler) error {
	return nil
}

// SaveRoundsInfo -
func (n *nilOutport) SaveRoundsInfo(_ []*indexer.RoundInfo) error {
	return nil
}

// SaveValidatorsPubKeys -
func (n *nilOutport) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) error {
	return nil
}

// SaveValidatorsRating -
func (n *nilOutport) SaveValidatorsRating(_ string, _ []*indexer.ValidatorRatingInfo) error {
	return nil
}

// SaveAccounts -
func (n *nilOutport) SaveAccounts(_ uint64, _ []data.UserAccountHandler) error {
	return nil
}

// FinalizedBlock -
func (n *nilOutport) FinalizedBlock(_ []byte) error {
	return nil
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
