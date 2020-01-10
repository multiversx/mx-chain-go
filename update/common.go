package update

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

// Type identifies the type of the export / import
type Type uint8

const (
	Transaction         Type = 0
	SmartContractResult Type = 1
	RewardTransaction   Type = 2
	MiniBlock           Type = 3
	Header              Type = 4
	MetaHeader          Type = 5
	AccountsTrie        Type = 6
	Unknown             Type = 100
)

var Types = []Type{Transaction, SmartContractResult, RewardTransaction, MiniBlock, Header, MetaHeader, AccountsTrie, Unknown}

const atSep = "@"

func GetTrieTypeAndShId(key string) (factory.Type, uint32, error) {
	splitString := strings.Split(key, atSep)
	if len(splitString) < 3 {
		return factory.UserAccount, 0, ErrUnknownType
	}

	accTypeUint64 := big.NewInt(0).SetBytes([]byte(splitString[1])).Uint64()
	accType := factory.UserAccount
	for currType := range factory.Types {
		if currType == int(accTypeUint64) {
			accType = currType
			break
		}
	}

	shId := uint32(big.NewInt(0).SetBytes([]byte(splitString[2])).Uint64())
	return accType, shId, nil
}

func CreateAccountKey(accType factory.Type, shId uint32, address string) string {
	key := CreateTrieIdentifier(shId, accType)
	return key + atSep + address
}

func CreateTrieIdentifier(shID uint32, accountType factory.Type) string {
	return fmt.Sprint("tr%s%d%s%d", atSep, shID, atSep, accountType)
}

func CreateMiniBlockKey(key string) string {
	return "mb" + atSep + key
}

func CreateTransactionKey(key string, tx data.TransactionHandler) string {
	_, ok := tx.(*transaction.Transaction)
	if ok {
		return "tx" + atSep + "nrm" + atSep + key
	}

	_, ok = tx.(*smartContractResult.SmartContractResult)
	if ok {
		return "tx" + atSep + "scr" + atSep + key
	}

	_, ok = tx.(*rewardTx.RewardTx)
	if ok {
		return "tx" + atSep + "rwd" + key
	}

	return "tx" + atSep + "ukw" + key
}

func getTransactionKeyTypeAndHash(splitString []string) (Type, []byte, error) {
	if len(splitString) < 2 {
		return Unknown, nil, ErrUnknownType
	}

	switch splitString[0] {
	case "nrm":
		return Transaction, []byte(splitString[1]), nil
	case "scr":
		return SmartContractResult, []byte(splitString[1]), nil
	case "rwd":
		return RewardTransaction, []byte(splitString[1]), nil
	}

	return Unknown, nil, ErrUnknownType
}

func getTrieTypeAndHash(splitString []string) (Type, []byte, error) {
	if len(splitString) < 3 {
		return Unknown, nil, ErrUnknownType
	}

	accTypeUint64 := big.NewInt(0).SetBytes([]byte(splitString[1])).Uint64()
	accType := factory.UserAccount
	for currType := range factory.Types {
		if currType == int(accTypeUint64) {
			accType = currType
			break
		}
	}

	return accType, []byte(splitString[2]), nil
}

func GetKeyTypeAndHash(key string) (Type, []byte, error) {
	splitString := strings.Split(key, atSep)

	if len(splitString) < 2 {
		return Unknown, nil, ErrUnknownType
	}

	switch splitString[0] {
	case "mb":
		return MiniBlock, []byte(splitString[1]), nil
	case "tx":
		return getTransactionKeyTypeAndHash(splitString[1:])
	case "tr":
		return getTrieTypeAndHash(splitString[1:])
	}

	return Unknown, nil, ErrUnknownType
}
