package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/stretchr/testify/assert"
)

func TestConsensusCore_WithNilBlockchainShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		nil,
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.BlsPrivateKey(),
		consensusCoreMock.BlsSingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.SyncTimer(),
		consensusCoreMock.ValidatorGroupSelector())

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestConsensusCore_WithNilBlockProcessorShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		nil,
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.BlsPrivateKey(),
		consensusCoreMock.BlsSingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.SyncTimer(),
		consensusCoreMock.ValidatorGroupSelector())

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestConsensusCore_WithNilBootstrapperShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		nil,
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.BlsPrivateKey(),
		consensusCoreMock.BlsSingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.SyncTimer(),
		consensusCoreMock.ValidatorGroupSelector())

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilBlootstraper, err)
}

func TestConsensusCore_WithNilChronologyShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		nil,
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.BlsPrivateKey(),
		consensusCoreMock.BlsSingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.SyncTimer(),
		consensusCoreMock.ValidatorGroupSelector())

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilChronologyHandler, err)
}

func TestConsensusCore_WithNilHasherShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.Chronology(),
		nil,
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.BlsPrivateKey(),
		consensusCoreMock.BlsSingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.SyncTimer(),
		consensusCoreMock.ValidatorGroupSelector())

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilHasher, err)
}

func TestConsensusCore_WithNilMarshalizerShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		nil,
		consensusCoreMock.BlsPrivateKey(),
		consensusCoreMock.BlsSingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.SyncTimer(),
		consensusCoreMock.ValidatorGroupSelector())

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestConsensusCore_WithNilBlsPrivateKeyShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		nil,
		consensusCoreMock.BlsSingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.SyncTimer(),
		consensusCoreMock.ValidatorGroupSelector())

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilBlsPrivateKey, err)
}

func TestConsensusCore_WithNilBlsSingleSignerShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.BlsPrivateKey(),
		nil,
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.SyncTimer(),
		consensusCoreMock.ValidatorGroupSelector())

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilBlsSingleSigner, err)
}

func TestConsensusCore_WithNilMultiSignerShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.BlsPrivateKey(),
		consensusCoreMock.BlsSingleSigner(),
		nil,
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.SyncTimer(),
		consensusCoreMock.ValidatorGroupSelector())

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilMultiSigner, err)
}

func TestConsensusCore_WithNilRounderShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.BlsPrivateKey(),
		consensusCoreMock.BlsSingleSigner(),
		consensusCoreMock.MultiSigner(),
		nil,
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.SyncTimer(),
		consensusCoreMock.ValidatorGroupSelector())

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilRounder, err)
}

func TestConsensusCore_WithNilShardCoordinatorShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.BlsPrivateKey(),
		consensusCoreMock.BlsSingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		nil,
		consensusCoreMock.SyncTimer(),
		consensusCoreMock.ValidatorGroupSelector())

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestConsensusCore_WithNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.BlsPrivateKey(),
		consensusCoreMock.BlsSingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		nil,
		consensusCoreMock.ValidatorGroupSelector())

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestConsensusCore_WithNilValidatorGroupSelectorShouldFail(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.BlsPrivateKey(),
		consensusCoreMock.BlsSingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.SyncTimer(),
		nil)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilValidatorGroupSelector, err)
}

func TestConsensusCore_CreateConsensusCoreShouldWork(t *testing.T) {
	t.Parallel()

	consensusCoreMock := mock.InitConsensusCore()

	consensusCore, err := spos.NewConsensusCore(
		consensusCoreMock.Blockchain(),
		consensusCoreMock.BlockProcessor(),
		consensusCoreMock.BootStrapper(),
		consensusCoreMock.Chronology(),
		consensusCoreMock.Hasher(),
		consensusCoreMock.Marshalizer(),
		consensusCoreMock.BlsPrivateKey(),
		consensusCoreMock.BlsSingleSigner(),
		consensusCoreMock.MultiSigner(),
		consensusCoreMock.Rounder(),
		consensusCoreMock.ShardCoordinator(),
		consensusCoreMock.SyncTimer(),
		consensusCoreMock.ValidatorGroupSelector())

	assert.NotNil(t, consensusCore)
	assert.Nil(t, err)
}
