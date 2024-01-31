package metachain

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/multiversx/mx-chain-go/state"
)

func (als *auctionListSelector) selectNodes(
	ownersData map[string]*ownerAuctionData,
	numAvailableSlots uint32,
	randomness []byte,
) []state.ValidatorInfoHandler {
	selectedFromAuction := make([]state.ValidatorInfoHandler, 0)
	validatorTopUpMap := make(map[string]*big.Int)

	pubKeyLen := getPubKeyLen(ownersData)
	normRand := calcNormalizedRandomness(randomness, pubKeyLen)

	for _, owner := range ownersData {
		sortListByPubKey(owner.auctionList)
		addQualifiedValidatorsTopUpInMap(owner, validatorTopUpMap)
		selectedFromAuction = append(selectedFromAuction, owner.auctionList[:owner.numQualifiedAuctionNodes]...)
	}

	als.displayOwnersSelectedNodes(ownersData)
	sortValidators(selectedFromAuction, validatorTopUpMap, normRand)
	als.displayAuctionList(selectedFromAuction, ownersData, numAvailableSlots)

	return selectedFromAuction[:numAvailableSlots]
}

func getPubKeyLen(ownersData map[string]*ownerAuctionData) int {
	for _, owner := range ownersData {
		return len(owner.auctionList[0].GetPublicKey())
	}

	return 0
}

func calcNormalizedRandomness(randomness []byte, expectedLen int) []byte {
	rand := randomness
	randLen := len(rand)

	if expectedLen > randLen {
		repeatedCt := expectedLen/randLen + 1
		rand = bytes.Repeat(randomness, repeatedCt)
	}

	rand = rand[:expectedLen]
	return rand
}

func sortListByPubKey(list []state.ValidatorInfoHandler) {
	sort.SliceStable(list, func(i, j int) bool {
		pubKey1 := list[i].GetPublicKey()
		pubKey2 := list[j].GetPublicKey()

		return bytes.Compare(pubKey1, pubKey2) > 0
	})
}

func addQualifiedValidatorsTopUpInMap(owner *ownerAuctionData, validatorTopUpMap map[string]*big.Int) {
	for i := int64(0); i < owner.numQualifiedAuctionNodes; i++ {
		validatorPubKey := string(owner.auctionList[i].GetPublicKey())
		validatorTopUpMap[validatorPubKey] = big.NewInt(0).SetBytes(owner.qualifiedTopUpPerNode.Bytes())
	}
}

func sortValidators(
	list []state.ValidatorInfoHandler,
	validatorTopUpMap map[string]*big.Int,
	randomness []byte,
) {
	sort.SliceStable(list, func(i, j int) bool {
		pubKey1 := list[i].GetPublicKey()
		pubKey2 := list[j].GetPublicKey()

		nodeTopUpPubKey1 := validatorTopUpMap[string(pubKey1)]
		nodeTopUpPubKey2 := validatorTopUpMap[string(pubKey2)]

		if nodeTopUpPubKey1.Cmp(nodeTopUpPubKey2) == 0 {
			return compareByXORWithRandomness(pubKey1, pubKey2, randomness)
		}

		return nodeTopUpPubKey1.Cmp(nodeTopUpPubKey2) > 0
	})
}

func compareByXORWithRandomness(pubKey1, pubKey2, randomness []byte) bool {
	xorLen := len(randomness)

	key1Xor := make([]byte, xorLen)
	key2Xor := make([]byte, xorLen)

	for idx := 0; idx < xorLen; idx++ {
		key1Xor[idx] = pubKey1[idx] ^ randomness[idx]
		key2Xor[idx] = pubKey2[idx] ^ randomness[idx]
	}

	return bytes.Compare(key1Xor, key2Xor) == 1
}
