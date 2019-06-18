package state

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

type AdapterTrie struct {
	trie2.Trie
}

func (AdapterTrie) SetCacheLimit(l uint16) {
	panic("implement me")
}

func (at AdapterTrie) Root() []byte {
	rootHash, err := at.Trie.Root()
	if err != nil {
		return nil
	}
	return rootHash
}

func (at AdapterTrie) Commit(onleaf trie.LeafCallback) (root []byte, err error) {
	root = at.Root()
	err = at.Trie.Commit()
	return root, err
}

func (at AdapterTrie) DBW() trie.DBWriteCacher {
	db := at.Trie.DBW()
	return adapterDB{db}
}

func (at AdapterTrie) Recreate(root []byte, dbw trie.DBWriteCacher) (trie.PatriciaMerkelTree, error) {
	tr, _ := at.Trie.Recreate(root, at.Trie.DBW())

	return AdapterTrie{tr}, nil
}

func (AdapterTrie) Copy() trie.PatriciaMerkelTree {
	panic("implement me")
}

type adapterDB struct {
	trie2.DBWriteCacher
}

func (adapterDB) Storer() storage.Storer {
	panic("implement me")
}

func (adapterDB) InsertBlob(hash []byte, blob []byte) {
	panic("implement me")
}

func (adapterDB) Node(hash []byte) ([]byte, error) {
	panic("implement me")
}

func (adapterDB) Reference(child []byte, parent []byte) {
	panic("implement me")
}

func (adapterDB) Dereference(root []byte) {
	panic("implement me")
}

func (adapterDB) Cap(limit float64) error {
	panic("implement me")
}

func (adapterDB) Commit(node []byte, report bool) error {
	panic("implement me")
}

func (adapterDB) Size() (float64, float64) {
	panic("implement me")
}

func (adapterDB) InsertWithLock(hash []byte, blob []byte, node trie.Node) {
	panic("implement me")
}

func (adapterDB) CachedNode(hash []byte, cachegen uint16) trie.Node {
	panic("implement me")
}
