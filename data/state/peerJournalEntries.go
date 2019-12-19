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
func (pje *PeerJournalEntryAddress) Revert() (AccountHandler, error) {
	pje.account.Address = pje.oldAddress

	return pje.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pje *PeerJournalEntryAddress) IsInterfaceNil() bool {
	if pje == nil {
		return true
	}
	return false
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
func (jens *PeerJournalEntrySchnorrPublicKey) Revert() (AccountHandler, error) {
	jens.account.SchnorrPublicKey = jens.oldSchnorrPubKey

	return jens.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (jens *PeerJournalEntrySchnorrPublicKey) IsInterfaceNil() bool {
	if jens == nil {
		return true
	}
	return false
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
func (pjeb *PeerJournalEntryBLSPublicKey) Revert() (AccountHandler, error) {
	pjeb.account.BLSPublicKey = pjeb.oldBLSPubKey

	return pjeb.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pjeb *PeerJournalEntryBLSPublicKey) IsInterfaceNil() bool {
	if pjeb == nil {
		return true
	}
	return false
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
func (pjes *PeerJournalEntryStake) Revert() (AccountHandler, error) {
	pjes.account.Stake = pjes.oldStake

	return pjes.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pjes *PeerJournalEntryStake) IsInterfaceNil() bool {
	if pjes == nil {
		return true
	}
	return false
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
func (pjej *PeerJournalEntryJailTime) Revert() (AccountHandler, error) {
	pjej.account.JailTime = pjej.oldJailTime

	return pjej.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pjej *PeerJournalEntryJailTime) IsInterfaceNil() bool {
	if pjej == nil {
		return true
	}
	return false
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
func (pjec *PeerJournalEntryCurrentShardId) Revert() (AccountHandler, error) {
	pjec.account.CurrentShardId = pjec.oldShardId

	return pjec.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pjec *PeerJournalEntryCurrentShardId) IsInterfaceNil() bool {
	if pjec == nil {
		return true
	}
	return false
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
func (pjen *PeerJournalEntryNextShardId) Revert() (AccountHandler, error) {
	pjen.account.NextShardId = pjen.oldShardId

	return pjen.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pjen *PeerJournalEntryNextShardId) IsInterfaceNil() bool {
	if pjen == nil {
		return true
	}
	return false
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
func (pjew *PeerJournalEntryInWaitingList) Revert() (AccountHandler, error) {
	pjew.account.NodeInWaitingList = pjew.oldNodeInWaitingList

	return pjew.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pjew *PeerJournalEntryInWaitingList) IsInterfaceNil() bool {
	if pjew == nil {
		return true
	}
	return false
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
func (pjev *PeerJournalEntryValidatorSuccessRate) Revert() (AccountHandler, error) {
	pjev.account.ValidatorSuccessRate = pjev.oldValidatorSuccessRate

	return pjev.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pjev *PeerJournalEntryValidatorSuccessRate) IsInterfaceNil() bool {
	if pjev == nil {
		return true
	}
	return false
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
func (pjel *PeerJournalEntryLeaderSuccessRate) Revert() (AccountHandler, error) {
	pjel.account.LeaderSuccessRate = pjel.oldLeaderSuccessRate

	return pjel.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pjel *PeerJournalEntryLeaderSuccessRate) IsInterfaceNil() bool {
	if pjel == nil {
		return true
	}
	return false
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
func (pjer *PeerJournalEntryRating) Revert() (AccountHandler, error) {
	pjer.account.Rating = pjer.oldRating

	return pjer.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pjer *PeerJournalEntryRating) IsInterfaceNil() bool {
	if pjer == nil {
		return true
	}
	return false
}

// PeerJournalEntryTempRating is used to revert a rating change
type PeerJournalEntryTempRating struct {
	account       *PeerAccount
	oldTempRating uint32
}

// NewPeerJournalEntryRating outputs a new PeerJournalEntryRating implementation used to revert a state change
func NewPeerJournalEntryTempRating(account *PeerAccount, oldTempRating uint32) (*PeerJournalEntryTempRating, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &PeerJournalEntryTempRating{
		account:       account,
		oldTempRating: oldTempRating,
	}, nil
}

// Revert applies undo operation
func (pjer *PeerJournalEntryTempRating) Revert() (AccountHandler, error) {
	pjer.account.TempRating = pjer.oldTempRating

	return pjer.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pjer *PeerJournalEntryTempRating) IsInterfaceNil() bool {
	return pjer == nil
}

// PeerJournalEntryUnStakedNonce is used to revert a unstaked nonce change
type PeerJournalEntryUnStakedNonce struct {
	account          *PeerAccount
	oldUnStakedNonce uint64
}

// PeerJournalEntryUnStakedNonce outputs a new PeerJournalEntryCurrentShardId implementation used to revert a state change
func NewPeerJournalEntryUnStakedNonce(account *PeerAccount, oldUnStakedNonce uint64) (*PeerJournalEntryUnStakedNonce, error) {
	if account == nil {
		return nil, ErrNilAccountHandler
	}

	return &PeerJournalEntryUnStakedNonce{
		account:          account,
		oldUnStakedNonce: oldUnStakedNonce,
	}, nil
}

// Revert applies undo operation
func (pjec *PeerJournalEntryUnStakedNonce) Revert() (AccountHandler, error) {
	pjec.account.UnStakedNonce = pjec.oldUnStakedNonce

	return pjec.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pjec *PeerJournalEntryUnStakedNonce) IsInterfaceNil() bool {
	if pjec == nil {
		return true
	}
	return false
}
