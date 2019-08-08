package state

import "math/big"

//------- PeerJournalEntryAddress

// PeerJournalEntryAddress is used to revert a round change
type PeerJournalEntryAddress struct {
	account    *PeerAccount
	oldAddress []byte
}

// NewPeerJournalEntryAddress outputs a new PeerJournalEntry implementation used to revert a round change
func NewPeerJournalEntryAddress(account *PeerAccount, oldAddress []byte) (*PeerJournalEntryAddress, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &PeerJournalEntryAddress{
		account:    account,
		oldAddress: oldAddress,
	}, nil
}

// Revert applies undo operation
func (jen *PeerJournalEntryAddress) Revert() (AccountHandler, error) {
	jen.account.Address = jen.oldAddress

	return jen.account, nil
}

//------- PeerJournalEntrySchnorrPublicKey

// PeerJournalEntrySchnorrPublicKey is used to revert a round change
type PeerJournalEntrySchnorrPublicKey struct {
	account          *PeerAccount
	oldSchnorrPubKey []byte
}

// NewPeerJournalEntrySchnorrPublicKey outputs a new PeerJournalEntrySchnorrPublicKey implementation used to revert a round change
func NewPeerJournalEntrySchnorrPublicKey(
	account *PeerAccount,
	oldSchnorrPubKey []byte,
) (*PeerJournalEntrySchnorrPublicKey, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &PeerJournalEntrySchnorrPublicKey{
		account:          account,
		oldSchnorrPubKey: oldSchnorrPubKey,
	}, nil
}

// Revert applies undo operation
func (jen *PeerJournalEntrySchnorrPublicKey) Revert() (AccountHandler, error) {
	jen.account.SchnorrPublicKey = jen.oldSchnorrPubKey

	return jen.account, nil
}

//------- PeerJournalEntryBLSPublicKey

// PeerJournalEntryBLSPublicKey is used to revert a round change
type PeerJournalEntryBLSPublicKey struct {
	account      *PeerAccount
	oldBLSPubKey []byte
}

// NewPeerJournalEntryBLSPublicKey outputs a new PeerJournalEntryBLSPublicKey implementation used to revert a round change
func NewPeerJournalEntryBLSPublicKey(account *PeerAccount, oldBLSPubKey []byte) (*PeerJournalEntryBLSPublicKey, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &PeerJournalEntryBLSPublicKey{
		account:      account,
		oldBLSPubKey: oldBLSPubKey,
	}, nil
}

// Revert applies undo operation
func (jen *PeerJournalEntryBLSPublicKey) Revert() (AccountHandler, error) {
	jen.account.BLSPublicKey = jen.oldBLSPubKey

	return jen.account, nil
}

//------- PeerJournalEntryStake

// PeerJournalEntryStake is used to revert a stake change
type PeerJournalEntryStake struct {
	account  *PeerAccount
	oldStake *big.Int
}

// NewPeerJournalEntryStake outputs a new PeerJournalEntryStake implementation used to revert a stake change
func NewPeerJournalEntryStake(account *PeerAccount, oldStake *big.Int) (*PeerJournalEntryStake, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &PeerJournalEntryStake{
		account:  account,
		oldStake: oldStake,
	}, nil
}

// Revert applies undo operation
func (jeb *PeerJournalEntryStake) Revert() (AccountHandler, error) {
	jeb.account.Stake = jeb.oldStake

	return jeb.account, nil
}

// PeerJournalEntryJailTime is used to revert a balance change
type PeerJournalEntryJailTime struct {
	account     *PeerAccount
	oldJailTime TimePeriod
}

// NewPeerJournalEntryJailTime outputs a new PeerJournalEntryJailTime implementation used to revert a state change
func NewPeerJournalEntryJailTime(account *PeerAccount, oldJailTime TimePeriod) (*PeerJournalEntryJailTime, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &PeerJournalEntryJailTime{
		account:     account,
		oldJailTime: oldJailTime,
	}, nil
}

// Revert applies undo operation
func (jeb *PeerJournalEntryJailTime) Revert() (AccountHandler, error) {
	jeb.account.JailTime = jeb.oldJailTime

	return jeb.account, nil
}

// PeerJournalEntryCurrentShardId is used to revert a shardId change
type PeerJournalEntryCurrentShardId struct {
	account    *PeerAccount
	oldShardId uint32
}

// NewPeerJournalEntryCurrentShardId outputs a new PeerJournalEntryCurrentShardId implementation used to revert a state change
func NewPeerJournalEntryCurrentShardId(account *PeerAccount, oldShardId uint32) (*PeerJournalEntryCurrentShardId, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &PeerJournalEntryCurrentShardId{
		account:    account,
		oldShardId: oldShardId,
	}, nil
}

// Revert applies undo operation
func (jeb *PeerJournalEntryCurrentShardId) Revert() (AccountHandler, error) {
	jeb.account.CurrentShardId = jeb.oldShardId

	return jeb.account, nil
}

// PeerJournalEntryNextShardId is used to revert a shardId change
type PeerJournalEntryNextShardId struct {
	account    *PeerAccount
	oldShardId uint32
}

// NewPeerJournalEntryNextShardId outputs a new PeerJournalEntryNextShardId implementation used to revert a state change
func NewPeerJournalEntryNextShardId(account *PeerAccount, oldShardId uint32) (*PeerJournalEntryNextShardId, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &PeerJournalEntryNextShardId{
		account:    account,
		oldShardId: oldShardId,
	}, nil
}

// Revert applies undo operation
func (jeb *PeerJournalEntryNextShardId) Revert() (AccountHandler, error) {
	jeb.account.NextShardId = jeb.oldShardId

	return jeb.account, nil
}

// PeerJournalEntryInWaitingList is used to revert a shardId change
type PeerJournalEntryInWaitingList struct {
	account              *PeerAccount
	oldNodeInWaitingList bool
}

// NewPeerJournalEntryInWaitingList outputs a new PeerJournalEntryInWaitingList implementation used to revert a state change
func NewPeerJournalEntryInWaitingList(
	account *PeerAccount,
	oldNodeInWaitingList bool,
) (*PeerJournalEntryInWaitingList, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &PeerJournalEntryInWaitingList{
		account:              account,
		oldNodeInWaitingList: oldNodeInWaitingList,
	}, nil
}

// Revert applies undo operation
func (jeb *PeerJournalEntryInWaitingList) Revert() (AccountHandler, error) {
	jeb.account.NodeInWaitingList = jeb.oldNodeInWaitingList

	return jeb.account, nil
}

// PeerJournalEntryValidatorSuccessRate is used to revert a success rate change
type PeerJournalEntryValidatorSuccessRate struct {
	account                 *PeerAccount
	oldValidatorSuccessRate SignRate
}

// NewPeerJournalEntryValidatorSuccessRate outputs a new PeerJournalEntryValidatorSuccessRate implementation used to revert a state change
func NewPeerJournalEntryValidatorSuccessRate(
	account *PeerAccount,
	oldValidatorSuccessRate SignRate,
) (*PeerJournalEntryValidatorSuccessRate, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &PeerJournalEntryValidatorSuccessRate{
		account:                 account,
		oldValidatorSuccessRate: oldValidatorSuccessRate,
	}, nil
}

// Revert applies undo operation
func (jeb *PeerJournalEntryValidatorSuccessRate) Revert() (AccountHandler, error) {
	jeb.account.ValidatorSuccessRate = jeb.oldValidatorSuccessRate

	return jeb.account, nil
}

// PeerJournalEntryLeaderSuccessRate is used to revert a success rate change
type PeerJournalEntryLeaderSuccessRate struct {
	account              *PeerAccount
	oldLeaderSuccessRate SignRate
}

// NewPeerJournalEntryLeaderSuccessRate outputs a new PeerJournalEntryLeaderSuccessRate implementation used to revert a state change
func NewPeerJournalEntryLeaderSuccessRate(
	account *PeerAccount,
	oldLeaderSuccessRate SignRate,
) (*PeerJournalEntryLeaderSuccessRate, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &PeerJournalEntryLeaderSuccessRate{
		account:              account,
		oldLeaderSuccessRate: oldLeaderSuccessRate,
	}, nil
}

// Revert applies undo operation
func (jeb *PeerJournalEntryLeaderSuccessRate) Revert() (AccountHandler, error) {
	jeb.account.LeaderSuccessRate = jeb.oldLeaderSuccessRate

	return jeb.account, nil
}

// PeerJournalEntryRating is used to revert a rating change
type PeerJournalEntryRating struct {
	account   *PeerAccount
	oldRating uint32
}

// NewPeerJournalEntryRating outputs a new PeerJournalEntryRating implementation used to revert a state change
func NewPeerJournalEntryRating(account *PeerAccount, oldRating uint32) (*PeerJournalEntryRating, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &PeerJournalEntryRating{
		account:   account,
		oldRating: oldRating,
	}, nil
}

// Revert applies undo operation
func (jeb *PeerJournalEntryRating) Revert() (AccountHandler, error) {
	jeb.account.Rating = jeb.oldRating

	return jeb.account, nil
}
