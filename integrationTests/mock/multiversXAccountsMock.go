package mock

import (
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/multiversx/mx-sdk-go/data"
)

type kleverchainAccountsMock struct {
	accounts map[string]*data.Account
}

func newKleverchainAccountsMock() *kleverchainAccountsMock {
	return &kleverchainAccountsMock{
		accounts: make(map[string]*data.Account),
	}
}

func (mock *kleverchainAccountsMock) getOrCreate(address address.Address) *data.Account {
	addrAsString := string(address.Bytes())
	acc, found := mock.accounts[addrAsString]
	if !found {
		acc = &data.Account{}
		mock.accounts[addrAsString] = acc
	}

	return acc
}

func (mock *kleverchainAccountsMock) updateNonce(address address.Address, nonce uint64) {
	acc := mock.getOrCreate(address)
	acc.Nonce = nonce
}
