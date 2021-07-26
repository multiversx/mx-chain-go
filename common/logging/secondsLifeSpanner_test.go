package logging

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/stretchr/testify/assert"
)

func TestSecondsLifeSpanner_NewWithSmallerTimeSpanShouldErr(t *testing.T) {
	t.Parallel()

	sls, err := newSecondsLifeSpanner(time.Millisecond)
	assert.Nil(t, sls)
	assert.True(t, errors.Is(err, core.ErrInvalidLogFileMinLifeSpan))
}

func TestSecondsLifeSpanner_NewShouldWork(t *testing.T) {
	t.Parallel()

	timeSpan := time.Minute
	sls, err := newSecondsLifeSpanner(timeSpan)
	assert.Nil(t, err)
	assert.NotNil(t, sls)
	assert.Equal(t, timeSpan, sls.timeSpanInSeconds)
}

func TestSecondsLifeSpanner_ChannelShouldCall(t *testing.T) {
	t.Parallel()

	timeSpan := time.Second
	sls, _ := newSecondsLifeSpanner(timeSpan)
	open := true
	recreateCalled := false
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, open = <-sls.lifeSpanChannel
		recreateCalled = true
		wg.Done()
	}()

	time.Sleep(1100 * time.Millisecond)

	wg.Wait()

	assert.True(t, open)
	assert.True(t, recreateCalled)
}

func TestSecondsLifeSpanner_CloseShouldWork(t *testing.T) {
	sls, _ := newSecondsLifeSpanner(time.Minute)
	open := true
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, open = <-sls.lifeSpanChannel
		wg.Done()
	}()

	_ = sls.Close()

	wg.Wait()

	assert.False(t, open)
}

func TestSecondsLifeSpanner_IsInterfaceNil(t *testing.T) {
	sls, _ := newSecondsLifeSpanner(time.Second)
	assert.False(t, sls.IsInterfaceNil())
}
