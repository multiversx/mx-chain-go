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
	importFolder  string
	files         map[string]*os.File
	dataReader    map[string]*bufio.Scanner
	importStorage storage.Storer
}

type ArgsNewMultiFileReader struct {
	ImportFolder  string
	ImportStorage storage.Storer
}

func NewMultiFileReader(args ArgsNewMultiFileReader) (*multiFileReader, error) {
	if len(args.ImportFolder) < 2 {
		return nil, update.ErrInvalidFolderName
	}
	if check.IfNil(args.ImportStorage) {
		return nil, update.ErrNilStorage
	}

	m := &multiFileReader{
		importFolder:  args.ImportFolder,
		files:         make(map[string]*os.File),
		dataReader:    make(map[string]*bufio.Scanner),
		importStorage: args.ImportStorage,
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
	value, err := m.importStorage.Get([]byte(key))
	if err != nil {
		return "", nil, err
	}

	return key, value, nil
}

func (m *multiFileReader) Finish() {
	for fileName, file := range m.files {
		err := file.Close()
		if err != nil {
			log.Warn("could not close file ", "fileName", fileName, "error", err)
		}
	}
}

func (m *multiFileReader) IsInterfaceNil() bool {
	return m == nil
}
