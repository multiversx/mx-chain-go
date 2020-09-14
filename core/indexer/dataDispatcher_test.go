package indexer

import (
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/stretchr/testify/require"
)

func TestNewDataDispatcher_InvalidCacheSize(t *testing.T) {
	t.Parallel()

	dispatcher, err := NewDataDispatcher(0)
	require.Equal(t, ErrInvalidCacheSize, err)
	require.Nil(t, dispatcher)
}

func TestNewDataDispatcher(t *testing.T) {
	t.Parallel()

	dispatcher, err := NewDataDispatcher(100)
	require.NoError(t, err)
	require.NotNil(t, dispatcher)
}

func TestDataDispatcher_StartIndexDataClose(t *testing.T) {
	t.Parallel()

	dispatcher, err := NewDataDispatcher(100)
	require.NoError(t, err)
	dispatcher.StartIndexData()

	called := false
	wg := sync.WaitGroup{}
	wg.Add(1)
	elasticProcessor := &mock.ElasticProcessorMock{
		SaveRoundsInfoCalled: func(infos []workItems.RoundInfo) error {
			called = true
			wg.Done()
			return nil
		},
	}
	dispatcher.Add(workItems.NewItemRounds(elasticProcessor, []workItems.RoundInfo{}))
	wg.Wait()

	require.True(t, called)

	err = dispatcher.Close()
	require.NoError(t, err)
}

func TestDataDispatcher_Add(t *testing.T) {
	t.Parallel()

	dispatcher, err := NewDataDispatcher(100)
	require.NoError(t, err)
	dispatcher.StartIndexData()

	calledCount := 0
	wg := sync.WaitGroup{}
	wg.Add(1)
	elasticProcessor := &mock.ElasticProcessorMock{
		SaveRoundsInfoCalled: func(infos []workItems.RoundInfo) error {
			if calledCount <= 1 {
				calledCount++
				return ErrBackOff
			}

			calledCount++
			wg.Done()
			return nil
		},
	}

	start := time.Now()
	dispatcher.Add(workItems.NewItemRounds(elasticProcessor, []workItems.RoundInfo{}))
	wg.Wait()

	timePassed := time.Since(start)
	require.Greater(t, 2*int64(timePassed), int64(backOffTime))

	require.Equal(t, 3, calledCount)

	err = dispatcher.Close()
	require.NoError(t, err)
}
