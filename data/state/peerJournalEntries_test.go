package state_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/stretchr/testify/assert"
)

func TestPeerJournalEntryAddress_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewPeerJournalEntryAddress(nil, nil)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestPeerJournalEntryAddress_ShouldWork(t *testing.T) {
	t.Parallel()
	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, err := state.NewPeerJournalEntryAddress(accnt, []byte("address"))

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestPeerJournalEntryAddress_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	shardRootHash := []byte("address")
	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewPeerJournalEntryAddress(accnt, shardRootHash)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, shardRootHash, accnt.Address)
}

func TestPeerJournalEntrySchnorrPublicKey_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewPeerJournalEntrySchnorrPublicKey(nil, nil)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestPeerJournalEntrySchnorrPublicKey_ShouldWork(t *testing.T) {
	t.Parallel()
	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, err := state.NewPeerJournalEntrySchnorrPublicKey(accnt, []byte("address"))

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestPeerJournalEntrySchnorrPublicKey_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	shardRootHash := []byte("address")
	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewPeerJournalEntrySchnorrPublicKey(accnt, shardRootHash)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, shardRootHash, accnt.SchnorrPublicKey)
}

func TestPeerJournalEntryBLSPublicKey_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewPeerJournalEntryBLSPublicKey(nil, nil)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestPeerJournalEntryBLSPublicKey_ShouldWork(t *testing.T) {
	t.Parallel()
	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, err := state.NewPeerJournalEntryBLSPublicKey(accnt, []byte("address"))

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestPeerJournalEntryBLSPublicKey_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	shardRootHash := []byte("address")
	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewPeerJournalEntryBLSPublicKey(accnt, shardRootHash)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, shardRootHash, accnt.BLSPublicKey)
}

func TestPeerJournalEntryStake_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewPeerJournalEntryStake(nil, nil)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestPeerJournalEntryStake_ShouldWork(t *testing.T) {
	t.Parallel()
	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, err := state.NewPeerJournalEntryStake(accnt, big.NewInt(9))

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestPeerJournalEntryStake_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	stake := big.NewInt(999)
	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewPeerJournalEntryStake(accnt, stake)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, stake.Uint64(), accnt.Stake.Uint64())
}

func TestPeerJournalEntryJailTime_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	startTime := state.TimeStamp{Round: 10, Epoch: 10}
	endTime := state.TimeStamp{Round: 11, Epoch: 10}
	jailTime := state.TimePeriod{StartTime: startTime, EndTime: endTime}

	entry, err := state.NewPeerJournalEntryJailTime(nil, jailTime)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestPeerJournalEntryJailTime_ShouldWork(t *testing.T) {
	t.Parallel()

	startTime := state.TimeStamp{Round: 10, Epoch: 10}
	endTime := state.TimeStamp{Round: 11, Epoch: 10}
	jailTime := state.TimePeriod{StartTime: startTime, EndTime: endTime}

	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, err := state.NewPeerJournalEntryJailTime(accnt, jailTime)

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestPeerJournalEntryJailTime_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	startTime := state.TimeStamp{Round: 10, Epoch: 10}
	endTime := state.TimeStamp{Round: 11, Epoch: 10}
	jailTime := state.TimePeriod{StartTime: startTime, EndTime: endTime}

	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewPeerJournalEntryJailTime(accnt, jailTime)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, jailTime, accnt.JailTime)
}

func TestPeerJournalEntryCurrentShardId_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewPeerJournalEntryCurrentShardId(nil, 0)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestPeerJournalEntryCurrentShardId_ShouldWork(t *testing.T) {
	t.Parallel()

	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, err := state.NewPeerJournalEntryCurrentShardId(accnt, 0)

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestPeerJournalEntryCurrentShardId_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewPeerJournalEntryCurrentShardId(accnt, 10)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, uint32(10), accnt.CurrentShardId)
}

func TestPeerJournalEntryNextShardId_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewPeerJournalEntryNextShardId(nil, 0)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestPeerJournalEntryNextShardId_ShouldWork(t *testing.T) {
	t.Parallel()

	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, err := state.NewPeerJournalEntryNextShardId(accnt, 0)

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestPeerJournalEntryNextShardId_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewPeerJournalEntryNextShardId(accnt, 10)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, uint32(10), accnt.NextShardId)
}

func TestPeerJournalEntryInWaitingList_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewPeerJournalEntryInWaitingList(nil, true)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestPeerJournalEntryInWaitingList_ShouldWork(t *testing.T) {
	t.Parallel()

	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, err := state.NewPeerJournalEntryInWaitingList(accnt, true)

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestPeerJournalEntryInWaitingList_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewPeerJournalEntryInWaitingList(accnt, true)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.True(t, accnt.NodeInWaitingList)
}

func TestPeerJournalEntryValidatorSuccessRate_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	successRate := state.SignRate{NrFailure: 10, NrSuccess: 10}

	entry, err := state.NewPeerJournalEntryValidatorSuccessRate(nil, successRate)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestPeerJournalEntryValidatorSuccessRate_ShouldWork(t *testing.T) {
	t.Parallel()

	successRate := state.SignRate{NrFailure: 10, NrSuccess: 10}
	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, err := state.NewPeerJournalEntryValidatorSuccessRate(accnt, successRate)

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestPeerJournalEntryValidatorSuccessRate_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	successRate := state.SignRate{NrFailure: 10, NrSuccess: 10}
	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewPeerJournalEntryValidatorSuccessRate(accnt, successRate)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, successRate, accnt.ValidatorSuccessRate)
}

func TestPeerJournalEntryLeaderSuccessRate_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	successRate := state.SignRate{NrFailure: 10, NrSuccess: 10}

	entry, err := state.NewPeerJournalEntryLeaderSuccessRate(nil, successRate)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestPeerJournalEntryLeaderSuccessRate_ShouldWork(t *testing.T) {
	t.Parallel()

	successRate := state.SignRate{NrFailure: 10, NrSuccess: 10}
	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, err := state.NewPeerJournalEntryLeaderSuccessRate(accnt, successRate)

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestPeerJournalEntryLeaderSuccessRate_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	successRate := state.SignRate{NrFailure: 10, NrSuccess: 10}
	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewPeerJournalEntryLeaderSuccessRate(accnt, successRate)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, successRate, accnt.LeaderSuccessRate)
}

func TestPeerJournalEntryRating_NilAccountShouldErr(t *testing.T) {
	t.Parallel()

	entry, err := state.NewPeerJournalEntryRating(nil, 10)

	assert.Nil(t, entry)
	assert.Equal(t, state.ErrNilAccountHandler, err)
}

func TestPeerJournalEntryRating_ShouldWork(t *testing.T) {
	t.Parallel()

	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, err := state.NewPeerJournalEntryRating(accnt, 10)

	assert.NotNil(t, entry)
	assert.Nil(t, err)
}

func TestPeerJournalEntryRating_RevertOkValsShouldWork(t *testing.T) {
	t.Parallel()

	accnt, _ := state.NewPeerAccount(mock.NewAddressMock(), &mock.AccountTrackerStub{})
	entry, _ := state.NewPeerJournalEntryRating(accnt, 10)
	_, err := entry.Revert()

	assert.Nil(t, err)
	assert.Equal(t, uint32(10), accnt.Rating)
}
