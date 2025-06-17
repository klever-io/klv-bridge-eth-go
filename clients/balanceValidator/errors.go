package balanceValidator

import "errors"

// ErrNilLogger signals that a nil logger has been provided
var ErrNilLogger = errors.New("nil logger")

// ErrNilKcClient signals that a nil Klever Blockchain client has been provided
var ErrNilKcClient = errors.New("nil Klever Blockchain client")

// ErrNilEthereumClient signals that a nil Ethereum client has been provided
var ErrNilEthereumClient = errors.New("nil Ethereum client")

// ErrInvalidDirection signals that an invalid direction was provided
var ErrInvalidDirection = errors.New("invalid direction")

// ErrInvalidSetup signals that an invalid setup was provided
var ErrInvalidSetup = errors.New("invalid setup")

// ErrNegativeAmount signals that a negative amount was provided
var ErrNegativeAmount = errors.New("negative amount")

// ErrBalanceMismatch signals that the balances are not expected
var ErrBalanceMismatch = errors.New("balance mismatch")
