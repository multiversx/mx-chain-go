package mock

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
)

var log = logger.GetOrCreate("node/mock")

// PeerHonestyHandlerStub -
type PeerHonestyHandlerStub struct {
	IncreaseCalled func(round int64, pk string, topic string, value float64)
	DecreaseCalled func(round int64, pk string, topic string, value float64)
}

// Increase -
func (phhs *PeerHonestyHandlerStub) Increase(round int64, pk string, topic string, value float64) {
	if phhs.IncreaseCalled != nil {
		phhs.IncreaseCalled(round, pk, topic, value)
		return
	}
}

// Decrease -
func (phhs *PeerHonestyHandlerStub) Decrease(round int64, pk string, topic string, value float64) {
	if phhs.DecreaseCalled != nil {
		phhs.DecreaseCalled(round, pk, topic, value)
		return
	}

	log.Warn("PeerHonestyHandlerStub.Decrease",
		"round", round,
		"topic", topic,
		"pk", core.GetTrimmedPk(hex.EncodeToString([]byte(pk))),
		"value", value)
}

// IsInterfaceNil -
func (phhs *PeerHonestyHandlerStub) IsInterfaceNil() bool {
	return phhs == nil
}
