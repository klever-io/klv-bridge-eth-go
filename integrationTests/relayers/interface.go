package relayers

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
)

type bridgeComponents interface {
	KleverRelayerAddress() address.Address
	EthereumRelayerAddress() common.Address
	Start() error
	Close() error
}
