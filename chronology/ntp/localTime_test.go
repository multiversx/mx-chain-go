package ntp_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/stretchr/testify/assert"
)

func TestLocalGetOffset(t *testing.T) {
	lt := ntp.LocalTime{}

	assert.Equal(t, lt.ClockOffset(), time.Millisecond*0)
}

func TestLocalCallQuery(t *testing.T) {
	lt := ntp.LocalTime{}

	assert.Equal(t, time.Duration(0), lt.ClockOffset())

	lt.SetClockOffset(100)
	assert.Equal(t, time.Duration(100), lt.ClockOffset())

	ct := lt.CurrentTime(lt.ClockOffset())

	//wait a few cycles
	time.Sleep(time.Millisecond * 10)

	assert.NotEqual(t, ct, lt.CurrentTime(lt.ClockOffset()))

	fmt.Printf("Last time: %v\nCurrent time: %v\n", lt.FormatTime(ct), lt.FormatedCurrentTime(lt.ClockOffset()))
}
