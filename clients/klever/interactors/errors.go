package interactors

import "errors"

// ErrNilProxy signals that a nil proxy was provided
var ErrNilProxy = errors.New("nil proxy")

// ErrInvalidValue signals that an invalid value was provided
var ErrInvalidValue = errors.New("invalid value")

// ErrNilAddress signals that the provided address is nil
var ErrNilAddress = errors.New("nil address")

// ErrNilTransaction signals that provided transaction is nil
var ErrNilTransaction = errors.New("nil transaction")

// ErrTxAlreadySent signals that a transaction was already sent
var ErrTxAlreadySent = errors.New("transaction already sent")

// ErrTxWithSameNonceAndGasPriceAlreadySent signals that a transaction with the same nonce & gas price was already sent
var ErrTxWithSameNonceAndGasPriceAlreadySent = errors.New("transaction with the same nonce & gas price was already sent")

// ErrGapNonce signals that a gap nonce between the lowest nonce of the transactions from the cache and the blockchain nonce has been detected
var ErrGapNonce = errors.New("gap nonce detected")
