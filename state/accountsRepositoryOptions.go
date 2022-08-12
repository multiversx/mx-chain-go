package state

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/holders"
)

func (repository *accountsRepository) checkAccountQueryOptions(options api.AccountQueryOptions) error {
	return nil
}

func (repository *accountsRepository) convertAccountQueryOptions(options api.AccountQueryOptions) (common.GetAccountsStateOptions, error) {
	blockRootHash, err := hex.DecodeString(options.BlockRootHash)
	if err != nil {
		return nil, err
	}

	return holders.NewGetAccountStateOptions(blockRootHash), nil
}
