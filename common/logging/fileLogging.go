package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-logger/lifespan"
	"github.com/ElrondNetwork/elrond-go-logger/redirects"
)

const defaultFileLifeSpan = 86400

var log = logger.GetOrCreate("common/logging")

// fileLogging is able to rotate the log files
type fileLogging struct {
	logLifeSpanner  logger.LogLifeSpanner
	mutFile         sync.Mutex
	currentFile     *os.File
	workingDir      string
	defaultLogsPath string
	logFilePrefix   string
	cancelFunc      func()
	mutIsClosed     sync.Mutex
	isClosed        bool
}

// NewFileLogging creates a file log watcher used to break the log file into multiple smaller files
func NewFileLogging(workingDir string, defaultLogsPath string, logFilePrefix string) (*fileLogging, error) {
	fl := &fileLogging{
		workingDir:      workingDir,
		defaultLogsPath: defaultLogsPath,
		logFilePrefix:   logFilePrefix,
		isClosed:        false,
	}
	fl.recreateLogFile("")

	//we need this function as to call file.Close() when the code panics and the defer func associated
	//with the file pointer in the main func will never be reached
	runtime.SetFinalizer(fl, func(fileLogHandler *fileLogging) {
		_ = fileLogHandler.currentFile.Close()
	})

	factory := lifespan.NewTypeLogLifeSpanFactory()

	secondsLifeSpanner, err := factory.CreateLogLifeSpanner(logger.LogLifeSpanFactoryArgs{
		LifeSpanType:  "second",
		RecreateEvery: defaultFileLifeSpan,
	})

	if err != nil {
		log.Debug("autoRecreateFile NewSecondsLifeSpanner failed", "err", err)
	}
	log.Debug("setting secondsLifeSpanner")
	fl.logLifeSpanner = secondsLifeSpanner

	ctx, cancelFunc := context.WithCancel(context.Background())
	go fl.autoRecreateFile(ctx)
	fl.cancelFunc = cancelFunc

	return fl, nil
}

func (fl *fileLogging) createFile(identifier string) (*os.File, error) {
	logDirectory := filepath.Join(fl.workingDir, fl.defaultLogsPath)

	if len(identifier) > 0 {
		identifier = fmt.Sprintf("-%s", identifier)
	}
	return core.CreateFile(
		core.ArgCreateFileArgument{
			Prefix:        fmt.Sprintf("%s%s", fl.logFilePrefix, identifier),
			Directory:     logDirectory,
			FileExtension: "log",
		},
	)
}

func (fl *fileLogging) recreateLogFile(identifier string) {
	newFile, err := fl.createFile(identifier)
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
	for {
		select {
		case <-ctx.Done():
			log.Info("closing fileLogging.autoRecreateFile go routine")
			return
		case identifier := <-fl.logLifeSpanner.GetChannel():
			log.Debug("autoRecreateFile - recreate log from lifespanner", "identifier", identifier)
			fl.recreateLogFile(identifier)
			sizeLogLifeSpanner, ok := fl.logLifeSpanner.(logger.SizeLogLifeSpanner)
			if ok {
				log.Debug("autoRecreateFile - found a sizeLogLifeSpanner", "new file", fl.currentFile.Name())
				newFile := fl.currentFile.Name()
				sizeLogLifeSpanner.SetCurrentFile(newFile)
			}
		}
	}
}

// ChangeFileLifeSpan changes the log file span
func (fl *fileLogging) ChangeFileLifeSpan(lifeSpanner logger.LogLifeSpanner) error {
	fl.mutIsClosed.Lock()
	defer fl.mutIsClosed.Unlock()

	if fl.isClosed {
		return core.ErrFileLoggingProcessIsClosed
	}

	if check.IfNil(lifeSpanner) {
		return fmt.Errorf("%w, error: nil lifespan", core.ErrInvalidLogFileMinLifeSpan)
	}

	fl.cancelFunc()

	fl.logLifeSpanner = lifeSpanner

	ctx, cancelFunc := context.WithCancel(context.Background())
	go fl.autoRecreateFile(ctx)
	fl.cancelFunc = cancelFunc

	sizeLogLifeSpanner, ok := lifeSpanner.(logger.SizeLogLifeSpanner)
	if ok {
		log.Info("Found a sizeLogLifeSpanner", "new file", fl.currentFile.Name())
		sizeLogLifeSpanner.SetCurrentFile(fl.currentFile.Name())
	}

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
