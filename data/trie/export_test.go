package trie

import "github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"

type Node interface {
	node
}

type NodeFlag struct {
	nodeFlag
}

func (nf *NodeFlag) Hash() *hashNode {
	return &nf.hash
}

func (nf *NodeFlag) Gen() uint16 {
	return nf.gen
}

func (nf *NodeFlag) Dirty() bool {
	return nf.dirty
}

type FullNode struct {
	fullNode
}

func (fn *FullNode) ChildrenNodes() []Node {
	nodes := make([]Node, 0)

	for i := 0; i < len(fn.Children); i++ {
		nodes = append(nodes, fn.Children[i])
	}

	return nodes
}

func (fn *FullNode) Flags() NodeFlag {
	nf := NodeFlag{fn.flags}
	return nf
}

type ShortNode struct {
	shortNode
}

func (sn *ShortNode) Flags() NodeFlag {
	nf := NodeFlag{sn.flags}
	return nf
}

//Database export the unexported fields needed in tests

func (db *Database) NodesTest() map[encoding.Hash]*cachedNode {
	return db.nodes
}

//Trie export the unexported fields needed in tests

func (t *Trie) DB() *Database {
	return t.db
}

func (t *Trie) RootNode() Node {
	return t.root
}

func (t *Trie) OriginalRoot() encoding.Hash {
	return t.originalRoot
}

func (t *Trie) Cachegen() uint16 {
	return t.cachegen
}

func (t *Trie) Cachelimit() uint16 {
	return t.cachelimit
}
