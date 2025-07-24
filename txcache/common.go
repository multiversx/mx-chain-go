package txcache

import "github.com/multiversx/mx-chain-go/common/holders"

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
