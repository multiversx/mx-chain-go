package update

import "errors"

var ErrUnknownType = errors.New("unknown type")

var ErrNilStateSyncer = errors.New("nil state syncer")

var ErrNoFileToImport = errors.New("no files to import")

var ErrEndOfFile = errors.New("end of file")

var ErrHashMissmatch = errors.New("hash missmatch")

var ErrNilDataWriter = errors.New("nil data writer")

var ErrNilDataReader = errors.New("nil data reader")

var ErrInvalidFolderName = errors.New("invalid folder name")

var ErrNilStorage = errors.New("nil storage")

var ErrNilActiveAccountHandlersContainer = errors.New("nil active account handlers container")

var ErrNilDataTrieContainer = errors.New("nil data trie container")

var ErrNotSynced = errors.New("not synced")

var ErrNilActiveTries = errors.New("nil active tries")

var ErrNilTrieSyncers = errors.New("nil trie syncers")

var ErrNotEpochStartBlock = errors.New("not epoch start block")
