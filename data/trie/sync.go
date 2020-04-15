package trie

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type trieSyncer struct {
	rootFound               bool
	shardId                 uint32
	topic                   string
	rootHash                []byte
	nodeHashes              map[string]bool
	receivedNodes           map[string]node
	waitTimeBetweenRequests time.Duration
	trie                    *patriciaMerkleTrie
	requestHandler          RequestHandler
	interceptedNodes        storage.Cacher
	nodeHashesMutex         sync.Mutex
	receivedNodesMutex      sync.Mutex
}

// NewTrieSyncer creates a new instance of trieSyncer
func NewTrieSyncer(
	requestHandler RequestHandler,
	interceptedNodes storage.Cacher,
	trie data.Trie,
	shardId uint32,
	topic string,
) (*trieSyncer, error) {
	if check.IfNil(requestHandler) {
		return nil, ErrNilRequestHandler
	}
	if check.IfNil(interceptedNodes) {
		return nil, data.ErrNilCacher
	}
	if check.IfNil(trie) {
		return nil, ErrNilTrie
	}
	if len(topic) == 0 {
		return nil, ErrInvalidTrieTopic
	}

	pmt, ok := trie.(*patriciaMerkleTrie)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	ts := &trieSyncer{
		requestHandler:          requestHandler,
		interceptedNodes:        interceptedNodes,
		trie:                    pmt,
		nodeHashes:              make(map[string]bool),
		receivedNodes:           make(map[string]node),
		topic:                   topic,
		shardId:                 shardId,
		waitTimeBetweenRequests: time.Second,
	}
	ts.interceptedNodes.RegisterHandler(ts.trieNodeIntercepted)

	return ts, nil
}

// StartSyncing completes the trie, asking for missing trie nodes on the network
func (ts *trieSyncer) StartSyncing(rootHash []byte, ctx context.Context) error {
	if len(rootHash) == 0 {
		return ErrInvalidHash
	}
	if ctx == nil {
		return ErrNilContext
	}

	ts.nodeHashesMutex.Lock()
	ts.nodeHashes = make(map[string]bool)
	ts.nodeHashes[string(rootHash)] = false
	ts.nodeHashesMutex.Unlock()

	ts.rootFound = false
	ts.rootHash = rootHash

	for {
		shouldRetryAfterRequest, err := ts.checkIfSynced()
		if err != nil {
			return err
		}

		numRequested := ts.requestNodes()
		if !shouldRetryAfterRequest && numRequested == 0 {
			err := ts.trie.Commit()
			if err != nil {
				return err
			}

			return nil
		}

		select {
		case <-time.After(ts.waitTimeBetweenRequests):
			continue
		case <-ctx.Done():
			return ErrTimeIsOut
		}
	}
}

func (ts *trieSyncer) checkIfSynced() (bool, error) {
	var currentNode node
	var err error
	var nextNodes []node
	missingNodes := make([][]byte, 0)
	currentMissingNodes := make([][]byte, 0)

	newElement := true
	shouldRetryAfterRequest := false

	ts.nodeHashesMutex.Lock()
	for newElement {
		newElement = false

		for nodeHash := range ts.nodeHashes {
			currentMissingNodes = currentMissingNodes[:0]

			currentNode, err = ts.getNode([]byte(nodeHash))
			if err != nil {
				continue
			}

			if !ts.rootFound && bytes.Equal([]byte(nodeHash), ts.rootHash) {
				ts.trie.root = currentNode
				ts.rootFound = true
			}

			currentMissingNodes, nextNodes, err = currentNode.loadChildren(ts.getNode)
			if err != nil {
				ts.nodeHashesMutex.Unlock()
				return false, err
			}

			if len(currentMissingNodes) > 0 {
				missingNodes = append(missingNodes, currentMissingNodes...)
				nextNodes = append(nextNodes, currentNode)
				tmpNewElement := ts.addNew(nextNodes)
				shouldRetryAfterRequest = shouldRetryAfterRequest || tmpNewElement
				continue
			}

			nextNodes, err = currentNode.getChildren(ts.trie.Database())
			if err != nil {
				ts.nodeHashesMutex.Unlock()
				return false, err
			}

			tmpNewElement := ts.addNew(nextNodes)
			newElement = newElement || tmpNewElement

			ts.deleteResolved(nodeHash)
		}
	}

	for _, missingNode := range missingNodes {
		ts.nodeHashes[string(missingNode)] = false
	}
	ts.nodeHashesMutex.Unlock()

	return shouldRetryAfterRequest, nil
}

func (ts *trieSyncer) deleteResolved(nodeHash string) {
	ts.receivedNodesMutex.Lock()
	delete(ts.receivedNodes, nodeHash)
	ts.receivedNodesMutex.Unlock()
	delete(ts.nodeHashes, nodeHash)
}

// adds new elements to needed hash map, lock ts.nodeHashesMutex before calling
func (ts *trieSyncer) addNew(nextNodes []node) bool {
	newElement := false
	ts.receivedNodesMutex.Lock()
	for _, nextNode := range nextNodes {
		nextHash := string(nextNode.getHash())
		if _, ok := ts.nodeHashes[nextHash]; !ok {
			ts.nodeHashes[nextHash] = true
			newElement = true
		}
		ts.receivedNodes[nextHash] = nextNode
	}
	ts.receivedNodesMutex.Unlock()

	return newElement
}

// Trie returns the synced trie
func (ts *trieSyncer) Trie() data.Trie {
	return ts.trie
}

func (ts *trieSyncer) getNode(hash []byte) (node, error) {
	ts.receivedNodesMutex.Lock()
	node, ok := ts.receivedNodes[string(hash)]
	ts.receivedNodesMutex.Unlock()
	if ok {
		return node, nil
	}

	n, ok := ts.interceptedNodes.Get(hash)
	if ok {
		return trieNode(n)
	}

	return nil, ErrNodeNotFound
}

func trieNode(data interface{}) (node, error) {
	n, ok := data.(*InterceptedTrieNode)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	return n.node, nil
}

func (ts *trieSyncer) requestNodes() uint32 {
	ts.nodeHashesMutex.Lock()
	numUnResolvedNodes := uint32(len(ts.nodeHashes))
	for hash, found := range ts.nodeHashes {
		if !found {
			ts.requestHandler.RequestTrieNodes(ts.shardId, []byte(hash), ts.topic)
		}
	}
	ts.nodeHashesMutex.Unlock()

	return numUnResolvedNodes
}

func (ts *trieSyncer) trieNodeIntercepted(hash []byte, val interface{}) {
	ts.nodeHashesMutex.Lock()
	_, ok := ts.nodeHashes[string(hash)]
	ts.nodeHashesMutex.Unlock()
	if !ok {
		return
	}

	node, err := trieNode(val)
	if err != nil {
		return
	}

	ts.receivedNodesMutex.Lock()
	ts.receivedNodes[string(hash)] = node
	ts.receivedNodesMutex.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ts *trieSyncer) IsInterfaceNil() bool {
	return ts == nil
}
