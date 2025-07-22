package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/state"
	testscommonState "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
)

var defaultBlockchainInfo = holders.NewBlockchainInfo(nil, 0)

var defaultSelectionSessionMock = txcachemocks.SelectionSessionMock{
	GetAccountStateCalled: func(address []byte) (state.UserAccountHandler, error) {
		return &testscommonState.StateUserAccountHandlerStub{
			GetBalanceCalled: func() *big.Int {
				return big.NewInt(20)
			},
			GetNonceCalled: func() uint64 {
				return uint64(1)
			},
		}, nil
	},
}
