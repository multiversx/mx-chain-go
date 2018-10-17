package ntp

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLocalGetOffset(t *testing.T) {
	lt := LocalTime{}

	assert.Equal(t, lt.clockOffset, time.Millisecond*0)
}

func TestLocalCallQuery(t *testing.T) {
	lt := LocalTime{}

	ct := lt.CurrentTime(lt.clockOffset)

	//wait a few cycles
	time.Sleep(time.Millisecond * 10)

	assert.NotEqual(t, ct, lt.CurrentTime(lt.clockOffset))

	fmt.Printf("Last time: %v\nCurrent time: %v\n", lt.FormatTime(ct), lt.FormatedCurrentTime(lt.clockOffset))
}
