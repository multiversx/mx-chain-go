package update

import "errors"

var ErrUnknownType = errors.New("unknown type")

var ErrNilStateSyncer = errors.New("nil state syncer")

var ErrNoFileToImport = errors.New("no files to import")

var ErrEndOfFile = errors.New("end of file")

var ErrHashMissmatch = errors.New("hash missmatch")

var ErrNotSynced = errors.New("not synced")

var ErrNilActiveTries = errors.New("nil active tries")

var ErrNilTrieSyncers = errors.New("nil trie syncers")

var ErrNotEpochStartBlock = errors.New("not epoch start block")
