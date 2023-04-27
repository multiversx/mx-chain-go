package trieIterators

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext/esdtSupply"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
)

type tokensSuppliesProcessor struct {
	storageService dataRetriever.StorageService
	marshaller     marshal.Marshalizer
	tokensSupplies map[string]*big.Int
}

// ArgsTokensSuppliesProcessor is the arguments struct for NewTokensSuppliesProcessor
type ArgsTokensSuppliesProcessor struct {
	StorageService dataRetriever.StorageService
	Marshaller     marshal.Marshalizer
}

// NewTokensSuppliesProcessor returns a new instance of tokensSuppliesProcessor
func NewTokensSuppliesProcessor(args ArgsTokensSuppliesProcessor) (*tokensSuppliesProcessor, error) {
	if check.IfNil(args.StorageService) {
		return nil, errNilStorageService
	}
	if check.IfNil(args.Marshaller) {
		return nil, errNilMarshaller
	}

	return &tokensSuppliesProcessor{
		storageService: args.StorageService,
		marshaller:     args.Marshaller,
		tokensSupplies: make(map[string]*big.Int),
	}, nil
}

// HandleTrieAccountIteration is the handler for the trie account iteration
func (t *tokensSuppliesProcessor) HandleTrieAccountIteration(userAccount state.UserAccountHandler) error {
	if check.IfNil(userAccount) {
		return errNilUserAccount
	}
	if bytes.Equal(core.SystemAccountAddress, userAccount.AddressBytes()) {
		log.Debug("repopulate tokens supplies: skipping system account address")
		return nil
	}
	rh := userAccount.GetRootHash()
	if len(rh) == 0 {
		return nil
	}

	dataTrie := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}

	errDataTrieGet := userAccount.DataTrie().GetAllLeavesOnChannel(dataTrie, context.Background(), rh, keyBuilder.NewKeyBuilder())
	if errDataTrieGet != nil {
		return errDataTrieGet
	}

	log.Trace("extractTokensSupplies - parsing account", "address", userAccount.AddressBytes())
	esdtPrefix := []byte(core.ProtectedKeyPrefix + core.ESDTKeyIdentifier)
	for userLeaf := range dataTrie.LeavesChan {
		if !bytes.HasPrefix(userLeaf.Key(), esdtPrefix) {
			continue
		}

		tokenKey := userLeaf.Key()
		lenESDTPrefix := len(esdtPrefix)
		suffix := append(userLeaf.Key(), userAccount.AddressBytes()...)
		value, errVal := userLeaf.ValueWithoutSuffix(suffix)
		if errVal != nil {
			log.Warn("cannot get value without suffix", "error", errVal, "key", userLeaf.Key())
			return errVal
		}
		var esToken esdt.ESDigitalToken
		err := t.marshaller.Unmarshal(&esToken, value)
		if err != nil {
			return err
		}

		tokenName := string(tokenKey)[lenESDTPrefix:]
		tokenID, nonce := common.ExtractTokenIDAndNonceFromTokenStorageKey([]byte(tokenName))
		tokenIDStr := string(tokenID)
		if nonce > 0 {
			tokenIDStr += fmt.Sprintf("-%d", nonce)
		}

		tokenSupply, found := t.tokensSupplies[tokenIDStr]
		if !found {
			t.tokensSupplies[tokenIDStr] = esToken.Value
		} else {
			tokenSupply = big.NewInt(0).Add(tokenSupply, esToken.Value)
			t.tokensSupplies[tokenIDStr] = tokenSupply
		}
	}

	err := dataTrie.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return fmt.Errorf("error while iterating over an account's trie: %w", err)
	}

	return nil
}

// SaveSupplies will store the recomputed tokens supplies into the database
func (t *tokensSuppliesProcessor) SaveSupplies() error {
	suppliesStorer, err := t.storageService.GetStorer(dataRetriever.ESDTSuppliesUnit)
	if err != nil {
		return err
	}

	for tokenName, supply := range t.tokensSupplies {
		log.Trace("repopulate tokens supplies", "token", tokenName, "supply", supply.String())
		supplyObj := &esdtSupply.SupplyESDT{
			Supply: supply,
		}
		supplyObjBytes, err := t.marshaller.Marshal(supplyObj)
		if err != nil {
			return err
		}

		err = suppliesStorer.Put([]byte(tokenName), supplyObjBytes)
		if err != nil {
			return fmt.Errorf("%w while saving recomputed supply of the token %s", err, tokenName)
		}
	}

	log.Debug("finished the repopulation of the tokens supplies", "num tokens", len(t.tokensSupplies))

	return nil
}
