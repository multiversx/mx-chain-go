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
	bulkSizeThreshold int,
	areESDTAccounts bool,
) ([]bytes.Buffer, error) {
	var err error

	var buff bytes.Buffer
	buffSlice := make([]bytes.Buffer, 0)
	for address, acc := range accounts {
		meta, serializedData, errPrepareAcc := prepareSerializedAccountInfo(address, acc, areESDTAccounts)
		if len(meta) == 0 {
			log.Warn("cannot prepare serialized account info", "error", errPrepareAcc)
			return nil, err
		}

		// append a newline for each element
		serializedData = append(serializedData, "\n"...)

		buffLenWithCurrentAcc := buff.Len() + len(meta) + len(serializedData)
		if buffLenWithCurrentAcc > bulkSizeThreshold && buff.Len() != 0 {
			buffSlice = append(buffSlice, buff)
			buff = bytes.Buffer{}
		}

		buff.Grow(len(meta) + len(serializedData))
		_, err = buff.Write(meta)
		if err != nil {
			log.Warn("elastic search: serialize bulk accounts, write meta", "error", err.Error())
			return nil, err
		}
		_, err = buff.Write(serializedData)
		if err != nil {
			log.Warn("elastic search: serialize bulk accounts, write serialized account", "error", err.Error())
			return nil, err
		}
	}

	// check if the last buffer contains data
	if buff.Len() != 0 {
		buffSlice = append(buffSlice, buff)
	}

	return buffSlice, nil
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
		log.Debug("indexer: marshal",
			"error", "could not serialize account, will skip indexing",
			"address", address)
		return nil, nil, err
	}

	return meta, serializedData, nil
}

// SerializeAccountsHistory will serialize accounts history in a way that Elastic Search expects a bulk request
func (ap *accountsProcessor) SerializeAccountsHistory(
	accounts map[string]*types.AccountBalanceHistory,
	bulkSizeThreshold int,
) ([]bytes.Buffer, error) {
	var err error

	var buff bytes.Buffer
	buffSlice := make([]bytes.Buffer, 0)
	for address, acc := range accounts {
		meta, serializedData, errPrepareAcc := prepareSerializedAccountBalanceHistory(address, acc)
		if errPrepareAcc != nil {
			log.Warn("cannot prepare serialized account balance history", "error", err)
			return nil, err
		}

		// append a newline for each element
		serializedData = append(serializedData, "\n"...)

		buffLenWithCurrentAccountHistory := buff.Len() + len(meta) + len(serializedData)
		if buffLenWithCurrentAccountHistory > bulkSizeThreshold && buff.Len() != 0 {
			buffSlice = append(buffSlice, buff)
			buff = bytes.Buffer{}
		}

		buff.Grow(len(meta) + len(serializedData))
		_, err = buff.Write(meta)
		if err != nil {
			log.Warn("elastic search: serialize bulk accounts history, write meta", "error", err.Error())
			return nil, err
		}
		_, err = buff.Write(serializedData)
		if err != nil {
			log.Warn("elastic search: serialize bulk accounts history, write serialized account history", "error", err.Error())
			return nil, err
		}
	}

	// check if the last buffer contains data
	if buff.Len() != 0 {
		buffSlice = append(buffSlice, buff)
	}

	return buffSlice, nil
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
		log.Debug("indexer: marshal",
			"error", "could not serialize account history entry, will skip indexing",
			"address", address)
		return nil, nil, err
	}

	return meta, serializedData, nil
}
