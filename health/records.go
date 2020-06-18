package health

import (
	"container/list"
	"os"
	"path"
)

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
