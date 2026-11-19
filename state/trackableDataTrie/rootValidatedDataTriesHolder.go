package trackableDataTrie

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
)

var _ common.TriesHolder = (*rootValidatedDataTriesHolder)(nil)

type growingDataTriesHolder interface {
	PutAndTrack(address []byte, tr common.Trie) common.Trie
}

func cacheDataTrie(holder common.TriesHolder, address []byte, tr common.Trie) common.Trie {
	if growingHolder, ok := holder.(growingDataTriesHolder); ok {
		return growingHolder.PutAndTrack(address, tr)
	}
	holder.Put(address, tr)
	return tr
}

// rootValidatedDataTriesHolder validates address-cache hits against the account's
// committed root. Other API requests may replace or reset the shared cache after
// this account leaves the adapter's root-selection lock.
type rootValidatedDataTriesHolder struct {
	common.TriesHolder
	rootHash []byte
}

func (holder *rootValidatedDataTriesHolder) Get(address []byte) common.Trie {
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

func (holder *rootValidatedDataTriesHolder) PutAndTrack(address []byte, tr common.Trie) common.Trie {
	return cacheDataTrie(holder.TriesHolder, address, tr)
}

func (holder *rootValidatedDataTriesHolder) IsInterfaceNil() bool {
	return holder == nil
}
