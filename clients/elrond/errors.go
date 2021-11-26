package elrond

import "errors"

var (
	errNilProxy          = errors.New("nil ElrondProxy")
	errNilAddressHandler = errors.New("nil address handler")
	errNilRequest        = errors.New("nil request")
	// TODO: use these
	// errInvalidNumberOfArguments = errors.New("invalid number of arguments")
	// errNotUint64Bytes           = errors.New("provided bytes do not represent a valid uint64 number")
)
