package logging

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// LogLifeSpanner defines a notification channel for the file logging lifespan
type LogLifeSpanner interface {
	GetChannel() <-chan string
	IsInterfaceNil() bool
}

// SizeLogLifeSpanner defines a notification channel for the file logging lifespan
type SizeLogLifeSpanner interface {
	LogLifeSpanner
	SetCurrentFile(string)
}

// EpochStartNotifier defines which actions should be done for handling new epoch's events
type EpochStartNotifier interface {
	RegisterHandler(handler epochStart.ActionHandler)
	UnregisterHandler(handler epochStart.ActionHandler)
	NotifyAll(hdr data.HeaderHandler)
	NotifyAllPrepare(metaHdr data.HeaderHandler, body data.BodyHandler)
	NotifyEpochChangeConfirmed(epoch uint32)
	IsInterfaceNil() bool
}

// EpochStartNotifierWithConfirm defines which actions should be done for handling new epoch's events and confirmation
type EpochStartNotifierWithConfirm interface {
	EpochStartNotifier
	RegisterForEpochChangeConfirmed(handler func(epoch uint32))
}

// LogLifeSpanFactory defines the methods for creating a log lifeSpanner
type LogLifeSpanFactory interface {
	CreateLogLifeSpanner(args LogLifeSpanFactoryArgs) (LogLifeSpanner, error)
}

// FileSizeCheckHandler defines the method needed for getting a file size
type FileSizeCheckHandler interface {
	GetSize(path string) (int64, error)
	IsInterfaceNil() bool
}
