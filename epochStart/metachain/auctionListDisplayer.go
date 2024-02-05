package metachain

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const maxPubKeyDisplayableLen = 20
const maxNumOfDecimalsToDisplay = 5

type auctionListDisplayer struct {
	softAuctionConfig *auctionConfig
}

func NewAuctionListDisplayer(auctionConfig config.SoftAuctionConfig, denomination int) (*auctionListDisplayer, error) {
	softAuctionConfig, err := getAuctionConfig(auctionConfig, denomination)
	if err != nil {
		return nil, err
	}

	return &auctionListDisplayer{
		softAuctionConfig: softAuctionConfig,
	}, nil
}

func (ald *auctionListDisplayer) DisplayMinRequiredTopUp(topUp *big.Int, startTopUp *big.Int) {
	if log.GetLevel() > logger.LogDebug {
		return
	}

	if topUp.Cmp(ald.softAuctionConfig.minTopUp) > 0 {
		topUp = big.NewInt(0).Sub(topUp, ald.softAuctionConfig.step)
	}

	iteratedValues := big.NewInt(0).Sub(topUp, startTopUp)
	iterations := big.NewInt(0).Div(iteratedValues, ald.softAuctionConfig.step).Int64()
	iterations++

	log.Debug("auctionListSelector: found min required",
		"topUp", getPrettyValue(topUp, ald.softAuctionConfig.denominator),
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

func (ald *auctionListDisplayer) DisplayOwnersData(ownersData map[string]*OwnerAuctionData) {
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
			getPrettyValue(owner.totalTopUp, ald.softAuctionConfig.denominator),
			getPrettyValue(owner.topUpPerNode, ald.softAuctionConfig.denominator),
			getShortDisplayableBlsKeys(owner.auctionList),
		}
		lines = append(lines, display.NewLineData(false, line))
	}

	displayTable(tableHeader, lines, "Initial nodes config in auction list")
}

func (ald *auctionListDisplayer) DisplayOwnersSelectedNodes(ownersData map[string]*OwnerAuctionData) {
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
			getPrettyValue(owner.topUpPerNode, ald.softAuctionConfig.denominator),
			getPrettyValue(owner.totalTopUp, ald.softAuctionConfig.denominator),
			strconv.Itoa(int(owner.numAuctionNodes)),
			strconv.Itoa(int(owner.numQualifiedAuctionNodes)),
			strconv.Itoa(int(owner.numActiveNodes)),
			getPrettyValue(owner.qualifiedTopUpPerNode, ald.softAuctionConfig.denominator),
			getShortDisplayableBlsKeys(owner.auctionList[:owner.numQualifiedAuctionNodes]),
		}
		lines = append(lines, display.NewLineData(false, line))
	}

	displayTable(tableHeader, lines, "Selected nodes config from auction list")
}

func (ald *auctionListDisplayer) DisplayAuctionList(
	auctionList []state.ValidatorInfoHandler,
	ownersData map[string]*OwnerAuctionData,
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
			getPrettyValue(qualifiedTopUp, ald.softAuctionConfig.denominator),
		})
		lines = append(lines, line)
	}

	displayTable(tableHeader, lines, "Final selected nodes from auction list")
}

func getBlsKeyOwnerMap(ownersData map[string]*OwnerAuctionData) map[string]string {
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

// IsInterfaceNil checks if the underlying pointer is nil
func (ald *auctionListDisplayer) IsInterfaceNil() bool {
	return ald == nil
}
