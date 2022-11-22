package tree

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/config"
)

type intervalTree struct {
	root *node
}

// NewIntervalTree returns a new instance of interval tree
func NewIntervalTree(cfg config.HardforkV2Config) *intervalTree {
	return createTree(cfg)
}

// Contains returns true if the provided value is part of any node
func (tree *intervalTree) Contains(value uint64) bool {
	return contains(tree.root, value)
}

func createTree(cfg config.HardforkV2Config) *intervalTree {
	intervals := make([]*interval, 0, len(cfg.BlocksExceptionsByRound))
	for _, blockExceptionInterval := range cfg.BlocksExceptionsByRound {
		intervals = append(intervals, newInterval(blockExceptionInterval.Low, blockExceptionInterval.High))
	}

	tree := &intervalTree{}
	for _, i := range intervals {
		n := newNode(i)
		tree.addNode(n)
	}
	computeMaxFields(tree.root)
	return tree
}

func (tree *intervalTree) addNode(n *node) {
	if check.IfNil(tree.root) {
		tree.root = n
		return
	}

	root := tree.root
	for !check.IfNil(root) {
		if n.low() <= root.low() {
			if check.IfNil(root.left) {
				root.left = n
				return
			}
			root = root.left
		} else {
			if check.IfNil(root.right) {
				root.right = n
				return
			}
			root = root.right
		}
	}
}

func contains(root *node, value uint64) bool {
	if check.IfNil(root) {
		return false
	}

	if root.contains(value) {
		return true
	}

	if !check.IfNil(root.left) {
		if root.left.max >= value {
			return contains(root.left, value)
		}
	}

	return contains(root.right, value)
}

func computeMaxFields(root *node) {
	if check.IfNil(root) {
		return
	}

	if root.isLeaf() {
		root.max = root.high()
		return
	}

	maxLeft := uint64(0)
	maxRight := uint64(0)
	if !check.IfNil(root.left) {
		computeMaxFields(root.left)
		maxLeft = root.left.max
	}
	if !check.IfNil(root.right) {
		computeMaxFields(root.right)
		maxRight = root.right.max
	}

	root.max = max(root.high(), max(maxLeft, maxRight))
}

func max(a, b uint64) uint64 {
	if a >= b {
		return a
	}
	return b
}

// IsInterfaceNil returns true if there is no value under the interface
func (tree *intervalTree) IsInterfaceNil() bool {
	return tree == nil
}
