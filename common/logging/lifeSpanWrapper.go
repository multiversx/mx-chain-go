package logging

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
)

type lifeSpanWrapper struct {
	mutLogLifeSpanner sync.Mutex
	lifeSpanner       factory.LogLifeSpanner
}

// SetLifeSpanner sets the current lifeSpanner
func (llsw *lifeSpanWrapper) SetLifeSpanner(spanner factory.LogLifeSpanner) {
	llsw.mutLogLifeSpanner.Lock()
	log.Debug("lifeSpanWrapper - set life spanner", "spanner", spanner)
	llsw.lifeSpanner = spanner
	llsw.mutLogLifeSpanner.Unlock()
}

// SetCurrentFile sets the new logger file
func (llsw *lifeSpanWrapper) SetCurrentFile(newFile string) {
	llsw.mutLogLifeSpanner.Lock()
	log.Debug("lifeSpanWrapper - set new file", "new file", newFile)
	llsw.lifeSpanner.SetCurrentFile(newFile)
	llsw.mutLogLifeSpanner.Unlock()
}

// GetNotificationChannel gets the notification channel for the log lifespan
func (llsw *lifeSpanWrapper) GetNotificationChannel() <-chan string {
	llsw.mutLogLifeSpanner.Lock()
	defer llsw.mutLogLifeSpanner.Unlock()
	return llsw.lifeSpanner.GetNotification()
}
