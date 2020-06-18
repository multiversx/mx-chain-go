package health

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
)

var _ record = (*memoryRecord)(nil)

type memoryRecord struct {
	stats    runtime.MemStats
	filename string
}

func newMemoryRecord(stats runtime.MemStats, parentFolder string) *memoryRecord {
	filename := formatMemoryRecordFilename(stats)
	filename = path.Join(parentFolder, filename)

	return &memoryRecord{
		stats:    stats,
		filename: filename,
	}
}

func formatMemoryRecordFilename(stats runtime.MemStats) string {
	timestamp := time.Now().Format("20060102150405")
	inUse := core.ConvertBytes(stats.HeapInuse)
	inUse = strings.ReplaceAll(inUse, " ", "_")
	inUse = strings.ReplaceAll(inUse, ".", "_")
	filename := fmt.Sprintf("mem__%s__%s.pprof", timestamp, inUse)
	return filename
}

func (record *memoryRecord) getFilename() string {
	return record.filename
}

func (record *memoryRecord) save() error {
	filename := path.Join(record.filename)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	log.Debug("memoryRecord.save()", "file", filename)

	err = pprof.WriteHeapProfile(file)
	if err != nil {
		return err
	}

	return file.Close()
}

func (record *memoryRecord) isMoreImportantThan(otherRecord record) bool {
	asMemoryRecord, ok := otherRecord.(*memoryRecord)
	if !ok {
		return false
	}

	return record.stats.HeapInuse > asMemoryRecord.stats.HeapInuse
}
