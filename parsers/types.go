package parsers

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// CallData defines the struct holding SC call data parameters
type CallData struct {
	Type      byte
	Function  string
	GasLimit  uint64
	Arguments []string
}

// ProxySCCompleteCallData defines the struct holding Proxy SC complete call data
type ProxySCCompleteCallData struct {
	RawCallData []byte
	From        common.Address
	To          address.Address
	Token       string
	Amount      *big.Int
	Nonce       uint64
}

// String returns the human-readable string version of the call data
func (callData ProxySCCompleteCallData) String() string {
	toString := "<nil>"
	if !check.IfNil(callData.To) {
		toString = callData.To.Bech32()
	}
	amountString := "<nil>"
	if callData.Amount != nil {
		amountString = callData.Amount.String()
	}

	return fmt.Sprintf("Eth address: %s, MvX address: %s, token: %s, amount: %s, nonce: %d, raw call data: %x",
		callData.From.String(),
		toString,
		callData.Token,
		amountString,
		callData.Nonce,
		callData.RawCallData,
	)
}
