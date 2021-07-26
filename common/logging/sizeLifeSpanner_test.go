package logging

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestSizeLifeSpanner_NewWithSmallerTimeSpanShouldErr(t *testing.T) {
	t.Parallel()

	fsc := &mock.FileSizeCheckStub{}
	sls, err := newSizeLifeSpanner(fsc, 0, time.Second)
	assert.Nil(t, sls)
	assert.True(t, errors.Is(err, core.ErrInvalidLogFileMinLifeSpan))
}

func TestSizeLifeSpanner_NewWithNilFileSizeCheckerShouldErr(t *testing.T) {
	t.Parallel()

	sls, err := newSizeLifeSpanner(nil, 2, time.Second)
	assert.Nil(t, sls)
	assert.True(t, errors.Is(err, core.ErrInvalidLogFileMinLifeSpan))
}

func TestSizeLifeSpanner_NewShouldWork(t *testing.T) {
	t.Parallel()

	fsc := &mock.FileSizeCheckStub{}
	sizeLifeSpanInMB := uint32(2)
	els, err := newSizeLifeSpanner(fsc, sizeLifeSpanInMB, time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, els)
	assert.NotNil(t, els.fileSizeChecker)
	assert.NotNil(t, els.GetChannel())
	sizeInMB := 1024 * 1024 * sizeLifeSpanInMB
	assert.Equal(t, sizeInMB, els.spanInMB)
}

func TestSizeLifeSpanner_ChannelShouldCall(t *testing.T) {
	t.Parallel()

	fsc := &mock.FileSizeCheckStub{}
	sizeLifeSpanInMB := uint32(2)
	sls, _ := newSizeLifeSpanner(fsc, sizeLifeSpanInMB, time.Second)
	open := true
	recreateCalled := 0
	currentFile := "test"
	sls.SetCurrentFile(currentFile)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, open = <-sls.GetChannel()
		recreateCalled++
		wg.Done()
	}()

	sizeLifeSpan := sizeLifeSpanInMB * 1024 * 1024
	fsc.GetSizeCalled = func(s string) (int64, error) {
		return int64(sizeLifeSpan + 10), nil
	}

	assert.Equal(t, 0, recreateCalled)

	time.Sleep(1100 * time.Millisecond)

	assert.Equal(t, 1, recreateCalled)

	wg.Wait()

	assert.Equal(t, 1, recreateCalled)
	assert.True(t, open)
}

func TestSizeLifeSpanner_CloseShouldWork(t *testing.T) {
	t.Parallel()

	fsc := &mock.FileSizeCheckStub{}
	sizeLifeSpanInMB := uint32(2)
	sls, _ := newSizeLifeSpanner(fsc, sizeLifeSpanInMB, time.Second)
	open := true
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, open = <-sls.GetChannel()
		wg.Done()
	}()

	sls.Close()

	wg.Wait()

	assert.False(t, open)
}

func TestSizeLifeSpanner_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	epochsLifeSpan := uint32(2)
	notifier := &testscommon.EpochStartNotifierStub{}
	els, _ := newEpochsLifeSpanner(notifier, epochsLifeSpan)
	assert.False(t, els.IsInterfaceNil())
}
