package statistics_test

import (
	"os"
	"testing"

	stats "github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/stretchr/testify/assert"
)

func TestResourceMonitor_NewResourceMonitorNilFileShouldErr(t *testing.T) {
	t.Parallel()

	resourceMonitor, err := stats.NewResourceMonitor(nil)

	assert.Nil(t, resourceMonitor)
	assert.Equal(t, stats.ErrNilFileToWriteStats, err)
}

func TestResourceMonitor_NewResourceMonitorShouldPass(t *testing.T) {
	t.Parallel()

	resourceMonitor, err := stats.NewResourceMonitor(&os.File{})

	assert.NotNil(t, resourceMonitor)

	assert.Nil(t, err)
}

func TestResourceMonitor_GenerateStatisticsShouldPass(t *testing.T) {
	t.Parallel()

	resourceMonitor, err := stats.NewResourceMonitor(&os.File{})

	assert.Nil(t, err)

	statistics := resourceMonitor.GenerateStatistics()

	assert.Nil(t, err)
	assert.NotNil(t, statistics)
}

func TestResourceMonitor_SaveStatisticsShouldPass(t *testing.T) {
	t.Parallel()

	file, err := os.Create("test1")

	assert.Nil(t, err)

	resourceMonitor, _ := stats.NewResourceMonitor(file)

	err = resourceMonitor.SaveStatistics()

	if _, errF := os.Stat("test1"); errF == nil {
		_ = os.Remove("test1")
	}

	assert.Nil(t, err)

}

func TestResourceMonitor_SaveStatisticsCloseFileBeforeSaveShouldErr(t *testing.T) {
	t.Parallel()

	file, err := os.Create("test2")

	assert.Nil(t, err)

	resourceMonitor, _ := stats.NewResourceMonitor(file)

	err = resourceMonitor.Close()

	assert.Nil(t, err)

	err = resourceMonitor.SaveStatistics()

	if _, errF := os.Stat("test2"); errF == nil {
		_ = os.Remove("test2")
	}

	assert.Equal(t, stats.ErrNilFileToWriteStats, err)
}

func TestResourceMonitor_CloseShouldPass(t *testing.T) {
	t.Parallel()

	file, err := os.Create("test3")

	assert.Nil(t, err)

	resourceMonitor, err := stats.NewResourceMonitor(file)

	assert.Nil(t, err)

	err = resourceMonitor.Close()

	if _, errF := os.Stat("test3"); errF == nil {
		_ = os.Remove("test3")
	}

	assert.Nil(t, err)
}
