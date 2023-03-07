package health

import (
	"context"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/require"
)

func TestHealthService_RegisterComponent_BadComponentsShouldErr(t *testing.T) {
	h := newHealthServiceToTest(42, 1)

	err := h.doRegisterComponent(&dummyNotDiagnosable{})
	require.Equal(t, errNotDiagnosableComponent, err)

	err = h.doRegisterComponent((*dummyDiagnosable)(nil))
	require.Equal(t, errNilComponent, err)
}

func TestHealthService_DiagnoseComponents(t *testing.T) {
	h := newHealthServiceToTest(42, 1)

	a := &dummyDiagnosable{}
	b := &dummyDiagnosable{}

	h.RegisterComponent(a)
	h.RegisterComponent(b)

	h.diagnoseComponents(false)
	h.diagnoseComponents(true)

	require.Equal(t, 1, int(a.numDeepDiagnoses.Get()))
	require.Equal(t, 1, int(b.numDeepDiagnoses.Get()))
	require.Equal(t, 1, int(a.numShallowDiagnoses.Get()))
	require.Equal(t, 1, int(b.numShallowDiagnoses.Get()))
}

func TestHealthService_MonitorMemory(t *testing.T) {
	h := newHealthServiceToTest(42, 1)

	// Memory still low, no record saved
	h.memory = newDummyMemory(41)
	h.monitorMemory()
	require.Equal(t, 0, h.records.len())

	// Memory is high now, record saved
	h.memory = newDummyMemory(43)
	h.monitorMemory()
	require.Equal(t, 1, h.records.len())

	// Check the saved record
	record := h.records.getMostImportant()
	recordAsMemoryUsageRecord, ok := record.(*memoryUsageRecord)
	filename := recordAsMemoryUsageRecord.getFilename()
	require.True(t, ok)
	require.NotNil(t, recordAsMemoryUsageRecord)
	require.Equal(t, uint64(43), recordAsMemoryUsageRecord.stats.HeapInuse)
	require.FileExists(t, filename)

	// Cleanup after test
	err := recordAsMemoryUsageRecord.delete()
	require.Nil(t, err)
}

func TestHealthService_MonitorContinuously(t *testing.T) {
	h := newHealthServiceToTest(42, 1)
	require.NotNil(t, h)

	// To keep things simpler, no memory records will be saved in this test (fixed to 41 bytes)
	memory := newDummyMemory(41)
	clock := newDummyClock()
	h.memory = memory
	h.clock = clock

	// Setup diagnosable components
	a := &dummyDiagnosable{}
	b := &dummyDiagnosable{}
	h.RegisterComponent(a)
	h.RegisterComponent(b)

	// Setup test synchronization
	chanMonitorBeginIteration := make(chan struct{})

	h.onMonitorContinuouslyBeginIteration = func() {
		chanMonitorBeginIteration <- dummySignal
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	go func() {
		h.monitorContinuously(ctx)
	}()

	for i := 0; i < 15; i++ {
		<-chanMonitorBeginIteration
		clock.tick()
	}

	<-chanMonitorBeginIteration
	cancelFunc()

	require.Equal(t, 10, int(memory.numGetStatsCalled.Get()))
	require.Equal(t, 3, int(a.numShallowDiagnoses.Get()))
	require.Equal(t, 2, int(a.numDeepDiagnoses.Get()))
}

func newHealthServiceToTest(highMemory int, intervalBase int) *healthService {
	return NewHealthService(
		config.HealthServiceConfig{
			MemoryUsageToCreateProfiles:               42,
			NumMemoryUsageRecordsToKeep:               10,
			IntervalVerifyMemoryInSeconds:             intervalBase * 1,
			IntervalDiagnoseComponentsInSeconds:       intervalBase * 5,
			IntervalDiagnoseComponentsDeeplyInSeconds: intervalBase * 7,
			FolderPath: ".",
		}, ".")
}
