package headersCache

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"sort"
	"time"
)

type headersMap map[uint64]timestampedListOfHeaders

func (hMap headersMap) addElement(nonce uint64, details timestampedListOfHeaders) {
	hMap[nonce] = details
}

func (hMap headersMap) getElement(nonce uint64) timestampedListOfHeaders {
	element, ok := hMap[nonce]
	if !ok {
		return timestampedListOfHeaders{
			headers:   make([]headerDetails, 0),
			timestamp: time.Now(),
		}
	}

	return element
}

func (hMap headersMap) appendElement(headerHash []byte, header data.HeaderHandler) {
	headerNonce := header.GetNonce()
	headersWithTimestamp := hMap.getElement(headerNonce)

	headerDetails := headerDetails{
		headerHash: headerHash,
		header:     header,
	}
	headersWithTimestamp.headers = append(headersWithTimestamp.headers, headerDetails)
	hMap.addElement(headerNonce, headersWithTimestamp)
}

func (hMap headersMap) removeElement(nonce uint64) {
	delete(hMap, nonce)
}

func (hMap headersMap) getNoncesSortedByTimestamp() []uint64 {
	noncesTimestampsSlice := make([]nonceTimestamp, 0)

	for key, value := range hMap {
		noncesTimestampsSlice = append(noncesTimestampsSlice, nonceTimestamp{nonce: key, timestamp: value.timestamp})
	}

	sort.Slice(noncesTimestampsSlice, func(i, j int) bool {
		return noncesTimestampsSlice[j].timestamp.After(noncesTimestampsSlice[i].timestamp)
	})

	nonces := make([]uint64, 0)
	for _, element := range noncesTimestampsSlice {
		nonces = append(nonces, element.nonce)
	}

	return nonces
}

// getHeadersByNonce will return a list of headers and update timestamp
func (hMap headersMap) getHeadersByNonce(hdrNonce uint64) (timestampedListOfHeaders, bool) {
	hdrsWithTimestamp := hMap.getElement(hdrNonce)
	if hdrsWithTimestamp.isEmpty() {
		return timestampedListOfHeaders{}, false
	}

	//update timestamp
	hdrsWithTimestamp.timestamp = time.Now()
	hMap.addElement(hdrNonce, hdrsWithTimestamp)

	return hdrsWithTimestamp, true
}

func (hMap headersMap) keys() []uint64 {
	nonces := make([]uint64, 0, len(hMap))

	for key := range hMap {
		nonces = append(nonces, key)
	}

	return nonces
}
