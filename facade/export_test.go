package facade

import (
	"github.com/ElrondNetwork/elrond-go/ntp"
)

// GetSyncer returns the current syncer
func (nf *nodeFacade) GetSyncer() ntp.SyncTimer {
	return nf.syncer
}
