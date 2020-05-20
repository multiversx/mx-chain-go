//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. log.proto
package transaction

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// GetLogEvents returns the interface for the underlying events of the log structure
func (l *Log) GetLogEvents() []data.EventHandler {
	events := make([]data.EventHandler, len(l.Events))
	for i, e := range l.Events {
		events[i] = e
	}
	return events
}

// IsInterfaceNil verifies if underlying object is nil
func (l *Log) IsInterfaceNil() bool {
	return l == nil
}

// IsInterfaceNil verifies if underlying object is nil
func (e *Event) IsInterfaceNil() bool {
	return e == nil
}
