package trie2

import (
	"fmt"
	"testing"
)

func TestNewTNodeIterator(t *testing.T) {
	tr := NewTrie()

	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("dogglesworth"), []byte("cat"))

	it := newNodeIterator(tr, nil)

	for it.Next(true) {
		fmt.Println(it.path)
	}
}
