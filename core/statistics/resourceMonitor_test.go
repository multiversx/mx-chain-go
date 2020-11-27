package statistics_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	stats "github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewResourceMonitor_NilConfigShouldErr(t *testing.T) {
	t.Parallel()

	resourceMonitor, err := stats.NewResourceMonitor(nil, &mock.PathManagerStub{}, "")

	assert.Equal(t, stats.ErrNilConfig, err)
	assert.Nil(t, resourceMonitor)
}

func TestNewResourceMonitor_NilPathManagerShouldErr(t *testing.T) {
	t.Parallel()

	resourceMonitor, err := stats.NewResourceMonitor(
		&config.Config{AccountsTrieStorage: config.StorageConfig{DB: config.DBConfig{}}},
		nil,
		"")

	assert.Equal(t, stats.ErrNilPathHandler, err)
	assert.Nil(t, resourceMonitor)
}

func TestResourceMonitor_NewResourceMonitorShouldPass(t *testing.T) {
	t.Parallel()

	resourceMonitor, err := stats.NewResourceMonitor(&config.Config{AccountsTrieStorage: config.StorageConfig{DB: config.DBConfig{}}}, &mock.PathManagerStub{}, "")

	assert.Nil(t, err)
	assert.NotNil(t, resourceMonitor)
}

func TestResourceMonitor_GenerateStatisticsShouldPass(t *testing.T) {
	t.Parallel()

	resourceMonitor, err := stats.NewResourceMonitor(&config.Config{AccountsTrieStorage: config.StorageConfig{DB: config.DBConfig{}}}, &mock.PathManagerStub{}, "")

	assert.Nil(t, err)
	statistics := resourceMonitor.GenerateStatistics()

	assert.NotNil(t, statistics)
}

func TestResourceMonitor_SaveStatisticsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("test should not have paniced: %v", r))
		}
	}()

	resourceMonitor, err := stats.NewResourceMonitor(&config.Config{AccountsTrieStorage: config.StorageConfig{DB: config.DBConfig{}}}, &mock.PathManagerStub{}, "")

	assert.Nil(t, err)
	resourceMonitor.SaveStatistics()
}
