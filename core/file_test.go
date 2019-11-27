package core_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	a, b int
}

func TestOpenFile_NoExistingFileShouldErr(t *testing.T) {
	t.Parallel()

	file, err := core.OpenFile("testFile1")

	assert.Nil(t, file)
	assert.Error(t, err)
}

func TestOpenFile_NoErrShouldPass(t *testing.T) {
	t.Parallel()

	fileName := "testFile2"
	_, err := os.Create(fileName)
	assert.Nil(t, err)

	file, err := core.OpenFile(fileName)
	if _, errF := os.Stat(fileName); errF == nil {
		_ = os.Remove(fileName)
	}

	assert.NotNil(t, file)
	assert.Nil(t, err)
}

func TestLoadTomlFile_NoExistingFileShouldErr(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}

	err := core.LoadTomlFile(cfg, "file")

	assert.Error(t, err)
}

func TestLoadTomlFile_FileExitsShouldPass(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}

	fileName := "testFile3"
	_, err := os.Create(fileName)
	assert.Nil(t, err)

	err = core.LoadTomlFile(cfg, fileName)
	if _, errF := os.Stat(fileName); errF == nil {
		_ = os.Remove(fileName)
	}

	assert.Nil(t, err)
}

func TestLoadJSonFile_NoExistingFileShouldErr(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}

	err := core.LoadJsonFile(cfg, "file")

	assert.Error(t, err)

}

func TestLoadJSonFile_FileExitsShouldPass(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}

	fileName := "testFile4"
	file, err := os.Create(fileName)
	assert.Nil(t, err)

	data, _ := json.MarshalIndent(TestStruct{a: 0, b: 0}, "", " ")

	_ = ioutil.WriteFile(fileName, data, 0644)

	err = file.Close()
	assert.Nil(t, err)

	err = core.LoadJsonFile(cfg, fileName)
	if _, errF := os.Stat(fileName); errF == nil {
		_ = os.Remove(fileName)
	}

	assert.Nil(t, err)
}

func TestLoadSkFromPemFile_InvalidSkIndexShouldErr(t *testing.T) {
	t.Parallel()

	data, err := core.LoadSkFromPemFile("testFile5", -1)

	assert.Nil(t, data)
	assert.Equal(t, core.ErrInvalidIndex, err)
}

func TestLoadSkFromPemFile_NoExistingFileShouldErr(t *testing.T) {
	t.Parallel()

	data, err := core.LoadSkFromPemFile("testFile6", 0)

	assert.Nil(t, data)
	assert.Error(t, err)
}

func TestLoadSkFromPemFile_EmptyFileShouldErr(t *testing.T) {
	t.Parallel()

	fileName := "testFile7"
	_, err := os.Create(fileName)

	data, err := core.LoadSkFromPemFile(fileName, 0)
	if _, errF := os.Stat(fileName); errF == nil {
		_ = os.Remove(fileName)
	}

	assert.Nil(t, data)
	assert.Equal(t, core.ErrEmptyFile, err)
}

func TestLoadSkFromPemFile_ShouldPass(t *testing.T) {
	t.Parallel()

	fileName := "testFile8"
	file, err := os.Create(fileName)
	assert.Nil(t, err)

	skBytes := make([]byte, 0)
	skBytes = append(skBytes, 10, 20, 30, 40, 50, 60)

	_, _ = file.WriteString("-----BEGIN PRIVATE KEY for data-----\n")
	_, _ = file.WriteString("ChQeKDI8\n")
	_, _ = file.WriteString("-----END PRIVATE KEY for data-----")

	data, err := core.LoadSkFromPemFile(fileName, 0)
	if _, errF := os.Stat(fileName); errF == nil {
		_ = os.Remove(fileName)
	}

	assert.Equal(t, data, skBytes)
	assert.Nil(t, err)
}

func TestLoadSkFromPemFile_InvalidPemFileShouldErr(t *testing.T) {
	t.Parallel()

	fileName := "testFile9"
	file, err := os.Create(fileName)
	assert.Nil(t, err)

	_, _ = file.WriteString("data")

	data, err := core.LoadSkFromPemFile(fileName, 0)
	if _, errF := os.Stat(fileName); errF == nil {
		_ = os.Remove(fileName)
	}

	assert.Nil(t, data)
	assert.Equal(t, core.ErrPemFileIsInvalid, err)
}

func TestLoadSkFromPemFile_InvalidIndexShouldErr(t *testing.T) {
	t.Parallel()

	fileName := "testFile10"
	file, err := os.Create(fileName)
	assert.Nil(t, err)

	skBytes := make([]byte, 0)
	skBytes = append(skBytes, 10, 20, 30, 40, 50, 60)

	_, _ = file.WriteString("-----BEGIN PRIVATE KEY for data-----\n")
	_, _ = file.WriteString("ChQeKDI8\n")
	_, _ = file.WriteString("-----END PRIVATE KEY for data-----")

	data, err := core.LoadSkFromPemFile(fileName, 1)
	if _, errF := os.Stat(fileName); errF == nil {
		_ = os.Remove(fileName)
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

	fileName := "testFile11"
	file, err := os.Create(fileName)
	assert.Nil(t, err)

	skBytes := make([]byte, 0)
	skBytes = append(skBytes, 10, 20, 30, 40, 50, 60)

	err = core.SaveSkToPemFile(file, "data", skBytes)
	if _, errF := os.Stat(fileName); errF == nil {
		_ = os.Remove(fileName)
	}

	assert.Nil(t, err)
}
