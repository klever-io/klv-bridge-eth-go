package parsers

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/stretchr/testify/assert"
)

func TestProxySCCompleteCallData_String(t *testing.T) {
	t.Parallel()

	t.Run("nil fields should work", func(t *testing.T) {
		t.Parallel()

		callData := ProxySCCompleteCallData{
			RawCallData: []byte{65, 66, 67},
			From:        common.Address{},
			To:          nil,
			Token:       "tkn",
			Amount:      nil,
			Nonce:       1,
		}

		expectedString := "Eth address: 0x0000000000000000000000000000000000000000, MvX address: <nil>, token: tkn, amount: <nil>, nonce: 1, raw call data: 414243"
		assert.Equal(t, expectedString, callData.String())
	})
	t.Run("not a Valid Klever address should work", func(t *testing.T) {
		t.Parallel()
		addr, _ := address.NewAddressFromBytes([]byte{0x1, 0x2})
		callData := ProxySCCompleteCallData{
			RawCallData: []byte{65, 66, 67},
			From:        common.Address{},
			To:          addr,
			Token:       "tkn",
			Nonce:       1,
		}

		expectedString := "Eth address: 0x0000000000000000000000000000000000000000, MvX address: <err>, token: tkn, amount: <nil>, nonce: 1, raw call data: 414243"
		assert.Equal(t, expectedString, callData.String())
	})
	t.Run("with valid data should work", func(t *testing.T) {
		t.Parallel()

		callData := ProxySCCompleteCallData{
			RawCallData: []byte{65, 66, 67},
			From:        common.Address{},
			Token:       "tkn",
			Amount:      big.NewInt(37),
			Nonce:       1,
		}
		ethUnhexed, _ := hex.DecodeString("880ec53af800b5cd051531672ef4fc4de233bd5d")
		callData.From.SetBytes(ethUnhexed)
		callData.To, _ = address.NewAddress("klv1qqqqqqqqqqqqqpgqh46r9zh78lry2py8tq723fpjdr4pp0zgsg8syf6mq0")

		expectedString := "Eth address: 0x880EC53Af800b5Cd051531672EF4fc4De233bD5d, MvX address: klv1qqqqqqqqqqqqqpgqh46r9zh78lry2py8tq723fpjdr4pp0zgsg8syf6mq0, token: tkn, amount: 37, nonce: 1, raw call data: 414243"
		assert.Equal(t, expectedString, callData.String())
	})
}
