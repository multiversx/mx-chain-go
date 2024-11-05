package trie

import "sync"

type modifiedChild struct {
	position int
	child    node
}
type modifiedChildren struct {
	children []*modifiedChild
	mutex    *sync.Mutex
}

// NewModifiedChildren creates a new instance of modifiedChildren
func NewModifiedChildren() *modifiedChildren {
	return &modifiedChildren{
		children: make([]*modifiedChild, 0),
		mutex:    &sync.Mutex{},
	}
}

// AddChild adds a new child to the list
func (mc *modifiedChildren) AddChild(position int, child node) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.children = append(mc.children, &modifiedChild{position, child})
}

// Range iterates over the children
func (mc *modifiedChildren) Range(f func(position int, child node)) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	for _, child := range mc.children {
		f(child.position, child.child)
	}
}
