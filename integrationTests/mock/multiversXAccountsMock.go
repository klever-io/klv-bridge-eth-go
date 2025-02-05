package mock

import (
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/multiversx/mx-sdk-go/data"
)

type multiversXAccountsMock struct {
	accounts map[string]*data.Account
}

func newMultiversXAccountsMock() *multiversXAccountsMock {
	return &multiversXAccountsMock{
		accounts: make(map[string]*data.Account),
	}
}

func (mock *multiversXAccountsMock) getOrCreate(address address.Address) *data.Account {
	addrAsString := string(address.Bytes())
	acc, found := mock.accounts[addrAsString]
	if !found {
		acc = &data.Account{}
		mock.accounts[addrAsString] = acc
	}

	return acc
}

func (mock *multiversXAccountsMock) updateNonce(address address.Address, nonce uint64) {
	acc := mock.getOrCreate(address)
	acc.Nonce = nonce
}
