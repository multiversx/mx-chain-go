package headersCache

import (
	"sort"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
)

type listOfHeadersByNonces map[uint64]timestampedListOfHeaders

func (hMap listOfHeadersByNonces) getListOfHeaders(nonce uint64) timestampedListOfHeaders {
	element, ok := hMap[nonce]
	if !ok {
		return timestampedListOfHeaders{
			items:     make([]headerDetails, 0),
			timestamp: time.Now(),
		}
	}

	return element
}

func (hMap listOfHeadersByNonces) appendHeaderToList(headerHash []byte, header data.HeaderHandler) {
	headerNonce := header.GetNonce()
	headersWithTimestamp := hMap.getListOfHeaders(headerNonce)

	hdrDetails := headerDetails{
		headerHash: headerHash,
		header:     header,
	}
	headersWithTimestamp.items = append(headersWithTimestamp.items, hdrDetails)
	hMap.setListOfHeaders(headerNonce, headersWithTimestamp)
}

func (hMap listOfHeadersByNonces) setListOfHeaders(nonce uint64, element timestampedListOfHeaders) {
	hMap[nonce] = element
}

func (hMap listOfHeadersByNonces) removeListOfHeaders(nonce uint64) {
	delete(hMap, nonce)
}

func (hMap listOfHeadersByNonces) getNoncesSortedByTimestamp() []uint64 {
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

// getHeadersByNonce will return a list of items and update timestamp
func (hMap listOfHeadersByNonces) getHeadersByNonce(hdrNonce uint64) (timestampedListOfHeaders, bool) {
	hdrsWithTimestamp := hMap.getListOfHeaders(hdrNonce)
	if hdrsWithTimestamp.isEmpty() {
		return timestampedListOfHeaders{}, false
	}

	//update timestamp
	hdrsWithTimestamp.timestamp = time.Now()
	hMap.setListOfHeaders(hdrNonce, hdrsWithTimestamp)

	return hdrsWithTimestamp, true
}

func (hMap listOfHeadersByNonces) keys() []uint64 {
	nonces := make([]uint64, 0, len(hMap))

	for key := range hMap {
		nonces = append(nonces, key)
	}

	return nonces
}
