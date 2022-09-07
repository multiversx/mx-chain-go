package logging

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const logsDirectory = "logs"

func createMockArgs(t *testing.T) ArgsFileLogging {
	return ArgsFileLogging{
		WorkingDir:      t.TempDir(),
		DefaultLogsPath: logsDirectory,
		LogFilePrefix:   "log",
	}
}

func TestNewFileLogging(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		args := createMockArgs(t)
		fl, err := NewFileLogging(args)

		assert.False(t, check.IfNil(fl))
		assert.Nil(t, err)
		_ = fl.Close()
	})
}

func TestNewFileLogging_CloseShouldStopCreatingLogFiles(t *testing.T) {
	t.Parallel()

	args := createMockArgs(t)
	fl, _ := NewFileLogging(args)
	_ = fl.ChangeFileLifeSpan(time.Second, 5)
	time.Sleep(time.Second*3 + time.Millisecond*200)

	err := fl.Close()
	assert.Nil(t, err)

	// wait to see if the generating go routine really stopped
	time.Sleep(time.Second * 2)

	dirToSearchFiles := filepath.Join(args.WorkingDir, logsDirectory)
	files, _ := ioutil.ReadDir(dirToSearchFiles)

	assert.Equal(t, 4, len(files))
}

func TestNewFileLogging_CloseCallTwiceShouldWork(t *testing.T) {
	t.Parallel()

	fl, _ := NewFileLogging(createMockArgs(t))

	err := fl.Close()
	assert.Nil(t, err)

	err = fl.Close()
	assert.Nil(t, err)
}

func TestFileLogging_ChangeFileLifeSpanInvalidValuesShouldErr(t *testing.T) {
	t.Parallel()

	fl, _ := NewFileLogging(createMockArgs(t))
	t.Run("invalid time life span", func(t *testing.T) {
		err := fl.ChangeFileLifeSpan(time.Millisecond, 5)

		assert.True(t, errors.Is(err, errInvalidParameter))
		assert.True(t, strings.Contains(err.Error(), "for the life span duration, minimum:"))
	})
	t.Run("under the minimum allowed size in MB", func(t *testing.T) {
		err := fl.ChangeFileLifeSpan(minFileLifeSpan, minSizeInMB-1)

		assert.True(t, errors.Is(err, errInvalidParameter))
		assert.True(t, strings.Contains(err.Error(), "for the life span in MB, minimum:"))
	})
	t.Run("over the maximum allowed size in MB", func(t *testing.T) {
		err := fl.ChangeFileLifeSpan(minFileLifeSpan, maxSizeInMB+1)

		assert.True(t, errors.Is(err, errInvalidParameter))
		assert.True(t, strings.Contains(err.Error(), "for the life span in MB, maximum:"))
	})

	_ = fl.Close()
}

func TestFileLogging_ChangeFileLifeSpanAfterCloseShouldErr(t *testing.T) {
	t.Parallel()

	fl, _ := NewFileLogging(createMockArgs(t))
	err := fl.ChangeFileLifeSpan(time.Second, 5)
	assert.Nil(t, err)

	_ = fl.Close()

	err = fl.ChangeFileLifeSpan(time.Second, 5)
	assert.True(t, errors.Is(err, core.ErrFileLoggingProcessIsClosed))
	_ = fl.Close()
}

func TestFileLogging_sizeReached(t *testing.T) {
	t.Parallel()

	f, err := core.CreateFile(core.ArgCreateFileArgument{
		Directory:     t.TempDir(),
		Prefix:        "",
		FileExtension: "test",
	})
	require.Nil(t, err)
	defer func() {
		_ = f.Close()
	}()

	fl := &fileLogging{}
	data := []byte("data")

	t.Run("nil file should return false", func(t *testing.T) {
		assert.False(t, fl.sizeReached())
	})
	fl.currentFile = f
	t.Run("file is new, life span size is 0", func(t *testing.T) {
		assert.True(t, fl.sizeReached())
	})
	t.Run("file contains data", func(t *testing.T) {
		_, err = f.Write(data)
		require.Nil(t, err)
		fl.lifeSpanSize = 0
		assert.True(t, fl.sizeReached())

		fl.lifeSpanSize = uint64(len(data))
		assert.True(t, fl.sizeReached())

		fl.lifeSpanSize = uint64(len(data) + 1)
		assert.False(t, fl.sizeReached())
	})
}
