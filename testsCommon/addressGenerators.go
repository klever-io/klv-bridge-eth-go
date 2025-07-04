package testsCommon

import (
	"crypto/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
)

// CreateRandomEthereumAddress will create a random Ethereum address
func CreateRandomEthereumAddress() common.Address {
	buff := make([]byte, len(common.Address{}))
	_, _ = rand.Read(buff)

	return common.BytesToAddress(buff)
}

// CreateRandomKCAddress will create a random Klever Blockchain address
func CreateRandomKCAddress() address.Address {
	buff := make([]byte, 32)
	_, _ = rand.Read(buff)

	addr, _ := address.NewAddressFromBytes(buff)

	return addr
}

// CreateRandomKCSCAddress will create a random Klever Blockchain smart contract address
func CreateRandomKCSCAddress() address.Address {
	buff := make([]byte, 22)
	_, _ = rand.Read(buff)

	firstPart := append(make([]byte, 8), []byte{5, 0}...)

	addr, _ := address.NewAddressFromBytes(append(firstPart, buff...))
	return addr
}
