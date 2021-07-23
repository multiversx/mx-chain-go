package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/state/temporary"
)

// CheckpointHashesHolderStub -
type CheckpointHashesHolderStub struct {
	PutCalled             func([]byte, temporary.ModifiedHashes) bool
	RemoveCommittedCalled func([]byte)
	RemoveCalled          func([]byte)
	ShouldCommitCalled    func([]byte) bool
}

// Put -
func (c *CheckpointHashesHolderStub) Put(rootHash []byte, hashes temporary.ModifiedHashes) bool {
	if c.PutCalled != nil {
		return c.PutCalled(rootHash, hashes)
	}

	return false
}

// RemoveCommitted -
func (c *CheckpointHashesHolderStub) RemoveCommitted(lastCommittedRootHash []byte) {
	if c.RemoveCommittedCalled != nil {
		c.RemoveCommittedCalled(lastCommittedRootHash)
	}
}

// Remove -
func (c *CheckpointHashesHolderStub) Remove(hash []byte) {
	if c.RemoveCalled != nil {
		c.RemoveCalled(hash)
	}
}

// ShouldCommit -
func (c *CheckpointHashesHolderStub) ShouldCommit(hash []byte) bool {
	if c.ShouldCommitCalled != nil {
		return c.ShouldCommitCalled(hash)
	}

	return true
}

// IsInterfaceNil -
func (c *CheckpointHashesHolderStub) IsInterfaceNil() bool {
	return c == nil
}
