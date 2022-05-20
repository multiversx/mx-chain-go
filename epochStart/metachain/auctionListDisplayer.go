package metachain

import (
	"fmt"
	"strconv"

	"github.com/ElrondNetwork/elrond-go-core/display"
	"github.com/ElrondNetwork/elrond-go/state"
)

const maxPubKeyDisplayableLen = 20

func getShortDisplayableBlsKeys(list []state.ValidatorInfoHandler) string {
	pubKeys := ""

	for idx, validator := range list {
		pubKey := validator.GetPublicKey()
		displayablePubKey := pubKey

		pubKeyLen := len(pubKey)
		if pubKeyLen > maxPubKeyDisplayableLen {
			displayablePubKey = make([]byte, 0)
			displayablePubKey = append(displayablePubKey, pubKey[:maxPubKeyDisplayableLen/2]...)
			displayablePubKey = append(displayablePubKey, []byte("...")...)
			displayablePubKey = append(displayablePubKey, pubKey[pubKeyLen-maxPubKeyDisplayableLen/2:]...)
		}

		pubKeys += string(displayablePubKey) // todo: hex here
		addDelimiter := idx != len(list)-1
		if addDelimiter {
			pubKeys += ", "
		}
	}

	return pubKeys
}

func (als *auctionListSelector) displayOwnersConfig(ownersData map[string]*ownerData) {
	//if log.GetLevel() > logger.LogDebug {
	//	return
	//}

	tableHeader := []string{
		"Owner",
		"Num active nodes",
		"Num auction nodes",
		"Num staked nodes",
		"Total top up",
		"Top up per node",
		"Auction list nodes",
	}
	lines := make([]*display.LineData, 0, len(ownersData))
	for ownerPubKey, owner := range ownersData {

		line := []string{
			(ownerPubKey),
			strconv.Itoa(int(owner.numActiveNodes)),
			strconv.Itoa(int(owner.numAuctionNodes)),
			strconv.Itoa(int(owner.numStakedNodes)),
			owner.totalTopUp.String(),
			owner.topUpPerNode.String(),
			getShortDisplayableBlsKeys(owner.auctionList),
		}
		lines = append(lines, display.NewLineData(false, line))
	}

	table, err := display.CreateTableString(tableHeader, lines)
	if err != nil {
		log.Error("could not create table", "error", err)
		return
	}

	message := fmt.Sprintf("Nodes config in auction list\n%s", table)
	log.Info(message)
}

func (als *auctionListSelector) displayAuctionList(auctionList []state.ValidatorInfoHandler, numOfSelectedNodes uint32) {
	//if log.GetLevel() > logger.LogDebug {
	//	return
	//}

	tableHeader := []string{"Owner", "Registered key", "TopUp per node"}
	lines := make([]*display.LineData, 0, len(auctionList))
	horizontalLine := false
	for idx, validator := range auctionList {
		pubKey := validator.GetPublicKey()

		owner, err := als.stakingDataProvider.GetBlsKeyOwner(pubKey)
		log.LogIfError(err)

		topUp, err := als.stakingDataProvider.GetNodeStakedTopUp(pubKey)
		log.LogIfError(err)

		horizontalLine = uint32(idx) == numOfSelectedNodes-1
		line := display.NewLineData(horizontalLine, []string{
			(owner),
			string(pubKey),
			topUp.String(),
		})
		lines = append(lines, line)
	}

	table, err := display.CreateTableString(tableHeader, lines)
	if err != nil {
		log.Error("could not create table", "error", err)
		return
	}

	message := fmt.Sprintf("Auction list\n%s", table)
	log.Info(message)
}
