package ntp_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	ntp2 "github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/beevik/ntp"
	"github.com/stretchr/testify/assert"
)

var responseMock1 *ntp.Response
var failNtpMock1 = false
var responseMock2 *ntp.Response
var failNtpMock2 = false
var responseMock3 *ntp.Response
var failNtpMock3 = false
var responseMock5 *ntp.Response
var failNtpMock5 = false

var errNtpMock = errors.New("NTP Mock generic error")
var queryMock4Call = 0

func queryMock1(host string) (*ntp.Response, error) {
	if failNtpMock1 {
		return nil, errNtpMock
	}

	return responseMock1, nil
}

func queryMock2(host string) (*ntp.Response, error) {
	if failNtpMock2 {
		return nil, errNtpMock
	}

	return responseMock2, nil
}

func queryMock3(host string) (*ntp.Response, error) {
	if failNtpMock3 {
		return nil, errNtpMock
	}

	return responseMock3, nil
}

func queryMock4(host string) (*ntp.Response, error) {
	queryMock4Call++

	return nil, errNtpMock
}

func queryMock5(host string) (*ntp.Response, error) {
	if failNtpMock5 {
		return nil, errNtpMock
	}

	return responseMock5, nil
}

func TestHandleErrorInDoSync(t *testing.T) {
	//rm := ntp.Response{Offset:23456}

	failNtpMock1 = true
	st := ntp2.SyncTime{}
	st.SetSyncPeriod(time.Millisecond)
	st.SetQuery(queryMock1)

	st.DoSync()

	assert.Equal(t, st.ClockOffset(), time.Millisecond*0)

	//manually put a value in Offset and observe if it goes to 0 as a result to error
	st.SetClockOffset(1234)

	st.DoSync()

	assert.Equal(t, st.ClockOffset(), time.Millisecond*0)

}

func TestValueInDoSync(t *testing.T) {
	responseMock2 = &ntp.Response{ClockOffset: 23456}

	failNtpMock2 = false
	st := ntp2.SyncTime{}
	st.SetSyncPeriod(time.Millisecond)
	st.SetQuery(queryMock2)

	assert.Equal(t, st.ClockOffset(), time.Millisecond*0)
	st.DoSync()
	assert.Equal(t, st.ClockOffset(), time.Nanosecond*23456)

	//manually put a value in Offset and observe if it goes to 0 as a result to error
	st.SetClockOffset(1234)

	st.DoSync()

	assert.Equal(t, st.ClockOffset(), time.Nanosecond*23456)
}

func TestGetOffset(t *testing.T) {
	responseMock3 = &ntp.Response{ClockOffset: 23456}

	failNtpMock3 = false
	st := ntp2.SyncTime{}
	st.SetSyncPeriod(time.Millisecond)
	st.SetQuery(queryMock3)

	assert.Equal(t, st.ClockOffset(), time.Millisecond*0)
	st.DoSync()
	assert.Equal(t, st.ClockOffset(), time.Nanosecond*23456)
	assert.Equal(t, st.ClockOffset(), time.Nanosecond*23456)
}

func TestCallQuery(t *testing.T) {
	st := ntp2.NewSyncTime(time.Millisecond, queryMock4)

	assert.NotNil(t, st.Query())
	assert.Equal(t, time.Millisecond, st.SyncPeriod())

	//wait a few cycles
	time.Sleep(time.Millisecond * 100)

	assert.NotEqual(t, queryMock4Call, 0)

	fmt.Printf("Current time: %v\n", st.FormatedCurrentTime(st.ClockOffset()))
}
