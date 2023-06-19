package trackableDataTrie

import "github.com/multiversx/mx-chain-core-go/core"

// DirtyData -
type DirtyData struct {
	Value      []byte
	NewVersion core.TrieNodeVersion
}

// DirtyData -
func (tdaw *trackableDataTrie) DirtyData() map[string]DirtyData {
	dd := make(map[string]DirtyData, len(tdaw.dirtyData))

	for key, value := range tdaw.dirtyData {
		dd[key] = DirtyData{
			Value:      value.value,
			NewVersion: value.newVersion,
		}
	}

	return dd
}
