package node

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/api"

	"github.com/multiversx/mx-chain-go/errors"
)

const (
	baseESDTKeyPrefix = core.ProtectedKeyPrefix + core.ESDTKeyIdentifier
)

type sovereignNode struct {
	*Node
	nativeESDT string
}

// NewSovereignNode creates a new sovereign node instance
func NewSovereignNode(node *Node, nativeESDT string) (*sovereignNode, error) {
	if check.IfNil(node) {
		return nil, errors.ErrNilNode
	}
	if len(nativeESDT) == 0 {
		return nil, ErrEmptyNativeEsdt
	}

	return &sovereignNode{
		Node:       node,
		nativeESDT: nativeESDT,
	}, nil
}

// GetAllIssuedESDTs returns all the issued esdt tokens, works only on metachain
func (sn *sovereignNode) GetAllIssuedESDTs(tokenType string, ctx context.Context) ([]string, error) {
	tokens, err := sn.baseGetAllIssuedESDTs(tokenType, ctx)
	if err != nil {
		return make([]string, 0), err
	}

	return sn.getTokensWithoutNativeESDT(tokens), nil
}

func (sn *sovereignNode) getTokensWithoutNativeESDT(tokens []string) []string {
	nativeEsdtWithBasePrefix := baseESDTKeyPrefix + sn.nativeESDT
	issuedTokens := make([]string, 0)
	for _, token := range tokens {
		if token != nativeEsdtWithBasePrefix {
			issuedTokens = append(issuedTokens, token)
		}
	}
	return issuedTokens
}

// GetNFTTokenIDsRegisteredByAddress returns all the token identifiers for semi or non fungible tokens registered by the address
func (sn *sovereignNode) GetNFTTokenIDsRegisteredByAddress(address string, options api.AccountQueryOptions, ctx context.Context) ([]string, api.BlockInfo, error) {
	addressBytes, err := sn.coreComponents.AddressPubKeyConverter().Decode(address)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	f := &getRegisteredNftsFilter{
		addressBytes: addressBytes,
	}
	return sn.baseGetTokensIDsWithFilter(f, options, ctx)
}

// GetESDTsWithRole returns all the tokens with the given role for the given address
func (sn *sovereignNode) GetESDTsWithRole(address string, role string, options api.AccountQueryOptions, ctx context.Context) ([]string, api.BlockInfo, error) {
	if !core.IsValidESDTRole(role) {
		return nil, api.BlockInfo{}, ErrInvalidESDTRole
	}

	addressBytes, err := sn.coreComponents.AddressPubKeyConverter().Decode(address)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	f := &getTokensWithRoleFilter{
		addressBytes: addressBytes,
		role:         role,
	}
	return sn.baseGetTokensIDsWithFilter(f, options, ctx)
}

// GetESDTsRoles returns all the tokens identifiers and roles for the given address
func (sn *sovereignNode) GetESDTsRoles(address string, options api.AccountQueryOptions, ctx context.Context) (map[string][]string, api.BlockInfo, error) {
	addressBytes, err := sn.coreComponents.AddressPubKeyConverter().Decode(address)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	tokensRoles := make(map[string][]string)

	f := &getAllTokensRolesFilter{
		addressBytes: addressBytes,
		outputRoles:  tokensRoles,
	}
	_, blockInfo, err := sn.baseGetTokensIDsWithFilter(f, options, ctx)
	if err != nil {
		return nil, api.BlockInfo{}, err
	}

	return tokensRoles, blockInfo, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sn *sovereignNode) IsInterfaceNil() bool {
	return sn == nil
}
