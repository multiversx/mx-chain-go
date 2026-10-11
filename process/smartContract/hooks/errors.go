package hooks

import "errors"

// ErrNotImplemented signals that a functionality can not be used as it is not implemented
var ErrNotImplemented = errors.New("not implemented")

// ErrEmptyCode signals that an account does not contain code
var ErrEmptyCode = errors.New("empty code in provided smart contract holding account")

// ErrAddressLengthNotCorrect signals that an account does not have the correct address
var ErrAddressLengthNotCorrect = errors.New("address length is not correct")

// ErrVMTypeLengthIsNotCorrect signals that the vm type length is not correct
var ErrVMTypeLengthIsNotCorrect = errors.New("vm type length is not correct")

// ErrNilBlockchainHookCounter signals that a nil blockchain hook counter was provided
var ErrNilBlockchainHookCounter = errors.New("nil blockchain hook counter")

// ErrNilMissingTrieNodesNotifier signals that a nil missing trie nodes notifier was provided
var ErrNilMissingTrieNodesNotifier = errors.New("nil missing trie nodes notifier")

// ErrNilEpochStartTriggerHandler signals that a nil epoch start trigger handler was provided
var ErrNilEpochStartTriggerHandler = errors.New("nil epoch start trigger handler")

// ErrNilRoundHandler signals that a nil round handler was provided
var ErrNilRoundHandler = errors.New("nil round handler")

// ErrNilCurrentHeader signals that a nil current header was provided
var ErrNilCurrentHeader = errors.New("nil current header")

// ErrNilEpochStartHeader signals that a nil epoch start header was provided
var ErrNilEpochStartHeader = errors.New("nil epoch start header")

// ErrNilLastCommitedEpochStartHdr signals that a nil last committed epoch start header was provided
var ErrNilLastCommitedEpochStartHdr = errors.New("nil last committed epoch start header")

// ErrLastCommitedEpochStartHdrMismatch signals that the current header epoch and last committed epoch start header epoch do not match
var ErrLastCommitedEpochStartHdrMismatch = errors.New("current header epoch and last committed epoch start header epoch do not match")
