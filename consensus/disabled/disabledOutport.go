package disabled

import (
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
)

type disabledOutport struct{}

// NewDisabledOutport will create a new instance of disabledOutport
func NewDisabledOutport() *disabledOutport {
	return new(disabledOutport)
}

// SaveRoundsInfo does nothing
func (n *disabledOutport) SaveRoundsInfo(_ []*indexer.RoundInfo) {
}

// HasDrivers does nothing
func (n *disabledOutport) HasDrivers() bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (n *disabledOutport) IsInterfaceNil() bool {
	return n == nil
}
