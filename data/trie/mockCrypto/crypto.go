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

package mockCrypto

import (
	"errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/mockCrypto/sha3"
	"math/big"
)

var (
	secp256k1N, _  = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	secp256k1halfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
)

var errInvalidPubkey = errors.New("invalid secp256k1 public key")

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

//
//
//// Keccak256 calculates and returns the Keccak256 hash of the input data.
//func Keccak256(data ...[]byte) []byte {
//	d := sha3.NewKeccak256()
//	for _, b := range data {
//		d.Write(b)
//	}
//	return d.Sum(nil)
//}
//
//// blake2b calculates and returns the blake2b hash of the input data.
//func Blake2b(data ...[]byte) []byte {
//	b2b := blake2b.Blake2b{}
//
//	buff := make([]byte, 0)
//
//	for _, b := range data {
//		buff = append(buff, b...)
//	}
//
//	return b2b.Compute(string(buff))
//}
//
//// sha256 calculates and returns the sha256 hash of the input data.
//func Sha256(data ...[]byte) []byte {
//	sha := sha256.Sha256{}
//
//	buff := make([]byte, 0)
//
//	for _, b := range data {
//		buff = append(buff, b...)
//	}
//
//	return sha.Compute(string(buff))
//}
//
//// Keccak256Hash calculates and returns the Keccak256 hash of the input data,
//// converting it to an internal Hash data structure.
//func Keccak256Hash(data ...[]byte) (h encoding.Hash) {
//	d := sha3.NewKeccak256()
//	for _, b := range data {
//		d.Write(b)
//	}
//	d.Sum(h[:0])
//	return h
//}
//
//// Blake2bHash calculates and returns the Blacke2b hash of the input data,
//// converting it to an internal Hash data structure.
//func Blake2bHash(data ...[]byte) (h encoding.Hash) {
//	b2b := blake2b.Blake2b{}
//
//	buff := make([]byte, 0)
//
//	for _, b := range data {
//		buff = append(buff, b...)
//	}
//
//	h.SetBytes(b2b.Compute(string(buff)))
//	return h
//}
//
//// Sha256 calculates and returns the Sha256 hash of the input data,
//// converting it to an internal Hash data structure.
//func Sha256Hash(data ...[]byte) (h encoding.Hash) {
//	sha256 := sha256.Sha256{}
//
//	buff := make([]byte, 0)
//
//	for _, b := range data {
//		buff = append(buff, b...)
//	}
//
//	h.SetBytes(sha256.Compute(string(buff)))
//	return h
//}
