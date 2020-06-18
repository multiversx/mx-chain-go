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

type memoryRecord struct {
	stats    runtime.MemStats
	filename string
}

func newMemoryRecord(stats runtime.MemStats) *memoryRecord {
	return &memoryRecord{
		stats:    stats,
		filename: formatMemoryRecordFilename(stats),
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

func (record *memoryRecord) isHigherThan(otherRecord *memoryRecord) bool {
	return record.stats.HeapInuse > otherRecord.stats.HeapInuse
}

func (record *memoryRecord) save(folderPath string) error {
	filename := path.Join(folderPath, record.filename)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	log.Info("record.save()", "file", filename)

	err = pprof.WriteHeapProfile(file)
	if err != nil {
		return err
	}

	return file.Close()
}
