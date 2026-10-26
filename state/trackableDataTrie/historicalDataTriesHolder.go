package trackableDataTrie

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
)

var _ common.TriesHolder = (*historicalDataTriesHolder)(nil)

// historicalDataTriesHolder validates address-cache hits against the account's
// committed root. Other historical requests may replace or reset the shared cache
// after this account leaves the historical adapter's root-selection lock.
type historicalDataTriesHolder struct {
	common.TriesHolder
	rootHash []byte
}

func (holder *historicalDataTriesHolder) Get(address []byte) common.Trie {
	tr := holder.TriesHolder.Get(address)
	if check.IfNil(tr) {
		return nil
	}
	rootHash, err := tr.RootHash()
	if err != nil || !bytes.Equal(rootHash, holder.rootHash) {
		return nil
	}
	return tr
}

func (holder *historicalDataTriesHolder) IsInterfaceNil() bool {
	return holder == nil
}
