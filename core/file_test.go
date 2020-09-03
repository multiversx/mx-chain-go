package core_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"
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

func TestLoadSkPkFromPemFile_InvalidSkIndexShouldErr(t *testing.T) {
	t.Parallel()

	dataSk, dataPk, err := core.LoadSkPkFromPemFile("testFile5", -1)

	assert.Nil(t, dataSk)
	assert.Empty(t, "", dataPk)
	assert.Equal(t, core.ErrInvalidIndex, err)
}

func TestLoadSkPkFromPemFile_NoExistingFileShouldErr(t *testing.T) {
	t.Parallel()

	dataSk, dataPk, err := core.LoadSkPkFromPemFile("testFile6", 0)

	assert.Nil(t, dataSk)
	assert.Empty(t, dataPk)
	assert.Error(t, err)
}

func TestLoadSkPkFromPemFile_EmptyFileShouldErr(t *testing.T) {
	t.Parallel()

	fileName := "testFile7"
	_, _ = os.Create(fileName)

	dataSk, dataPk, err := core.LoadSkPkFromPemFile(fileName, 0)
	if _, errF := os.Stat(fileName); errF == nil {
		_ = os.Remove(fileName)
	}

	assert.Nil(t, dataSk)
	assert.Empty(t, dataPk)
	assert.True(t, errors.Is(err, core.ErrEmptyFile))
}

func TestLoadSkPkFromPemFile_ShouldPass(t *testing.T) {
	t.Parallel()

	fileName := "testFile8"
	file, err := os.Create(fileName)
	assert.Nil(t, err)

	skBytes := []byte{10, 20, 30, 40, 50, 60}
	pkString := "ABCD"

	_, _ = file.WriteString("-----BEGIN PRIVATE KEY for " + pkString + "-----\n")
	_, _ = file.WriteString("ChQeKDI8\n")
	_, _ = file.WriteString("-----END PRIVATE KEY for " + pkString + "-----")

	dataSk, dataPk, err := core.LoadSkPkFromPemFile(fileName, 0)
	if _, errF := os.Stat(fileName); errF == nil {
		_ = os.Remove(fileName)
	}

	assert.Equal(t, dataSk, skBytes)
	assert.Equal(t, dataPk, pkString)
	assert.Nil(t, err)
}

func TestLoadSkPkFromPemFile_IncorrectHeaderShoukldErr(t *testing.T) {
	t.Parallel()

	fileName := "testFile9"
	file, err := os.Create(fileName)
	assert.Nil(t, err)

	_, _ = file.WriteString("-----BEGIN INCORRECT HEADER ABCD-----\n")
	_, _ = file.WriteString("ChQeKDI8\n")
	_, _ = file.WriteString("-----END INCORRECT HEADER ABCD-----")

	dataSk, dataPk, err := core.LoadSkPkFromPemFile(fileName, 0)
	if _, errF := os.Stat(fileName); errF == nil {
		_ = os.Remove(fileName)
	}

	assert.Nil(t, dataSk)
	assert.Empty(t, dataPk)
	assert.True(t, errors.Is(err, core.ErrPemFileIsInvalid))
}

func TestLoadSkPkFromPemFile_InvalidPemFileShouldErr(t *testing.T) {
	t.Parallel()

	fileName := "testFile10"
	file, err := os.Create(fileName)
	assert.Nil(t, err)

	_, _ = file.WriteString("data")

	dataSk, dataPk, err := core.LoadSkPkFromPemFile(fileName, 0)
	if _, errF := os.Stat(fileName); errF == nil {
		_ = os.Remove(fileName)
	}

	assert.Nil(t, dataSk)
	assert.Empty(t, dataPk)
	assert.True(t, errors.Is(err, core.ErrPemFileIsInvalid))
}

func TestLoadSkPkFromPemFile_InvalidIndexShouldErr(t *testing.T) {
	t.Parallel()

	fileName := "testFile11"
	file, err := os.Create(fileName)
	assert.Nil(t, err)

	_, _ = file.WriteString("-----BEGIN PRIVATE KEY for data-----\n")
	_, _ = file.WriteString("ChQeKDI8\n")
	_, _ = file.WriteString("-----END PRIVATE KEY for data-----")

	dataSk, dataPk, err := core.LoadSkPkFromPemFile(fileName, 1)
	if _, errF := os.Stat(fileName); errF == nil {
		_ = os.Remove(fileName)
	}

	assert.Nil(t, dataSk)
	assert.Empty(t, dataPk)
	assert.True(t, errors.Is(err, core.ErrInvalidIndex))
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

	fileName := "testFile12"
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

func TestCreateFile(t *testing.T) {
	t.Parallel()

	arg := core.ArgCreateFileArgument{
		Directory:     "subdir",
		Prefix:        "prefix",
		FileExtension: "extension",
	}

	file, err := core.CreateFile(arg)
	assert.Nil(t, err)
	assert.NotNil(t, file)

	assert.True(t, strings.Contains(file.Name(), arg.Prefix))
	assert.True(t, strings.Contains(file.Name(), arg.FileExtension))
	if _, errF := os.Stat(file.Name()); errF == nil {
		_ = os.Remove(file.Name())
		_ = os.Remove(arg.Directory)
	}
}
