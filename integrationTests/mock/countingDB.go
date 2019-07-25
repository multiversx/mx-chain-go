package mock

import (
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
)

type countingDB struct {
	db      *memorydb.DB
	nrOfPut int
}

func NewCountingDB() *countingDB {
	db, _ := memorydb.New()
	return &countingDB{db, 0}
}

func (cdb *countingDB) Put(key, val []byte) error {
	cdb.db.Put(key, val)
	cdb.nrOfPut++
	return nil
}

func (cdb *countingDB) Get(key []byte) ([]byte, error) {
	return cdb.db.Get(key)
}

func (cdb *countingDB) Has(key []byte) error {
	return cdb.db.Has(key)
}

func (cdb *countingDB) Init() error {
	return cdb.db.Init()
}

func (cdb *countingDB) Close() error {
	return cdb.db.Close()
}

func (cdb *countingDB) Remove(key []byte) error {
	return cdb.db.Remove(key)
}

func (cdb *countingDB) Destroy() error {
	return cdb.db.Destroy()
}

func (cdb *countingDB) Reset() {
	cdb.nrOfPut = 0
}

func (cdb *countingDB) GetCounter() int {
	return cdb.nrOfPut
}
