package state

// ModifyingDataAccountWrap wraps SimpleAccountWrap adding modifying data capabilities
type ModifyingDataAccountWrap struct {
	SimpleAccountWrapper

	originalData map[string][]byte
	dirtyData    map[string][]byte
}

// NewModifyingDataAccountWrap returns a ModifyindDataAccountWrap that wraps SimpleAccountWrap
// adding modifying data capabilities
func NewModifyingDataAccountWrap(simpleAccountWrapper SimpleAccountWrapper) (*ModifyingDataAccountWrap, error) {
	if simpleAccountWrapper == nil {
		return nil, ErrNilSimpleAccountWrapper
	}

	return &ModifyingDataAccountWrap{
		SimpleAccountWrapper: simpleAccountWrapper,
		originalData:         make(map[string][]byte),
		dirtyData:            make(map[string][]byte),
	}, nil
}

// ClearDataCaches empties the dirtyData map and original map
func (mdaw *ModifyingDataAccountWrap) ClearDataCaches() {
	mdaw.dirtyData = make(map[string][]byte)
	mdaw.originalData = make(map[string][]byte)
}

// DirtyData returns the map of (key, value) pairs that contain the data needed to be saved in the data trie
func (mdaw *ModifyingDataAccountWrap) DirtyData() map[string][]byte {
	return mdaw.dirtyData
}

// OriginalValue returns the value for a key stored in originalData map which is acting like a cache
func (mdaw *ModifyingDataAccountWrap) OriginalValue(key []byte) []byte {
	return mdaw.originalData[string(key)]
}

// RetrieveValue fetches the value from a particular key searching the account data store
// The search starts with dirty map, continues with original map and ends with the trie
// Data must have been retrieved from its trie
func (mdaw *ModifyingDataAccountWrap) RetrieveValue(key []byte) ([]byte, error) {
	if mdaw.DataTrie() == nil {
		return nil, ErrNilDataTrie
	}

	strKey := string(key)

	//search in dirty data cache
	data, found := mdaw.dirtyData[strKey]
	if found {
		return data, nil
	}

	//search in original data cache
	data, found = mdaw.originalData[strKey]
	if found {
		return data, nil
	}

	//ok, not in cache, retrieve from trie
	data, err := mdaw.DataTrie().Get(key)
	if err != nil {
		return nil, err
	}

	//got the value, put it originalData cache as the next fetch will run faster
	mdaw.originalData[string(key)] = data
	return data, nil
}

// SaveKeyValue stores in dirtyData the data keys "touched"
// It does not care if the data is really dirty as calling this check here will be sub-optimal
func (mdaw *ModifyingDataAccountWrap) SaveKeyValue(key []byte, value []byte) {
	mdaw.dirtyData[string(key)] = value
}
