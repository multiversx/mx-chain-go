// Copyright 2015 The go-ethereum Authors
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

package state

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/stretchr/testify/assert"
)

var testAddrHasher = &mock.HasherMock{}

func TestAddress_FromPubKeyBytes_LowNoOfBytes_ShouldErr(t *testing.T) {
	buff := []byte("ABCDEF")

	_, err := FromPubKeyBytes(buff)
	assert.NotNil(t, err)
	fmt.Println(err.Error())
}

func TestAddress_FromPubKeyBytes_Values_ShouldWork(t *testing.T) {
	buff := []byte("ABCDEFGHIJKLMNOPQRSTUVXYZ124567890")

	adr, err := FromPubKeyBytes(buff)
	assert.Nil(t, err)

	fmt.Println(adr.Hex(testAddrHasher))
}

func TestAddress_Hash_Values_ShouldWork(t *testing.T) {
	buff := []byte("ABCDEFGHIJKLMNOPQRSTUVXYZ124567890")

	adr, err := FromPubKeyBytes(buff)
	assert.Nil(t, err)

	assert.Equal(t, adr.Hash(testAddrHasher), testAddrHasher.Compute(string(adr.Bytes())))
}

func TestAddress_IsHexAddress_Values_ShouldWork(t *testing.T) {
	buff := []byte("ABCDEFGHIJKLMNOPQRSTUVXYZ124567890")

	adr, err := FromPubKeyBytes(buff)
	assert.Nil(t, err)

	assert.True(t, IsHexAddress(adr.Hex(testAddrHasher)))
	fmt.Printf("Address: %v\n", adr.Hex(testAddrHasher))
}

func TestAddress_IsHexAddress_BadAddrs_ShouldRetFalse(t *testing.T) {
	//invalid characters
	assert.False(t, IsHexAddress("ABCDEFGH"))
	//odd numbers of hexa chars
	assert.False(t, IsHexAddress("0x434445464748494A4B4c4d4e4F5051525354555658595A31323435363738393"))
}

func TestAddress_HexToAddress_Values_ShouldWork(t *testing.T) {
	buff := []byte("ABCDEFGHIJKLMNOPQRSTUVXYZ124567890")

	adr, err := FromPubKeyBytes(buff)
	assert.Nil(t, err)

	adr2 := HexToAddress(adr.Hex(testAddrHasher))

	assert.Equal(t, adr, adr2)
}

func TestIsHexAddress(t *testing.T) {
	if AdrLen != 20 {
		t.Skip("Test not valid on a different address length of 20!")
	}

	tests := []struct {
		str string
		exp bool
	}{
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed", true},
		{"5aaeb6053f3e94c9b9a09f33669435e7ef1beaed", true},
		{"0X5aaeb6053f3e94c9b9a09f33669435e7ef1beaed", true},
		{"0XAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", true},
		{"0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", true},
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed1", false},
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beae", false},
		{"5aaeb6053f3e94c9b9a09f33669435e7ef1beaed11", false},
		{"0xxaaeb6053f3e94c9b9a09f33669435e7ef1beaed", false},
		{"0xxaaeb6053f3e94c9b9a09f33669435e7ef1beaedd", false},
	}

	for _, test := range tests {
		result := IsHexAddress(test.str)
		assert.Equal(t, test.exp, result)
	}
}

func TestAddressHexChecksum(t *testing.T) {
	if AdrLen != 20 {
		t.Skip("Test not valid on a different address length of 20!")
	}

	var tests = []struct {
		Input  string
		Output string
	}{
		{"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed", "0x5AAEb6053f3e94c9B9A09F33669435e7ef1BEaEd"},
		{"0xfb6916095ca1df60bb79ce92ce3ea74c37c5d359", "0xFB6916095cA1dF60bB79Ce92cE3EA74c37C5d359"},
		{"0xdbf03b407c01e7cd3cbea99509d93f8dddc8c6fb", "0xdbf03b407c01e7cd3CBEA99509D93F8DddC8C6FB"},
		{"0xd1220a0cf47c7b9be7a2e6ba89f429762e7b9adb", "0xd1220A0cf47c7B9BE7a2e6Ba89F429762E7B9Adb"},
		// Ensure that non-standard length input values are handled correctly
		{"0xa", "0x000000000000000000000000000000000000000A"},
		{"0x0a", "0x000000000000000000000000000000000000000A"},
		{"0x00a", "0x000000000000000000000000000000000000000A"},
		{"0x000000000000000000000000000000000000000a", "0x000000000000000000000000000000000000000A"},
	}
	for i, test := range tests {
		output := HexToAddress(test.Input).Hex(mock.HasherMock{})
		if output != test.Output {
			t.Errorf("test #%d: failed to match when it should (%s != %s)", i, output, test.Output)
		}
	}
}

func TestAddressFromPubKey(t *testing.T) {
	if AdrLen != 20 {
		t.Skip("Test not valid on a different address length of 20!")
	}

	//test error
	_, err := FromPubKeyBytes([]byte{45, 56})

	switch e := err.(type) {
	case *ErrorWrongSize:
		fmt.Println(e.Error())
		break
	default:
		assert.Fail(t, "Should have errored")
	}

	//test trim
	buff := []byte("ABCDEFGHIJKLMNOPQRSTUVXYZ")

	adr, err := FromPubKeyBytes(buff)
	assert.Nil(t, err)

	assert.Equal(t, []byte("FGHIJKLMNOPQRSTUVXYZ"), adr.Bytes())

}
