package framework

import (
	"encoding/hex"
	"testing"

	"github.com/klever-io/klever-go-sdk/core/address"
	"github.com/stretchr/testify/require"
)

// MvxAddress holds the different forms a MultiversX address might have
type MvxAddress struct {
	address.Address
	bytes  []byte
	bech32 string
	hex    string
}

// NewMvxAddressFromBytes return a new instance of MvxAddress from bytes
func NewMvxAddressFromBytes(tb testing.TB, bytes []byte) *MvxAddress {
	addr, err := address.NewAddressFromBytes(bytes)
	require.Nil(tb, err)
	address := &MvxAddress{
		bytes:   make([]byte, len(bytes)),
		hex:     hex.EncodeToString(bytes),
		Address: addr,
	}

	copy(address.bytes, bytes)
	address.bech32, err = addressPubkeyConverter.Encode(bytes)
	require.Nil(tb, err)

	return address
}

// NewMvxAddressFromBech32 return a new instance of MvxAddress from the bech32 string
func NewMvxAddressFromBech32(tb testing.TB, bech32 string) *MvxAddress {
	addressHandler, err := address.NewAddress(bech32)
	require.Nil(tb, err)

	return &MvxAddress{
		bytes:   addressHandler.Bytes(),
		hex:     hex.EncodeToString(addressHandler.Bytes()),
		bech32:  bech32,
		Address: addressHandler,
	}
}

// Bytes returns the bytes format address
func (address *MvxAddress) Bytes() []byte {
	return address.bytes
}

// Bech32 returns the bech32 string format address
func (address *MvxAddress) Bech32() string {
	return address.bech32
}

// Hex returns the hex string format address
func (address *MvxAddress) Hex() string {
	return address.hex
}

// String returns the address in bech32 format
func (address *MvxAddress) String() string {
	return address.bech32
}
