// Copyright 2014 The go-ethereum Authors
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
package trie_try01

import "fmt"

var indices = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "[17]"}

type fullNode struct {
	Children [17]dataNode // Actual trie node data to encode/decode (needs custom encoder)
	flags    nodeFlag
}

type shortNode struct {
	Key   []byte
	Val   dataNode
	flags nodeFlag
}

type hashNode []byte
type valueNode []byte

func (n *fullNode) copy() *fullNode   { cpy := *n; return &cpy }
func (n *shortNode) copy() *shortNode { cpy := *n; return &cpy }

func (n *fullNode) canUnload(gen, limit uint16) bool  { return n.flags.canUnload(gen, limit) }
func (n *shortNode) canUnload(gen, limit uint16) bool { return n.flags.canUnload(gen, limit) }
func (n hashNode) canUnload(uint16, uint16) bool      { return false }
func (n valueNode) canUnload(uint16, uint16) bool     { return false }

func (n *fullNode) cache() ([]byte, bool)  { return n.flags.hash, n.flags.dirty }
func (n *shortNode) cache() ([]byte, bool) { return n.flags.hash, n.flags.dirty }
func (n hashNode) cache() ([]byte, bool)   { return nil, true }
func (n valueNode) cache() ([]byte, bool)  { return nil, true }

func (n *fullNode) String() string  { return n.fstring("") }
func (n *shortNode) String() string { return n.fstring("") }
func (n hashNode) String() string   { return n.fstring("") }
func (n valueNode) String() string  { return n.fstring("") }

func (n *fullNode) fstring(ind string) string {
	resp := fmt.Sprintf("[\n%s  ", ind)
	for i, node := range &n.Children {
		if node == nil {
			resp += fmt.Sprintf("%s: <nil> ", indices[i])
		} else {
			resp += fmt.Sprintf("%s: %v", indices[i], node.fstring(ind+"  "))
		}
	}
	return resp + fmt.Sprintf("\n%s] ", ind)
}

func (n *shortNode) fstring(ind string) string {
	return fmt.Sprintf("{%x: %v} ", n.Key, n.Val.fstring(ind+"  "))
}

func (n hashNode) fstring(ind string) string {
	return fmt.Sprintf("<%x> ", []byte(n))
}

func (n valueNode) fstring(ind string) string {
	return fmt.Sprintf("%x ", []byte(n))
}

//func (n *fullNode) Encode() ([]byte, error) {
//	var nodes [17]dataNode
//
//	for i, child := range &n.Children {
//		if child != nil {
//			nodes[i] = child
//		} else {
//			nodes[i] = nilValueNode
//		}
//	}
//	return rlp.Encode(w, nodes)
//}
