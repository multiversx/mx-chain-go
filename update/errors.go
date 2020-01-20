package update

import "errors"

// ErrUnknownType signals that type is unknown
var ErrUnknownType = errors.New("unknown type")

// ErrNilStateSyncer signals that state syncer is nil
var ErrNilStateSyncer = errors.New("nil state syncer")

// ErrNoFileToImport signals that there are no files to import
var ErrNoFileToImport = errors.New("no files to import")

// ErrEndOfFile signals that end of file was reached
var ErrEndOfFile = errors.New("end of file")

// ErrHashMissmatch signals that received hash is not equal with processed one
var ErrHashMissmatch = errors.New("hash missmatch")

// ErrNilDataWriter signals that data writer is nil
var ErrNilDataWriter = errors.New("nil data writer")

// ErrNilDataReader signals that data reader is nil
var ErrNilDataReader = errors.New("nil data reader")

// ErrInvalidFolderName signals that folder name is nil
var ErrInvalidFolderName = errors.New("invalid folder name")

// ErrNilStorage signals that storage is nil
var ErrNilStorage = errors.New("nil storage")

// ErrNilDataTrieContainer signals that data trie container is nil
var ErrNilDataTrieContainer = errors.New("nil data trie container")

// ErrNotSynced signals that syncing has not been finished yet
var ErrNotSynced = errors.New("not synced")

// ErrNilActiveTries signals that active tries container is nil
var ErrNilActiveTries = errors.New("nil active tries")

// ErrNilTrieSyncers signals that trie syncers container is nil
var ErrNilTrieSyncers = errors.New("nil trie syncers")

// ErrNotEpochStartBlock signals that block is not of type epoch start
var ErrNotEpochStartBlock = errors.New("not epoch start block")

// ErrNilContainerElement signals when trying to add a nil element in the container
var ErrNilContainerElement = errors.New("element cannot be nil")

// ErrInvalidContainerKey signals that an element does not exist in the container's map
var ErrInvalidContainerKey = errors.New("element does not exist in container")

// ErrContainerKeyAlreadyExists signals that an element was already set in the container's map
var ErrContainerKeyAlreadyExists = errors.New("provided key already exists in container")

// ErrLenMismatch signals that 2 or more slices have different lengths
var ErrLenMismatch = errors.New("lengths mismatch")

// ErrWrongTypeInContainer signals that a wrong type of object was found in container
var ErrWrongTypeInContainer = errors.New("wrong type of object inside container")

// ErrNilWhiteListHandler signals that white list handler is nil
var ErrNilWhiteListHandler = errors.New("nil white list handler")
