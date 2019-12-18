package headersCache

import (
	"sort"
	"sync"
	"time"
)

type nonceTimestamp struct {
	nonce     uint64
	timestamp time.Time
}

type headersMap struct {
	hdrsMap    map[uint64]headerListDetails
	mutHdrsMap sync.RWMutex
}

func newHeadersMap() *headersMap {
	return &headersMap{
		hdrsMap:    make(map[uint64]headerListDetails),
		mutHdrsMap: sync.RWMutex{},
	}
}

func (h *headersMap) addElement(nonce uint64, details headerListDetails) {
	h.mutHdrsMap.Lock()
	h.hdrsMap[nonce] = details
	h.mutHdrsMap.Unlock()
}

func (h *headersMap) getElement(nonce uint64) headerListDetails {
	h.mutHdrsMap.RLock()
	defer h.mutHdrsMap.RUnlock()

	element, ok := h.hdrsMap[nonce]
	if !ok {
		return headerListDetails{
			headerList: make([]headerDetails, 0),
			timestamp:  time.Now(),
		}
	}

	return element
}

func (h *headersMap) removeElement(nonce uint64) {
	h.mutHdrsMap.Lock()
	delete(h.hdrsMap, nonce)
	h.mutHdrsMap.Unlock()
}

func (h *headersMap) getNoncesTimestampSorted() []uint64 {
	noncesTimestampsSlice := make([]nonceTimestamp, 0)
	h.mutHdrsMap.RLock()
	for key, value := range h.hdrsMap {
		noncesTimestampsSlice = append(noncesTimestampsSlice, nonceTimestamp{nonce: key, timestamp: value.timestamp})
	}
	h.mutHdrsMap.RUnlock()

	sort.Slice(noncesTimestampsSlice, func(i, j int) bool {
		return noncesTimestampsSlice[j].timestamp.After(noncesTimestampsSlice[i].timestamp)
	})

	nonceSlice := make([]uint64, 0)
	for _, d := range noncesTimestampsSlice {
		nonceSlice = append(nonceSlice, d.nonce)
	}

	return nonceSlice
}

func (h *headersMap) getHeadersDetailsListFromSMap(hdrNonce uint64) (headerListDetails, bool) {
	headersListD := h.getElement(hdrNonce)
	if len(headersListD.headerList) == 0 {
		return headerListDetails{}, false
	}

	//update timestamp
	headersListD.timestamp = time.Now()
	h.addElement(hdrNonce, headersListD)

	return headersListD, true
}

func (h *headersMap) keys() []uint64 {
	h.mutHdrsMap.RLock()
	defer h.mutHdrsMap.RUnlock()

	nonces := make([]uint64, len(h.hdrsMap))

	for key := range h.hdrsMap {
		nonces = append(nonces, key)
	}

	return nonces
}
