package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeMeasure_Start(t *testing.T) {
	t.Parallel()

	tm := NewTimeMeasure()
	identifier := "identifier"

	tm.Start(identifier)

	_, has := tm.started[identifier]

	assert.True(t, has)
	assert.Equal(t, identifier, tm.identifiers[0])
}

func TestTimeMeasure_DoubleStartShouldNotReAddInIdentifiers(t *testing.T) {
	t.Parallel()

	tm := NewTimeMeasure()
	identifier1 := "identifier1"
	identifier2 := "identifier2"

	tm.Start(identifier1)
	tm.Start(identifier2)
	tm.Start(identifier1)

	assert.Equal(t, identifier1, tm.identifiers[0])
	assert.Equal(t, identifier2, tm.identifiers[1])
	assert.Equal(t, 2, len(tm.identifiers))
}

func TestTimeMeasure_FinishNoStartShouldNotAddDuration(t *testing.T) {
	t.Parallel()

	tm := NewTimeMeasure()
	identifier := "identifier"

	tm.Finish(identifier)

	_, has := tm.elapsed[identifier]

	assert.False(t, has)
}

func TestTimeMeasure_FinishWithStartShouldAddDuration(t *testing.T) {
	t.Parallel()

	tm := NewTimeMeasure()
	identifier := "identifier"

	tm.Start(identifier)
	tm.Finish(identifier)

	_, has := tm.elapsed[identifier]

	assert.True(t, has)
}

func TestTimeMeasure_GetMeasurementsNotFinishedShouldOmit(t *testing.T) {
	t.Parallel()

	tm := NewTimeMeasure()
	identifier := "identifier"

	tm.Start(identifier)

	measurements := tm.GetMeasurements()
	log.Info("measurements", measurements...)

	assert.Equal(t, 0, len(measurements))
}

func TestTimeMeasure_GetMeasurementsShouldWork(t *testing.T) {
	t.Parallel()

	tm := NewTimeMeasure()
	identifier := "identifier"

	tm.Start(identifier)
	tm.Finish(identifier)

	measurements := tm.GetMeasurements()
	log.Info("measurements", measurements...)

	assert.Equal(t, 2, len(measurements))
	assert.Equal(t, identifier, measurements[0])
}

func TestTimeMeasure_AddShouldWork(t *testing.T) {
	t.Parallel()

	identifier1 := "identifier1"
	duration1 := time.Duration(5)
	identifier2 := "identifier2"
	duration2 := time.Duration(7)

	tmSrc := NewTimeMeasure()
	tmSrc.identifiers = []string{identifier1, identifier2}
	tmSrc.elapsed[identifier1] = duration1
	tmSrc.elapsed[identifier2] = duration2

	tm := NewTimeMeasure()

	tm.Add(tmSrc)

	data, _ := tm.GetContainingDuration()
	assert.Equal(t, duration1, data[identifier1])
	assert.Equal(t, duration2, data[identifier2])

	tm.Add(tmSrc)

	data, _ = tm.GetContainingDuration()
	assert.Equal(t, duration1*2, data[identifier1])
	assert.Equal(t, duration2*2, data[identifier2])
}
