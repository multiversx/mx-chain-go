package files

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
)

var _ update.MultiFileReader = (*multiFileReader)(nil)

type multiFileReader struct {
	importFolder string
	files        map[string]io.Closer
	dataReader   map[string]update.DataReader
	filenames    map[string]string
	importStore  storage.Storer
}

// ArgsNewMultiFileReader defines the arguments needed to create a multi file reader
type ArgsNewMultiFileReader struct {
	ImportFolder string
	ImportStore  storage.Storer
}

// NewMultiFileReader creates a multi file reader and opens all the files from a give folder
func NewMultiFileReader(args ArgsNewMultiFileReader) (*multiFileReader, error) {
	if len(args.ImportFolder) < 2 {
		return nil, update.ErrInvalidFolderName
	}
	if check.IfNil(args.ImportStore) {
		return nil, update.ErrNilStorage
	}

	m := &multiFileReader{
		importFolder: args.ImportFolder,
		files:        make(map[string]io.Closer),
		dataReader:   make(map[string]update.DataReader),
		filenames:    make(map[string]string),
		importStore:  args.ImportStore,
	}

	err := m.readAllFileNames()
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *multiFileReader) readAllFileNames() error {
	files, err := ioutil.ReadDir(m.importFolder)
	if err != nil {
		return err
	}

	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			continue
		}

		name := fileInfo.Name()
		m.filenames[name] = m.importFolder + "/" + name
	}

	return nil
}

// GetFileNames returns the list of opened files
func (m *multiFileReader) GetFileNames() []string {
	fileNames := make([]string, 0)
	for fileName := range m.filenames {
		fileNames = append(fileNames, fileName)
	}

	sort.Slice(fileNames, func(i, j int) bool {
		return fileNames[i] < fileNames[j]
	})

	return fileNames
}

// ReadNextItem returns the next elements from a file if that exist
func (m *multiFileReader) ReadNextItem(fileName string) (string, []byte, error) {
	scanner, err := m.getDataReader(fileName)
	if err != nil {
		return "", nil, err
	}

	hexEncoded := scanner.Text()

	key, err := hex.DecodeString(hexEncoded)
	if err != nil {
		return "", nil, err
	}

	formattedKey := []byte(string(key) + fileName)
	value, err := m.importStore.Get(formattedKey)
	if err != nil {
		return "", nil, err
	}
	log.Trace("import", "key", string(key), "value", value)

	return string(key), value, nil
}

func (m *multiFileReader) getDataReader(fileName string) (update.DataReader, error) {
	scanner, ok := m.dataReader[fileName]
	if !ok {
		var importFileName string
		importFileName, ok = m.filenames[fileName]
		if !ok {
			return nil, fmt.Errorf("%w for file: %s", update.ErrMissingFile, fileName)
		}

		log.Debug("file opened", "filename", fileName)
		file, err := os.OpenFile(importFileName, os.O_RDONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("%w for file: %s", err, importFileName)
		}

		scanner = bufio.NewScanner(file)
		m.files[fileName] = file
		m.dataReader[fileName] = scanner
	}

	canRead := scanner.Scan()
	if canRead {
		return scanner, nil
	}

	err := scanner.Err()
	if err != nil {
		return nil, err
	}

	// is end of file because cannot read from file and err is nil (scanner.Err() return nil if it end of file
	return nil, update.ErrEndOfFile
}

// CloseFile will close the file and dataWriter
func (m *multiFileReader) CloseFile(fileName string) {
	file, ok := m.files[fileName]
	if ok {
		err := file.Close()
		log.LogIfError(err, "closeFile multiFileWriter file close", fileName)
	}

	m.fileClosed(fileName)
}

// Finish closes all the opened files
func (m *multiFileReader) Finish() {
	for fileName, file := range m.files {
		err := file.Close()
		if err != nil {
			log.Trace("could not close file ", "fileName", fileName, "error", err)
		}

		m.fileClosed(fileName)
	}

	err := m.importStore.Close()
	if err != nil {
		log.Trace("could not close import store ", "error", err)
	}
}

func (m *multiFileReader) fileClosed(fileName string) {
	log.Debug("file closed", "filename", fileName)
	delete(m.files, fileName)
	delete(m.dataReader, fileName)
}

// IsInterfaceNil returns true if underlying object is nil
func (m *multiFileReader) IsInterfaceNil() bool {
	return m == nil
}
