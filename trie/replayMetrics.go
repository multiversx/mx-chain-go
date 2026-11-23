package trie

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/replaymetrics"
)

type replayTrieStorage struct {
	common.TrieStorageInteractor
	kind replaymetrics.Kind
}

// SetReplayMetricsKind marks an owned trie before it is published to processing.
func (tr *patriciaMerkleTrie) SetReplayMetricsKind(kind replaymetrics.Kind) {
	if tr.replayStorage == nil {
		tr.replayStorage = &replayTrieStorage{TrieStorageInteractor: tr.trieStorage, kind: kind}
	}
}

func (tr *patriciaMerkleTrie) processingStorage() common.TrieStorageInteractor {
	if tr.replayStorage != nil {
		return tr.replayStorage
	}
	return tr.trieStorage
}

func replayStorageScope(db common.TrieStorageInteractor) replaymetrics.Scope {
	if tracked, ok := db.(*replayTrieStorage); ok {
		return replaymetrics.Capture(tracked.kind)
	}
	return replaymetrics.Scope{}
}

func recordResolution(db common.TrieStorageInteractor, size int, err error) {
	stats := replayStorageScope(db)
	if err != nil {
		stats.Add(replaymetrics.ResolveErrors, 1)
		return
	}
	stats.Add(replaymetrics.NodesResolved, 1)
	stats.Add(replaymetrics.ResolvedBytes, uint64(size))
}

func (db *replayTrieStorage) Put(key, value []byte) error {
	err := db.TrieStorageInteractor.Put(key, value)
	if err == nil {
		stats := replaymetrics.Capture(db.kind)
		stats.Add(replaymetrics.NodeWrites, 1)
		stats.Add(replaymetrics.NodeWriteBytes, uint64(len(value)))
	}
	return err
}
