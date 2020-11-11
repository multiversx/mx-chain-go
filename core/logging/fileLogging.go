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

// fileLogging is able to rotate the log files
type fileLogging struct {
	chLifeSpanChanged chan time.Duration
	mutFile           sync.Mutex
	currentFile       *os.File
	workingDir        string
	defaultLogsPath   string
	logFilePrefix     string
	cancelFunc        func()
	mutIsClosed       sync.Mutex
	isClosed          bool
}

// NewFileLogging creates a file log watcher used to break the log file into multiple smaller files
func NewFileLogging(workingDir string, defaultLogsPath string, logFilePrefix string) (*fileLogging, error) {
	fl := &fileLogging{
		workingDir:        workingDir,
		defaultLogsPath:   defaultLogsPath,
		logFilePrefix:     logFilePrefix,
		chLifeSpanChanged: make(chan time.Duration),
		isClosed:          false,
	}
	fl.recreateLogFile()

	//we need this function as to call file.Close() when the code panics and the defer func associated
	//with the file pointer in the main func will never be reached
	runtime.SetFinalizer(fl, func(fileLogHandler *fileLogging) {
		_ = fileLogHandler.currentFile.Close()
	})

	ctx, cancelFunc := context.WithCancel(context.Background())
	go fl.autoRecreateFile(ctx)
	fl.cancelFunc = cancelFunc

	return fl, nil
}

func (fl *fileLogging) createFile() (*os.File, error) {
	logDirectory := filepath.Join(fl.workingDir, fl.defaultLogsPath)

	return core.CreateFile(
		core.ArgCreateFileArgument{
			Prefix:        fl.logFilePrefix,
			Directory:     logDirectory,
			FileExtension: "log",
		},
	)
}

func (fl *fileLogging) recreateLogFile() {
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
	log.LogIfError(errNotCritical, "step", "redirecting std error")

	fl.currentFile = newFile

	if oldFile == nil {
		return
	}

	errNotCritical = oldFile.Close()
	log.LogIfError(errNotCritical, "step", "closing old log file")

	errNotCritical = logger.RemoveLogObserver(oldFile)
	log.LogIfError(errNotCritical, "step", "removing old log observer")
}

func (fl *fileLogging) autoRecreateFile(ctx context.Context) {
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
func (fl *fileLogging) ChangeFileLifeSpan(newDuration time.Duration) error {
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
func (fl *fileLogging) Close() error {
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

// IsInterfaceNil returns true if there is no value under the interface
func (fl *fileLogging) IsInterfaceNil() bool {
	return fl == nil
}
