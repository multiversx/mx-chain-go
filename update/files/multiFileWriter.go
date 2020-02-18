package files

import (
	"bufio"
	"encoding/hex"
	"io"
	"os"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
)

var log = logger.GetOrCreate("update/files")

type multiFileWriter struct {
	exportFolder string

	files       map[string]io.Closer
	dataWriters map[string]update.DataWriter
	exportStore storage.Storer
}

// ArgsNewMultiFileWriter defines the arguments needed by the multi file writer
type ArgsNewMultiFileWriter struct {
	ExportFolder string
	ExportStore  storage.Storer
}

// NewMultiFileWriter creates a new multi file writer, returns error if arguments are nil
func NewMultiFileWriter(args ArgsNewMultiFileWriter) (*multiFileWriter, error) {
	if check.IfNil(args.ExportStore) {
		return nil, update.ErrNilStorage
	}
	if len(args.ExportFolder) < 2 {
		return nil, update.ErrInvalidFolderName
	}

	m := &multiFileWriter{
		exportFolder: args.ExportFolder,
		files:        make(map[string]io.Closer),
		dataWriters:  make(map[string]update.DataWriter),
		exportStore:  args.ExportStore,
	}

	return m, nil
}

// NewFile creates a file with a given name
func (m *multiFileWriter) NewFile(fileName string) error {
	if _, ok := m.files[fileName]; ok {
		return nil
	}

	file, err := os.OpenFile(m.exportFolder+"/"+fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Debug("unable to open file")
		return err
	}

	dataWriter := bufio.NewWriter(file)
	m.files[fileName] = file
	m.dataWriters[fileName] = dataWriter

	return nil
}

// Write appends the next key, value pair to the selected file, creates a file if that does not exist
func (m *multiFileWriter) Write(fileName string, key string, value []byte) error {
	dataWriter, err := m.getOrCreateDataWriter(fileName)
	if err != nil {
		return err
	}

	hexEncoded := hex.EncodeToString([]byte(key))
	_, err = dataWriter.WriteString(hexEncoded + "\n")
	if err != nil {
		return err
	}

	err = m.exportStore.Put([]byte(key), value)
	if err != nil {
		return err
	}

	return nil
}

func (m *multiFileWriter) getOrCreateDataWriter(fileName string) (update.DataWriter, error) {
	var dataWriter update.DataWriter
	var ok bool

	if dataWriter, ok = m.dataWriters[fileName]; ok {
		return dataWriter, nil
	}

	if err := m.NewFile(fileName); err != nil {
		return nil, err
	}

	if dataWriter, ok = m.dataWriters[fileName]; ok {
		return dataWriter, nil

	}

	return nil, update.ErrNilDataWriter
}

// Finish flushes all the data to the files and closes the opened files
func (m *multiFileWriter) Finish() {
	for fileName, dataWriter := range m.dataWriters {
		err := dataWriter.Flush()
		if err != nil {
			log.Warn("could not flush data for ", "fileName", fileName, "error", err)
		}
	}

	for fileName, file := range m.files {
		err := file.Close()
		if err != nil {
			log.Warn("could not close file ", "fileName", fileName, "error", err)
		}
	}
}

// IsInterfaceNil returns true if underlying object is nil
func (m *multiFileWriter) IsInterfaceNil() bool {
	return m == nil
}
