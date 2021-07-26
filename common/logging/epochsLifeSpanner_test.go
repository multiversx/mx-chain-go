package logging

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestEpochsLifeSpanner_NewWithSmallerTimeSpanShouldErr(t *testing.T) {
	t.Parallel()

	esnwc := &testscommon.EpochStartNotifierStub{}
	els, err := newEpochsLifeSpanner(esnwc, 0)
	assert.Nil(t, els)
	assert.True(t, errors.Is(err, core.ErrInvalidLogFileMinLifeSpan))
}

func TestEpochsLifeSpanner_NewWithNilEpochStartNotifierShouldErr(t *testing.T) {
	t.Parallel()

	els, err := newEpochsLifeSpanner(nil, 3)
	assert.Nil(t, els)
	assert.True(t, errors.Is(err, core.ErrInvalidLogFileMinLifeSpan))
}

func TestEpochsLifeSpanner_NewShouldWork(t *testing.T) {
	t.Parallel()

	epochsLifeSpan := uint32(2)
	notifier := &testscommon.EpochStartNotifierStub{}
	els, err := newEpochsLifeSpanner(notifier, epochsLifeSpan)
	assert.Nil(t, err)
	assert.NotNil(t, els)
	assert.NotNil(t, els.GetChannel())
	assert.Equal(t, epochsLifeSpan, els.spanInEpochs)
}

func TestEpochsLifeSpanner_ChannelShouldCall(t *testing.T) {
	t.Parallel()

	epochsLifeSpan := uint32(2)
	notifier := &testscommon.EpochStartNotifierStub{}
	sls, _ := newEpochsLifeSpanner(notifier, epochsLifeSpan)
	open := true
	recreateCalled := 0
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, open = <-sls.GetChannel()
		recreateCalled++
		wg.Done()
	}()

	notifier.NotifyEpochChangeConfirmed(1)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 0, recreateCalled)

	notifier.NotifyEpochChangeConfirmed(epochsLifeSpan)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, recreateCalled)

	notifier.NotifyEpochChangeConfirmed(3)
	time.Sleep(10 * time.Millisecond)

	wg.Wait()

	assert.Equal(t, 1, recreateCalled)
	assert.True(t, open)
}

func TestEpochsLifeSpanner_CloseShouldWork(t *testing.T) {
	t.Parallel()

	epochsLifeSpan := uint32(1)
	notifier := &testscommon.EpochStartNotifierStub{}
	els, _ := newEpochsLifeSpanner(notifier, epochsLifeSpan)
	open := true
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, open = <-els.GetChannel()
		wg.Done()
	}()

	els.Close()

	wg.Wait()

	assert.False(t, open)
}

func TestEpochsLifeSpanner_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	epochsLifeSpan := uint32(2)
	notifier := &testscommon.EpochStartNotifierStub{}
	els, _ := newEpochsLifeSpanner(notifier, epochsLifeSpan)
	assert.False(t, els.IsInterfaceNil())
}
