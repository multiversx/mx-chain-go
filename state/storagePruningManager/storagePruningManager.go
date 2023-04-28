package storagePruningManager

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/pruningBuffer"
	logger "github.com/multiversx/mx-chain-logger-go"
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
	oldHashes common.ModifiedHashes,
	newHashes common.ModifiedHashes,
) error {
	if bytes.Equal(newRoot, oldRoot) {
		log.Trace("old root and new root are identical", "rootHash", newRoot)
		return nil
	}

	log.Trace("trie hashes sizes", "newHashes", len(newHashes), "oldHashes", len(oldHashes))
	removeDuplicatedKeys(oldHashes, newHashes)

	err := spm.markForEviction(newRoot, newHashes, state.NewRoot)
	if err != nil {
		return err
	}

	return spm.markForEviction(oldRoot, oldHashes, state.OldRoot)
}

func (spm *storagePruningManager) markForEviction(
	rootHash []byte,
	hashes map[string]struct{},
	identifier state.TriePruningIdentifier,
) error {
	if len(rootHash) == 0 || len(hashes) == 0 {
		return nil
	}

	rootHash = append(rootHash, byte(identifier))

	err := spm.dbEvictionWaitingList.Put(rootHash, hashes)
	if err != nil {
		return err
	}

	logMapWithTrace(fmt.Sprintf("MarkForEviction %d", identifier), "hash", hashes)
	return nil
}

func removeDuplicatedKeys(oldHashes map[string]struct{}, newHashes map[string]struct{}) {
	for key := range oldHashes {
		_, ok := newHashes[key]
		if ok {
			delete(oldHashes, key)
			delete(newHashes, key)
			log.Trace("found in newHashes and oldHashes", "hash", []byte(key))
		}
	}
}

func logMapWithTrace(message string, paramName string, hashes common.ModifiedHashes) {
	if log.GetLevel() == logger.LogTrace {
		for key := range hashes {
			log.Trace(message, paramName, []byte(key))
		}
	}
}

// PruneTrie removes old values from the trie database
func (spm *storagePruningManager) PruneTrie(
	rootHash []byte,
	identifier state.TriePruningIdentifier,
	tsm common.StorageManager,
	handler state.PruningHandler,
) {
	rootHash = append(rootHash, byte(identifier))

	if tsm.IsPruningBlocked() {
		if identifier == state.NewRoot {
			spm.cancelPrune(rootHash)
			return
		}

		rootHash = append(rootHash, byte(prune))
		spm.pruningBuffer.Add(rootHash)

		return
	}

	oldHashes := spm.pruningBuffer.RemoveAll()
	spm.resolveBufferedHashes(oldHashes, tsm, handler)
	spm.prune(rootHash, tsm, handler)
}

// CancelPrune clears the evictionWaitingList at the given hash
func (spm *storagePruningManager) CancelPrune(rootHash []byte, identifier state.TriePruningIdentifier, tsm common.StorageManager) {
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

func (spm *storagePruningManager) resolveBufferedHashes(oldHashes [][]byte, tsm common.StorageManager, handler state.PruningHandler) {
	for _, rootHash := range oldHashes {
		lastBytePos := len(rootHash) - 1
		if lastBytePos < 0 {
			continue
		}

		pruneOperation := pruningOperation(rootHash[lastBytePos])
		rootHash = rootHash[:lastBytePos]

		switch pruneOperation {
		case prune:
			spm.prune(rootHash, tsm, handler)
		case cancelPrune:
			spm.cancelPrune(rootHash)
		default:
			log.Error("invalid pruning operation", "operation id", pruneOperation)
		}
	}
}

func (spm *storagePruningManager) prune(rootHash []byte, tsm common.StorageManager, handler state.PruningHandler) {
	log.Trace("trie storage manager prune", "root", rootHash)

	err := spm.removeFromDb(rootHash, tsm, handler)
	if err != nil {
		if errors.IsClosingError(err) {
			log.Debug("did not remove hash", "rootHash", rootHash, "error", err)
			return
		}

		log.Error("trie storage manager remove from db", "error", err, "rootHash", hex.EncodeToString(rootHash))
	}
}

func (spm *storagePruningManager) removeFromDb(
	rootHash []byte,
	tsm common.StorageManager,
	handler state.PruningHandler,
) error {
	hashes, err := spm.dbEvictionWaitingList.Evict(rootHash)
	if err != nil {
		return err
	}

	if !handler.IsPruningEnabled() {
		log.Debug("trie removeFromDb", "skipping remove for rootHash", rootHash)
		return nil
	}

	log.Debug("trie removeFromDb", "rootHash", rootHash)

	lastBytePos := len(rootHash) - 1
	if lastBytePos < 0 {
		return state.ErrInvalidIdentifier
	}
	identifier := state.TriePruningIdentifier(rootHash[lastBytePos])

	sw := core.NewStopWatch()
	sw.Start("removeFromDb")
	defer func() {
		sw.Stop("removeFromDb")
		log.Debug("trieStorageManager.removeFromDb", sw.GetMeasurements()...)
	}()

	for key := range hashes {
		shouldKeepHash, errShouldKeep := spm.dbEvictionWaitingList.ShouldKeepHash(key, identifier)
		if errShouldKeep != nil {
			return errShouldKeep
		}
		if shouldKeepHash {
			continue
		}

		hash := []byte(key)
		log.Trace("remove hash from trie db", "hash", key)
		errRemove := tsm.Remove(hash)
		if errRemove != nil {
			return errRemove
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
