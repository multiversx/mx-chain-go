package node

import (
	"context"
	"errors"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/config"
)

var ErrTrieOperationsLimit = errors.New("trie collection limit exceeded; use paginated retrieval where available")

// WithTrieOperations configures API collection limits before the node starts serving requests.
func WithTrieOperations(cfg config.TrieOperationsConfig) Option {
	return func(n *Node) error {
		n.trieOperationsConfig = cfg.WithDefaults()
		return nil
	}
}

type trieLeafSource interface {
	GetAllLeaves(*common.TrieIteratorChannels, context.Context) error
}

type trieCollectionBudget struct {
	bytesRemaining  uint64
	leavesRemaining uint64
}

func (b *trieCollectionBudget) chargeBytes(size uint64) error {
	if size > b.bytesRemaining {
		return ErrTrieOperationsLimit
	}
	b.bytesRemaining -= size
	return nil
}

func (b *trieCollectionBudget) chargeLeaf(leaf core.KeyValueHolder) error {
	if b.leavesRemaining == 0 {
		return ErrTrieOperationsLimit
	}
	b.leavesRemaining--
	// Account for hex encoding and per-entry overhead, including filtered leaves.
	if err := b.chargeBytes(128); err != nil {
		return err
	}
	if err := b.chargeBytes(2 * uint64(len(leaf.Key()))); err != nil {
		return err
	}
	return b.chargeBytes(2 * uint64(len(leaf.Value())))
}

func (n *Node) collectTrieLeaves(ctx context.Context, source trieLeafSource, visit func(core.KeyValueHolder, *trieCollectionBudget) error) error {
	n.trieOperationsOnce.Do(func() {
		n.trieOperationsConfig = n.trieOperationsConfig.WithDefaults()
		n.trieOperationsSlots = make(chan struct{}, n.trieOperationsConfig.MaxConcurrentRequests)
	})
	if ctx.Err() != nil {
		return ErrTrieOperationsTimeout
	}
	select {
	case n.trieOperationsSlots <- struct{}{}:
	default:
		return common.ErrTrieOperationsBusy
	}
	defer func() { <-n.trieOperationsSlots }()

	scanCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	channels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	if err := source.GetAllLeaves(channels, scanCtx); err != nil {
		return err
	}
	// Hold admission until the producer exits, including on callback/limit errors.
	defer func() {
		cancel()
		for range channels.LeavesChan {
		}
	}()
	budget := &trieCollectionBudget{bytesRemaining: n.trieOperationsConfig.MaxBytes, leavesRemaining: n.trieOperationsConfig.MaxLeaves}
	for {
		select {
		case <-scanCtx.Done():
			return ErrTrieOperationsTimeout
		case leaf, ok := <-channels.LeavesChan:
			if scanCtx.Err() != nil {
				return ErrTrieOperationsTimeout
			}
			if !ok {
				return channels.ErrChan.ReadFromChanNonBlocking()
			}
			if err := budget.chargeLeaf(leaf); err != nil {
				return err
			}
			if err := visit(leaf, budget); err != nil {
				return err
			}
		}
	}
}
