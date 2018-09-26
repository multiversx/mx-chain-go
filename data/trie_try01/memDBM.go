package trie_try01

import (
	"github.com/pkg/errors"
	"sync"
)

type MemDBM struct {
	nodes map[string]*dataNode
	lock  sync.RWMutex
	marsh Marshalizer
}

func NewMemDBM(marsh Marshalizer) (DatabaseManager, error) {
	if marsh == nil {
		errors.New("MemDBM.NewMemDBM called without a marshalizer")
	}

	return &MemDBM{marsh: marsh}, nil
}

func (mdb *MemDBM) Node(hash []byte, cachegen uint16) dataNode {
	mdb.lock.RLock()
	defer mdb.lock.RUnlock()

	val, ok := mdb.nodes[string(hash)]
	if !ok {
		return nil
	}

	mdb.marsh.Unmarshal()

	return
}

func (mdb *MemDBM) Lock() {
	panic("implement me")
}

func (mdb *MemDBM) Unlock() {
	panic("implement me")
}

func (mdb *MemDBM) Insert(hash []byte, blob []byte, node dataNode) {
	panic("implement me")
}
