package framework

import (
	"encoding/hex"
	"testing"

	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/stretchr/testify/require"
)

// KlvAddress holds the different forms a Klever Blockchain address might have
type KlvAddress struct {
	address.Address
	bytes  []byte
	bech32 string
	hex    string
}

// NewKlvAddressFromBytes return a new instance of KlvAddress from bytes
func NewKlvAddressFromBytes(tb testing.TB, bytes []byte) *KlvAddress {
	addr, err := address.NewAddressFromBytes(bytes)
	require.Nil(tb, err)
	address := &KlvAddress{
		bytes:   make([]byte, len(bytes)),
		hex:     hex.EncodeToString(bytes),
		Address: addr,
	}

	copy(address.bytes, bytes)
	address.bech32, err = addressPubkeyConverter.Encode(bytes)
	require.Nil(tb, err)

	return address
}

// NewKlvAddressFromBech32 return a new instance of KlvAddress from the bech32 string
func NewKlvAddressFromBech32(tb testing.TB, bech32 string) *KlvAddress {
	addressHandler, err := address.NewAddress(bech32)
	require.Nil(tb, err)

	return &KlvAddress{
		bytes:   addressHandler.Bytes(),
		hex:     hex.EncodeToString(addressHandler.Bytes()),
		bech32:  bech32,
		Address: addressHandler,
	}
}

// Bytes returns the bytes format address
func (address *KlvAddress) Bytes() []byte {
	return address.bytes
}

// Bech32 returns the bech32 string format address
func (address *KlvAddress) Bech32() string {
	return address.bech32
}

// Hex returns the hex string format address
func (address *KlvAddress) Hex() string {
	return address.hex
}

// String returns the address in bech32 format
func (address *KlvAddress) String() string {
	return address.bech32
}
