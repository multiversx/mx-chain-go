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

// TODO: remove this file once we completely migrate to Elrond Trie implementation
// crypto_test.go needs to be removed as well as the sha3 folder
package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/mock/sha3"
)

type Keccak256 struct {
}

func (*Keccak256) Compute(input string) []byte {
	d := sha3.NewKeccak256()
	d.Write([]byte(input))
	return d.Sum(nil)
}

func (k *Keccak256) EmptyHash() []byte {
	return k.Compute("")
}

func (*Keccak256) Size() int {
	return 32
}
