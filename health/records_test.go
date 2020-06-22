package health

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecords_AddRecord(t *testing.T) {
	records := newRecords(10)
	records.addRecord(newDummyRecord(1))
	records.addRecord(newDummyRecord(2))
	records.addRecord(newDummyRecord(3))
	require.Equal(t, 3, records.len())
}

func TestRecords_AddRecordDoesNotAddLessImportantWhenFullCapacity(t *testing.T) {
	records := newRecords(3)
	a := newDummyRecord(42)
	b := newDummyRecord(42)
	c := newDummyRecord(42)
	d := newDummyRecord(41)

	records.addRecord(a)
	records.addRecord(b)
	records.addRecord(c)
	records.addRecord(d)

	require.True(t, a.saved)
	require.True(t, b.saved)
	require.True(t, c.saved)
	require.False(t, d.saved)
	require.Equal(t, 3, records.len())
}

func TestRecords_AddRecordEvictsLessImportantWhenFullCapacity(t *testing.T) {
	records := newRecords(3)
	a := newDummyRecord(42)
	b := newDummyRecord(41)
	c := newDummyRecord(42)
	d := newDummyRecord(100)

	records.addRecord(a)
	records.addRecord(b)
	records.addRecord(c)
	records.addRecord(d)

	require.True(t, a.saved && !a.deleted)
	require.True(t, b.saved && b.deleted)
	require.True(t, c.saved && !c.deleted)
	require.True(t, d.saved && !d.deleted)
	require.Equal(t, 3, records.len())
}

func TestRecords_ConcurrentAddition(t *testing.T) {
	records := newRecords(100)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for i := 0; i < 1000; i++ {
			records.addRecord(newDummyRecord(i))
		}

		wg.Done()
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			records.addRecord(newDummyRecord(i))
		}

		wg.Done()
	}()

	wg.Wait()
	require.Equal(t, 100, records.len())
}
