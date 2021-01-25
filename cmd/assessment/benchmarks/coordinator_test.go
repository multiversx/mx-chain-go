package benchmarks

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/cmd/assessment/mock"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCoordinator_NilSliceShouldErr(t *testing.T) {
	t.Parallel()

	c, err := NewCoordinator(nil)

	assert.True(t, check.IfNil(c))
	assert.True(t, errors.Is(err, ErrEmptyBenchmarksSlice))
}

func TestNewCoordinator_NilBenchmarkShouldErr(t *testing.T) {
	t.Parallel()

	c, err := NewCoordinator([]BenchmarkRunner{
		&mock.BenchmarkStub{},
		nil,
		&mock.BenchmarkStub{},
	})

	assert.True(t, check.IfNil(c))
	assert.True(t, errors.Is(err, ErrNilBenchmark))
}

func TestNewCoordinator_ShouldWork(t *testing.T) {
	t.Parallel()

	c, err := NewCoordinator([]BenchmarkRunner{
		&mock.BenchmarkStub{},
		&mock.BenchmarkStub{},
	})

	assert.False(t, check.IfNil(c))
	assert.Nil(t, err)
}

func TestCoordinator_RunAllOneErrorsShouldRetErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	c, _ := NewCoordinator([]BenchmarkRunner{
		&mock.BenchmarkStub{
			RunCalled: func() (time.Duration, error) {
				return 2, nil
			},
		},
		&mock.BenchmarkStub{
			RunCalled: func() (time.Duration, error) {
				return 3, expectedErr
			},
		},
	})

	result := c.RunAllTests()
	assert.Equal(t, expectedErr, result.Error)
	assert.Equal(t, time.Duration(5), result.TotalDuration)
	require.Equal(t, 2, len(result.Results))
	assert.Equal(t, time.Duration(2), result.Results[0].Duration)
	assert.Equal(t, time.Duration(3), result.Results[1].Duration)
	assert.Equal(t, expectedErr, result.Results[1].Error)
}

func TestCoordinator_RunAllShouldWork(t *testing.T) {
	t.Parallel()

	c, _ := NewCoordinator([]BenchmarkRunner{
		&mock.BenchmarkStub{
			RunCalled: func() (time.Duration, error) {
				return 2, nil
			},
		},
		&mock.BenchmarkStub{
			RunCalled: func() (time.Duration, error) {
				return 3, nil
			},
		},
	})

	result := c.RunAllTests()
	assert.Nil(t, result.Error)
	assert.Equal(t, time.Duration(5), result.TotalDuration)
	require.Equal(t, 2, len(result.Results))
	assert.Equal(t, time.Duration(2), result.Results[0].Duration)
	assert.Equal(t, time.Duration(3), result.Results[1].Duration)
}
