package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// HeaderWrapper is a wrapper for StateBlockBody, adding functionality for validity and integrity checks
// and as well for signature verification
type HeaderWrapper struct {
	*block.Header
}

// StateBlockBodyWrapper is a wrapper for StateBlockBody, adding functionality for validity and integrity checks
type StateBlockBodyWrapper struct {
	*block.StateBlockBody
}

// PeerBlockBodyWrapper is a wrapper for StateBlockBody, adding functionality for validity and integrity checks
type PeerBlockBodyWrapper struct {
	*block.PeerBlockBody
}

// TxBlockBodyWrapper is a wrapper for StateBlockBody, adding functionality for validity and integrity checks
type TxBlockBodyWrapper struct {
	*block.TxBlockBody
}

// InterceptedHeader represents the wrapper over HeaderWrapper struct.
// It implements Newer and Hashed interfaces
type InterceptedHeader struct {
	*HeaderWrapper
	hash []byte
}

// InterceptedPeerBlockBody represents the wrapper over PeerBlockBodyWrapper struct.
type InterceptedPeerBlockBody struct {
	*PeerBlockBodyWrapper
	hash []byte
}

// InterceptedStateBlockBody represents the wrapper over InterceptedStateBlockBody struct.
type InterceptedStateBlockBody struct {
	*StateBlockBodyWrapper
	hash []byte
}

// InterceptedTxBlockBody represents the wrapper over TxBlockBodyWrapper struct.
type InterceptedTxBlockBody struct {
	*TxBlockBodyWrapper
	hash []byte
}

//------- HeaderWrapper

// NewHeaderWrapper creates a new HeaderWrapper instance
func NewHeaderWrapper() *HeaderWrapper {
	return &HeaderWrapper{Header: &block.Header{}}
}

// Integrity checks the integrity of the state block wrapper
func (hWrapper *HeaderWrapper) Integrity(coordinator sharding.ShardCoordinator) error {
	if coordinator == nil {
		return process.ErrNilShardCoordinator
	}

	if hWrapper.Header == nil {
		return process.ErrNilBlockHeader
	}

	if hWrapper.BlockBodyHash == nil {
		return process.ErrNilBlockBodyHash
	}

	if hWrapper.PubKeysBitmap == nil {
		return process.ErrNilPubKeysBitmap
	}

	if hWrapper.ShardId >= coordinator.NoShards() {
		return process.ErrInvalidShardId
	}

	if hWrapper.PrevHash == nil {
		return process.ErrNilPreviousBlockHash
	}

	if hWrapper.Signature == nil {
		return process.ErrNilSignature
	}

	if hWrapper.Commitment == nil {
		return process.ErrNilCommitment
	}

	switch hWrapper.BlockBodyType {
	case block.PeerBlock:
	case block.StateBlock:
	case block.TxBlock:
	default:
		return process.ErrInvalidBlockBodyType
	}

	return nil
}

// IntegrityAndValidity checks the integrity and validity of a block header wrapper
func (hWrapper *HeaderWrapper) IntegrityAndValidity(coordinator sharding.ShardCoordinator) error {
	err := hWrapper.Integrity(coordinator)
	if err != nil {
		return err
	}

	return hWrapper.validityCheck()
}

func (hWrapper *HeaderWrapper) validityCheck() error {
	// TODO: need to check epoch is round - timestamp - epoch - nonce - requires chronology
	return nil
}

// VerifySig verifies a signature
func (hWrapper *HeaderWrapper) VerifySig() error {
	// TODO: Check block signature after multisig will be implemented
	return nil
}

//------- StateBlockBodyWrapper

// NewStateBlockBodyWrapper creates a new StateBlockBodyWrapper instance
func NewStateBlockBodyWrapper() *StateBlockBodyWrapper {
	return &StateBlockBodyWrapper{StateBlockBody: &block.StateBlockBody{}}
}

// Integrity checks the integrity of the state block
func (sbWrapper StateBlockBodyWrapper) Integrity(coordinator sharding.ShardCoordinator) error {
	if coordinator == nil {
		return process.ErrNilShardCoordinator
	}

	if sbWrapper.StateBlockBody == nil {
		return process.ErrNilStateBlockBody
	}

	if sbWrapper.ShardID >= coordinator.NoShards() {
		return process.ErrInvalidShardId
	}

	if sbWrapper.RootHash == nil {
		return process.ErrNilRootHash
	}

	return nil
}

// IntegrityAndValidity checks the integrity and validity of a state block wrapper
func (sbWrapper StateBlockBodyWrapper) IntegrityAndValidity(coordinator sharding.ShardCoordinator) error {
	err := sbWrapper.Integrity(coordinator)

	if err != nil {
		return err
	}

	return sbWrapper.validityCheck(coordinator)
}

func (sbWrapper StateBlockBodyWrapper) validityCheck(coordinator sharding.ShardCoordinator) error {
	return nil
}

//------- PeerBlockBodyWrapper

// NewPeerBlockBodyWrapper creates a new PeerBlockBodyWrapper instance
func NewPeerBlockBodyWrapper() *PeerBlockBodyWrapper {
	return &PeerBlockBodyWrapper{PeerBlockBody: &block.PeerBlockBody{}}
}

// Integrity checks the integrity of the state block wrapper
func (pbWrapper PeerBlockBodyWrapper) Integrity(coordinator sharding.ShardCoordinator) error {
	if coordinator == nil {
		return process.ErrNilShardCoordinator
	}

	if pbWrapper.PeerBlockBody == nil {
		return process.ErrNilPeerBlockBody
	}

	stateBlockWrapper := StateBlockBodyWrapper{
		StateBlockBody: &pbWrapper.StateBlockBody,
	}

	err := stateBlockWrapper.Integrity(coordinator)
	if err != nil {
		return err
	}

	if pbWrapper.Changes == nil {
		return process.ErrNilPeerChanges
	}

	for _, change := range pbWrapper.Changes {
		if change.ShardIdDest >= coordinator.NoShards() {
			return process.ErrInvalidShardId
		}

		if change.PubKey == nil {
			return process.ErrNilPublicKey
		}
	}

	return nil
}

// IntegrityAndValidity checks the integrity and validity of a peer block wrapper
func (pbWrapper PeerBlockBodyWrapper) IntegrityAndValidity(coordinator sharding.ShardCoordinator) error {
	err := pbWrapper.Integrity(coordinator)
	if err != nil {
		return err
	}

	return pbWrapper.validityCheck()
}

func (pbWrapper PeerBlockBodyWrapper) validityCheck() error {
	// TODO: check that the peer changes received are equal with what has been calculated

	return nil
}

//------- TxBlockBodyWrapper

// NewTxBlockBodyWrapper creates a new PeerBlockBodyWrapper instance
func NewTxBlockBodyWrapper() *TxBlockBodyWrapper {
	return &TxBlockBodyWrapper{TxBlockBody: &block.TxBlockBody{}}
}

// Integrity checks the integrity of the state block wrapper
func (txbWrapper TxBlockBodyWrapper) Integrity(coordinator sharding.ShardCoordinator) error {
	if coordinator == nil {
		return process.ErrNilShardCoordinator
	}

	if txbWrapper.TxBlockBody == nil {
		return process.ErrNilTxBlockBody
	}

	stateBlockWrapper := StateBlockBodyWrapper{
		StateBlockBody: &txbWrapper.StateBlockBody,
	}

	err := stateBlockWrapper.Integrity(coordinator)
	if err != nil {
		return err
	}

	if txbWrapper.MiniBlocks == nil {
		return process.ErrNilMiniBlocks
	}

	for _, miniBlock := range txbWrapper.MiniBlocks {
		if miniBlock.TxHashes == nil {
			return process.ErrNilTxHashes
		}

		if miniBlock.ShardID >= coordinator.NoShards() {
			return process.ErrInvalidShardId
		}

		for _, txHash := range miniBlock.TxHashes {
			if txHash == nil {
				return process.ErrNilTxHash
			}
		}
	}

	return nil
}

// IntegrityAndValidity checks the integrity of a transactions block
func (txbWrapper TxBlockBodyWrapper) IntegrityAndValidity(coordinator sharding.ShardCoordinator) error {
	err := txbWrapper.Integrity(coordinator)

	if err != nil {
		return err
	}

	return txbWrapper.validityCheck()
}

func (txbWrapper TxBlockBodyWrapper) validityCheck() error {
	// TODO: update with validity checks

	return nil
}

//------- InterceptedHeader

// NewInterceptedHeader creates a new instance of InterceptedHeader struct
func NewInterceptedHeader() *InterceptedHeader {
	return &InterceptedHeader{
		HeaderWrapper: NewHeaderWrapper(),
	}
}

// SetHash sets the hash of this header. The hash will also be the ID of this object
func (inHdr *InterceptedHeader) SetHash(hash []byte) {
	inHdr.hash = hash
}

// Hash gets the hash of this header
func (inHdr *InterceptedHeader) Hash() []byte {
	return inHdr.hash
}

// Create returns a new instance of this struct (used in topics)
func (inHdr *InterceptedHeader) Create() p2p.Creator {
	return NewInterceptedHeader()
}

// ID returns the ID of this object. Set to return the hash of the header
func (inHdr *InterceptedHeader) ID() string {
	return string(inHdr.hash)
}

// Shard returns the shard ID for which this header is addressed
func (inHdr *InterceptedHeader) Shard() uint32 {
	return inHdr.ShardId
}

// GetHeader returns the Header pointer that holds the data
func (inHdr *InterceptedHeader) GetHeader() *block.Header {
	return inHdr.Header
}

//------- InterceptedPeerBlockBody

// NewInterceptedPeerBlockBody creates a new instance of InterceptedPeerBlockBody struct
func NewInterceptedPeerBlockBody() *InterceptedPeerBlockBody {
	return &InterceptedPeerBlockBody{
		PeerBlockBodyWrapper: NewPeerBlockBodyWrapper(),
	}
}

// SetHash sets the hash of this peer block body. The hash will also be the ID of this object
func (inPeerBlkBdy *InterceptedPeerBlockBody) SetHash(hash []byte) {
	inPeerBlkBdy.hash = hash
}

// Hash gets the hash of this peer block body
func (inPeerBlkBdy *InterceptedPeerBlockBody) Hash() []byte {
	return inPeerBlkBdy.hash
}

// Create returns a new instance of this struct (used in topics)
func (inPeerBlkBdy *InterceptedPeerBlockBody) Create() p2p.Creator {
	return NewInterceptedPeerBlockBody()
}

// ID returns the ID of this object. Set to return the hash of the peer block body
func (inPeerBlkBdy *InterceptedPeerBlockBody) ID() string {
	return string(inPeerBlkBdy.hash)
}

// Shard returns the shard ID for which this body is addressed
func (inPeerBlkBdy *InterceptedPeerBlockBody) Shard() uint32 {
	return inPeerBlkBdy.ShardID
}

//------- InterceptedStateBlockBody

// NewInterceptedStateBlockBody creates a new instance of InterceptedStateBlockBody struct
func NewInterceptedStateBlockBody() *InterceptedStateBlockBody {
	return &InterceptedStateBlockBody{
		StateBlockBodyWrapper: NewStateBlockBodyWrapper()}
}

// SetHash sets the hash of this state block body. The hash will also be the ID of this object
func (inStateBlkBdy *InterceptedStateBlockBody) SetHash(hash []byte) {
	inStateBlkBdy.hash = hash
}

// Hash gets the hash of this state block body
func (inStateBlkBdy *InterceptedStateBlockBody) Hash() []byte {
	return inStateBlkBdy.hash
}

// Create returns a new instance of this struct (used in topics)
func (inStateBlkBdy *InterceptedStateBlockBody) Create() p2p.Creator {
	return NewInterceptedStateBlockBody()
}

// ID returns the ID of this object. Set to return the hash of the state block body
func (inStateBlkBdy *InterceptedStateBlockBody) ID() string {
	return string(inStateBlkBdy.hash)
}

// Shard returns the shard ID for which this body is addressed
func (inStateBlkBdy *InterceptedStateBlockBody) Shard() uint32 {
	return inStateBlkBdy.ShardID
}

//------- InterceptedTxBlockBody

// NewInterceptedTxBlockBody creates a new instance of InterceptedTxBlockBody struct
func NewInterceptedTxBlockBody() *InterceptedTxBlockBody {
	return &InterceptedTxBlockBody{
		TxBlockBodyWrapper: NewTxBlockBodyWrapper()}
}

// SetHash sets the hash of this transaction block body. The hash will also be the ID of this object
func (inTxBlkBdy *InterceptedTxBlockBody) SetHash(hash []byte) {
	inTxBlkBdy.hash = hash
}

// Hash gets the hash of this transaction block body
func (inTxBlkBdy *InterceptedTxBlockBody) Hash() []byte {
	return inTxBlkBdy.hash
}

// Create returns a new instance of this struct (used in topics)
func (inTxBlkBdy *InterceptedTxBlockBody) Create() p2p.Creator {
	return NewInterceptedTxBlockBody()
}

// ID returns the ID of this object. Set to return the hash of the transaction block body
func (inTxBlkBdy *InterceptedTxBlockBody) ID() string {
	return string(inTxBlkBdy.hash)
}

// Shard returns the shard ID for which this body is addressed
func (inTxBlkBdy *InterceptedTxBlockBody) Shard() uint32 {
	return inTxBlkBdy.ShardID
}
