package block

import "errors"

var errInvalidNumOutGoingMBInMetaHdrProposal = errors.New("invalid number of outgoing miniblocks in meta header proposal, should be zero")

var errInvalidNumOutGoingTxsInMetaHdrProposal = errors.New("invalid number of outgoing transactions in meta header proposal, should be zero")

var errInvalidNumPendingMiniBlocksInMetaHdrProposal = errors.New("invalid number of pending miniblocks in meta header proposal, should be zero")
