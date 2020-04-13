package files

import (
	"bufio"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"sort"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
)

type multiFileReader struct {
	importFolder string
	files        map[string]io.Closer
	dataReader   map[string]update.DataReader
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
		importStore:  args.ImportStore,
	}

	err := m.readAllFiles()
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *multiFileReader) readAllFiles() error {
	files, err := ioutil.ReadDir(m.importFolder)
	if err != nil {
		return err
	}

	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			continue
		}

		name := fileInfo.Name()
		file, err := os.OpenFile(m.importFolder+"/"+name, os.O_RDONLY, 0644)
		if err != nil {
			log.Warn("unable to open file", "fileName", name, "error", err)
			continue
		}

		scanner := bufio.NewScanner(file)
		m.files[name] = file
		m.dataReader[name] = scanner
	}

	return nil
}

// GetFileNames returns the list of opened files
func (m *multiFileReader) GetFileNames() []string {
	fileNames := make([]string, 0)
	for fileName := range m.files {
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

	value, err := m.importStore.Get(key)
	if err != nil {
		return "", nil, err
	}

	return string(key), value, nil
}

func (m *multiFileReader) getDataReader(fileName string) (update.DataReader, error) {
	scanner, ok := m.dataReader[fileName]
	if !ok {
		return nil, update.ErrNilDataReader
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

// Finish closes all the opened files
func (m *multiFileReader) Finish() {
	for fileName, file := range m.files {
		err := file.Close()
		if err != nil {
			log.Warn("could not close file ", "fileName", fileName, "error", err)
		}
	}
}

// IsInterfaceNil returns true if underlying object is nil
func (m *multiFileReader) IsInterfaceNil() bool {
	return m == nil
}
