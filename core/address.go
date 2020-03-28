package core

import (
	"bytes"

	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// NumInitCharactersForScAddress numbers of characters for smart contract address identifier
const NumInitCharactersForScAddress = 10

// ShardIdentiferLen number of characters for shard identifier in an address
const ShardIdentiferLen = 2

const metaChainShardIdentifier uint8 = 255
const numInitCharactersForOnMetachainSC = 15

// IsSmartContractAddress verifies if a set address is of type smart contract
func IsSmartContractAddress(rcvAddress []byte) bool {
	if len(rcvAddress) <= NumInitCharactersForScAddress {
		return false
	}

	isEmptyAddress := bytes.Equal(rcvAddress, make([]byte, len(rcvAddress)))
	if isEmptyAddress {
		return true
	}

	numOfZeros := NumInitCharactersForScAddress - vmcommon.VMTypeLen
	isSCAddress := bytes.Equal(rcvAddress[:numOfZeros], make([]byte, numOfZeros))
	return isSCAddress
}

// IsMetachainIdentifier verifies if the identifier is of type metachain
func IsMetachainIdentifier(identifier []byte) bool {
	if len(identifier) == 0 {
		return false
	}

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

	leftSide := rcvAddress[NumInitCharactersForScAddress:(NumInitCharactersForScAddress + numInitCharactersForOnMetachainSC)]
	isOnMetaChainSCAddress := bytes.Equal(leftSide,
		make([]byte, numInitCharactersForOnMetachainSC))
	return isOnMetaChainSCAddress
}
