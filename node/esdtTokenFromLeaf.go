package node

import (
	"bytes"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state/dataTrieValue"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// esdtLeafParser parses the data trie leaves walked for the ESDT tokens of an account as the account's own parser
// does, validating the envelope of an auto-balanced leaf the way a lookup of its key through the account would: a
// leaf whose key is not the one it is stored under, or (once the leaf validation fix is active) whose address is not
// the account's, is parsed as a key without data, which the lookup would have reported too
type esdtLeafParser struct {
	inner               common.TrieLeafParser
	address             []byte
	hasher              hashing.Hasher
	marshaller          marshal.Marshalizer
	enableEpochsHandler common.EnableEpochsHandler
}

func (n *Node) newESDTLeafParser(inner common.TrieLeafParser, address []byte) *esdtLeafParser {
	return &esdtLeafParser{
		inner:               inner,
		address:             address,
		hasher:              n.coreComponents.Hasher(),
		marshaller:          n.coreComponents.InternalMarshalizer(),
		enableEpochsHandler: n.coreComponents.EnableEpochsHandler(),
	}
}

// ParseLeaf returns the key and value of the leaf
func (p *esdtLeafParser) ParseLeaf(trieKey []byte, trieVal []byte, version core.TrieNodeVersion) (core.KeyValueHolder, error) {
	isAutoBalancedLeaf := version == core.AutoBalanceEnabled && p.enableEpochsHandler.IsFlagEnabled(common.AutoBalanceDataTriesFlag)
	if !isAutoBalancedLeaf {
		return p.inner.ParseLeaf(trieKey, trieVal, version)
	}

	data := &dataTrieValue.TrieLeafData{}
	err := p.marshaller.Unmarshal(data, trieVal)
	if err != nil {
		return nil, err
	}

	validEnvelope := bytes.Equal(p.hasher.Compute(string(data.Key)), trieKey)
	if validEnvelope && p.enableEpochsHandler.IsFlagEnabled(common.FixDataTrieLeafValidationFlag) {
		validEnvelope = bytes.Equal(data.Address, p.address)
	}
	if !validEnvelope {
		return keyValStorage.NewKeyValStorage(data.Key, nil), nil
	}

	return keyValStorage.NewKeyValStorage(data.Key, data.Value), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *esdtLeafParser) IsInterfaceNil() bool {
	return p == nil
}

// esdtTokenFromLeaf decodes the token stored in a data trie leaf of a user account and completes it from the system
// account when its metadata lives there (a token with a nonce whose type keeps the metadata in the system account,
// once the save-to-system-account activation is on). It resolves the token exactly as the ESDT storage handler does
// for the same key, without reading the user account data trie again
func (n *Node) esdtTokenFromLeaf(leaf core.KeyValueHolder, nonce uint64, systemAccount vmcommon.UserAccountHandler) (*esdt.ESDigitalToken, error) {
	esdtToken := &esdt.ESDigitalToken{
		Value: big.NewInt(0),
		Type:  uint32(core.Fungible),
	}
	if len(leaf.Value()) == 0 {
		return esdtToken, nil
	}

	marshaller := n.coreComponents.InternalMarshalizer()
	err := marshaller.Unmarshal(esdtToken, leaf.Value())
	if err != nil {
		return nil, err
	}

	metadataOnSystemAccount := n.coreComponents.EnableEpochsHandler().IsFlagEnabled(common.SaveToSystemAccountFlag) &&
		nonce != 0 && !esdtMetaDataOnUserAccount(esdtToken.Type)
	if !metadataOnSystemAccount {
		return esdtToken, nil
	}

	marshalledSystemData, _, err := systemAccount.AccountDataHandler().RetrieveValue(leaf.Key())
	if core.IsGetNodeFromDBError(err) {
		return nil, err
	}
	if err != nil || len(marshalledSystemData) == 0 {
		return esdtToken, nil
	}

	systemData := &esdt.ESDigitalToken{}
	err = marshaller.Unmarshal(systemData, marshalledSystemData)
	if err != nil {
		return nil, err
	}
	if systemData.TokenMetaData != nil {
		esdtToken.TokenMetaData = systemData.TokenMetaData
		if core.IsDynamicESDT(esdtToken.Type) {
			esdtToken.Reserved = systemData.Reserved
		}
	}

	return esdtToken, nil
}

// esdtMetaDataOnUserAccount tells whether the metadata of a token of the given type is stored on the user account
// (and not in the system account); same rule as the ESDT storage handler
func esdtMetaDataOnUserAccount(esdtType uint32) bool {
	return esdtType == uint32(core.NonFungibleV2) ||
		esdtType == uint32(core.DynamicNFT) ||
		esdtType == uint32(core.Fungible)
}
