package leveldb

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func openLevelDB(path string, options *opt.Options) (*leveldb.DB, error) {
	db, errOpen := leveldb.OpenFile(path, options)
	if errOpen != nil {
		var errRecover error
		log.Warn("error opening DB file",
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
	}

	return db, nil
}
