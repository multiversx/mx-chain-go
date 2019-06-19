package core

import (
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/pelletier/go-toml"
)

var errPemFileIsInvalid = errors.New("pem file is invalid")
var errNilFile = errors.New("nil file provided")
var errEmptyFile = errors.New("empty file provided")
var errInvalidIndex = errors.New("invalid private key index")

// OpenFile method opens the file from given path - does not close the file
func OpenFile(relativePath string, log *logger.Logger) (*os.File, error) {
	path, err := filepath.Abs(relativePath)
	fmt.Println(path)
	if err != nil {
		log.Error("cannot create absolute path for the provided file", err.Error())
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// LoadTomlFile method to open and decode toml file
func LoadTomlFile(dest interface{}, relativePath string, log *logger.Logger) error {
	f, err := OpenFile(relativePath, log)
	if err != nil {
		return err
	}

	defer func() {
		err = f.Close()
		if err != nil {
			log.Error("cannot close file: ", err.Error())
		}
	}()

	return toml.NewDecoder(f).Decode(dest)
}

// LoadJsonFile method to open and decode json file
func LoadJsonFile(dest interface{}, relativePath string, log *logger.Logger) error {
	f, err := OpenFile(relativePath, log)
	if err != nil {
		return err
	}

	defer func() {
		err = f.Close()
		if err != nil {
			log.Error("cannot close file: ", err.Error())
		}
	}()

	return json.NewDecoder(f).Decode(dest)
}

// CreateFile opens or creates a file relative to the default path
func CreateFile(prefix string, subfolder string, fileExtension string) (*os.File, error) {
	absPath, err := filepath.Abs(subfolder)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(absPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	fileName := time.Now().Format("2006-02-01-15-04-05")
	if prefix != "" {
		fileName = prefix + "-" + fileName
	}

	return os.OpenFile(
		filepath.Join(absPath, fileName+"."+fileExtension),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0666)
}

// LoadSkFromPemFile loads the secret key bytes stored in the file
func LoadSkFromPemFile(relativePath string, log *logger.Logger, skIndex int) ([]byte, error) {
	if skIndex < 0 {
		return nil, errInvalidIndex
	}

	file, err := OpenFile(relativePath, log)
	if err != nil {
		return nil, err
	}

	defer func() {
		cerr := file.Close()
		log.LogIfError(cerr)
	}()

	buff, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if len(buff) == 0 {
		return nil, errEmptyFile
	}

	var blkRecovered *pem.Block

	for i := 0; i <= skIndex; i++ {
		if len(buff) == 0 {
			//less private keys present in the file than required
			return nil, errInvalidIndex
		}

		blkRecovered, buff = pem.Decode(buff)
		if blkRecovered == nil {
			return nil, errPemFileIsInvalid
		}
	}

	return blkRecovered.Bytes, nil
}

// SaveSkToPemFile saves secret key bytes in the file
func SaveSkToPemFile(file *os.File, identifier string, skBytes []byte) error {
	if file == nil {
		return errNilFile
	}

	blk := pem.Block{
		Type:  "PRIVATE KEY for " + identifier,
		Bytes: skBytes,
	}

	return pem.Encode(file, &blk)
}
