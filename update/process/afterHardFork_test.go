package process

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/assert"
)

func createMockArgsAfterHardFork() ArgsAfterHardFork {
	return ArgsAfterHardFork{
		MapBlockProcessors: make(map[uint32]update.HardForkBlockProcessor),
		ImportHandler:      &mock.ImportHandlerStub{},
		ShardCoordinator:   mock.NewOneShardCoordinatorMock(),
		Hasher:             &mock.HasherMock{},
		Marshalizer:        &mock.MarshalizerMock{},
	}
}

func TestNewAfterHardForkBlockCreation(t *testing.T) {
	t.Parallel()

	args := createMockArgsAfterHardFork()

	hardForkBlockCreator, err := NewAfterHardForkBlockCreation(args)
	assert.NoError(t, err)
	assert.False(t, check.IfNil(hardForkBlockCreator))
}

func TestCreateAllBlocksAfterHardfork(t *testing.T) {
	t.Parallel()

	args := createMockArgsAfterHardFork()
	args.ShardCoordinator = &mock.CoordinatorStub{
		NumberOfShardsCalled: func() uint32 {
			return 1
		},
	}

	hdr1 := &block.Header{}
	hdr2 := &block.Header{}
	body1 := &block.Body{}
	body2 := &block.Body{}
	args.MapBlockProcessors[0] = &mock.HardForkBlockProcessor{
		CreateBodyCalled: func() (*block.Body, []*update.MbInfo, error) {
			return body1, nil, nil
		},
		CreateBlockCalled: func(body *block.Body, chainID string, round uint64, nonce uint64, epoch uint32) (data.HeaderHandler, error) {
			return hdr1, nil
		},
	}
	args.MapBlockProcessors[core.MetachainShardId] = &mock.HardForkBlockProcessor{
		CreateBodyCalled: func() (*block.Body, []*update.MbInfo, error) {
			return body2, nil, nil
		},
		CreateBlockCalled: func(body *block.Body, chainID string, round uint64, nonce uint64, epoch uint32) (data.HeaderHandler, error) {
			return hdr2, nil
		},
	}

	hardForkBlockCreator, _ := NewAfterHardForkBlockCreation(args)

	expectedHeaders := map[uint32]data.HeaderHandler{
		0: hdr1, core.MetachainShardId: hdr2,
	}
	expectedBodies := map[uint32]*block.Body{
		0: body1, core.MetachainShardId: body2,
	}
	chainID, round, nonce, epoch := "chainId", uint64(100), uint64(90), uint32(2)
	headers, bodies, err := hardForkBlockCreator.CreateAllBlocksAfterHardfork(chainID, round, nonce, epoch)
	assert.NoError(t, err)
	assert.Equal(t, expectedHeaders, headers)
	assert.Equal(t, expectedBodies, bodies)

}
