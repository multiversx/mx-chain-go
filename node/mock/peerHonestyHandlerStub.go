package mock

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
)

var log = logger.GetOrCreate("node/mock")

// PeerHonestyHandlerStub -
type PeerHonestyHandlerStub struct {
	IncreaseCalled func(pk string, topic string, value float64)
	DecreaseCalled func(pk string, topic string, value float64)
}

// Increase -
func (phhs *PeerHonestyHandlerStub) Increase(pk string, topic string, value float64) {
	if phhs.IncreaseCalled != nil {
		phhs.IncreaseCalled(pk, topic, value)
		return
	}
}

// Decrease -
func (phhs *PeerHonestyHandlerStub) Decrease(pk string, topic string, value float64) {
	if phhs.DecreaseCalled != nil {
		phhs.DecreaseCalled(pk, topic, value)
		return
	}

	log.Warn("PeerHonestyHandlerStub.Decrease",
		"topic", topic,
		"pk", core.GetTrimmedPk(hex.EncodeToString([]byte(pk))),
		"value", value)
}

// IsInterfaceNil -
func (phhs *PeerHonestyHandlerStub) IsInterfaceNil() bool {
	return phhs == nil
}
