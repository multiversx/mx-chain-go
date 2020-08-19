package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-logger/redirects"
	"github.com/ElrondNetwork/elrond-go/core"
)

const minFileLifeSpan = time.Second
const defaultFileLifeSpan = time.Hour * 24

var log = logger.GetOrCreate("core/logging")

// FileLogging is able to rotate the log files
type FileLogging struct {
	chLifeSpanChanged chan time.Duration
	mutFile           sync.Mutex
	currentFile       *os.File
	workingDir        string
	defaultLogsPath   string
	cancelFunc        func()
	mutIsClosed       sync.Mutex
	isClosed          bool
}

// NewFileLogging creates a file log watcher used to break the log file into multiple smaller files
func NewFileLogging(workingDir string, defaultLogsPath string) (*FileLogging, error) {
	fl := &FileLogging{
		workingDir:        workingDir,
		defaultLogsPath:   defaultLogsPath,
		chLifeSpanChanged: make(chan time.Duration),
		isClosed:          false,
	}
	fl.recreateLogFile()

	//we need this function as to call file.Close() when the code panics and the defer func associated
	//with the file pointer in the main func will never be reached
	runtime.SetFinalizer(fl, func(fileLogHandler *FileLogging) {
		_ = fileLogHandler.currentFile.Close()
	})

	ctx, cancelFunc := context.WithCancel(context.Background())
	go fl.autoRecreateFile(ctx)
	fl.cancelFunc = cancelFunc

	return fl, nil
}

func (fl *FileLogging) createFile() (*os.File, error) {
	logDirectory := filepath.Join(fl.workingDir, fl.defaultLogsPath)

	return core.CreateFile("elrond-go", logDirectory, "log")
}

func (fl *FileLogging) recreateLogFile() {
	newFile, err := fl.createFile()
	if err != nil {
		log.Error("error creating new log file", "error", err)
		return
	}

	fl.mutFile.Lock()
	defer fl.mutFile.Unlock()

	oldFile := fl.currentFile
	err = logger.AddLogObserver(newFile, &logger.PlainFormatter{})
	if err != nil {
		log.Error("error adding log observer", "error", err)
		return
	}

	errNotCritical := redirects.RedirectStderr(newFile)
	if errNotCritical != nil {
		log.Error("error redirecting std error, moving on", "error", errNotCritical)
	}

	fl.currentFile = newFile

	if oldFile == nil {
		return
	}

	errNotCritical = oldFile.Close()
	if errNotCritical != nil {
		log.Error("error closing old log file, moving on...", "error", errNotCritical)
	}

	errNotCritical = logger.RemoveLogObserver(oldFile)
	if errNotCritical != nil {
		log.Error("error removing old log observer, moving on...", "error", errNotCritical)
	}
}

func (fl *FileLogging) autoRecreateFile(ctx context.Context) {
	fileLifeSpan := defaultFileLifeSpan
	for {
		select {
		case <-ctx.Done():
			log.Debug("closing fileLogging.autoRecreateFile go routine")
			return
		case <-time.After(fileLifeSpan):
			fl.recreateLogFile()
		case fileLifeSpan = <-fl.chLifeSpanChanged:
			log.Debug("changed log file span", "new value", fileLifeSpan)
		}
	}
}

// ChangeFileLifeSpan changes the log file span
func (fl *FileLogging) ChangeFileLifeSpan(newDuration time.Duration) error {
	if newDuration < minFileLifeSpan {
		return fmt.Errorf("%w, provided %v", core.ErrInvalidLogFileMinLifeSpan, newDuration)
	}

	fl.mutIsClosed.Lock()
	defer fl.mutIsClosed.Unlock()

	if fl.isClosed {
		return core.ErrFileLoggingProcessIsClosed
	}

	fl.chLifeSpanChanged <- newDuration
	return nil
}

// Close closes the file logging handler
func (fl *FileLogging) Close() error {
	fl.mutIsClosed.Lock()
	if fl.isClosed {
		fl.mutIsClosed.Unlock()
		return nil
	}

	fl.isClosed = true
	fl.mutIsClosed.Unlock()

	fl.mutFile.Lock()
	err := fl.currentFile.Close()
	fl.mutFile.Unlock()

	fl.cancelFunc()

	return err
}
