package address_test

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcutil/base58"
	"github.com/klever-io/klv-bridge-eth-go/clients/tron/address"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromBytes_InvalidLength(t *testing.T) {
	_, err := address.FromBytes([]byte{1, 2, 3})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid raw address length")
}

func TestFromBytesAndBase58_RoundTrip(t *testing.T) {
	rawDigest, err := hex.DecodeString("d95f3f6e5b9b0f5afbd6be6c9fcb5f4f97f6c17a")
	require.NoError(t, err)

	addr, err := address.FromBytes(rawDigest)
	require.NoError(t, err)
	require.Equal(t, byte(0x41), addr.Bytes()[0])

	decoded, err := address.FromBase58(addr.Base58())
	require.NoError(t, err)
	assert.Equal(t, addr, decoded)

	reconstructed, err := address.FromBytes(addr.Bytes()[1:])
	require.NoError(t, err)
	assert.Equal(t, addr, reconstructed)
}

func TestFromBase58_KnownVector(t *testing.T) {
	const genesisAddr = "TN3W4H6rK2ce4vX9YnFQHwKENnHjoxb3m9"

	addr, err := address.FromBase58(genesisAddr)
	require.NoError(t, err)

	assert.Equal(t, genesisAddr, addr.Base58())
	assert.Equal(t, genesisAddr, addr.String())
	assert.Len(t, addr.Bytes(), 21)
	assert.Equal(t, byte(0x41), addr.Bytes()[0])
}

func TestFromBase58_InvalidInputs(t *testing.T) {
	_, err := address.FromBase58("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty string")

	_, err = address.FromBase58("!!!not-base58!!!")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid base58 address")

	wrongPrefix := base58.CheckEncode(make([]byte, 20), 0x00)
	_, err = address.FromBase58(wrongPrefix)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid prefix")
}

func TestHexAndInterfaceNil(t *testing.T) {
	rawDigest, err := hex.DecodeString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	require.NoError(t, err)

	addr, err := address.FromBytes(rawDigest)
	require.NoError(t, err)

	hexAddr := addr.Hex()
	assert.Contains(t, hexAddr, "0x")
	assert.Len(t, hexAddr, 44)

	var nilAddr *address.TronAddress
	assert.True(t, nilAddr.IsInterfaceNil())
}
