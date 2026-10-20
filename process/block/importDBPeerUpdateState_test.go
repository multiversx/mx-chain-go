package block_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	processBlock "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/stretchr/testify/require"
)

type importPeerUpdateObserver struct {
	*stateMock.AccountsStub
	observe func(data.HeaderHandler, []byte, uint64, []byte)
}

func (o *importPeerUpdateObserver) ProcessImportDBPeerSnapshot(header data.HeaderHandler, hash []byte, finalNonce uint64, finalHash []byte) {
	o.observe(header, hash, finalNonce, finalHash)
}

func TestImportDBPeerUpdateStateUsesSettledExecutionInsteadOfHeaderFinality(t *testing.T) {
	result := func(nonce uint64, hash string) *block.MetaExecutionResultInfo {
		return &block.MetaExecutionResultInfo{ExecutionResult: &block.BaseMetaExecutionResult{
			BaseExecutionResult:    &block.BaseExecutionResult{HeaderNonce: nonce, HeaderHash: []byte(hash), RootHash: []byte("root")},
			ValidatorStatsRootHash: []byte("peer-root"),
		}}
	}
	settled := &block.MetaBlockV3{Nonce: 8, Round: 8, LastExecutionResult: result(5, "executed5")}
	committed := &block.MetaBlockV3{Nonce: 9, Round: 9, PrevHash: []byte("header8"), LastExecutionResult: result(6, "executed6")}
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
	coreComponents.EnableEpochsHandlerField = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, _ uint32) bool {
			return flag == common.SupernovaFlag || flag == common.AndromedaFlag
		},
	}
	dataPool := initDataPool()
	dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{GetHeaderByHashCalled: func([]byte) (data.HeaderHandler, error) { return settled, nil }}
	}
	dataComponents.DataPool = dataPool
	var finalNonce uint64
	var finalHash []byte
	dataComponents.BlockChain = &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled:  func() data.HeaderHandler { return createGenesisMetaBlock() },
		SetFinalBlockInfoCalled: func(nonce uint64, hash, _ []byte) { finalNonce, finalHash = nonce, hash },
		GetFinalBlockInfoCalled: func() (uint64, []byte, []byte) { return finalNonce, finalHash, nil },
	}
	args := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	args.ForkDetector = &mock.ForkDetectorMock{GetHighestSettledBlockInfoCalled: func() (uint64, []byte) { return 8, []byte("header8") }}
	calls := 0
	args.AccountsDB[state.PeerAccountsState] = &importPeerUpdateObserver{
		AccountsStub: &stateMock.AccountsStub{},
		observe: func(header data.HeaderHandler, hash []byte, nonce uint64, blockHash []byte) {
			calls++
			require.Same(t, committed, header)
			require.Equal(t, []byte("header9"), hash)
			require.Equal(t, uint64(5), nonce)
			require.Equal(t, []byte("executed5"), blockHash)
		},
	}
	mp, err := processBlock.NewMetaProcessor(args)
	require.NoError(t, err)
	mp.UpdateState(committed, []byte("header9"))
	require.Equal(t, 1, calls)
}
