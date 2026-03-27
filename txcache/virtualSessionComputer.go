package txcache

import (
	"math/big"
	"runtime"

	"github.com/multiversx/mx-chain-core-go/core"
	"golang.org/x/sync/errgroup"
)

type virtualSessionComputer struct {
	session                  SelectionSession
	virtualAccountsByAddress map[string]*virtualAccountRecord
}

func newVirtualSessionComputer(session SelectionSession) *virtualSessionComputer {
	return &virtualSessionComputer{
		session:                  session,
		virtualAccountsByAddress: make(map[string]*virtualAccountRecord),
	}
}

// createVirtualSelectionSession iterates over the global breadcrumbs of the selection tracker.
// If the global breadcrumb of an account is continuous with the session nonce,
// the virtual record of that account is created or updated.
// NOTE: The createVirtualSelectionSession method should receive a deep copy of the globalAccountBreadcrumbs because it mutates them.
func (computer *virtualSessionComputer) createVirtualSelectionSession(
	globalAccountBreadcrumbs map[string]*globalAccountBreadcrumb,
) (*virtualSelectionSession, error) {
	err := computer.handleGlobalAccountBreadcrumbs(globalAccountBreadcrumbs)
	if err != nil {
		return nil, err
	}

	virtualSession := newVirtualSelectionSession(computer.session, computer.virtualAccountsByAddress)
	log.Debug("virtualSessionComputer.createVirtualSelectionSession",
		"num of global accounts breadcrumbs", len(globalAccountBreadcrumbs),
		"num of actual virtual records", len(virtualSession.virtualAccountsByAddress),
	)
	return virtualSession, nil
}

const parallelBreadcrumbThreshold = 16

// handleGlobalAccountBreadcrumbs processes each account breadcrumb into a virtual record.
// For >= 16 accounts, fetches state and builds records in parallel using errgroup.
func (computer *virtualSessionComputer) handleGlobalAccountBreadcrumbs(
	globalAccountBreadcrumbs map[string]*globalAccountBreadcrumb,
) error {
	if len(globalAccountBreadcrumbs) < parallelBreadcrumbThreshold {
		for addr, bc := range globalAccountBreadcrumbs {
			nonce, balance, _, err := computer.session.GetAccountNonceAndBalance([]byte(addr))
			if err != nil {
				return err
			}
			record, err := computer.buildRecord(nonce, balance, bc)
			if err != nil {
				return err
			}
			computer.virtualAccountsByAddress[addr] = record
		}
		return nil
	}

	type entry struct {
		addr string
		bc   *globalAccountBreadcrumb
	}
	entries := make([]entry, 0, len(globalAccountBreadcrumbs))
	for addr, bc := range globalAccountBreadcrumbs {
		entries = append(entries, entry{addr, bc})
	}

	records := make([]*virtualAccountRecord, len(entries))
	g := new(errgroup.Group)
	g.SetLimit(runtime.GOMAXPROCS(0))

	for i := range entries {
		idx := i
		g.Go(func() error {
			nonce, balance, _, err := computer.session.GetAccountNonceAndBalance([]byte(entries[idx].addr))
			if err != nil {
				return err
			}
			rec, err := computer.buildRecord(nonce, balance, entries[idx].bc)
			if err != nil {
				return err
			}
			records[idx] = rec
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	for i, rec := range records {
		computer.virtualAccountsByAddress[entries[i].addr] = rec
	}
	return nil
}

// buildRecord creates a virtual account record from on-chain state and a breadcrumb.
// Pure computation — no shared state access, safe for concurrent use.
func (computer *virtualSessionComputer) buildRecord(
	accountNonce uint64,
	accountBalance *big.Int,
	bc *globalAccountBreadcrumb,
) (*virtualAccountRecord, error) {
	nonce := core.OptionalUint64{Value: accountNonce, HasValue: true}

	if !bc.isContinuousWithSessionNonce(accountNonce) {
		nonce = core.OptionalUint64{HasValue: false}
	} else if bc.isUser() {
		nonce = core.OptionalUint64{Value: bc.lastNonce.Value + 1, HasValue: true}
	}

	record, err := newVirtualAccountRecord(nonce, accountBalance)
	if err != nil {
		return nil, err
	}

	record.accumulateConsumedBalance(bc.consumedBalance)
	return record, nil
}
