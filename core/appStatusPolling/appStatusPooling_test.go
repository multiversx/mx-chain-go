package appStatusPolling_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/appStatusPolling"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestNewAppStatusPooling_NilAppStatusHandlerShouldErr(t *testing.T) {
	_, err := appStatusPolling.NewAppStatusPolling(nil, 10)
	assert.Equal(t, err, appStatusPolling.NilAppStatusHandler)
}

func TestNewAppStatusPooling_NegativePollingDurationShouldErr(t *testing.T) {
	_, err := appStatusPolling.NewAppStatusPolling(&statusHandler.NilStatusHandler{}, -10)
	assert.Equal(t, err, appStatusPolling.PollingDurationNegative)
}

func TestNewAppStatusPooling_OkValsShouldPass(t *testing.T) {
	_, err := appStatusPolling.NewAppStatusPolling(&statusHandler.NilStatusHandler{}, 10)
	assert.Nil(t, err)
}

func TestAppStatusPolling_SetConnectedAddressesNilShouldErr(t *testing.T) {
	asp, err := appStatusPolling.NewAppStatusPolling(&statusHandler.NilStatusHandler{}, 10)
	assert.Nil(t, err)
	err = asp.SetConnectedAddresses(nil)
	assert.Equal(t, err, appStatusPolling.NilConnectedAddressesHandler)
}

func TestAppStatusPolling_SetConnectedAddressesOkValsShouldPass(t *testing.T) {
	asp, err := appStatusPolling.NewAppStatusPolling(&statusHandler.NilStatusHandler{}, 10)
	assert.Nil(t, err)
	err = asp.SetConnectedAddresses(&mock.ConnectedAddressesMock{})
	assert.Nil(t, err)
}

func TestAppStatusPolling_Poll_TestNoConnectedAddressesCalled(t *testing.T) {
	t.Parallel()

	pollingDuration := time.Second * 1
	chDone := make(chan struct{})
	ash := mock.AppStatusHandlerStub{
		SetInt64ValueHandler: func(key string, value int64) {
			chDone <- struct{}{}
		},
	}
	asp, err := appStatusPolling.NewAppStatusPolling(&ash, pollingDuration)
	assert.Nil(t, err)

	err = asp.SetConnectedAddresses(&mock.ConnectedAddressesMock{})
	assert.Nil(t, err)

	asp.Poll()

	select {
	case <-chDone:
	case <-time.After(pollingDuration * 2):
		assert.Fail(t, "timeout calling SetInt64Value")
	}
}
