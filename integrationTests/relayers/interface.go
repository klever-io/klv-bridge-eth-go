package relayers

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klever-go-sdk/core/address"
)

type bridgeComponents interface {
	KleverRelayerAddress() address.Address
	EthereumRelayerAddress() common.Address
	Start() error
	Close() error
}
