package logs_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/api/logs"
	"github.com/stretchr/testify/assert"
)

func TestNewLogWriter(t *testing.T) {
	t.Parallel()

	lw := logs.NewLogWriter()

	assert.NotNil(t, lw)
}

func TestLogWriter_WriteNilShouldReturnNilAndZero(t *testing.T) {
	t.Parallel()

	lw := logs.NewLogWriter()

	n, err := lw.Write(nil)

	assert.Nil(t, err)
	assert.Equal(t, 0, n)
}

func TestLogWriter_WriteNilMultipleTimeNoReadShouldReturnNilAndZero(t *testing.T) {
	t.Parallel()

	lw := logs.NewLogWriter()

	numWritesNil := 10
	for i := 0; i < numWritesNil; i++ {
		n, err := lw.Write(nil)

		assert.Nil(t, err)
		assert.Equal(t, 0, n)
	}
}

func TestLogWriter_FirstWriteShouldWork(t *testing.T) {
	t.Parallel()

	lw := logs.NewLogWriter()
	data := []byte("data")

	n, err := lw.Write(data)

	assert.Nil(t, err)
	assert.Equal(t, len(data), n)
}

func TestLogWriter_SecondWriteNoReadShouldErr(t *testing.T) {
	t.Parallel()

	lw := logs.NewLogWriter()
	data := []byte("data")

	_, _ = lw.Write(data)
	n, err := lw.Write(data)

	assert.Equal(t, logs.ErrWriterBusy, err)
	assert.Equal(t, 0, n)
}

func TestLogWriter_ReadShouldWork(t *testing.T) {
	t.Parallel()

	lw := logs.NewLogWriter()
	data := []byte("data")
	_, _ = lw.Write(data)

	recoveredData, ok := lw.ReadBlocking()

	assert.Equal(t, data, recoveredData)
	assert.True(t, ok)
}

func TestLogWriter_SecondWriteWithReadShouldWork(t *testing.T) {
	t.Parallel()

	lw := logs.NewLogWriter()
	data := []byte("data")

	_, _ = lw.Write(data)
	_, _ = lw.ReadBlocking()
	n, err := lw.Write(data)

	assert.Nil(t, err)
	assert.Equal(t, len(data), n)
}

//-------- Close

func TestLogWriter_CloseShouldWork(t *testing.T) {
	t.Parallel()

	lw := logs.NewLogWriter()

	err := lw.Close()

	assert.Nil(t, err)
}

func TestLogWriter_DoubleClosingShouldErr(t *testing.T) {
	t.Parallel()

	lw := logs.NewLogWriter()

	_ = lw.Close()
	err := lw.Close()

	assert.Equal(t, logs.ErrWriterClosed, err)
}

func TestLogWriter_ClosedWriterShouldNotWriteShouldErr(t *testing.T) {
	t.Parallel()

	lw := logs.NewLogWriter()
	_ = lw.Close()

	n, err := lw.Write([]byte("data"))

	assert.Equal(t, logs.ErrWriterClosed, err)
	assert.Equal(t, 0, n)
}
