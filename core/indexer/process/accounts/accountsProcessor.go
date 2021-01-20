package accounts

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

var log = logger.GetOrCreate("indexer/process/accounts")

const numDecimalsInFloatBalance = 10

// AccountESDT is a structure that is needed for ESDT accounts
type AccountESDT struct {
	Account         state.UserAccountHandler
	TokenIdentifier string
	IsSender        bool
}

// AlteredAccount is a structure that holds information about an altered account
type AlteredAccount struct {
	IsSender        bool
	IsESDTOperation bool
	TokenIdentifier string
}

// AccountsProcessor a is structure responsible for processing accounts
type AccountsProcessor struct {
	dividerForDenomination float64
	balancePrecision       float64
	internalMarshalizer    marshal.Marshalizer
	addressPubkeyConverter core.PubkeyConverter
	accountsDB             state.AccountsAdapter
}

// NewAccountsProcessor will create a new instance of accounts processor
func NewAccountsProcessor(
	denomination int,
	marshalizer marshal.Marshalizer,
	addressPubkeyConverter core.PubkeyConverter,
	accountsDB state.AccountsAdapter,
) *AccountsProcessor {
	return &AccountsProcessor{
		internalMarshalizer:    marshalizer,
		addressPubkeyConverter: addressPubkeyConverter,
		balancePrecision:       math.Pow(10, float64(numDecimalsInFloatBalance)),
		dividerForDenomination: math.Pow(10, float64(core.MaxInt(denomination, 0))),
		accountsDB:             accountsDB,
	}
}

// GetAccounts will get accounts for egld operations and esdt operations
func (ap *AccountsProcessor) GetAccounts(alteredAccounts map[string]*AlteredAccount) ([]*types.AccountEGLD, []*AccountESDT) {
	accountsToIndexEGLD := make([]*types.AccountEGLD, 0)
	accountsToIndexESDT := make([]*AccountESDT, 0)
	for address, info := range alteredAccounts {
		addressBytes, err := ap.addressPubkeyConverter.Decode(address)
		if err != nil {
			log.Warn("cannot decode address", "address", address, "error", err)
			continue
		}

		account, err := ap.accountsDB.LoadAccount(addressBytes)
		if err != nil {
			log.Warn("cannot load account", "address bytes", addressBytes, "error", err)
			continue
		}

		userAccount, ok := account.(state.UserAccountHandler)
		if !ok {
			log.Warn("cannot cast AccountHandler to type UserAccountHandler")
			continue
		}

		if info.IsESDTOperation {
			accountsToIndexESDT = append(accountsToIndexESDT, &AccountESDT{
				Account:         userAccount,
				TokenIdentifier: info.TokenIdentifier,
				IsSender:        info.IsSender,
			})
		}

		if info.IsESDTOperation && !info.IsSender {
			// should continue because he have an esdt transfer and the current account is not the sender
			// this transfer will not affect the egld balance of the account
			continue
		}

		accountsToIndexEGLD = append(accountsToIndexEGLD, &types.AccountEGLD{
			Account:  userAccount,
			IsSender: info.IsSender,
		})
	}

	return accountsToIndexEGLD, accountsToIndexESDT
}

// PrepareAccountsMapEGLD will prepare a map of accounts with egld
func (ap *AccountsProcessor) PrepareAccountsMapEGLD(accounts []*types.AccountEGLD) map[string]*types.AccountInfo {
	accountsMap := make(map[string]*types.AccountInfo)
	for _, userAccount := range accounts {
		balanceAsFloat := ap.computeBalanceAsFloat(userAccount.Account.GetBalance())
		acc := &types.AccountInfo{
			Nonce:      userAccount.Account.GetNonce(),
			Balance:    userAccount.Account.GetBalance().String(),
			BalanceNum: balanceAsFloat,
			IsSender:   userAccount.IsSender,
		}
		address := ap.addressPubkeyConverter.Encode(userAccount.Account.AddressBytes())
		accountsMap[address] = acc
	}

	return accountsMap
}

// PrepareAccountsMapESDT will prepare a map of accounts with ESDT tokens
func (ap *AccountsProcessor) PrepareAccountsMapESDT(accounts []*AccountESDT) map[string]*types.AccountInfo {
	accountsESDTMap := make(map[string]*types.AccountInfo)
	for _, accountESDT := range accounts {
		address := ap.addressPubkeyConverter.Encode(accountESDT.Account.AddressBytes())
		balance, properties, err := ap.getESDTInfo(accountESDT)
		if err != nil {
			log.Warn("cannot get esdt info from account",
				"address", address,
				"error", err.Error())
			continue
		}

		acc := &types.AccountInfo{
			Address:         address,
			TokenIdentifier: accountESDT.TokenIdentifier,
			Balance:         balance.String(),
			BalanceNum:      ap.computeBalanceAsFloat(balance),
			Properties:      properties,
			IsSender:        accountESDT.IsSender,
		}

		accountsESDTMap[address] = acc
	}

	return accountsESDTMap
}

// PrepareAccountsHistory will prepare a map of accounts history balance from a map of accounts
func (ap *AccountsProcessor) PrepareAccountsHistory(accounts map[string]*types.AccountInfo) map[string]*types.AccountBalanceHistory {
	currentTimestamp := time.Now().Unix()
	accountsMap := make(map[string]*types.AccountBalanceHistory)
	for address, userAccount := range accounts {
		acc := &types.AccountBalanceHistory{
			Address:         address,
			Balance:         userAccount.Balance,
			Timestamp:       currentTimestamp,
			TokenIdentifier: userAccount.TokenIdentifier,
			IsSender:        userAccount.IsSender,
		}
		addressKey := fmt.Sprintf("%s_%d", address, currentTimestamp)
		accountsMap[addressKey] = acc
	}

	return accountsMap
}

func (ap *AccountsProcessor) getESDTInfo(accountESDT *AccountESDT) (*big.Int, string, error) {
	if accountESDT.TokenIdentifier == "" {
		return nil, "", nil
	}

	tokenKey := core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier + accountESDT.TokenIdentifier
	valueBytes, err := accountESDT.Account.DataTrieTracker().RetrieveValue([]byte(tokenKey))
	if err != nil {
		return nil, "", err
	}

	esdtToken := &esdt.ESDigitalToken{}
	err = ap.internalMarshalizer.Unmarshal(esdtToken, valueBytes)
	if err != nil {
		return nil, "", err
	}

	return esdtToken.Value, hex.EncodeToString(esdtToken.Properties), nil
}

func (ap *AccountsProcessor) computeBalanceAsFloat(balance *big.Int) float64 {
	balanceBigFloat := big.NewFloat(0).SetInt(balance)
	balanceFloat64, _ := balanceBigFloat.Float64()

	bal := balanceFloat64 / ap.dividerForDenomination
	balanceFloatWithDecimals := math.Round(bal*ap.balancePrecision) / ap.balancePrecision

	return core.MaxFloat64(balanceFloatWithDecimals, 0)
}
