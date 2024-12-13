package converters

import (
	"encoding/hex"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
)

const hexPrefix = "0x"
const hrp = "klv"

type addressConverter struct {
	converter core.PubkeyConverter
}

// NewAddressConverter will create an address converter instance
func NewAddressConverter() (*addressConverter, error) {
	var err error
	ac := &addressConverter{}
	ac.converter, err = pubkeyConverter.NewBech32PubkeyConverter(sdkCore.AddressBytesLen, hrp)
	if err != nil {
		return nil, err
	}

	return ac, nil
}

// ToHexString will convert the addressBytes to the hex representation
func (ac *addressConverter) ToHexString(addressBytes []byte) string {
	return hex.EncodeToString(addressBytes)
}

// ToHexStringWithPrefix will convert the addressBytes to the hex representation adding the hex prefix
func (ac *addressConverter) ToHexStringWithPrefix(addressBytes []byte) string {
	return hexPrefix + hex.EncodeToString(addressBytes)
}

// ToBech32String will convert the addressBytes to the bech32 representation
func (ac *addressConverter) ToBech32String(addressBytes []byte) (string, error) {
	return ac.converter.Encode(addressBytes)
}

// ToBech32StringSilent will try to convert the addressBytes to the bech32 representation
func (ac *addressConverter) ToBech32StringSilent(addressBytes []byte) string {
	bech32Address, _ := ac.converter.Encode(addressBytes)

	return bech32Address
}

// IsInterfaceNil returns true if there is no value under the interface
func (ac *addressConverter) IsInterfaceNil() bool {
	return ac == nil
}

// TrimWhiteSpaceCharacters will remove the white spaces from the input string
func TrimWhiteSpaceCharacters(input string) string {
	cutset := "\n\t "

	return strings.Trim(input, cutset)
}
