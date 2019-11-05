package logger_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/stretchr/testify/assert"
)

func TestSetLogLevel_WrongStringParameterShouldErr(t *testing.T) {
	err := logger.SetLogLevel("wrong string")

	assert.Equal(t, logger.ErrInvalidLogLevelPattern, err)
}

func TestSetLogLevel_WrongLogLevelShouldErr(t *testing.T) {
	err := logger.SetLogLevel("WRONG_LEVEL|*")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unknown log level")
}

func TestSetLogLevel_NonWildcardPatternShouldNotError(t *testing.T) {
	log1 := logger.Get("pattern1")
	log2 := logger.Get("pattern2")

	err := logger.SetLogLevel("DEBuG|pattern1")

	assert.Nil(t, err)
	assert.Equal(t, logger.LogDebug, log1.LogLevel())
	assert.Equal(t, logger.LogInfo, log2.LogLevel())

	err = logger.SetLogLevel("tRace|pattern")

	assert.Nil(t, err)
	assert.Equal(t, logger.LogTrace, log1.LogLevel())
	assert.Equal(t, logger.LogTrace, log2.LogLevel())

	//rollback to the default value
	_ = logger.SetLogLevel("INFO|*")
}

func TestSetLogLevel_WildcardPatternShouldWork(t *testing.T) {
	log1 := logger.Get("1")
	log2 := logger.Get("2")

	err := logger.SetLogLevel("DEBuG|*")

	assert.Nil(t, err)
	assert.Equal(t, logger.LogDebug, log1.LogLevel())
	assert.Equal(t, logger.LogDebug, log2.LogLevel())
	assert.Equal(t, logger.LogDebug, *logger.DefaultLogLevel)

	//rollback to the default value
	_ = logger.SetLogLevel("INFO|*")
}
