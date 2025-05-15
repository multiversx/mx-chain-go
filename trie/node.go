//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. node.proto
package trie

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

const (
	nrOfChildren         = 17
	firstByte            = 0
	pointerSizeInBytes   = 8
	mutexSizeInBytes     = 24
	dirtyFlagSizeInBytes = 1
	pollingIdleNode      = time.Millisecond
)

type branchNode struct {
	CollapsedBn
	children [nrOfChildren]node
	*baseNode
	childrenMutexes [nrOfChildren]sync.RWMutex
}

type extensionNode struct {
	CollapsedEn
	child node
	*baseNode
	childMutex sync.RWMutex
}

type leafNode struct {
	CollapsedLn
	*baseNode
}

func encodeNodeAndGetHash(n node, trieCtx common.TrieContext) ([]byte, error) {
	encNode, err := n.getEncodedNode(trieCtx)
	if err != nil {
		return nil, err
	}

	hash := trieCtx.Compute(string(encNode))

	return hash, nil
}

// encodeNodeAndCommitToDB will encode and save provided node. It returns the node's value in bytes
func encodeNodeAndCommitToDB(n node, trieCtx common.TrieContext) (int, error) {
	val, err := n.getEncodedNode(trieCtx)
	if err != nil {
		return 0, err
	}
	key := trieCtx.Compute(string(val))

	// test point encodeNodeAndCommitToDB

	err = trieCtx.Put(key, val)

	return len(val), err
}

func getNodeFromDBAndDecode(n []byte, trieCtx common.TrieContext) (node, []byte, error) {
	encodedNode, err := trieCtx.Get(n)
	if err != nil {
		treatLogError(log, err, n)

		return nil, nil, core.NewGetNodeFromDBErrWithKey(n, err, trieCtx.GetIdentifier())
	}

	decodedNode, err := decodeNode(encodedNode, trieCtx)
	if err != nil {
		return nil, nil, err
	}

	return decodedNode, encodedNode, nil
}

func treatLogError(logInstance logger.Logger, err error, key []byte) {
	if logInstance.GetLevel() != logger.LogTrace {
		return
	}

	logInstance.Trace(core.GetNodeFromDBErrorString, "error", err, "key", key, "stack trace", string(debug.Stack()))
}

func handleStorageInteractorStats(db common.TrieStorageInteractor) {
	if db != nil {
		db.GetStateStatsHandler().IncrementTrie()
	}
}

func concat(s1 []byte, s2 ...byte) []byte {
	r := make([]byte, len(s1)+len(s2))
	copy(r, s1)
	copy(r[len(s1):], s2)

	return r
}

func decodeNode(encNode []byte, trieCtx common.TrieContext) (node, error) {
	if encNode == nil || len(encNode) < 1 {
		return nil, ErrInvalidEncoding
	}

	nodeType := encNode[len(encNode)-1]
	encNode = encNode[:len(encNode)-1]

	newNode, err := getEmptyNodeOfType(nodeType)
	if err != nil {
		return nil, err
	}

	err = trieCtx.Unmarshal(newNode, encNode)
	if err != nil {
		return nil, err
	}

	return newNode, nil
}

func getEmptyNodeOfType(t byte) (node, error) {
	switch t {
	case extension:
		return &extensionNode{baseNode: &baseNode{}}, nil
	case leaf:
		return &leafNode{baseNode: &baseNode{}}, nil
	case branch:
		return &branchNode{baseNode: &baseNode{}}, nil
	default:
		return nil, ErrInvalidNode
	}
}

func childPosOutOfRange(pos byte) bool {
	return pos >= nrOfChildren
}

// prefixLen returns the length of the common prefix of a and b.
func prefixLen(a, b []byte) int {
	i := 0
	length := len(a)
	if len(b) < length {
		length = len(b)
	}

	for ; i < length; i++ {
		if a[i] != b[i] {
			break
		}
	}

	return i
}

func shouldStopIfContextDoneBlockingIfBusy(ctx context.Context, idleProvider IdleNodeProvider) bool {
	for {
		select {
		case <-ctx.Done():
			return true
		default:
		}

		if idleProvider.IsIdle() {
			return false
		}

		select {
		case <-ctx.Done():
			return true
		case <-time.After(pollingIdleNode):
		}
	}
}

func treatCommitSnapshotError(err error, hash []byte, missingNodesChan chan []byte) (nodeIsMissing bool, error error) {
	if err == nil {
		return false, nil
	}

	if !core.IsGetNodeFromDBError(err) {
		return false, err
	}

	log.Error("error during trie snapshot", "err", err.Error(), "hash", hash)
	missingNodesChan <- hash
	return true, nil
}

func shouldMigrateCurrentNode(
	currentNode node,
	migrationArgs vmcommon.ArgsMigrateDataTrieLeaves,
) (bool, error) {
	version, err := currentNode.getVersion()
	if err != nil {
		return false, err
	}

	if version == migrationArgs.NewVersion {
		return false, nil
	}

	if version != migrationArgs.OldVersion && version != core.NotSpecified {
		return false, nil
	}

	return true, nil
}

func saveDirtyNodeToStorage(
	n node,
	goRoutinesManager common.TrieGoroutinesManager,
	hashesCollector common.TrieHashesCollector,
	trieCtx common.TrieContext,
) bool {
	n.setDirty(false)
	encNode, err := n.getEncodedNode(trieCtx)
	if err != nil {
		goRoutinesManager.SetError(err)
		return false
	}
	hash := trieCtx.Compute(string(encNode))
	hashesCollector.AddDirtyHash(hash)

	// test point encodeNodeAndCommitToDB

	err = trieCtx.Put(hash, encNode)
	if err != nil {
		goRoutinesManager.SetError(err)
		return false
	}
	return true
}
