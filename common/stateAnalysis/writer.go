package stateAnalysis

import (
	"compress/gzip"
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

// csvFile is a gzip-compressed csv file with a header, flushed after every row
type csvFile struct {
	mut  sync.Mutex
	f    *os.File
	gz   *gzip.Writer
	csv  *csv.Writer
	path string
}

func newCsvFile(path string, header []string) (*csvFile, error) {
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return nil, err
	}
	// append so that a component restart (e.g. shuffle out) does not truncate the rows of the previous cycle;
	// each cycle appends a new gzip member, which zcat and the python gzip module read as one stream
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	gz := gzip.NewWriter(f)
	w := csv.NewWriter(gz)
	c := &csvFile{f: f, gz: gz, csv: w, path: path}
	if info.Size() == 0 {
		err = c.write(header)
		if err != nil {
			_ = c.close()
			return nil, err
		}
	}
	return c, nil
}

func (c *csvFile) write(row []string) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	err := c.csv.Write(row)
	if err != nil {
		return err
	}
	c.csv.Flush()
	return c.gz.Flush()
}

func (c *csvFile) close() error {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.csv.Flush()
	errGz := c.gz.Close()
	errF := c.f.Close()
	if errGz != nil {
		return errGz
	}
	return errF
}

func i64s(v int64) string {
	return strconv.FormatInt(v, 10)
}

func u64s(v uint64) string {
	return strconv.FormatUint(v, 10)
}
