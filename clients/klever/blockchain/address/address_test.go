package address_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddress_ZeroAddress(t *testing.T) {
	addr := address.ZeroAddress()
	assert.Equal(t, addr.Bech32(), "klv1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpgm89z")
	assert.Equal(t, "0000000000000000000000000000000000000000000000000000000000000000", hex.EncodeToString(addr.Bytes()))
}

func TestAddress_InvalidSize(t *testing.T) {
	_, err := address.NewAddressFromBytes([]byte("000"))
	assert.Contains(t, err.Error(), "decoding address, expected length 32")

	_, err = address.NewAddressFromHex("000")
	assert.Contains(t, err.Error(), "encoding/hex: odd length hex string")

	_, err = address.NewAddressFromHex("0000")
	assert.Contains(t, err.Error(), "decoding address, expected length")
}

func TestAddress_Address_ShouldFail(t *testing.T) {
	// invalid size
	_, err := address.NewAddressFromHex("000000000000000000000000000000000000000000000000000000000000000")
	assert.NotNil(t, err)

	// invalid hex
	_, err = address.NewAddressFromHex("X000000000000000000000000000000000000000000000000000000000000000")
	assert.NotNil(t, err)

	// invalid bech32
	_, err = address.NewAddress("klv1qy352eufzqg3yyc5z5v3wxqeyqsjygeyy5nzw2pfxqcnyve5x5mqfrkqfh")
	assert.NotNil(t, err)

	// invalid bech32 prefix
	_, err = address.NewAddress("kfi1qy352eufzqg3yyc5z5v3wxqeyqsjygeyy5nzw2pfxqcnyve5x5mq7ze5xk")
	assert.NotNil(t, err)

	// invalid decoded len
	_, err = address.NewAddress("klv1d05ju9jaj6u99zph0ant9jjv3jkq")
	assert.Contains(t, err.Error(), "invalid incomplete group")

	_, err = address.NewAddress("klv1xqcrqt6vdma")
	assert.Contains(t, err.Error(), "decoding address, expected length 32")

	// invalid checksum
	_, err = address.NewAddress("klv1d05ju9jaj6u99zph0ant9jh7gksg")
	assert.Contains(t, err.Error(), "invalid checksum")

}

func TestAddress_Address_ShouldWork(t *testing.T) {
	hexString := "0123456789101112131415191718192021222324252627282930313233343536"
	bytes, _ := hex.DecodeString(hexString)
	bech32Addr := "klv1qy352eufzqg3yyc5z5v3wxqeyqsjygeyy5nzw2pfxqcnyve5x5mqfrkqfg"

	// invalid size
	addr, err := address.NewAddressFromHex(hexString)
	assert.Nil(t, err)
	assert.Equal(t, addr.Bech32(), bech32Addr)
	assert.Equal(t, addr.Hex(), hexString)

	addr, err = address.NewAddressFromBytes(bytes)
	assert.Nil(t, err)
	assert.Equal(t, addr.Bech32(), bech32Addr)
	assert.Equal(t, addr.Hex(), hexString)

	addr, err = address.NewAddress(bech32Addr)
	assert.Nil(t, err)
	assert.Equal(t, addr.Bech32(), bech32Addr)
	assert.Equal(t, addr.Hex(), hexString)
}

func TestAddress_AdressSlice_ShouldWork(t *testing.T) {
	bech32Addr := "klv1mge94r8n3q44hcwu2tk9afgjcxcawmutycu0cwkap7m6jnktjlvq58355l"
	hexString := "da325a8cf3882b5be1dc52ec5ea512c1b1d76f8b2638fc3add0fb7a94ecb97d8"
	expectedBytes, err := hex.DecodeString(hexString)
	require.NoError(t, err)

	addr, err := address.NewAddress(bech32Addr)
	require.NoError(t, err)

	addrSlice := addr.AddressSlice()
	bytes.Equal(addrSlice[:], expectedBytes)
}
