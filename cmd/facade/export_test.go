package facade

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
)

// GetLogger returns the current logger
func (ef *ElrondNodeFacade) GetLogger() *logger.Logger {
	return ef.log
}

// GetSyncer returns the current syncer
func (ef *ElrondNodeFacade) GetSyncer() ntp.SyncTimer {
	return ef.syncer
}
