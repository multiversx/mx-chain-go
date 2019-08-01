package statistics_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	stats "github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/stretchr/testify/assert"
)

var errNilFileToWriteStats = errors.New("nil file to write statistics")

func TestResourceMonitor_NewResourceMonitorNilFileShouldErr(t *testing.T) {
	t.Parallel()

	resourceMonitor, err := stats.NewResourceMonitor(nil)

	assert.Nil(t, resourceMonitor)
	assert.Equal(t, errNilFileToWriteStats, err)
}

func TestResourceMonitor_NewResourceMonitorShouldPass(t *testing.T) {
	t.Parallel()

	resourceMonitor, err := stats.NewResourceMonitor(&os.File{})

	fmt.Println(resourceMonitor)

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

	file, err := os.Create("test")

	assert.Nil(t, err)

	resourceMonitor, _ := stats.NewResourceMonitor(file)

	err = resourceMonitor.SaveStatistics()

	if _, errF := os.Stat("test"); errF == nil {
		_ = os.Remove("test")
	}

	assert.Nil(t, err)

}

func TestResourceMonitor_SaveStatisticsCloseFileBeforeSaveShouldErr(t *testing.T) {
	t.Parallel()

	file, err := os.Create("test")

	assert.Nil(t, err)

	resourceMonitor, _ := stats.NewResourceMonitor(file)

	err = resourceMonitor.Close()

	assert.Nil(t, err)

	err = resourceMonitor.SaveStatistics()

	if _, errF := os.Stat("test"); errF == nil {
		_ = os.Remove("test")
	}

	assert.Equal(t, errNilFileToWriteStats, err)
}

func TestResourceMonitor_CloseShouldPass(t *testing.T) {
	t.Parallel()

	file, err := os.Create("test")

	assert.Nil(t, err)

	resourceMonitor, err := stats.NewResourceMonitor(file)

	assert.Nil(t, err)

	err = resourceMonitor.Close()

	if _, errF := os.Stat("test"); errF == nil {
		_ = os.Remove("test")
	}

	assert.Nil(t, err)
}
