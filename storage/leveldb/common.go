package leveldb

import (
	"fmt"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const resourceUnavailable = "resource temporarily unavailable"
const maxRetries = 10
const timeBetweenRetries = time.Second

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

		log.Debug("error opening db",
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
