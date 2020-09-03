package statistics_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	stats "github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/stretchr/testify/assert"
)

func TestResourceMonitor_NewResourceMonitorShouldPass(t *testing.T) {
	t.Parallel()

	resourceMonitor := stats.NewResourceMonitor()

	assert.NotNil(t, resourceMonitor)
}

func TestResourceMonitor_GenerateStatisticsShouldPass(t *testing.T) {
	t.Parallel()

	resourceMonitor := stats.NewResourceMonitor()

	statistics := resourceMonitor.GenerateStatistics(&config.Config{AccountsTrieStorage: config.StorageConfig{DB: config.DBConfig{}}}, &mock.PathManagerStub{}, "")

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

	resourceMonitor := stats.NewResourceMonitor()

	resourceMonitor.SaveStatistics(&config.Config{AccountsTrieStorage: config.StorageConfig{DB: config.DBConfig{}}}, &mock.PathManagerStub{}, "")
}
