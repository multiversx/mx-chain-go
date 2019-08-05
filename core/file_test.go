package core_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	a, b int
}

func TestOpenFile_NoExistingFileShouldErr(t *testing.T) {
	t.Parallel()

	file, err := core.OpenFile("testFile1", logger.DefaultLogger())

	assert.Nil(t, file)
	assert.Error(t, err)
}

func TestOpenFile_NoErrShouldPass(t *testing.T) {
	t.Parallel()

	_, err := os.Create("testFile2")

	assert.Nil(t, err)

	file, err := core.OpenFile("testFile2", logger.DefaultLogger())

	if _, errF := os.Stat("testFile2"); errF == nil {
		_ = os.Remove("testFile2")
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

	_, err := os.Create("testFile3")

	assert.Nil(t, err)

	err = core.LoadTomlFile(cfg, "testFile3", logger.DefaultLogger())

	if _, errF := os.Stat("testFile3"); errF == nil {
		_ = os.Remove("testFile3")
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

	file, err := os.Create("testFile4")

	assert.Nil(t, err)

	data, _ := json.MarshalIndent(TestStruct{}, "", " ")

	_ = ioutil.WriteFile("testFile4", data, 0644)

	err = file.Close()

	assert.Nil(t, err)

	err = core.LoadJsonFile(cfg, "testFile4", logger.DefaultLogger())

	if _, errF := os.Stat("testFile4"); errF == nil {
		_ = os.Remove("testFile4")
	}

	assert.Nil(t, err)
}

func TestLoadSkFromPemFile_InvalidSkIndexShouldErr(t *testing.T) {
	t.Parallel()

	data, err := core.LoadSkFromPemFile("testFile5", logger.DefaultLogger(), -1)

	assert.Nil(t, data)
	assert.Equal(t, core.ErrInvalidIndex, err)
}

func TestLoadSkFromPemFile_NoExistingFileShouldErr(t *testing.T) {
	t.Parallel()

	data, err := core.LoadSkFromPemFile("testFile6", logger.DefaultLogger(), 0)

	assert.Nil(t, data)
	assert.Error(t, err)
}

func TestLoadSkFromPemFile_EmptyFileShouldErr(t *testing.T) {
	t.Parallel()

	_, err := os.Create("testFile7")

	data, err := core.LoadSkFromPemFile("testFile7", logger.DefaultLogger(), 0)

	if _, errF := os.Stat("testFile7"); errF == nil {
		_ = os.Remove("testFile7")
	}

	assert.Nil(t, data)
	assert.Equal(t, core.ErrEmptyFile, err)
}

func TestLoadSkFromPemFile_ShouldPass(t *testing.T) {
	t.Parallel()

	file, err := os.Create("testFile8")

	assert.Nil(t, err)

	skBytes := make([]byte, 0)
	skBytes = append(skBytes, 10, 20, 30, 40, 50, 60)

	_, _ = file.WriteString("-----BEGIN PRIVATE KEY for data-----\n")
	_, _ = file.WriteString("ChQeKDI8\n")
	_, _ = file.WriteString("-----END PRIVATE KEY for data-----")

	data, err := core.LoadSkFromPemFile("testFile8", logger.DefaultLogger(), 0)

	if _, errF := os.Stat("testFile8"); errF == nil {
		_ = os.Remove("testFile8")
	}

	//assert.Nil(t, data)
	assert.Equal(t, data, skBytes)
	assert.Nil(t, err)
}

func TestLoadSkFromPemFile_InvalidPemFileShouldErr(t *testing.T) {
	t.Parallel()

	file, err := os.Create("testFile9")

	assert.Nil(t, err)

	_, _ = file.WriteString("data")

	data, err := core.LoadSkFromPemFile("testFile9", logger.DefaultLogger(), 0)

	if _, errF := os.Stat("testFile9"); errF == nil {
		_ = os.Remove("testFile9")
	}

	assert.Nil(t, data)
	assert.Equal(t, core.ErrPemFileIsInvalid, err)
}

func TestLoadSkFromPemFile_InvalidIndexShouldErr(t *testing.T) {
	t.Parallel()

	file, err := os.Create("testFile10")

	assert.Nil(t, err)

	skBytes := make([]byte, 0)
	skBytes = append(skBytes, 10, 20, 30, 40, 50, 60)

	_, _ = file.WriteString("-----BEGIN PRIVATE KEY for data-----\n")
	_, _ = file.WriteString("ChQeKDI8\n")
	_, _ = file.WriteString("-----END PRIVATE KEY for data-----")

	data, err := core.LoadSkFromPemFile("testFile10", logger.DefaultLogger(), 1)

	if _, errF := os.Stat("testFile10"); errF == nil {
		_ = os.Remove("testFile10")
	}

	assert.Nil(t, data)
	assert.Equal(t, core.ErrInvalidIndex, err)
}

func TestSaveSkToPemFile_NilFileShouldErr(t *testing.T) {
	t.Parallel()

	skBytes := make([]byte, 0)
	skBytes = append(skBytes, 10, 20, 30)

	err := core.SaveSkToPemFile(nil, "data", skBytes)

	assert.Equal(t, core.ErrNilFile, err)
}

func TestSaveSkToPemFile_ShouldPass(t *testing.T) {
	t.Parallel()

	file, err := os.Create("testFile11")

	assert.Nil(t, err)

	skBytes := make([]byte, 0)
	skBytes = append(skBytes, 10, 20, 30, 40, 50, 60)

	err = core.SaveSkToPemFile(file, "data", skBytes)

	if _, errF := os.Stat("testFile11"); errF == nil {
		_ = os.Remove("testFile11")
	}

	assert.Nil(t, err)
}
