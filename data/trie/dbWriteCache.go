// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package trie

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/trie/encoding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// DBWriteCache is an intermediate write layer between the trie data structures and
// the disk database. The aim is to accumulate trie writes in-memory and only
// periodically flush a couple tries to disk, garbage collecting the remainder.
//TODO check if it is really needed batching writes
type DBWriteCache struct {
	storer storage.Storer // Persistent storage for matured trie nodes

	nodes  map[encoding.Hash]*cachedNode // Data and references relationships of a node
	oldest encoding.Hash                 // Oldest tracked node, flush-list head
	newest encoding.Hash                 // Newest tracked node, flush-list tail

	preimages map[encoding.Hash][]byte       // Preimages of nodes from the secure trie
	seckeybuf [encoding.SecureKeyLength]byte // Ephemeral buffer for calculating preimage keys

	gctime  time.Duration        // Time spent on garbage collection since last commit
	gcnodes uint64               // Nodes garbage collected since last commit
	gcsize  encoding.StorageSize // Data storage garbage collected since last commit

	flushtime  time.Duration        // Time spent on data flushing since last commit
	flushnodes uint64               // Nodes flushed since last commit
	flushsize  encoding.StorageSize // Data storage flushed since last commit

	nodesSize     encoding.StorageSize // Storage size of the nodes cache (exc. flushlist)
	preimagesSize encoding.StorageSize // Storage size of the preimages cache

	lock sync.RWMutex
}

// NewDBWriteCache creates a new trie database to store ephemeral trie content before
// its written out to disk or garbage collected.
func NewDBWriteCache(storer storage.Storer) (*DBWriteCache, error) {
	if storer == nil {
		return nil, errors.New("nil storer")
	}

	return &DBWriteCache{
		storer:    storer,
		nodes:     map[encoding.Hash]*cachedNode{{}: {}},
		preimages: make(map[encoding.Hash][]byte),
	}, nil
}

// Storer retrieves the persistent storage backing the trie database.
func (db *DBWriteCache) Storer() storage.Storer {
	return db.storer
}

// InsertBlob writes a new reference tracked blob to the memory database if it's
// yet unknown. This method should only be used for non-trie nodes that require
// reference counting, since trie nodes are garbage collected directly through
// their embedded children.
func (db *DBWriteCache) InsertBlob(hash []byte, blob []byte) {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.insert(encoding.BytesToHash(hash), blob, rawNode(blob))
}

// insert inserts a collapsed trie node into the memory database. This method is
// a more generic version of InsertBlob, supporting both raw blob insertions as
// well ex trie node insertions. The blob must always be specified to allow proper
// size tracking.
func (db *DBWriteCache) insert(hash encoding.Hash, blob []byte, node Node) {
	// If the node's already cached, skip
	if _, ok := db.nodes[hash]; ok {
		return
	}
	// Create the cached entry for this node
	entry := &cachedNode{
		node:      simplifyNode(node),
		size:      uint16(len(blob)),
		flushPrev: db.newest,
	}
	for _, child := range entry.childs() {
		if c := db.nodes[child]; c != nil {
			c.parents++
		}
	}
	db.nodes[hash] = entry

	// Update the flush-list endpoints
	if db.oldest == (encoding.Hash{}) {
		db.oldest, db.newest = hash, hash
	} else {
		db.nodes[db.newest].flushNext, db.newest = hash, hash
	}
	db.nodesSize += encoding.StorageSize(encoding.HashLength + entry.size)
}

// insertPreimage writes a new trie node pre-image to the memory database if it's
// yet unknown. The method will make a copy of the slice.
//
// Note, this method assumes that the database's lock is held!
func (db *DBWriteCache) insertPreimage(hash encoding.Hash, preimage []byte) {
	if _, ok := db.preimages[hash]; ok {
		return
	}
	db.preimages[hash] = encoding.CopyBytes(preimage)
	db.preimagesSize += encoding.StorageSize(encoding.HashLength + len(preimage))
}

// CachedNode retrieves a cached trie node from memory, or returns nil if none can be
// found in the memory cache.
func (db *DBWriteCache) CachedNode(hash []byte, cachegen uint16) Node {
	// Retrieve the node from cache if available
	h := encoding.BytesToHash(hash)

	db.lock.RLock()
	node := db.nodes[h]
	db.lock.RUnlock()

	if node != nil {
		return node.obj(h, cachegen)
	}
	// Content unavailable in memory, attempt to retrieve from disk
	enc, err := db.storer.Get(h[:])
	if err != nil || enc == nil {
		return nil
	}
	return mustDecodeNode(h[:], enc, cachegen)
}

// Node retrieves an encoded cached trie node from memory. If it cannot be found
// cached, the method queries the persistent database for the content.
func (db *DBWriteCache) Node(hash []byte) ([]byte, error) {
	// Retrieve the node from cache if available
	db.lock.RLock()
	node := db.nodes[encoding.BytesToHash(hash)]
	db.lock.RUnlock()

	if node != nil {
		return node.rlp(), nil
	}
	// Content unavailable in memory, attempt to retrieve from disk
	return db.storer.Get(hash[:])
}

// preimage retrieves a cached trie node pre-image from memory. If it cannot be
// found cached, the method queries the persistent database for the content.
func (db *DBWriteCache) preimage(hash encoding.Hash) ([]byte, error) {
	// Retrieve the node from cache if available
	db.lock.RLock()
	preimage := db.preimages[hash]
	db.lock.RUnlock()

	if preimage != nil {
		return preimage, nil
	}
	// Content unavailable in memory, attempt to retrieve from disk
	return db.storer.Get(db.secureKey(hash[:]))
}

// secureKey returns the database key for the preimage of key, as an ephemeral
// buffer. The caller must not hold onto the return value because it will become
// invalid on the next call.
func (db *DBWriteCache) secureKey(key []byte) []byte {
	buf := append(db.seckeybuf[:0], encoding.SecureKeyPrefix...)
	buf = append(buf, key...)
	return buf
}

// Nodes retrieves the hashes of all the nodes cached within the memory database.
// This method is extremely expensive and should only be used to validate internal
// states in test code.
func (db *DBWriteCache) nodesUnused() []encoding.Hash {
	db.lock.RLock()
	defer db.lock.RUnlock()

	var hashes = make([]encoding.Hash, 0, len(db.nodes))
	for hash := range db.nodes {
		if hash != (encoding.Hash{}) { // Special case for "root" references/nodes
			hashes = append(hashes, hash)
		}
	}

	return hashes
}

// Reference adds a new reference from a parent node to a child node.
func (db *DBWriteCache) Reference(child []byte, parent []byte) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	db.reference(encoding.BytesToHash(child), encoding.BytesToHash(parent))
}

// reference is the private locked version of Reference.
func (db *DBWriteCache) reference(child encoding.Hash, parent encoding.Hash) {
	// If the node does not exist, it's a node pulled from disk, skip
	node, ok := db.nodes[child]
	if !ok {
		return
	}
	// If the reference already exists, only duplicate for roots
	if db.nodes[parent].children == nil {
		db.nodes[parent].children = make(map[encoding.Hash]uint16)
	} else if _, ok = db.nodes[parent].children[child]; ok && parent != (encoding.Hash{}) {
		return
	}
	node.parents++
	db.nodes[parent].children[child]++
}

// Dereference removes an existing reference from a root node.
func (db *DBWriteCache) Dereference(root []byte) {
	// Sanity check to ensure that the meta-root is not removed
	rHash := encoding.BytesToHash(root)
	if rHash == (encoding.Hash{}) {
		//log.Error("Attempted to dereference the trie cache meta root")
		return
	}
	db.lock.Lock()
	defer db.lock.Unlock()

	nodes, storeSize, start := len(db.nodes), db.nodesSize, time.Now()
	db.dereference(rHash, encoding.Hash{})

	db.gcnodes += uint64(nodes - len(db.nodes))
	db.gcsize += storeSize - db.nodesSize
	db.gctime += time.Since(start)
}

// dereference is the private locked version of Dereference.
func (db *DBWriteCache) dereference(child encoding.Hash, parent encoding.Hash) {
	// Dereference the parent-child
	node := db.nodes[parent]

	if node.children != nil && node.children[child] > 0 {
		node.children[child]--
		if node.children[child] == 0 {
			delete(node.children, child)
		}
	}
	// If the child does not exist, it's a previously committed node.
	node, ok := db.nodes[child]
	if !ok {
		return
	}
	// If there are no more references to the child, delete it and cascade
	if node.parents > 0 {
		// This is a special cornercase where a node loaded from disk (i.e. not in the
		// memcache any more) gets reinjected as a new node (short node split into full,
		// then reverted into short), causing a cached node to have no parents. That is
		// no problem in itself, but don't make maxint parents out of it.
		node.parents--
	}
	if node.parents == 0 {
		// Remove the node from the flush-list
		switch child {
		case db.oldest:
			db.oldest = node.flushNext
			db.nodes[node.flushNext].flushPrev = encoding.Hash{}
		case db.newest:
			db.newest = node.flushPrev
			db.nodes[node.flushPrev].flushNext = encoding.Hash{}
		default:
			db.nodes[node.flushPrev].flushNext = node.flushNext
			db.nodes[node.flushNext].flushPrev = node.flushPrev
		}
		// Dereference all children and delete the node
		for _, hash := range node.childs() {
			db.dereference(hash, child)
		}
		delete(db.nodes, child)
		db.nodesSize -= encoding.StorageSize(encoding.HashLength + int(node.size))
	}
}

// Cap iteratively flushes old but still referenced trie nodes until the total
// memory usage goes below the given threshold.
func (db *DBWriteCache) Cap(limit float64) error {
	// Create a database batch to flush persistent data out. It is important that
	// outside code doesn't see an inconsistent state (referenced data removed from
	// memory cache during commit but not yet in persistent storage). This is ensured
	// by only uncaching existing data when the database write finalizes.
	db.lock.RLock()

	nodes, storeSize, start := len(db.nodes), db.nodesSize, time.Now()

	// db.nodesSize only contains the useful data in the cache, but when reporting
	// the total memory consumption, the maintenance metadata is also needed to be
	// counted. For every useful node, we track 2 extra hashes as the flushlist.
	size := db.nodesSize + encoding.StorageSize((len(db.nodes)-1)*2*encoding.HashLength)

	for hash, preimage := range db.preimages {
		if err := db.storer.Put(db.secureKey(hash[:]), preimage); err != nil {
			db.lock.RUnlock()
			return err
		}
	}
	// Keep committing nodes from the flush-list until we're below allowance
	oldest := db.oldest

	for size > encoding.StorageSize(limit) && oldest != (encoding.Hash{}) {
		// Fetch the oldest referenced node and push into the batch
		node := db.nodes[oldest]
		if err := db.storer.Put(oldest[:], node.rlp()); err != nil {
			db.lock.RUnlock()
			return err
		}
		// Iterate to the next flush item, or abort if the size cap was achieved. Size
		// is the total size, including both the useful cached data (hash -> blob), as
		// well as the flushlist metadata (2*hash). When flushing items from the cache,
		// we need to reduce both.
		size -= encoding.StorageSize(3*encoding.HashLength + int(node.size))
		oldest = node.flushNext
	}

	db.lock.RUnlock()

	// Write successful, clear out the flushed data
	db.lock.Lock()
	defer db.lock.Unlock()

	db.preimages = make(map[encoding.Hash][]byte)
	db.preimagesSize = 0

	for db.oldest != oldest {
		node := db.nodes[db.oldest]
		delete(db.nodes, db.oldest)
		db.oldest = node.flushNext

		db.nodesSize -= encoding.StorageSize(encoding.HashLength + int(node.size))
	}
	if db.oldest != (encoding.Hash{}) {
		db.nodes[db.oldest].flushPrev = encoding.Hash{}
	}
	db.flushnodes += uint64(nodes - len(db.nodes))
	db.flushsize += storeSize - db.nodesSize
	db.flushtime += time.Since(start)

	return nil
}

// Commit iterates over all the children of a particular node, writes them out
// to disk, forcefully tearing down all references in both directions.
//
// As a side effect, all pre-images accumulated up to this point are also written.
func (db *DBWriteCache) Commit(node []byte, report bool) error {
	isEmpty := true
	for _, b := range node {
		if b != 0 {
			isEmpty = false
			break
		}
	}
	if isEmpty {
		return nil
	}

	// Create a database batch to flush persistent data out. It is important that
	// outside code doesn't see an inconsistent state (referenced data removed from
	// memory cache during commit but not yet in persistent storage). This is ensured
	// by only uncaching existing data when the database write finalizes.
	db.lock.RLock()

	// Move all of the accumulated preimages into a write batch
	for hash, preimage := range db.preimages {
		if err := db.storer.Put(db.secureKey(hash[:]), preimage); err != nil {
			db.lock.RUnlock()
			return err
		}
	}
	// Write the trie itself
	if err := db.commit(encoding.BytesToHash(node)); err != nil {
		//log.Error("Failed to commit trie from trie database", "err", err)
		db.lock.RUnlock()
		return err
	}
	db.lock.RUnlock()

	// Write successful, clear out the flushed data
	db.lock.Lock()
	defer db.lock.Unlock()

	db.preimages = make(map[encoding.Hash][]byte)
	db.preimagesSize = 0

	db.uncache(encoding.BytesToHash(node))

	// Reset the garbage collection statistics
	db.gcnodes, db.gcsize, db.gctime = 0, 0, 0
	db.flushnodes, db.flushsize, db.flushtime = 0, 0, 0

	return nil
}

// commit is the private locked version of Commit.
func (db *DBWriteCache) commit(hash encoding.Hash) error {
	// If the node does not exist, it's a previously committed node
	node, ok := db.nodes[hash]
	if !ok {
		return nil
	}
	for _, child := range node.childs() {
		if err := db.commit(child); err != nil {
			return err
		}
	}

	if err := db.storer.Put(hash[:], node.rlp()); err != nil {
		return err
	}

	return nil
}

// uncache is the post-processing step of a commit operation where the already
// persisted trie is removed from the cache. The reason behind the two-phase
// commit is to ensure consistent data availability while moving from memory
// to disk.
func (db *DBWriteCache) uncache(hash encoding.Hash) {
	// If the node does not exist, we're done on this path
	node, ok := db.nodes[hash]
	if !ok {
		return
	}
	// Node still exists, remove it from the flush-list
	switch hash {
	case db.oldest:
		db.oldest = node.flushNext
		db.nodes[node.flushNext].flushPrev = encoding.Hash{}
	case db.newest:
		db.newest = node.flushPrev
		db.nodes[node.flushPrev].flushNext = encoding.Hash{}
	default:
		db.nodes[node.flushPrev].flushNext = node.flushNext
		db.nodes[node.flushNext].flushPrev = node.flushPrev
	}
	// Uncache the node's subtries and remove the node itself too
	for _, child := range node.childs() {
		db.uncache(child)
	}
	delete(db.nodes, hash)
	db.nodesSize -= encoding.StorageSize(encoding.HashLength + int(node.size))
}

// Size returns the current storage size of the memory cache in front of the
// persistent database layer.
func (db *DBWriteCache) Size() (float64, float64) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	// db.nodesSize only contains the useful data in the cache, but when reporting
	// the total memory consumption, the maintenance metadata is also needed to be
	// counted. For every useful node, we track 2 extra hashes as the flushlist.
	var flushlistSize = encoding.StorageSize((len(db.nodes) - 1) * 2 * encoding.HashLength)
	return float64(db.nodesSize + flushlistSize), float64(db.preimagesSize)
}

// verifyIntegrity is a debug method to iterate over the entire trie stored in
// memory and check whether every node is reachable from the meta root. The goal
// is to find any errors that might cause memory leaks and or trie nodes to go
// missing.
//
// This method is extremely CPU and memory intensive, only use when must.
func (db *DBWriteCache) verifyIntegrity() {
	// Iterate over all the cached nodes and accumulate them into a set
	reachable := map[encoding.Hash]struct{}{{}: {}}

	for child := range db.nodes[encoding.Hash{}].children {
		db.accumulate(child, reachable)
	}
	// Find any unreachable but cached nodes
	unreachable := make([]string, 0)
	for hash, node := range db.nodes {
		if _, ok := reachable[hash]; !ok {
			unreachable = append(unreachable, fmt.Sprintf("%x: {Node: %v, Parents: %d, Prev: %x, Next: %x}",
				hash, node.node, node.parents, node.flushPrev, node.flushNext))
		}
	}
	if len(unreachable) != 0 {
		panic(fmt.Sprintf("trie cache memory leak: %v", unreachable))
	}
}

// accumulate iterates over the trie defined by hash and accumulates all the
// cached children found in memory.
func (db *DBWriteCache) accumulate(hash encoding.Hash, reachable map[encoding.Hash]struct{}) {
	// Mark the node reachable if present in the memory cache
	node, ok := db.nodes[hash]
	if !ok {
		return
	}
	reachable[hash] = struct{}{}

	// Iterate over all the children and accumulate them too
	for _, child := range node.childs() {
		db.accumulate(child, reachable)
	}
}

// InsertWithLock inserts data but is concurrent safe
func (db *DBWriteCache) InsertWithLock(hash []byte, blob []byte, node Node) {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.insert(encoding.BytesToHash(hash), blob, node)
}
