package ntp

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var responseMock1 *Response
var failNtpMock1 = false
var responseMock2 *Response
var failNtpMock2 = false
var responseMock3 *Response
var failNtpMock3 = false
var responseMock5 *Response
var failNtpMock5 = false

var errNtpMock = errors.New("NTP Mock generic error")
var queryMock4Call = 0

func queryMock1(host string) (*Response, error) {
	if failNtpMock1 {
		return nil, errNtpMock
	}

	return responseMock1, nil
}

func queryMock2(host string) (*Response, error) {
	if failNtpMock2 {
		return nil, errNtpMock
	}

	return responseMock2, nil
}

func queryMock3(host string) (*Response, error) {
	if failNtpMock3 {
		return nil, errNtpMock
	}

	return responseMock3, nil
}

func queryMock4(host string) (*Response, error) {
	queryMock4Call++

	return nil, errNtpMock
}

func queryMock5(host string) (*Response, error) {
	if failNtpMock5 {
		return nil, errNtpMock
	}

	return responseMock5, nil
}

func TestHandleErrorInDoSync(t *testing.T) {
	//rm := Response{ClockOffset:23456}

	failNtpMock1 = true
	st := SyncTime{clockOffset: 0, syncPeriod: time.Millisecond, query: queryMock1}

	st.doSync()
	assert.Equal(t, st.clockOffset, time.Millisecond*0)

	//manually put a value in clockOffset and observe if it goes to 0 as a result to error
	st.mut.Lock()
	st.clockOffset = 1234
	st.mut.Unlock()

	st.doSync()

	st.mut.Lock()
	assert.Equal(t, st.clockOffset, time.Millisecond*0)
	st.mut.Unlock()

}

func TestValueInDoSync(t *testing.T) {
	responseMock2 = &Response{ClockOffset: 23456}

	failNtpMock2 = false
	st := SyncTime{clockOffset: 0, syncPeriod: time.Millisecond, query: queryMock2}

	assert.Equal(t, st.clockOffset, time.Millisecond*0)
	st.doSync()
	assert.Equal(t, st.clockOffset, time.Nanosecond*23456)

	//manually put a value in clockOffset and observe if it goes to 0 as a result to error
	st.mut.Lock()
	st.clockOffset = 1234
	st.mut.Unlock()

	st.doSync()

	st.mut.Lock()
	assert.Equal(t, st.clockOffset, time.Nanosecond*23456)
	st.mut.Unlock()
}

func TestGetOffset(t *testing.T) {
	responseMock3 = &Response{ClockOffset: 23456}

	failNtpMock3 = false
	st := SyncTime{clockOffset: 0, syncPeriod: time.Millisecond, query: queryMock3}

	assert.Equal(t, st.clockOffset, time.Millisecond*0)
	st.doSync()
	assert.Equal(t, st.clockOffset, time.Nanosecond*23456)
	assert.Equal(t, st.GetClockOffset(), time.Nanosecond*23456)
}

func TestCallQuery(t *testing.T) {
	st := NewSyncTime(time.Millisecond, queryMock4)

	//wait a few cycles
	time.Sleep(time.Millisecond * 100)

	assert.NotEqual(t, queryMock4Call, 0)

	fmt.Printf("Current time: %v\n", st.GetFormatedCurrentTime(st.clockOffset))
}
