package block_test

import (
	"testing"

	block2 "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/stretchr/testify/assert"
)

//------- InterceptedHeader

//Check()

func TestInterceptedHeader_CheckNilHeaderShouldRetFalse(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()
	hdr.Header = nil

	assert.False(t, hdr.Check())
}

func TestInterceptedHeader_CheckNilPrevHashShouldRetFalse(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = nil
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.BlockBodyPeer
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)

	assert.False(t, hdr.Check())
}

func TestInterceptedHeader_CheckNilPubKeysBitmapShouldRetFalse(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = nil
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.BlockBodyPeer
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)

	assert.False(t, hdr.Check())
}

func TestInterceptedHeader_CheckNilBlockBodyHashShouldRetFalse(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = nil
	hdr.BlockBodyType = block2.BlockBodyPeer
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)

	assert.False(t, hdr.Check())
}

func TestInterceptedHeader_CheckInvalidBlockBodyPeerShouldRetFalse(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.BlockBodyPeer + 1
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)

	assert.False(t, hdr.Check())
}

func TestInterceptedHeader_CheckNilSignatureShouldRetFalse(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.BlockBodyPeer
	hdr.Signature = nil
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)

	assert.False(t, hdr.Check())
}

func TestInterceptedHeader_CheckNilCommitmentShouldRetFalse(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.BlockBodyPeer
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = nil
	hdr.RootHash = make([]byte, 0)

	assert.False(t, hdr.Check())
}

func TestInterceptedHeader_CheckNilRootHashShouldRetFalse(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.BlockBodyPeer
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = nil

	assert.False(t, hdr.Check())
}

func TestInterceptedHeader_CheckOkValsShouldRetTrue(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.BlockBodyPeer
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)

	assert.True(t, hdr.Check())
}

//Getters and Setters

func TestInterceptedHeader_AllGettersAndSettersShouldWork(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.BlockBodyPeer
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.ShardId = 56

	hash := []byte("aaaa")
	hdr.SetHash(hash)
	assert.Equal(t, hash, hdr.Hash())
	assert.Equal(t, string(hash), hdr.ID())

	newHdr := hdr.New()
	assert.NotNil(t, newHdr)
	assert.NotNil(t, newHdr.(*block.InterceptedHeader).Header)

	assert.Equal(t, uint32(56), hdr.Shard())
}

//------- InterceptedPeerBlockBody

//Check

func TestInterceptedPeerBlockBody_CheckNilPeerBlodyShouldRetTrue(t *testing.T) {
	t.Parallel()

	inBB := block.NewInterceptedPeerBlockBody()
	inBB.PeerBlockBody = nil
	assert.False(t, inBB.Check())
}

func TestInterceptedPeerBlockBody_CheckNilChangesShouldRetFalse(t *testing.T) {
	t.Parallel()

	inBB := block.NewInterceptedPeerBlockBody()
	inBB.Changes = nil
	assert.False(t, inBB.Check())
}

func TestInterceptedPeerBlockBody_CheckEmptyChangessShouldRetFalse(t *testing.T) {
	t.Parallel()

	inBB := block.NewInterceptedPeerBlockBody()
	inBB.Changes = make([]block2.PeerChange, 0)
	assert.False(t, inBB.Check())
}

func TestInterceptedPeerBlockBody_CheckEmptyPubKeyListShouldRetFalse(t *testing.T) {
	t.Parallel()

	inBB := block.NewInterceptedPeerBlockBody()

	change := block2.PeerChange{}

	inBB.Changes = make([]block2.PeerChange, 0)
	inBB.Changes = append(inBB.Changes, change)

	assert.False(t, inBB.Check())
}

func TestInterceptedPeerBlockBody_CheckOkValsShouldRetTrue(t *testing.T) {
	t.Parallel()

	inBB := block.NewInterceptedPeerBlockBody()

	change := block2.PeerChange{}
	change.PubKey = []byte{65}

	inBB.Changes = make([]block2.PeerChange, 0)
	inBB.Changes = append(inBB.Changes, change)

	assert.True(t, inBB.Check())
}

//Getters and Setters

func TestInterceptedPeerBlockBody_AllGettersAndSettersShouldWork(t *testing.T) {
	t.Parallel()

	peerBlockBody := block.NewInterceptedPeerBlockBody()

	peerBlockBody.PeerBlockBody.ShardId = 45
	peerBlockBody.PeerBlockBody.Changes = make([]block2.PeerChange, 0)

	assert.Equal(t, uint32(45), peerBlockBody.Shard())
	assert.Equal(t, 0, len(peerBlockBody.Changes))

	hash := []byte("aaaa")
	peerBlockBody.SetHash(hash)
	assert.Equal(t, hash, peerBlockBody.Hash())
	assert.Equal(t, string(hash), peerBlockBody.ID())

	newPeerBB := peerBlockBody.New()
	assert.NotNil(t, newPeerBB)
	assert.NotNil(t, newPeerBB.(*block.InterceptedPeerBlockBody).PeerBlockBody)

	assert.Equal(t, uint32(45), peerBlockBody.Shard())
}

//------- InterceptedStateBlockBody

//Check

func TestInterceptedStateBlockBody_CheckNilStateBlodyShouldRetTrue(t *testing.T) {
	t.Parallel()

	inBB := block.NewInterceptedStateBlockBody()
	inBB.StateBlockBody = nil
	assert.False(t, inBB.Check())
}

func TestInterceptedStateBlockBody_CheckNilRootHashShouldRetFalse(t *testing.T) {
	t.Parallel()

	inBB := block.NewInterceptedStateBlockBody()
	inBB.RootHash = nil
	assert.False(t, inBB.Check())
}

func TestInterceptedStateBlockBody_CheckEmptyRootHashShouldRetFalse(t *testing.T) {
	t.Parallel()

	inBB := block.NewInterceptedStateBlockBody()
	inBB.RootHash = make([]byte, 0)
	assert.False(t, inBB.Check())
}

func TestInterceptedStateBlockBody_CheckOkValsShouldRetTrue(t *testing.T) {
	t.Parallel()

	inBB := block.NewInterceptedStateBlockBody()

	inBB.RootHash = []byte("aaaa")
	assert.True(t, inBB.Check())
}

//Getters and Setters

func TestInterceptedStateBlockBody_AllGettersAndSettersShouldWork(t *testing.T) {
	t.Parallel()

	stateBlockBody := block.NewInterceptedStateBlockBody()

	stateBlockBody.ShardId = 45
	stateBlockBody.RootHash = []byte("aaa")

	assert.Equal(t, uint32(45), stateBlockBody.Shard())

	hash := []byte("aaaa")
	stateBlockBody.SetHash(hash)
	assert.Equal(t, hash, stateBlockBody.Hash())
	assert.Equal(t, string(hash), stateBlockBody.ID())

	newBB := stateBlockBody.New()
	assert.NotNil(t, newBB)
	assert.NotNil(t, newBB.(*block.InterceptedStateBlockBody).StateBlockBody)

	assert.Equal(t, uint32(45), stateBlockBody.Shard())
}

//------- InterceptedTxBlock

//Check

func TestInterceptedTxBlockBody_CheckNilTxBlodyShouldRetTrue(t *testing.T) {
	t.Parallel()

	inBB := block.NewInterceptedTxBlockBody()
	inBB.TxBlockBody = nil
	assert.False(t, inBB.Check())
}

func TestInterceptedTxBlockBody_CheckNilMiniBlocksShouldRetFalse(t *testing.T) {
	t.Parallel()

	inBB := block.NewInterceptedTxBlockBody()
	inBB.MiniBlocks = nil
	assert.False(t, inBB.Check())
}

func TestInterceptedTxBlockBody_CheckEmptyMiniBlocksShouldRetFalse(t *testing.T) {
	t.Parallel()

	inBB := block.NewInterceptedTxBlockBody()
	inBB.MiniBlocks = make([]block2.MiniBlock, 0)
	assert.False(t, inBB.Check())
}

func TestInterceptedTxBlockBody_CheckEmptyMiniBlockShouldRetFalse(t *testing.T) {
	t.Parallel()

	inBB := block.NewInterceptedTxBlockBody()

	miniBlock := block2.MiniBlock{}

	inBB.MiniBlocks = make([]block2.MiniBlock, 0)
	inBB.MiniBlocks = append(inBB.MiniBlocks, miniBlock)
	assert.False(t, inBB.Check())
}

func TestInterceptedTxBlockBody_CheckOkValsShouldRetTrue(t *testing.T) {
	t.Parallel()

	inBB := block.NewInterceptedTxBlockBody()

	miniBlock := block2.MiniBlock{}
	miniBlock.TxHashes = append(miniBlock.TxHashes, []byte{65})

	inBB.MiniBlocks = make([]block2.MiniBlock, 0)
	inBB.MiniBlocks = append(inBB.MiniBlocks, miniBlock)
	assert.True(t, inBB.Check())
}

//Getters and Setters

func TestInterceptedTxBlockBody_AllGettersAndSettersShouldWork(t *testing.T) {
	t.Parallel()

	txBlockBody := block.NewInterceptedTxBlockBody()

	txBlockBody.TxBlockBody.ShardId = 45
	txBlockBody.TxBlockBody.MiniBlocks = make([]block2.MiniBlock, 0)

	assert.Equal(t, uint32(45), txBlockBody.Shard())
	assert.Equal(t, 0, len(txBlockBody.MiniBlocks))

	hash := []byte("aaaa")
	txBlockBody.SetHash(hash)
	assert.Equal(t, hash, txBlockBody.Hash())
	assert.Equal(t, string(hash), txBlockBody.ID())

	newTxBB := txBlockBody.New()
	assert.NotNil(t, newTxBB)
	assert.NotNil(t, newTxBB.(*block.InterceptedTxBlockBody).TxBlockBody)

	assert.Equal(t, uint32(45), txBlockBody.Shard())
}
