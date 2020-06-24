package health

import (
	"container/list"
	"sync"
)

type records struct {
	capacity int
	items    *list.List
	mutex    sync.RWMutex
}

func newRecords(capacity int) *records {
	return &records{
		capacity: capacity,
		items:    list.New(),
	}
}

func (records *records) addRecord(incomingRecord record) {
	records.mutex.Lock()
	defer records.mutex.Unlock()

	if !records.makeRoomForRecordNoLock(incomingRecord) {
		return
	}

	insertionPlace := records.findInsertionPlaceNoLock(incomingRecord)
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

func (records *records) makeRoomForRecordNoLock(incomingRecord record) bool {
	capacityReached := records.items.Len() >= records.capacity
	if capacityReached {
		if records.getLeastImportantNoLock().isMoreImportantThan(incomingRecord) {
			return false
		}

		records.evictLeastImportantNoLock()
	}

	return true
}

func (records *records) getLeastImportantNoLock() record {
	return records.items.Back().Value.(record)
}

func (records *records) evictLeastImportantNoLock() {
	leastImportantElement := records.items.Back()
	leastImportantRecord := leastImportantElement.Value.(record)

	records.items.Remove(leastImportantElement)
	err := leastImportantRecord.delete()
	if err != nil {
		log.Error("records.evictLeastImportantNoLock()", "err", err)
	}
}

func (records *records) findInsertionPlaceNoLock(incomingRecord record) *list.Element {
	for element := records.items.Front(); element != nil; element = element.Next() {
		record := element.Value.(record)

		if incomingRecord.isMoreImportantThan(record) {
			return element
		}
	}

	return nil
}

func (records *records) len() int {
	records.mutex.RLock()
	defer records.mutex.RUnlock()

	return records.items.Len()
}

func (records *records) getMostImportant() record {
	records.mutex.RLock()
	defer records.mutex.RUnlock()

	return records.items.Front().Value.(record)
}
