package indexer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/indexer/client"
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewDataDispatcher_InvalidCacheSize(t *testing.T) {
	t.Parallel()

	dataDist, err := NewDataDispatcher(-1)

	require.Nil(t, dataDist)
	require.Equal(t, ErrNegativeCacheSize, err)
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
	elasticProc := &testscommon.ElasticProcessorStub{
		SaveRoundsInfoCalled: func(infos []types.RoundInfo) error {
			called = true
			wg.Done()
			return nil
		},
		SaveAccountsCalled: func(acc []*types.AccountEGLD) error {
			time.Sleep(7 * time.Second)
			return nil
		},
		SaveValidatorsRatingCalled: func(index string, validatorsRatingInfo []types.ValidatorRatingInfo) error {
			time.Sleep(6 * time.Second)
			return nil
		},
	}
	dispatcher.Add(workItems.NewItemRounds(elasticProc, []types.RoundInfo{}))
	wg.Wait()

	require.True(t, called)

	dispatcher.Add(workItems.NewItemAccounts(elasticProc, nil))
	wg.Add(1)
	dispatcher.Add(workItems.NewItemRounds(elasticProc, []types.RoundInfo{}))
	dispatcher.Add(workItems.NewItemRating(elasticProc, "", nil))
	wg.Add(1)
	dispatcher.Add(workItems.NewItemRounds(elasticProc, []types.RoundInfo{}))
	err = dispatcher.Close()
	require.NoError(t, err)
}

func TestDataDispatcher_Add(t *testing.T) {
	t.Parallel()

	dispatcher, err := NewDataDispatcher(100)
	require.NoError(t, err)
	dispatcher.StartIndexData()

	calledCount := uint32(0)
	wg := sync.WaitGroup{}
	wg.Add(1)
	elasticProc := &testscommon.ElasticProcessorStub{
		SaveRoundsInfoCalled: func(infos []types.RoundInfo) error {
			if calledCount < 2 {
				atomic.AddUint32(&calledCount, 1)
				return fmt.Errorf("%w: wrapped error", client.ErrBackOff)
			}

			atomic.AddUint32(&calledCount, 1)
			wg.Done()
			return nil
		},
	}

	start := time.Now()
	dispatcher.Add(workItems.NewItemRounds(elasticProc, []types.RoundInfo{}))
	wg.Wait()

	timePassed := time.Since(start)
	require.Greater(t, 2*int64(timePassed), int64(backOffTime))

	require.Equal(t, uint32(3), atomic.LoadUint32(&calledCount))

	err = dispatcher.Close()
	require.NoError(t, err)
}

func TestDataDispatcher_AddWithErrorShouldRetryTheReprocessing(t *testing.T) {
	t.Parallel()

	dispatcher, err := NewDataDispatcher(100)
	require.NoError(t, err)
	dispatcher.StartIndexData()

	calledCount := uint32(0)
	wg := sync.WaitGroup{}
	wg.Add(1)
	elasticProc := &testscommon.ElasticProcessorStub{
		SaveRoundsInfoCalled: func(infos []types.RoundInfo) error {
			if calledCount < 2 {
				atomic.AddUint32(&calledCount, 1)
				return errors.New("generic error")
			}

			atomic.AddUint32(&calledCount, 1)
			wg.Done()
			return nil
		},
	}

	start := time.Now()
	dispatcher.Add(workItems.NewItemRounds(elasticProc, []types.RoundInfo{}))
	wg.Wait()

	timePassed := time.Since(start)
	require.Greater(t, int64(timePassed), int64(2*durationBetweenErrorRetry))

	require.Equal(t, uint32(3), atomic.LoadUint32(&calledCount))

	err = dispatcher.Close()
	require.NoError(t, err)
}

func TestDataDispatcher_Close(t *testing.T) {
	t.Parallel()

	dispatcher, err := NewDataDispatcher(100)
	require.NoError(t, err)
	dispatcher.StartIndexData()

	elasticProc := &testscommon.ElasticProcessorStub{
		SaveRoundsInfoCalled: func(infos []types.RoundInfo) error {
			time.Sleep(1000*time.Millisecond + 200*time.Microsecond)
			return nil
		},
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func(c context.Context, w *sync.WaitGroup) {
		count := 0
		for {
			select {
			case <-c.Done():
				return
			default:
				count++
				if count == 105 {
					w.Done()
				}
				dispatcher.Add(workItems.NewItemRounds(elasticProc, []types.RoundInfo{}))
				time.Sleep(50 * time.Millisecond)
			}
		}
	}(ctx, wg)

	wg.Wait()

	err = dispatcher.Close()
	require.NoError(t, err)

	cancelFunc()
}
