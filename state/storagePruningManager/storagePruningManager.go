package storagePruningManager

import (
	"bytes"
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/pruningBuffer"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
)

type pruningOperation byte

const (
	cancelPrune pruningOperation = 0
	prune       pruningOperation = 1
)

var log = logger.GetOrCreate("state/storagePruningManager")

type storagePruningManager struct {
	dbEvictionWaitingList state.DBRemoveCacher
	pruningBuffer         state.AtomicBuffer
}

// NewStoragePruningManager creates a new instance of storagePruningManager
func NewStoragePruningManager(
	evictionWaitingList state.DBRemoveCacher,
	pruningBufferLen uint32,
) (*storagePruningManager, error) {
	if check.IfNil(evictionWaitingList) {
		return nil, state.ErrNilEvictionWaitingList
	}

	return &storagePruningManager{
		dbEvictionWaitingList: evictionWaitingList,
		pruningBuffer:         pruningBuffer.NewPruningBuffer(pruningBufferLen),
	}, nil
}

// MarkForEviction adds the given hashes to the underlying evictionWaitingList
func (spm *storagePruningManager) MarkForEviction(
	oldRoot []byte,
	newRoot []byte,
	oldHashes temporary.ModifiedHashes,
	newHashes temporary.ModifiedHashes,
) error {
	if bytes.Equal(newRoot, oldRoot) {
		log.Trace("old root and new root are identical", "rootHash", newRoot)
		return nil
	}

	log.Trace("trie hashes sizes", "newHashes", len(newHashes), "oldHashes", len(oldHashes))
	removeDuplicatedKeys(oldHashes, newHashes)

	if len(newHashes) > 0 && len(newRoot) > 0 {
		newRoot = append(newRoot, byte(temporary.NewRoot))
		err := spm.dbEvictionWaitingList.Put(newRoot, newHashes)
		if err != nil {
			return err
		}

		logMapWithTrace("MarkForEviction newHashes", "hash", newHashes)
	}

	if len(oldHashes) > 0 && len(oldRoot) > 0 {
		oldRoot = append(oldRoot, byte(temporary.OldRoot))
		err := spm.dbEvictionWaitingList.Put(oldRoot, oldHashes)
		if err != nil {
			return err
		}

		logMapWithTrace("MarkForEviction oldHashes", "hash", oldHashes)
	}
	return nil
}

func removeDuplicatedKeys(oldHashes map[string]struct{}, newHashes map[string]struct{}) {
	for key := range oldHashes {
		_, ok := newHashes[key]
		if ok {
			delete(oldHashes, key)
			delete(newHashes, key)
			log.Trace("found in newHashes and oldHashes", "hash", key)
		}
	}
}

func logMapWithTrace(message string, paramName string, hashes temporary.ModifiedHashes) {
	if log.GetLevel() == logger.LogTrace {
		for key := range hashes {
			log.Trace(message, paramName, key)
		}
	}
}

// PruneTrie removes old values from the trie database
func (spm *storagePruningManager) PruneTrie(
	rootHash []byte,
	identifier temporary.TriePruningIdentifier,
	tsm temporary.StorageManager,
) {
	rootHash = append(rootHash, byte(identifier))

	if tsm.IsPruningBlocked() {
		if identifier == temporary.NewRoot {
			spm.cancelPrune(rootHash)
			return
		}

		rootHash = append(rootHash, byte(prune))
		spm.pruningBuffer.Add(rootHash)

		return
	}

	oldHashes := spm.pruningBuffer.RemoveAll()
	spm.resolveBufferedHashes(oldHashes, tsm)
	spm.prune(rootHash, tsm)
}

// CancelPrune clears the evictionWaitingList at the given hash
func (spm *storagePruningManager) CancelPrune(rootHash []byte, identifier temporary.TriePruningIdentifier, tsm temporary.StorageManager) {
	rootHash = append(rootHash, byte(identifier))

	if tsm.IsPruningBlocked() || spm.pruningBuffer.Len() != 0 {
		rootHash = append(rootHash, byte(cancelPrune))
		spm.pruningBuffer.Add(rootHash)

		return
	}

	spm.cancelPrune(rootHash)
}

func (spm *storagePruningManager) cancelPrune(rootHash []byte) {
	log.Trace("trie storage manager cancel prune", "root", rootHash)
	_, _ = spm.dbEvictionWaitingList.Evict(rootHash)
}

func (spm *storagePruningManager) resolveBufferedHashes(oldHashes [][]byte, tsm temporary.StorageManager) {
	for _, rootHash := range oldHashes {
		lastBytePos := len(rootHash) - 1
		if lastBytePos < 0 {
			continue
		}

		pruneOperation := pruningOperation(rootHash[lastBytePos])
		rootHash = rootHash[:lastBytePos]

		switch pruneOperation {
		case prune:
			spm.prune(rootHash, tsm)
		case cancelPrune:
			spm.cancelPrune(rootHash)
		default:
			log.Error("invalid pruning operation", "operation id", pruneOperation)
		}
	}
}

func (spm *storagePruningManager) prune(rootHash []byte, tsm temporary.StorageManager) {
	log.Trace("trie storage manager prune", "root", rootHash)

	err := spm.removeFromDb(rootHash, tsm)
	if err != nil {
		log.Error("trie storage manager remove from db", "error", err, "rootHash", hex.EncodeToString(rootHash))
	}
}

func (spm *storagePruningManager) removeFromDb(
	rootHash []byte,
	tsm temporary.StorageManager,
) error {
	hashes, err := spm.dbEvictionWaitingList.Evict(rootHash)
	if err != nil {
		return err
	}

	log.Debug("trie removeFromDb", "rootHash", rootHash)

	lastBytePos := len(rootHash) - 1
	if lastBytePos < 0 {
		return state.ErrInvalidIdentifier
	}
	identifier := temporary.TriePruningIdentifier(rootHash[lastBytePos])

	sw := core.NewStopWatch()
	sw.Start("removeFromDb")
	defer func() {
		sw.Stop("removeFromDb")
		log.Debug("trieStorageManager.removeFromDb", sw.GetMeasurements()...)
	}()

	for key := range hashes {
		shouldKeepHash, err := spm.dbEvictionWaitingList.ShouldKeepHash(key, identifier)
		if err != nil {
			return err
		}
		if shouldKeepHash {
			continue
		}

		hash := []byte(key)

		log.Trace("remove hash from trie db", "hash", hex.EncodeToString(hash))
		err = tsm.Remove(hash)
		if err != nil {
			return err
		}
	}

	return nil
}

// Close will handle the closing of the underlying components
func (spm *storagePruningManager) Close() error {
	return spm.dbEvictionWaitingList.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (spm *storagePruningManager) IsInterfaceNil() bool {
	return spm == nil
}
