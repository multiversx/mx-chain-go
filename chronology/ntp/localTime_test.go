package ntp

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLocalGetOffset(t *testing.T) {
	lt := LocalTime{ClockOffset: 0}

	assert.Equal(t, lt.GetClockOffset(), time.Millisecond*0)
}

func TestLocalCallQuery(t *testing.T) {
	lt := LocalTime{ClockOffset: 0}

	ct := lt.GetCurrentTime(lt.GetClockOffset())

	//wait a few cycles
	time.Sleep(time.Millisecond * 10)

	assert.NotEqual(t, ct, lt.GetCurrentTime(lt.GetClockOffset()))

	fmt.Printf("Last time: %v\nCurrent time: %v\n", lt.FormatTime(ct), lt.GetFormatedCurrentTime(lt.GetClockOffset()))
}
