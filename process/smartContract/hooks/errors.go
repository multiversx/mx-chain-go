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
