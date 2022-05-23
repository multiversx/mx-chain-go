package metachain

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/ElrondNetwork/elrond-go-core/display"
	"github.com/ElrondNetwork/elrond-go/state"
)

const maxPubKeyDisplayableLen = 20

func displayRequiredTopUp(topUp *big.Int, min *big.Int, step *big.Int) {
	//if log.GetLevel() > logger.LogDebug {
	//	return
	//}

	minPossible := big.NewInt(minEGLD)
	if !(topUp.Cmp(minPossible) == 0) {
		topUp = big.NewInt(0).Sub(topUp, step)
	}

	valToIterate := big.NewInt(0).Sub(topUp, min)
	iterations := big.NewInt(0).Div(valToIterate, step)

	log.Info("auctionListSelector: found min required",
		"topUp", topUp.String(),
		"after num of iterations", iterations.String(),
	)
}

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
		"Num staked nodes",
		"Num active nodes",
		"Num auction nodes",
		"Total top up",
		"Top up per node",
		"Auction list nodes",
	}
	lines := make([]*display.LineData, 0, len(ownersData))
	for ownerPubKey, owner := range ownersData {

		line := []string{
			(ownerPubKey),
			strconv.Itoa(int(owner.numStakedNodes)),
			strconv.Itoa(int(owner.numActiveNodes)),
			strconv.Itoa(int(owner.numAuctionNodes)),
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

	message := fmt.Sprintf("Initial nodes config in auction list\n%s", table)
	log.Info(message)
}

func (als *auctionListSelector) displayOwnersSelectedNodes(ownersData2 map[string]*ownerData) {
	//if log.GetLevel() > logger.LogDebug {
	//	return
	//}
	ownersData := copyOwnersData(ownersData2)
	tableHeader := []string{
		"Owner",
		"Num staked nodes",
		"TopUp per node",
		"Total top up",
		"Num auction nodes",
		"Num qualified auction nodes",
		"Num active nodes",
		"Qualified top up per node",
		"Selected auction list nodes",
	}
	lines := make([]*display.LineData, 0, len(ownersData))
	for ownerPubKey, owner := range ownersData {
		selectedFromAuction := owner.auctionList[:owner.numQualifiedAuctionNodes]

		line := []string{
			(ownerPubKey),
			strconv.Itoa(int(owner.numStakedNodes)),
			owner.topUpPerNode.String(),
			owner.totalTopUp.String(),
			strconv.Itoa(int(owner.numAuctionNodes)),
			strconv.Itoa(int(owner.numQualifiedAuctionNodes)),
			strconv.Itoa(int(owner.numActiveNodes)),
			owner.qualifiedTopUpPerNode.String(),
			getShortDisplayableBlsKeys(selectedFromAuction),
		}
		lines = append(lines, display.NewLineData(false, line))
	}

	table, err := display.CreateTableString(tableHeader, lines)
	if err != nil {
		log.Error("could not create table", "error", err)
		return
	}

	message := fmt.Sprintf("Selected nodes config in auction list\n%s", table)
	log.Info(message)
}

func (als *auctionListSelector) displayAuctionList(auctionList []state.ValidatorInfoHandler, ownersData map[string]*ownerData, numOfSelectedNodes uint32) {
	//if log.GetLevel() > logger.LogDebug {
	//	return
	//}

	tableHeader := []string{"Owner", "Registered key", "Qualified TopUp per node"}
	lines := make([]*display.LineData, 0, len(auctionList))
	horizontalLine := false
	for idx, validator := range auctionList {
		pubKey := validator.GetPublicKey()

		owner, err := als.stakingDataProvider.GetBlsKeyOwner(pubKey)
		log.LogIfError(err)

		topUp := ownersData[owner].qualifiedTopUpPerNode
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

	message := fmt.Sprintf("Final selected nodes from auction list\n%s", table)
	log.Info(message)
}
