package factory

import "errors"

// ErrorNilDelegatedListHandler signal that returned account is wrong
var ErrorNilDelegatedListHandler = errors.New("nil DelegatedListHandler")

// ErrorNilDirectStakedListHandler signal that returned account is wrong
var ErrorNilDirectStakedListHandler = errors.New("nil DirectStakedListHandler")
