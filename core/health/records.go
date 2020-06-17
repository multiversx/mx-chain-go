package health

import (
	"container/list"
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
	timestamp := time.Now().Format("20060102150405")
	inUse := core.ConvertBytes(stats.HeapInuse)
	inUse = strings.ReplaceAll(inUse, " ", "_")
	inUse = strings.ReplaceAll(inUse, ".", "_")
	filename := fmt.Sprintf("mem__%s__%s.pprof", timestamp, inUse)

	return &memoryRecord{
		stats:    stats,
		filename: filename,
	}
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

type records struct {
	capacity      int
	folderPath    string
	memoryRecords *list.List
}

func newRecords(capacity int, folderPath string) *records {
	return &records{
		capacity:      capacity,
		folderPath:    folderPath,
		memoryRecords: list.New(),
	}
}

func (records *records) addMemoryRecord(incomingRecord *memoryRecord) {
	if !records.makeRoomForRecord(incomingRecord) {
		return
	}

	insertionPlace := records.findInsertionPlace(incomingRecord)
	if insertionPlace == nil {
		records.memoryRecords.PushBack(incomingRecord)
	} else {
		records.memoryRecords.InsertBefore(incomingRecord, insertionPlace)
	}

	err := incomingRecord.save(records.folderPath)
	if err != nil {
		log.Error("records.addMemoryRecord()", "err", err)
	}
}

func (records *records) makeRoomForRecord(incomingRecord *memoryRecord) bool {
	capacityReached := records.memoryRecords.Len() > records.capacity
	if capacityReached {
		lowestRecord := records.getLowestMemoryRecord()
		if lowestRecord.isHigherThan(incomingRecord) {
			return false
		}

		records.removeLowestMemoryRecord()
	}

	return true
}

func (records *records) getLowestMemoryRecord() *memoryRecord {
	return records.memoryRecords.Back().Value.(*memoryRecord)
}

func (records *records) removeLowestMemoryRecord() {
	lowestElement := records.memoryRecords.Back()
	lowestRecord := lowestElement.Value.(*memoryRecord)

	records.memoryRecords.Remove(lowestElement)
	err := os.Remove(records.getRecordPath(lowestRecord.filename))
	if err != nil {
		log.Error("records.removeLowestMemoryRecord()", "file", lowestRecord.filename, "err", err)
	}
}

func (records *records) getRecordPath(filename string) string {
	return path.Join(records.folderPath, filename)
}

func (records *records) findInsertionPlace(incomingRecord *memoryRecord) *list.Element {
	for element := records.memoryRecords.Front(); element != nil; element = element.Next() {
		record := element.Value.(*memoryRecord)

		if incomingRecord.isHigherThan(record) {
			return element
		}
	}

	return nil
}
