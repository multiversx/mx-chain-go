package txcache

import "errors"

var errNilMempoolHost = errors.New("nil mempool host")
var errNilSelectionSession = errors.New("nil selection session")
var errItemAlreadyInCache = errors.New("item already in cache")
var errEmptyBunchOfTransactions = errors.New("empty bunch of transactions")
var errNilBlockBody = errors.New("nil block body")
var errNilHeaderHandler = errors.New("nil header handler")
var errNilBlockHash = errors.New("nil block hash")
var errPreviousBlockNotFound = errors.New("previous block not found")
var errDiscontinuousBlockNonce = errors.New("discontinuous block nonce")
var errNilTxCache = errors.New("nil tx cache")
var errNotFoundTx = errors.New("tx not found")
var errDiscontinuousBreadcrumbs = errors.New("discontinuous breadcrumbs")
var errNonceGap = errors.New("nonce gap")
var errExceededBalance = errors.New("exceeded balance")
var errNonceNotSet = errors.New("nonce not set")
var errReceivedLastNonceNotSet = errors.New("received last nonce not set")
