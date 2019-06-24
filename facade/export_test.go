package facade

import (
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/ntp"
)

// GetLogger returns the current logger
func (ef *ElrondNodeFacade) GetLogger() *logger.Logger {
	return ef.log
}

// GetSyncer returns the current syncer
func (ef *ElrondNodeFacade) GetSyncer() ntp.SyncTimer {
	return ef.syncer
}
