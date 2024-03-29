package metachain

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-go/config"
	errorsCommon "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
)

const maxPubKeyDisplayableLen = 20
const maxNumOfDecimalsToDisplay = 5

type auctionListDisplayer struct {
	softAuctionConfig        *auctionConfig
	tableDisplayer           TableDisplayHandler
	validatorPubKeyConverter core.PubkeyConverter
	addressPubKeyConverter   core.PubkeyConverter
}

// ArgsAuctionListDisplayer is a struct placeholder for arguments needed to create an auction list displayer
type ArgsAuctionListDisplayer struct {
	TableDisplayHandler      TableDisplayHandler
	ValidatorPubKeyConverter core.PubkeyConverter
	AddressPubKeyConverter   core.PubkeyConverter
	AuctionConfig            config.SoftAuctionConfig
	Denomination             int
}

// NewAuctionListDisplayer creates an auction list data displayer, useful for debugging purposes during selection process
func NewAuctionListDisplayer(args ArgsAuctionListDisplayer) (*auctionListDisplayer, error) {
	softAuctionConfig, err := getAuctionConfig(args.AuctionConfig, args.Denomination)
	if err != nil {
		return nil, err
	}

	err = checkDisplayerNilArgs(args)
	if err != nil {
		return nil, err
	}

	return &auctionListDisplayer{
		softAuctionConfig:        softAuctionConfig,
		tableDisplayer:           args.TableDisplayHandler,
		validatorPubKeyConverter: args.ValidatorPubKeyConverter,
		addressPubKeyConverter:   args.AddressPubKeyConverter,
	}, nil
}

func checkDisplayerNilArgs(args ArgsAuctionListDisplayer) error {
	if check.IfNil(args.TableDisplayHandler) {
		return errNilTableDisplayHandler
	}
	if check.IfNil(args.ValidatorPubKeyConverter) {
		return errorsCommon.ErrNilValidatorPublicKeyConverter
	}
	if check.IfNil(args.AddressPubKeyConverter) {
		return errorsCommon.ErrNilAddressPublicKeyConverter
	}

	return nil
}

// DisplayOwnersData will display initial owners data for auction selection
func (ald *auctionListDisplayer) DisplayOwnersData(ownersData map[string]*OwnerAuctionData) {

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
			ald.addressPubKeyConverter.SilentEncode([]byte(ownerPubKey), log),
			strconv.Itoa(int(owner.numStakedNodes)),
			strconv.Itoa(int(owner.numActiveNodes)),
			strconv.Itoa(int(owner.numAuctionNodes)),
			getPrettyValue(owner.totalTopUp, ald.softAuctionConfig.denominator),
			getPrettyValue(owner.topUpPerNode, ald.softAuctionConfig.denominator),
			ald.getShortDisplayableBlsKeys(owner.auctionList),
		}
		lines = append(lines, display.NewLineData(false, line))
	}

	ald.tableDisplayer.DisplayTable(tableHeader, lines, "Initial nodes config in auction list")
}

func (ald *auctionListDisplayer) DisplayExtraOwnersData(ownersData map[string]*OwnerAuctionData) {
	tableHeader := []string{
		"Owner",
		"Num staked nodes",
		"Num active nodes",
		"Num auction nodes",
		"Total top up",
		"Top up per node",
	}

	lines := make([]*display.LineData, 0, len(ownersData))
	for ownerPubKey, owner := range ownersData {
		line := []string{
			ald.addressPubKeyConverter.SilentEncode([]byte(ownerPubKey), log),
			strconv.Itoa(int(owner.numStakedNodes)),
			strconv.Itoa(int(owner.numActiveNodes)),
			strconv.Itoa(int(owner.numAuctionNodes)),
			getPrettyValue(owner.totalTopUp, ald.softAuctionConfig.denominator),
			getPrettyValue(owner.topUpPerNode, ald.softAuctionConfig.denominator),
		}
		lines = append(lines, display.NewLineData(false, line))
	}

	ald.tableDisplayer.DisplayTable(tableHeader, lines, "Extra nodes config not in auction list")
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

func (ald *auctionListDisplayer) getShortDisplayableBlsKeys(list []state.ValidatorInfoHandler) string {
	pubKeys := ""

	for idx, validator := range list {
		pubKeys += ald.getShortKey(validator.GetPublicKey())
		addDelimiter := idx != len(list)-1
		if addDelimiter {
			pubKeys += ", "
		}
	}

	return pubKeys
}

func (ald *auctionListDisplayer) getShortKey(pubKey []byte) string {
	pubKeyHex := ald.validatorPubKeyConverter.SilentEncode(pubKey, log)
	displayablePubKey := pubKeyHex

	pubKeyLen := len(displayablePubKey)
	if pubKeyLen > maxPubKeyDisplayableLen {
		displayablePubKey = pubKeyHex[:maxPubKeyDisplayableLen/2] + "..." + pubKeyHex[pubKeyLen-maxPubKeyDisplayableLen/2:]
	}

	return displayablePubKey
}

// DisplayOwnersSelectedNodes will display owners' selected nodes
func (ald *auctionListDisplayer) DisplayOwnersSelectedNodes(ownersData map[string]*OwnerAuctionData) {

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
			ald.addressPubKeyConverter.SilentEncode([]byte(ownerPubKey), log),
			strconv.Itoa(int(owner.numStakedNodes)),
			getPrettyValue(owner.topUpPerNode, ald.softAuctionConfig.denominator),
			getPrettyValue(owner.totalTopUp, ald.softAuctionConfig.denominator),
			strconv.Itoa(int(owner.numAuctionNodes)),
			strconv.Itoa(int(owner.numQualifiedAuctionNodes)),
			strconv.Itoa(int(owner.numActiveNodes)),
			getPrettyValue(owner.qualifiedTopUpPerNode, ald.softAuctionConfig.denominator),
			ald.getShortDisplayableBlsKeys(owner.auctionList[:owner.numQualifiedAuctionNodes]),
		}
		lines = append(lines, display.NewLineData(false, line))
	}

	ald.tableDisplayer.DisplayTable(tableHeader, lines, "Selected nodes config from auction list")
}

// DisplayAuctionList will display the final selected auction nodes
func (ald *auctionListDisplayer) DisplayAuctionList(
	auctionList []state.ValidatorInfoHandler,
	ownersData map[string]*OwnerAuctionData,
	numOfSelectedNodes uint32,
) {

	tableHeader := []string{"Owner", "Registered key", "Qualified TopUp per node"}
	lines := make([]*display.LineData, 0, len(auctionList))
	blsKeysOwnerMap := getBlsKeyOwnerMap(ownersData)
	for idx, validator := range auctionList {
		pubKey := validator.GetPublicKey()
		pubKeyEncoded := ald.validatorPubKeyConverter.SilentEncode(pubKey, log)
		owner, found := blsKeysOwnerMap[string(pubKey)]
		if !found {
			log.Error("auctionListSelector.displayAuctionList could not find owner for",
				"bls key", pubKeyEncoded)
			continue
		}

		qualifiedTopUp := ownersData[owner].qualifiedTopUpPerNode
		horizontalLine := uint32(idx) == numOfSelectedNodes-1
		line := display.NewLineData(horizontalLine, []string{
			ald.addressPubKeyConverter.SilentEncode([]byte(owner), log),
			pubKeyEncoded,
			getPrettyValue(qualifiedTopUp, ald.softAuctionConfig.denominator),
		})
		lines = append(lines, line)
	}

	ald.tableDisplayer.DisplayTable(tableHeader, lines, "Final selected nodes from auction list")
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

// IsInterfaceNil checks if the underlying pointer is nil
func (ald *auctionListDisplayer) IsInterfaceNil() bool {
	return ald == nil
}
