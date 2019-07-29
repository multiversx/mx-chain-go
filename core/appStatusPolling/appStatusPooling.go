package appStatusPolling

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
)

// AppStatusPolling will update an AppStatusHandler by polling components at a predefined interval
type AppStatusPolling struct {
	pollingDuration           time.Duration
	connectedAddressesHandler core.ConnectedAddressesHandler
	appStatusHandler          core.AppStatusHandler
}

// NewAppStatusPolling will return an instance of AppStatusPolling
func NewAppStatusPolling(appStatusHandler core.AppStatusHandler, pollingDuration time.Duration) (*AppStatusPolling, error) {
	if appStatusHandler == nil {
		return nil, NilAppStatusHandler
	}
	if pollingDuration < 0 {
		return nil, PollingDurationNegative
	}
	return &AppStatusPolling{
		pollingDuration:  pollingDuration,
		appStatusHandler: appStatusHandler,
	}, nil
}

// SetConnectedAddresses will receive a handler for the connected addresses
func (asp *AppStatusPolling) SetConnectedAddresses(connectedAddressesHandler core.ConnectedAddressesHandler) error {
	if connectedAddressesHandler == nil {
		return NilConnectedAddressesHandler
	}
	asp.connectedAddressesHandler = connectedAddressesHandler
	return nil
}

// Poll will notify the AppStatusHandler at a given time
func (asp *AppStatusPolling) Poll() {
	go func() {
		for {
			<-time.After(asp.pollingDuration)
			if asp.connectedAddressesHandler != nil {
				asp.appStatusHandler.SetInt64Value(core.MetricNumConnectedPeers, int64(len(asp.connectedAddressesHandler.ConnectedAddresses())))
			}
		}
	}()
}
