package state

// TrackableDataAccountWrap wraps SimpleAccountWrap adding modifying data capabilities
type TrackableDataAccountWrap struct {
	AccountWrapper

	originalData map[string][]byte
	dirtyData    map[string][]byte
}

// NewTrackableDataAccountWrap returns a ModifyindDataAccountWrap that wraps AccountWrap
// adding modifying data capabilities
func NewTrackableDataAccountWrap(accountWrapper AccountWrapper) (*TrackableDataAccountWrap, error) {
	if accountWrapper == nil {
		return nil, ErrNilSimpleAccountWrapper
	}

	return &TrackableDataAccountWrap{
		AccountWrapper: accountWrapper,
		originalData:   make(map[string][]byte),
		dirtyData:      make(map[string][]byte),
	}, nil
}

// ClearDataCaches empties the dirtyData map and original map
func (tdaw *TrackableDataAccountWrap) ClearDataCaches() {
	tdaw.dirtyData = make(map[string][]byte)
	tdaw.originalData = make(map[string][]byte)
}

// DirtyData returns the map of (key, value) pairs that contain the data needed to be saved in the data trie
func (tdaw *TrackableDataAccountWrap) DirtyData() map[string][]byte {
	return tdaw.dirtyData
}

// OriginalValue returns the value for a key stored in originalData map which is acting like a cache
func (tdaw *TrackableDataAccountWrap) OriginalValue(key []byte) []byte {
	return tdaw.originalData[string(key)]
}

// RetrieveValue fetches the value from a particular key searching the account data store
// The search starts with dirty map, continues with original map and ends with the trie
// Data must have been retrieved from its trie
func (tdaw *TrackableDataAccountWrap) RetrieveValue(key []byte) ([]byte, error) {
	if tdaw.DataTrie() == nil {
		return nil, ErrNilDataTrie
	}

	strKey := string(key)

	//search in dirty data cache
	data, found := tdaw.dirtyData[strKey]
	if found {
		return data, nil
	}

	//search in original data cache
	data, found = tdaw.originalData[strKey]
	if found {
		return data, nil
	}

	//ok, not in cache, retrieve from trie
	data, err := tdaw.DataTrie().Get(key)
	if err != nil {
		return nil, err
	}

	//got the value, put it originalData cache as the next fetch will run faster
	tdaw.originalData[string(key)] = data
	return data, nil
}

// SaveKeyValue stores in dirtyData the data keys "touched"
// It does not care if the data is really dirty as calling this check here will be sub-optimal
func (tdaw *TrackableDataAccountWrap) SaveKeyValue(key []byte, value []byte) {
	tdaw.dirtyData[string(key)] = value
}
