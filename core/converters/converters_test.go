package converters

import (
	"encoding/hex"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestConvertFromByteSliceToArray(t *testing.T) {
	t.Parallel()

	buff := []byte("12345678901234567890123456789012")

	result := data.NewAddressFromBytes(buff).AddressSlice()
	assert.Equal(t, buff, result[:])
}

func TestTrimWhiteSpaceCharacters(t *testing.T) {
	t.Parallel()

	dataField := "aaII139HSAh32q782!$#*$(nc"

	input := " " + dataField
	assert.Equal(t, dataField, TrimWhiteSpaceCharacters(input))

	input = "\t " + dataField
	assert.Equal(t, dataField, TrimWhiteSpaceCharacters(input))

	input = "\t " + dataField + "\n"
	assert.Equal(t, dataField, TrimWhiteSpaceCharacters(input))

	input = "\t\n " + dataField + "\n\n\n\n\t"
	assert.Equal(t, dataField, TrimWhiteSpaceCharacters(input))
}

func TestAddressConverter_ToBech32String(t *testing.T) {
	t.Parallel()

	addrConv, err := NewAddressConverter()
	require.Nil(t, err)
	assert.False(t, check.IfNil(addrConv))

	t.Run("invalid bytes should return empty", func(t *testing.T) {
		str := addrConv.ToBech32String([]byte("invalid"))
		assert.Empty(t, str)
	})
	t.Run("should work", func(t *testing.T) {
		expected := "klv1mge94r8n3q44hcwu2tk9afgjcxcawmutycu0cwkap7m6jnktjlvq58355l"
		bytes, _ := hex.DecodeString("da325a8cf3882b5be1dc52ec5ea512c1b1d76f8b2638fc3add0fb7a94ecb97d8")
		bech32Address := addrConv.ToBech32String(bytes)
		assert.Equal(t, expected, bech32Address)
	})
}

func TestAddressConverter_ToHexString(t *testing.T) {
	t.Parallel()

	addrConv, err := NewAddressConverter()
	require.Nil(t, err)
	assert.False(t, check.IfNil(addrConv))

	expected := "627974657320746f20656e636f6465"
	bytes := []byte("bytes to encode")
	assert.Equal(t, expected, addrConv.ToHexString(bytes))
}

func TestAddressConverter_ToHexStringWithPrefix(t *testing.T) {
	t.Parallel()

	addrConv, err := NewAddressConverter()
	require.Nil(t, err)
	assert.False(t, check.IfNil(addrConv))

	expected := "0x627974657320746f20656e636f6465"
	bytes := []byte("bytes to encode")
	assert.Equal(t, expected, addrConv.ToHexStringWithPrefix(bytes))
}
