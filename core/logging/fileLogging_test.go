package logging

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

const logsDirectory = "logs"

func TestNewFileLogging_ShouldWork(t *testing.T) {
	t.Parallel()

	dir, _ := ioutil.TempDir("", "file_logging")
	defer func() {
		err := os.RemoveAll(dir)
		log.LogIfError(err)
	}()

	fl, err := NewFileLogging(dir, logsDirectory)

	assert.False(t, check.IfNil(fl))
	assert.Nil(t, err)
}

func TestNewFileLogging_CloseShouldStopCreatingLogFiles(t *testing.T) {
	t.Parallel()

	dir, _ := ioutil.TempDir("", "file_logging")
	defer func() {
		err := os.RemoveAll(dir)
		log.LogIfError(err)
	}()

	fl, _ := NewFileLogging(dir, logsDirectory)
	_ = fl.ChangeFileLifeSpan(time.Second)
	time.Sleep(time.Second*3 + time.Millisecond*200)

	err := fl.Close()
	assert.Nil(t, err)

	//wait to see if the generating go routine really stopped
	time.Sleep(time.Second * 2)

	dirToSearchFiles := filepath.Join(dir, logsDirectory)
	files, _ := ioutil.ReadDir(dirToSearchFiles)

	assert.Equal(t, 4, len(files))
}

func TestNewFileLogging_CloseCallTwiceShouldWork(t *testing.T) {
	t.Parallel()

	dir, _ := ioutil.TempDir("", "file_logging")
	defer func() {
		err := os.RemoveAll(dir)
		log.LogIfError(err)
	}()

	fl, _ := NewFileLogging(dir, logsDirectory)

	err := fl.Close()
	assert.Nil(t, err)

	err = fl.Close()
	assert.Nil(t, err)
}

func TestFileLogging_ChangeFileLifeSpanInvalidValueShouldErr(t *testing.T) {
	t.Parallel()

	dir, _ := ioutil.TempDir("", "file_logging")
	defer func() {
		err := os.RemoveAll(dir)
		log.LogIfError(err)
	}()

	fl, _ := NewFileLogging(dir, logsDirectory)
	err := fl.ChangeFileLifeSpan(time.Millisecond)

	assert.True(t, errors.Is(err, core.ErrInvalidLogFileMinLifeSpan))
}

func TestFileLogging_ChangeFileLifeSpanAfterCloseShouldErr(t *testing.T) {
	t.Parallel()

	dir, _ := ioutil.TempDir("", "file_logging")
	defer func() {
		err := os.RemoveAll(dir)
		log.LogIfError(err)
	}()

	fl, _ := NewFileLogging(dir, logsDirectory)
	err := fl.ChangeFileLifeSpan(time.Second)
	assert.Nil(t, err)

	_ = fl.Close()

	err = fl.ChangeFileLifeSpan(time.Second)
	assert.True(t, errors.Is(err, core.ErrFileLoggingProcessIsClosed))
}
