package leveldb

import (
	"fmt"
	"reflect"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func openLevelDB(path string, options *opt.Options) (*leveldb.DB, error) {
	db, errOpen := leveldb.OpenFile(path, options)
	if errOpen == nil {
		return db, nil
	}

	if reflect.TypeOf(errOpen) == reflect.TypeOf(&errors.ErrCorrupted{}) {
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
