package node

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
)

type tokensFilterFunction func(tokenIdentifier string, esdtData *systemSmartContracts.ESDTData) bool

func getRegisteredNftsFilter(addressBytes []byte) tokensFilterFunction {
	return func(_ string, esdtData *systemSmartContracts.ESDTData) bool {
		return !bytes.Equal(esdtData.TokenType, []byte(core.FungibleESDT)) && bytes.Equal(esdtData.OwnerAddress, addressBytes)
	}
}

func getTokensWithRoleFilter(addressBytes []byte, role string) tokensFilterFunction {
	return func(tokenIdentifier string, esdtData *systemSmartContracts.ESDTData) bool {
		for _, esdtRoles := range esdtData.SpecialRoles {
			if !bytes.Equal(esdtRoles.Address, addressBytes) {
				continue
			}

			for _, specialRole := range esdtRoles.Roles {
				if bytes.Equal(specialRole, []byte(role)) {
					return true
				}
			}
		}

		return false
	}
}

func getAllTokensRolesFilter(addressBytes []byte, outputRoles map[string][]string) tokensFilterFunction {
	return func(tokenIdentifier string, esdtData *systemSmartContracts.ESDTData) bool {
		for _, esdtRoles := range esdtData.SpecialRoles {
			if !bytes.Equal(esdtRoles.Address, addressBytes) {
				continue
			}

			rolesStr := make([]string, 0, len(esdtRoles.Roles))
			for _, roleBytes := range esdtRoles.Roles {
				rolesStr = append(rolesStr, string(roleBytes))
			}

			outputRoles[tokenIdentifier] = rolesStr
			return true
		}
		return false
	}
}
