package tree

import "github.com/ElrondNetwork/elrond-go-core/core/check"

type node struct {
	interval *interval
	left     *node
	right    *node
	max      uint64
}

func newNode(interval *interval) *node {
	return &node{
		interval: interval,
	}
}

func (node *node) contains(value uint64) bool {
	return node.interval.contains(value)
}

func (node *node) low() uint64 {
	return node.interval.low
}

func (node *node) high() uint64 {
	return node.interval.high
}

func (node *node) isLeaf() bool {
	return check.IfNil(node.left) && check.IfNil(node.right)
}

// IsInterfaceNil returns true if there is no value under the interface
func (node *node) IsInterfaceNil() bool {
	return node == nil
}
