package accounts

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
)

// SerializeAccounts will serialize the provided accounts in a way that Elastic Search expects a bulk request
func (ap *accountsProcessor) SerializeAccounts(
	accounts map[string]*types.AccountInfo,
	areESDTAccounts bool,
) ([]*bytes.Buffer, error) {
	var err error

	buffSlice := types.NewBufferSlice()
	for address, acc := range accounts {
		meta, serializedData, errPrepareAcc := prepareSerializedAccountInfo(address, acc, areESDTAccounts)
		if len(meta) == 0 {
			log.Warn("accountsProcessor.SerializeAccounts: cannot prepare serialized account info", "error", errPrepareAcc)
			return nil, err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			log.Warn("accountsProcessor.SerializeAccounts: cannot put data in buffer", "error", err.Error())
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

func prepareSerializedAccountInfo(
	address string,
	account *types.AccountInfo,
	isESDTAccount bool,
) ([]byte, []byte, error) {
	id := address
	if isESDTAccount {
		id += fmt.Sprintf("_%s", account.TokenIdentifier)
	}

	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, id, "\n"))
	serializedData, err := json.Marshal(account)
	if err != nil {
		log.Debug("prepareSerializedAccountInfo marshal could not serialize account", "address", address)
		return nil, nil, err
	}

	return meta, serializedData, nil
}

// SerializeAccountsHistory will serialize accounts history in a way that Elastic Search expects a bulk request
func (ap *accountsProcessor) SerializeAccountsHistory(
	accounts map[string]*types.AccountBalanceHistory,
) ([]*bytes.Buffer, error) {
	var err error

	buffSlice := types.NewBufferSlice()
	for address, acc := range accounts {
		meta, serializedData, errPrepareAcc := prepareSerializedAccountBalanceHistory(address, acc)
		if errPrepareAcc != nil {
			log.Warn("accountsProcessor.SerializeAccountsHistory: cannot prepare serialized account balance history", "error", err)
			return nil, err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			log.Warn("accountsProcessor.SerializeAccountsHistory: cannot put data in buffer", "error", err.Error())
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

func prepareSerializedAccountBalanceHistory(
	address string,
	account *types.AccountBalanceHistory,
) ([]byte, []byte, error) {
	// no '_id' is specified because an elastic client would never search after the identifier for this index.
	// this is also an improvement: more details here:
	// https://www.elastic.co/guide/en/elasticsearch/reference/master/tune-for-indexing-speed.html#_use_auto_generated_ids
	meta := []byte(fmt.Sprintf(`{ "index" : { } }%s`, "\n"))

	serializedData, err := json.Marshal(account)
	if err != nil {
		log.Debug("prepareSerializedAccountBalanceHistory could not serialize account history entry", "address", address, "err", err)
		return nil, nil, err
	}

	return meta, serializedData, nil
}
