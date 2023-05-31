package errors

import (
	"errors"
)

// ErrInvalidAppContext signals an invalid context passed to the routing system
var ErrInvalidAppContext = errors.New("invalid app context")

// ErrInvalidJSONRequest signals an error in json request formatting
var ErrInvalidJSONRequest = errors.New("invalid json request")

// ErrCouldNotGetAccount signals that a requested account could not be retrieved
var ErrCouldNotGetAccount = errors.New("could not get requested account")

// ErrGetBalance signals an error in getting the balance for an account
var ErrGetBalance = errors.New("get balance error")

// ErrGetUsername signals an error in getting the username for an account
var ErrGetUsername = errors.New("get username error")

// ErrGetCodeHash signals an error in getting the code hash for an account
var ErrGetCodeHash = errors.New("get code hash error")

// ErrGetValueForKey signals an error in getting the value of a key for an account
var ErrGetValueForKey = errors.New("get value for key error")

// ErrGetKeyValuePairs signals an error in getting the key-value pairs of a key for an account
var ErrGetKeyValuePairs = errors.New("get key-value pairs error")

// ErrGetESDTBalance signals an error in getting esdt balance for given address
var ErrGetESDTBalance = errors.New("get esdt balance for account error")

// ErrGetGuardianData signals an error in getting the guardian data for given address
var ErrGetGuardianData = errors.New("get guardian data for account error")

// ErrGetRolesForAccount signals an error in getting esdt tokens and roles for a given address
var ErrGetRolesForAccount = errors.New("get roles for account error")

// ErrGetESDTNFTData signals an error in getting esdt nft data for given address, tokenID and nonce
var ErrGetESDTNFTData = errors.New("get esdt nft data for account error")

// ErrEmptyAddress signals that an empty address was provided
var ErrEmptyAddress = errors.New("address is empty")

// ErrEmptyKey signals that an empty key was provided
var ErrEmptyKey = errors.New("key is empty")

// ErrEmptyTokenIdentifier signals that an empty token identifier was provided
var ErrEmptyTokenIdentifier = errors.New("token identifier is empty")

// ErrEmptyRole signals that an empty role was provided
var ErrEmptyRole = errors.New("role is empty")

// ErrNonceInvalid signals that nonce is invalid
var ErrNonceInvalid = errors.New("nonce is invalid")

// ErrValidation signals an error in validation
var ErrValidation = errors.New("validation error")

// ErrTxGenerationFailed signals an error generating a transaction
var ErrTxGenerationFailed = errors.New("transaction generation failed")

// ErrValidationEmptyTxHash signals that an empty tx hash was provided
var ErrValidationEmptyTxHash = errors.New("TxHash is empty")

// ErrInvalidBlockNonce signals that an invalid block nonce was provided
var ErrInvalidBlockNonce = errors.New("invalid block nonce")

// ErrInvalidBlockRound signals that an invalid block round was provided
var ErrInvalidBlockRound = errors.New("invalid block round")

// ErrInvalidEpoch signals that an invalid epoch parameter was provided
var ErrInvalidEpoch = errors.New("invalid epoch parameter")

// ErrValidationEmptyBlockHash signals an empty block hash was provided
var ErrValidationEmptyBlockHash = errors.New("block hash is empty")

// ErrGetTransaction signals an error happening when trying to fetch a transaction
var ErrGetTransaction = errors.New("getting transaction failed")

// ErrGetBlock signals an error happening when trying to fetch a block
var ErrGetBlock = errors.New("getting block failed")

// ErrGetValidatorsInfo signals an error happening when trying to fetch validators info
var ErrGetValidatorsInfo = errors.New("validators info failed")

// ErrGetAlteredAccountsForBlock signals an error happening when trying to fetch the altered accounts for a block
var ErrGetAlteredAccountsForBlock = errors.New("getting altered accounts for block failed")

// ErrQueryError signals a general query error
var ErrQueryError = errors.New("query error")

// ErrGetPidInfo signals that an error occurred while getting peer ID info
var ErrGetPidInfo = errors.New("error getting peer id info")

// ErrGetEpochStartData signals that an error occurred while getting the epoch start data for a provided epoch
var ErrGetEpochStartData = errors.New("error getting epoch start data for epoch")

// ErrTooManyRequests signals that too many requests were simultaneously received
var ErrTooManyRequests = errors.New("too many requests")

// ErrValidationEmptyRootHash signals that an empty root hash was provided
var ErrValidationEmptyRootHash = errors.New("rootHash is empty")

// ErrValidationEmptyAddress signals that an empty address was provided
var ErrValidationEmptyAddress = errors.New("address is empty")

// ErrValidationEmptyKey signals that an empty key was provided
var ErrValidationEmptyKey = errors.New("key is empty")

// ErrGetProof signals an error happening when trying to compute a Merkle proof
var ErrGetProof = errors.New("getting proof failed")

// ErrVerifyProof signals an error happening when trying to verify a Merkle proof
var ErrVerifyProof = errors.New("verifying proof failed")

// ErrNilHttpServer signals that a nil http server has been provided
var ErrNilHttpServer = errors.New("nil http server")

// ErrCannotCreateGinWebServer signals that the gin web server cannot be created
var ErrCannotCreateGinWebServer = errors.New("cannot create gin web server")

// ErrNilFacadeHandler signals that a nil facade handler has been provided
var ErrNilFacadeHandler = errors.New("nil facade handler")

// ErrFacadeWrongTypeAssertion signals that a type conversion to a facade type failed
var ErrFacadeWrongTypeAssertion = errors.New("facade - wrong type assertion")

// ErrGetGenesisNodes signals that an error occurred while trying to fetch genesis nodes config
var ErrGetGenesisNodes = errors.New("getting genesis nodes failed")

// ErrGetGenesisBalances signals that an error happened when trying to fetch genesis balances config
var ErrGetGenesisBalances = errors.New("getting genesis balances failed")

// ErrBadUrlParams signals one or more incorrectly provided URL params (generic error)
var ErrBadUrlParams = errors.New("bad url parameter(s)")

// ErrGetGasConfigs signals that an error occurred while trying to fetch gas configs
var ErrGetGasConfigs = errors.New("getting gas configs failed")

// ErrEmptySenderToGetLatestNonce signals that an error happened when trying to fetch latest nonce
var ErrEmptySenderToGetLatestNonce = errors.New("empty sender to get latest nonce")

// ErrEmptySenderToGetNonceGaps signals that an error happened when trying to fetch nonce gaps
var ErrEmptySenderToGetNonceGaps = errors.New("empty sender to get nonce gaps")

// ErrFetchingLatestNonceCannotIncludeFields signals that an error happened when trying to fetch latest nonce
var ErrFetchingLatestNonceCannotIncludeFields = errors.New("fetching latest nonce cannot include fields")

// ErrFetchingNonceGapsCannotIncludeFields signals that an error happened when trying to fetch nonce gaps
var ErrFetchingNonceGapsCannotIncludeFields = errors.New("fetching nonce gaps cannot include fields")

// ErrInvalidFields signals that invalid fields were provided
var ErrInvalidFields = errors.New("invalid fields")

// ErrGetESDTTokensWithRole signals an error in getting the esdt tokens with the given role for given address
var ErrGetESDTTokensWithRole = errors.New("getting esdt tokens with role error")

// ErrRegisteredNFTTokenIDs signals an error in getting the registered nft token ids by the given address
var ErrRegisteredNFTTokenIDs = errors.New("getting registered nft token ids error")

// ErrInvalidRole signals that an invalid role was provided
var ErrInvalidRole = errors.New("invalid role")
