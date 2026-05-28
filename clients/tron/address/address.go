package address

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	gotronAddress "github.com/fbsobreira/gotron-sdk/pkg/address"
)

const (
	tronAddressPrefix = byte(0x41)
	rawLen            = 20
	encodedLen        = 21
)

// TronAddress holds the internal TRON address form: 0x41 prefix + 20 digest bytes.
type TronAddress [encodedLen]byte

// FromBytes constructs a TronAddress from a 20-byte digest.
func FromBytes(b []byte) (TronAddress, error) {
	if len(b) != rawLen {
		return TronAddress{}, fmt.Errorf("invalid raw address length: expected %d, got %d", rawLen, len(b))
	}

	var addr TronAddress
	addr[0] = tronAddressPrefix
	copy(addr[1:], b)

	return addr, nil
}

// FromBase58 decodes a TRON Base58Check address string.
func FromBase58(s string) (TronAddress, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return TronAddress{}, errors.New("invalid base58 address: empty string")
	}

	sdkAddr, err := gotronAddress.Base58ToAddress(trimmed)
	if err != nil {
		return TronAddress{}, fmt.Errorf("invalid base58 address %q: %w", trimmed, err)
	}

	if !sdkAddr.IsValid() {
		return TronAddress{}, fmt.Errorf("invalid base58 address %q: checksum or format validation failed", trimmed)
	}

	if len(sdkAddr) != encodedLen {
		return TronAddress{}, fmt.Errorf("invalid decoded address length: expected %d, got %d", encodedLen, len(sdkAddr))
	}

	if sdkAddr[0] != tronAddressPrefix {
		return TronAddress{}, fmt.Errorf("invalid address prefix: expected 0x%X, got 0x%X", tronAddressPrefix, sdkAddr[0])
	}

	var addr TronAddress
	copy(addr[:], sdkAddr)

	return addr, nil
}

// Bytes returns the 21-byte address representation.
func (a TronAddress) Bytes() []byte {
	result := make([]byte, encodedLen)
	copy(result, a[:])

	return result
}

// Hex returns the 0x-prefixed hex string of the 21-byte address.
func (a TronAddress) Hex() string {
	return "0x" + hex.EncodeToString(a[:])
}

// Base58 returns the TRON Base58Check representation.
func (a TronAddress) Base58() string {
	return gotronAddress.Address(a[:]).String()
}

// String implements fmt.Stringer and returns the Base58 representation.
func (a TronAddress) String() string {
	return a.Base58()
}

// IsInterfaceNil allows interface nil checks in repository conventions.
func (a *TronAddress) IsInterfaceNil() bool {
	return a == nil
}
