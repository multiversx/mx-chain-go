package main

import (
	"fmt"
	"os"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/cmd/node/factory"
	"github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-logger-go/file"
)

type nodeLogging struct {
	fileLogging     factory.FileLoggingHandler
	fileLoggingJson factory.FileLoggingHandler
}

func (logging *nodeLogging) setup(log logger.Logger, flagsConfig *config.ContextFlagsConfig) error {
	err := logging.handleSaveLogFile(flagsConfig)
	if err != nil {
		return err
	}

	err = logging.handleSaveLogFileJson(flagsConfig)
	if err != nil {
		return err
	}

	err = logger.SetDisplayByteSlice(logger.ToHex)
	if err != nil {
		return err
	}

	err = logger.SetLogLevel(flagsConfig.LogLevel)
	if err != nil {
		return err
	}

	err = logging.handleDisableAnsiColor(flagsConfig)
	if err != nil {
		return err
	}

	logger.ToggleCorrelation(flagsConfig.EnableLogCorrelation)
	logger.ToggleLoggerName(flagsConfig.EnableLogName)

	log.Debug("logging set up",
		"level", flagsConfig.LogLevel,
		"save log file", flagsConfig.SaveLogFile,
		"save log file (JSON)", flagsConfig.SaveLogFileJson,
		"disable ANSI color", flagsConfig.DisableAnsiColor,
	)

	return nil
}

func (logging *nodeLogging) handleSaveLogFile(flagsConfig *config.ContextFlagsConfig) error {
	if !flagsConfig.SaveLogFile {
		return nil
	}

	args := file.ArgsFileLogging{
		WorkingDir:      flagsConfig.LogsDir,
		DefaultLogsPath: defaultLogsPath,
		LogFilePrefix:   logFilePrefix,
	}

	fileLogging, err := file.NewFileLogging(args)
	if err != nil {
		return fmt.Errorf("%w creating log file", err)
	}

	logging.fileLogging = fileLogging
	return nil
}

func (logging *nodeLogging) handleSaveLogFileJson(flagsConfig *config.ContextFlagsConfig) error {
	if !flagsConfig.SaveLogFileJson {
		return nil
	}

	args := file.ArgsFileLogging{
		WorkingDir:      flagsConfig.LogsDir,
		DefaultLogsPath: defaultLogsPath,
		LogFilePrefix:   logFilePrefix,
		LogAsJson:       true,
	}

	fileLoggingJson, err := file.NewFileLogging(args)
	if err != nil {
		return fmt.Errorf("%w creating log file (JSON)", err)
	}

	logging.fileLoggingJson = fileLoggingJson
	return nil
}

func (logging *nodeLogging) handleDisableAnsiColor(flagsConfig *config.ContextFlagsConfig) error {
	if !flagsConfig.DisableAnsiColor {
		return nil
	}

	err := logger.RemoveLogObserver(os.Stdout)
	if err != nil {
		return err
	}

	err = logger.AddLogObserver(os.Stdout, &logger.PlainFormatter{})
	if err != nil {
		return err
	}

	return nil
}

func (logging *nodeLogging) applyFileLifeSpan(log logger.Logger, config *config.Config) error {
	durationLogLifeSpan := time.Second * time.Duration(config.Logs.LogFileLifeSpanInSec)
	sizeLogLifeSpanInMB := uint64(config.Logs.LogFileLifeSpanInMB)

	if !check.IfNil(logging.fileLogging) {
		err := logging.fileLogging.ChangeFileLifeSpan(durationLogLifeSpan, sizeLogLifeSpanInMB)
		if err != nil {
			return err
		}

		log.Debug("log file life span changed", "duration", durationLogLifeSpan, "size", sizeLogLifeSpanInMB)
	}

	if !check.IfNil(logging.fileLoggingJson) {
		err := logging.fileLoggingJson.ChangeFileLifeSpan(durationLogLifeSpan, sizeLogLifeSpanInMB)
		if err != nil {
			return err
		}

		log.Debug("log file life span changed (JSON)", "duration", durationLogLifeSpan, "size", sizeLogLifeSpanInMB)
	}

	return nil
}

func (logging *nodeLogging) closeFiles(log logger.Logger) {
	if !check.IfNil(logging.fileLogging) {
		err := logging.fileLogging.Close()
		log.LogIfError(err)
	}

	if !check.IfNil(logging.fileLoggingJson) {
		err := logging.fileLoggingJson.Close()
		log.LogIfError(err)
	}
}
