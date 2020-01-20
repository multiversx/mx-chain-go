package files

import (
	"bufio"
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
	dataReader   map[string]*bufio.Scanner
	importStore  storage.Storer
}

// ArgsNewMultiFileReader defines the arguments needed to create a multi file reader
type ArgsNewMultiFileReader struct {
	ImportFolder string
	ImportStore  storage.Storer
}

// NewMultiFileReaders creates a multi file reader and opens all the files from a give folder
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
		dataReader:   make(map[string]*bufio.Scanner),
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
		file, err := os.OpenFile(name, os.O_RDONLY, 0644)
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
	scanner, ok := m.dataReader[fileName]
	if !ok {
		return "", nil, update.ErrNilDataReader
	}

	canRead := scanner.Scan()
	if !canRead {
		err := scanner.Err()
		if err == io.EOF {
			return "", nil, update.ErrEndOfFile
		}
		return "", nil, err
	}

	key := scanner.Text()
	value, err := m.importStore.Get([]byte(key))
	if err != nil {
		return "", nil, err
	}

	return key, value, nil
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
