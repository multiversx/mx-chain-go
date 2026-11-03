package snapshotAnalysis

import (
	"compress/gzip"
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
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
	// append mode: a component restart inside one process (e.g. shuffle out) appends a new gzip member,
	// which zcat and the python gzip module read as a single stream
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
	c := &csvFile{f: f, gz: gz, csv: csv.NewWriter(gz), path: path}
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

// row is a helper that builds a csv row with typed appends
type row struct {
	cols []string
}

func newRow(capacity int) *row {
	return &row{cols: make([]string, 0, capacity)}
}

func (r *row) str(v string) *row {
	r.cols = append(r.cols, v)
	return r
}

func (r *row) i64(v int64) *row {
	r.cols = append(r.cols, strconv.FormatInt(v, 10))
	return r
}

func (r *row) u64(v uint64) *row {
	r.cols = append(r.cols, strconv.FormatUint(v, 10))
	return r
}

func (r *row) f(v float64) *row {
	r.cols = append(r.cols, strconv.FormatFloat(v, 'f', 3, 64))
	return r
}

func (r *row) dur(d time.Duration) *row {
	return r.i64(int64(d))
}

func (r *row) b(v bool) *row {
	if v {
		return r.str("1")
	}
	return r.str("0")
}
