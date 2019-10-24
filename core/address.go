package core

import (
	"bytes"
)

// NumInitCharactersForScAddress numbers of characters for smart contract address identifier
const NumInitCharactersForScAddress = 10

// VMTypeLen number of characters with VMType identifier in an address, these are the last 2 characters from the
// initial identifier
const VMTypeLen = 2

// ShardIdentiferLen number of characters for shard identifier in an address
const ShardIdentiferLen = 2

const metaChainShardIdentifier uint8 = 255
const numInitCharactersForOnMetachainSC = 5

// IsSmartContractAddress verifies if a set address is of type smart contract
func IsSmartContractAddress(rcvAddress []byte) bool {
	if len(rcvAddress) <= NumInitCharactersForScAddress {
		return false
	}

	isEmptyAddress := bytes.Equal(rcvAddress, make([]byte, len(rcvAddress)))
	if isEmptyAddress {
		return true
	}

	isSCAddress := bytes.Equal(rcvAddress[:(NumInitCharactersForScAddress-VMTypeLen)],
		make([]byte, NumInitCharactersForScAddress-VMTypeLen))
	if isSCAddress {
		return true
	}

	return false
}

// IsMetachainIdentifier verifies if the identifier is of type metachain
func IsMetachainIdentifier(identifier []byte) bool {
	for i := 0; i < len(identifier); i++ {
		if identifier[i] != metaChainShardIdentifier {
			return false
		}
	}

	return true
}

// IsSmartContractOnMetachain verifies if an address is smart contract on metachain
func IsSmartContractOnMetachain(identifier []byte, rcvAddress []byte) bool {
	if len(rcvAddress) <= NumInitCharactersForScAddress+numInitCharactersForOnMetachainSC {
		return false
	}

	if !IsMetachainIdentifier(identifier) {
		return false
	}

	if !IsSmartContractAddress(rcvAddress) {
		return false
	}

	isOnMetaChainSCAddress := bytes.Equal(rcvAddress[NumInitCharactersForScAddress:(NumInitCharactersForScAddress+numInitCharactersForOnMetachainSC)],
		make([]byte, numInitCharactersForOnMetachainSC))
	return isOnMetaChainSCAddress
}
