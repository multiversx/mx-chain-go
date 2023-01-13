package metachain

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const maxPubKeyDisplayableLen = 20
const maxNumOfDecimalsToDisplay = 5

func (als *auctionListSelector) displayMinRequiredTopUp(topUp *big.Int, startTopUp *big.Int) {
	if log.GetLevel() > logger.LogDebug {
		return
	}

	if topUp.Cmp(als.softAuctionConfig.minTopUp) > 0 {
		topUp = big.NewInt(0).Sub(topUp, als.softAuctionConfig.step)
	}

	iteratedValues := big.NewInt(0).Sub(topUp, startTopUp)
	iterations := big.NewInt(0).Div(iteratedValues, als.softAuctionConfig.step).Int64()
	iterations++

	log.Debug("auctionListSelector: found min required",
		"topUp", getPrettyValue(topUp, als.softAuctionConfig.denominator),
		"after num of iterations", iterations,
	)
}

func getShortKey(pubKey []byte) string {
	pubKeyHex := hex.EncodeToString(pubKey)
	displayablePubKey := pubKeyHex

	pubKeyLen := len(displayablePubKey)
	if pubKeyLen > maxPubKeyDisplayableLen {
		displayablePubKey = pubKeyHex[:maxPubKeyDisplayableLen/2] + "..." + pubKeyHex[pubKeyLen-maxPubKeyDisplayableLen/2:]
	}

	return displayablePubKey
}

func getShortDisplayableBlsKeys(list []state.ValidatorInfoHandler) string {
	pubKeys := ""

	for idx, validator := range list {
		pubKeys += getShortKey(validator.GetPublicKey())
		addDelimiter := idx != len(list)-1
		if addDelimiter {
			pubKeys += ", "
		}
	}

	return pubKeys
}

func getPrettyValue(val *big.Int, denominator *big.Int) string {
	first := big.NewInt(0).Div(val, denominator).String()
	decimals := big.NewInt(0).Mod(val, denominator).String()

	zeroesCt := (len(denominator.String()) - len(decimals)) - 1
	zeroesCt = core.MaxInt(zeroesCt, 0)
	zeroes := strings.Repeat("0", zeroesCt)

	second := zeroes + decimals
	if len(second) > maxNumOfDecimalsToDisplay {
		second = second[:maxNumOfDecimalsToDisplay]
	}

	return first + "." + second
}

func (als *auctionListSelector) displayOwnersData(ownersData map[string]*ownerAuctionData) {
	if log.GetLevel() > logger.LogDebug {
		return
	}

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
			hex.EncodeToString([]byte(ownerPubKey)),
			strconv.Itoa(int(owner.numStakedNodes)),
			strconv.Itoa(int(owner.numActiveNodes)),
			strconv.Itoa(int(owner.numAuctionNodes)),
			getPrettyValue(owner.totalTopUp, als.softAuctionConfig.denominator),
			getPrettyValue(owner.topUpPerNode, als.softAuctionConfig.denominator),
			getShortDisplayableBlsKeys(owner.auctionList),
		}
		lines = append(lines, display.NewLineData(false, line))
	}

	displayTable(tableHeader, lines, "Initial nodes config in auction list")
}

func (als *auctionListSelector) displayOwnersSelectedNodes(ownersData map[string]*ownerAuctionData) {
	if log.GetLevel() > logger.LogDebug {
		return
	}

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
		line := []string{
			hex.EncodeToString([]byte(ownerPubKey)),
			strconv.Itoa(int(owner.numStakedNodes)),
			getPrettyValue(owner.topUpPerNode, als.softAuctionConfig.denominator),
			getPrettyValue(owner.totalTopUp, als.softAuctionConfig.denominator),
			strconv.Itoa(int(owner.numAuctionNodes)),
			strconv.Itoa(int(owner.numQualifiedAuctionNodes)),
			strconv.Itoa(int(owner.numActiveNodes)),
			getPrettyValue(owner.qualifiedTopUpPerNode, als.softAuctionConfig.denominator),
			getShortDisplayableBlsKeys(owner.auctionList[:owner.numQualifiedAuctionNodes]),
		}
		lines = append(lines, display.NewLineData(false, line))
	}

	displayTable(tableHeader, lines, "Selected nodes config from auction list")
}

func (als *auctionListSelector) displayAuctionList(
	auctionList []state.ValidatorInfoHandler,
	ownersData map[string]*ownerAuctionData,
	numOfSelectedNodes uint32,
) {
	if log.GetLevel() > logger.LogDebug {
		return
	}

	tableHeader := []string{"Owner", "Registered key", "Qualified TopUp per node"}
	lines := make([]*display.LineData, 0, len(auctionList))
	blsKeysOwnerMap := getBlsKeyOwnerMap(ownersData)
	for idx, validator := range auctionList {
		pubKey := validator.GetPublicKey()
		owner, found := blsKeysOwnerMap[string(pubKey)]
		if !found {
			log.Error("auctionListSelector.displayAuctionList could not find owner for",
				"bls key", hex.EncodeToString(pubKey))
			continue
		}

		qualifiedTopUp := ownersData[owner].qualifiedTopUpPerNode
		horizontalLine := uint32(idx) == numOfSelectedNodes-1
		line := display.NewLineData(horizontalLine, []string{
			hex.EncodeToString([]byte(owner)),
			hex.EncodeToString(pubKey),
			getPrettyValue(qualifiedTopUp, als.softAuctionConfig.denominator),
		})
		lines = append(lines, line)
	}

	displayTable(tableHeader, lines, "Final selected nodes from auction list")
}

func getBlsKeyOwnerMap(ownersData map[string]*ownerAuctionData) map[string]string {
	ret := make(map[string]string)
	for ownerPubKey, owner := range ownersData {
		for _, blsKey := range owner.auctionList {
			ret[string(blsKey.GetPublicKey())] = ownerPubKey
		}
	}

	return ret
}

func displayTable(tableHeader []string, lines []*display.LineData, message string) {
	table, err := display.CreateTableString(tableHeader, lines)
	if err != nil {
		log.Error("could not create table", "error", err)
		return
	}

	msg := fmt.Sprintf("%s\n%s", message, table)
	log.Debug(msg)
}
