package core_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/stretchr/testify/assert"
)

var errInvalidIndex = errors.New("invalid private key index")
var errPemFileIsInvalid = errors.New("pem file is invalid")
var errEmptyFile = errors.New("empty file provided")
var errNilFile = errors.New("nil file provided")

type TestStruct struct {
	a, b int
}

func TestOpenFile_NoExistingFileShouldErr(t *testing.T) {
	t.Parallel()

	file, err := core.OpenFile("testFile", logger.DefaultLogger())

	assert.Nil(t, file)
	assert.Error(t, err)
}

func TestOpenFile_NoErrShouldPass(t *testing.T) {
	t.Parallel()

	_, err := os.Create("testFile")

	assert.Nil(t, err)

	file, err := core.OpenFile("testFile", logger.DefaultLogger())

	if _, errF := os.Stat("testFile"); errF == nil {
		_ = os.Remove("testFile")
	}

	assert.NotNil(t, file)
	assert.Nil(t, err)
}

func TestLoadTomlFile_NoExistingFileShouldErr(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}

	err := core.LoadTomlFile(cfg, "file", logger.DefaultLogger())

	assert.Error(t, err)
}

func TestLoadTomlFile_FileExitsShouldPass(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}

	_, err := os.Create("testFile")

	assert.Nil(t, err)

	err = core.LoadTomlFile(cfg, "testFile", logger.DefaultLogger())

	if _, errF := os.Stat("testFile"); errF == nil {
		_ = os.Remove("testFile")
	}

	assert.Nil(t, err)
}

func TestLoadJSonFile_NoExistingFileShouldErr(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}

	err := core.LoadJsonFile(cfg, "file", logger.DefaultLogger())

	assert.Error(t, err)

}

func TestLoadJSonFile_FileExitsShouldPass(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}

	file, err := os.Create("testFile")

	assert.Nil(t, err)

	data, _ := json.MarshalIndent(TestStruct{}, "", " ")

	_ = ioutil.WriteFile("testFile", data, 0644)

	err = file.Close()

	assert.Nil(t, err)

	err = core.LoadJsonFile(cfg, "testFile", logger.DefaultLogger())

	if _, errF := os.Stat("testFile"); errF == nil {
		_ = os.Remove("testFile")
	}

	assert.Nil(t, err)
}

func TestLoadSkFromPemFile_InvalidSkIndexShouldErr(t *testing.T) {
	t.Parallel()

	data, err := core.LoadSkFromPemFile("testFile", logger.DefaultLogger(), -1)

	assert.Nil(t, data)
	assert.Equal(t, errInvalidIndex, err)
}

func TestLoadSkFromPemFile_NoExistingFileShouldErr(t *testing.T) {
	t.Parallel()

	data, err := core.LoadSkFromPemFile("testFile", logger.DefaultLogger(), 0)

	assert.Nil(t, data)
	assert.Error(t, err)
}

func TestLoadSkFromPemFile_EmptyFileShouldErr(t *testing.T) {
	t.Parallel()

	_, err := os.Create("testFile")

	data, err := core.LoadSkFromPemFile("testFile", logger.DefaultLogger(), 0)

	if _, errF := os.Stat("testFile"); errF == nil {
		_ = os.Remove("testFile")
	}

	assert.Nil(t, data)
	assert.Equal(t, errEmptyFile, err)
}

func TestLoadSkFromPemFile_ShouldPass(t *testing.T) {
	t.Parallel()

	file, err := os.Create("testFile")

	assert.Nil(t, err)

	skBytes := make([]byte, 0)
	skBytes = append(skBytes, 10, 20, 30, 40, 50, 60)

	_, _ = file.WriteString("-----BEGIN PRIVATE KEY for data-----\n")
	_, _ = file.WriteString("ChQeKDI8\n")
	_, _ = file.WriteString("-----END PRIVATE KEY for data-----")

	data, err := core.LoadSkFromPemFile("testFile", logger.DefaultLogger(), 0)

	if _, errF := os.Stat("testFile"); errF == nil {
		_ = os.Remove("testFile")
	}

	//assert.Nil(t, data)
	assert.Equal(t, data, skBytes)
	assert.Nil(t, err)
}

func TestLoadSkFromPemFile_InvalidPemFileShouldErr(t *testing.T) {
	t.Parallel()

	file, err := os.Create("testFile")

	assert.Nil(t, err)

	_, _ = file.WriteString("data")

	data, err := core.LoadSkFromPemFile("testFile", logger.DefaultLogger(), 0)

	if _, errF := os.Stat("testFile"); errF == nil {
		_ = os.Remove("testFile")
	}

	assert.Nil(t, data)
	assert.Equal(t, errPemFileIsInvalid, err)
}

func TestLoadSkFromPemFile_InvalidIndexShouldErr(t *testing.T) {
	t.Parallel()

	file, err := os.Create("testFile")

	assert.Nil(t, err)

	skBytes := make([]byte, 0)
	skBytes = append(skBytes, 10, 20, 30, 40, 50, 60)

	_, _ = file.WriteString("-----BEGIN PRIVATE KEY for data-----\n")
	_, _ = file.WriteString("ChQeKDI8\n")
	_, _ = file.WriteString("-----END PRIVATE KEY for data-----")

	data, err := core.LoadSkFromPemFile("testFile", logger.DefaultLogger(), 1)

	if _, errF := os.Stat("testFile"); errF == nil {
		_ = os.Remove("testFile")
	}

	assert.Nil(t, data)
	assert.Equal(t, errInvalidIndex, err)
}

func TestSaveSkToPemFile_NilFileShouldErr(t *testing.T) {
	t.Parallel()

	skBytes := make([]byte, 0)
	skBytes = append(skBytes, 10, 20, 30)

	err := core.SaveSkToPemFile(nil, "data", skBytes)

	assert.Equal(t, errNilFile, err)
}

func TestSaveSkToPemFile_ShouldPass(t *testing.T) {
	t.Parallel()

	file, err := os.Create("testFile")

	assert.Nil(t, err)

	skBytes := make([]byte, 0)
	skBytes = append(skBytes, 10, 20, 30, 40, 50, 60)

	err = core.SaveSkToPemFile(file, "data", skBytes)

	if _, errF := os.Stat("testFile"); errF == nil {
		_ = os.Remove("testFile")
	}

	assert.Nil(t, err)
}
