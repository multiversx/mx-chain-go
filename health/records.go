package health

import (
	"container/list"
	"os"
)

type records struct {
	capacity int
	items    *list.List
}

func newRecords(capacity int) *records {
	return &records{
		capacity: capacity,
		items:    list.New(),
	}
}

func (records *records) addRecord(incomingRecord record) {
	if !records.makeRoomForRecord(incomingRecord) {
		return
	}

	insertionPlace := records.findInsertionPlace(incomingRecord)
	if insertionPlace == nil {
		records.items.PushBack(incomingRecord)
	} else {
		records.items.InsertBefore(incomingRecord, insertionPlace)
	}

	err := incomingRecord.save()
	if err != nil {
		log.Error("records.addRecord()", "err", err)
	}
}

func (records *records) makeRoomForRecord(incomingRecord record) bool {
	capacityReached := records.items.Len() > records.capacity
	if capacityReached {
		if records.getLeastImportant().isMoreImportantThan(incomingRecord) {
			return false
		}

		records.removeLeastImportant()
	}

	return true
}

func (records *records) getLeastImportant() record {
	return records.items.Back().Value.(record)
}

func (records *records) removeLeastImportant() {
	leastImportantElement := records.items.Back()
	leastImportantRecord := leastImportantElement.Value.(record)

	records.items.Remove(leastImportantElement)
	err := os.Remove(leastImportantRecord.getFilename())
	if err != nil {
		log.Error("records.removeLeastImportant()", "file", leastImportantRecord.getFilename(), "err", err)
	}
}

func (records *records) findInsertionPlace(incomingRecord record) *list.Element {
	for element := records.items.Front(); element != nil; element = element.Next() {
		record := element.Value.(record)

		if incomingRecord.isMoreImportantThan(record) {
			return element
		}
	}

	return nil
}
