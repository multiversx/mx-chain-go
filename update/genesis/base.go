package genesis

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/update"
)

// MetaBlockFileName is the constant which defines the export/import filename for metablock
const MetaBlockFileName = "metaBlock"

// MetaBlockFileName is the constant which defines the export/import filename for miniblocks
const TransactionsFileName = "transactions"

// MetaBlockFileName is the constant which defines the export/import filename for transactions
const MiniBlocksFileName = "miniBlocks"

// TrieFileName is the constant which defines the export/import filename for tries
const TrieFileName = "trie"

// Type identifies the type of the export / import
type Type uint8

const (
	Unknown Type = iota
	Transaction
	SmartContractResult
	RewardTransaction
	MiniBlock
	Header
	MetaHeader
	AccountsTrie
	RootHash
	UserAccount
	PeerAccount
	MetaAccount
)

var Types = []Type{Transaction, SmartContractResult, RewardTransaction, MiniBlock, Header, MetaHeader, AccountsTrie, Unknown}

const atSep = "@"

func NewObject(objType Type) (interface{}, error) {
	switch objType {
	case Transaction:
		return &transaction.Transaction{}, nil
	case SmartContractResult:
		return &smartContractResult.SmartContractResult{}, nil
	case RewardTransaction:
		return &rewardTx.RewardTx{}, nil
	case MiniBlock:
		return &block.MiniBlock{}, nil
	case Header:
		return &block.Header{}, nil
	case MetaHeader:
		return &block.MetaBlock{}, nil
	case RootHash:
		return make([]byte, 0), nil
	}
	return nil, update.ErrUnknownType
}

func NewEmptyAccount(accType factory.Type) (state.AccountHandler, error) {
	switch accType {
	case factory.UserAccount:
		return &state.Account{}, nil
	case factory.ShardStatistics:
		return &state.MetaAccount{}, nil
	case factory.ValidatorAccount:
		return &state.PeerAccount{}, nil

	}
	return nil, update.ErrUnknownType
}

func GetTrieTypeAndShId(key string) (factory.Type, uint32, error) {
	splitString := strings.Split(key, atSep)
	if len(splitString) < 3 {
		return factory.UserAccount, 0, update.ErrUnknownType
	}

	accTypeUint64 := big.NewInt(0).SetBytes([]byte(splitString[1])).Uint64()
	accType := getAccountType(int(accTypeUint64))

	shId := uint32(big.NewInt(0).SetBytes([]byte(splitString[2])).Uint64())
	return accType, shId, nil
}

func getTransactionKeyTypeAndHash(splitString []string) (Type, []byte, error) {
	if len(splitString) < 2 {
		return Unknown, nil, update.ErrUnknownType
	}

	switch splitString[0] {
	case "nrm":
		return Transaction, []byte(splitString[1]), nil
	case "scr":
		return SmartContractResult, []byte(splitString[1]), nil
	case "rwd":
		return RewardTransaction, []byte(splitString[1]), nil
	}

	return Unknown, nil, update.ErrUnknownType
}

func getAccountType(intType int) factory.Type {
	accType := factory.UserAccount
	for currType := range factory.SupportedAccountTypes {
		if currType == intType {
			accType = factory.Type(currType)
			break
		}
	}
	return accType
}

func getTrieTypeAndHash(splitString []string) (Type, []byte, error) {
	if len(splitString) < 3 {
		return Unknown, nil, update.ErrUnknownType
	}

	accTypeUint64 := big.NewInt(0).SetBytes([]byte(splitString[1])).Uint64()
	accType := getAccountType(int(accTypeUint64))

	convertedType := Unknown
	switch accType {
	case factory.UserAccount:
		convertedType = UserAccount
	case factory.ShardStatistics:
		convertedType = MetaAccount
	case factory.ValidatorAccount:
		convertedType = PeerAccount
	}

	return convertedType, []byte(splitString[2]), nil
}

func GetKeyTypeAndHash(key string) (Type, []byte, error) {
	splitString := strings.Split(key, atSep)

	if len(splitString) < 2 {
		return Unknown, nil, update.ErrUnknownType
	}

	switch splitString[0] {
	case "mb":
		return MiniBlock, []byte(splitString[1]), nil
	case "tx":
		return getTransactionKeyTypeAndHash(splitString[1:])
	case "tr":
		return getTrieTypeAndHash(splitString[1:])
	case "rt":
		return RootHash, []byte(key), nil
	}

	return Unknown, nil, update.ErrUnknownType
}

func CreateVersionKey(meta *block.MetaBlock) string {
	return string(meta.ChainID)
}

func CreateAccountKey(accType factory.Type, shId uint32, address string) string {
	key := CreateTrieIdentifier(shId, accType)
	return key + atSep + address
}

func CreateRootHashKey(trieIdentifier string) string {
	return "rt" + atSep + trieIdentifier
}

func CreateTrieIdentifier(shID uint32, accountType factory.Type) string {
	return fmt.Sprint("tr", atSep, shID, atSep, accountType)
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
