package node

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
)

type getTokensOfTypeFilter struct {
	tokenType string
}

func (f *getTokensOfTypeFilter) filter(_ string, esdtData *systemSmartContracts.ESDTDataV2) (bool, uint64) {
	return tokenTypeEquals(esdtData.TokenType, f.tokenType), 0
}

type getRegisteredNftsFilter struct {
	addressBytes []byte
}

func (f *getRegisteredNftsFilter) filter(_ string, esdtData *systemSmartContracts.ESDTDataV2) (bool, uint64) {
	return !bytes.Equal(esdtData.TokenType, []byte(core.FungibleESDT)) && bytes.Equal(esdtData.OwnerAddress, f.addressBytes), 0
}

type getTokensWithRoleFilter struct {
	addressBytes []byte
	role         string
}

func (f *getTokensWithRoleFilter) filter(_ string, esdtData *systemSmartContracts.ESDTDataV2) (bool, uint64) {
	for _, esdtRoles := range esdtData.SpecialRoles {
		if !bytes.Equal(esdtRoles.Address, f.addressBytes) {
			continue
		}

		for _, specialRole := range esdtRoles.Roles {
			if bytes.Equal(specialRole, []byte(f.role)) {
				return true, 0
			}
		}
	}

	return false, 0
}

type getAllTokensRolesFilter struct {
	addressBytes []byte
	outputRoles  map[string][]string
}

func (f *getAllTokensRolesFilter) filter(tokenIdentifier string, esdtData *systemSmartContracts.ESDTDataV2) (bool, uint64) {
	for _, esdtRoles := range esdtData.SpecialRoles {
		if !bytes.Equal(esdtRoles.Address, f.addressBytes) {
			continue
		}

		rolesSize := uint64(0)
		rolesStr := make([]string, 0, len(esdtRoles.Roles))
		for _, roleBytes := range esdtRoles.Roles {
			rolesStr = append(rolesStr, string(roleBytes))
			rolesSize += uint64(len(roleBytes))
		}

		f.outputRoles[tokenIdentifier] = rolesStr
		return true, rolesSize
	}
	return false, 0
}
