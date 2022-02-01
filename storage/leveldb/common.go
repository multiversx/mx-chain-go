package leveldb

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const resourceUnavailable = "resource temporarily unavailable"
const maxRetries = 10
const timeBetweenRetries = time.Second

// loggingDBCounter this variable should be used only used in logging prints
var loggingDBCounter = uint32(0)

func openLevelDB(path string, options *opt.Options) (*leveldb.DB, error) {
	retries := 0
	for {
		db, err := openOneTime(path, options)
		if err == nil {
			return db, nil
		}
		if err.Error() != resourceUnavailable {
			return nil, err
		}

		log.Debug("error opening DB",
			"error", err,
			"path", path,
			"retry", retries,
		)

		time.Sleep(timeBetweenRetries)
		retries++
		if retries > maxRetries {
			return nil, fmt.Errorf("%w, retried %d number of times", err, maxRetries)
		}
	}
}

func openOneTime(path string, options *opt.Options) (*leveldb.DB, error) {
	db, errOpen := leveldb.OpenFile(path, options)
	if errOpen == nil {
		return db, nil
	}

	if errors.IsCorrupted(errOpen) {
		var errRecover error
		log.Warn("corrupted DB file",
			"path", path,
			"error", errOpen,
		)
		db, errRecover = leveldb.RecoverFile(path, options)
		if errRecover != nil {
			return nil, fmt.Errorf("%w while recovering DB %s, after the initial failure %s",
				errRecover,
				path,
				errOpen.Error(),
			)
		}
		log.Info("DB file recovered",
			"path", path,
		)

		return db, nil
	}

	return nil, errOpen
}

type baseLevelDb struct {
	mutDb sync.RWMutex
	path  string
	db    *leveldb.DB
}

func (bldb *baseLevelDb) getDbPointer() *leveldb.DB {
	bldb.mutDb.RLock()
	defer bldb.mutDb.RUnlock()

	return bldb.db
}

func (bldb *baseLevelDb) makeDbPointerNilReturningLast() *leveldb.DB {
	bldb.mutDb.Lock()
	defer bldb.mutDb.Unlock()

	if bldb.db != nil {
		crtCounter := atomic.AddUint32(&loggingDBCounter, ^uint32(0)) // subtract 1
		log.Debug("makeDbPointerNilReturningLast", "path", bldb.path, "nilled pointer", fmt.Sprintf("%p", bldb.db), "global db counter", crtCounter)
	}

	db := bldb.db
	bldb.db = nil

	return db
}

// RangeKeys will call the handler function for each (key, value) pair
// If the handler returns true, the iteration will continue, otherwise will stop
func (bldb *baseLevelDb) RangeKeys(handler func(key []byte, value []byte) bool) {
	if handler == nil {
		return
	}

	db := bldb.getDbPointer()
	if db == nil {
		return
	}

	iterator := db.NewIterator(nil, nil)
	for {
		if !iterator.Next() {
			break
		}

		key := iterator.Key()
		clonedKey := make([]byte, len(key))
		copy(clonedKey, key)

		val := iterator.Value()
		clonedVal := make([]byte, len(val))
		copy(clonedVal, val)

		shouldContinue := handler(clonedKey, clonedVal)
		if !shouldContinue {
			break
		}
	}

	iterator.Release()
}
