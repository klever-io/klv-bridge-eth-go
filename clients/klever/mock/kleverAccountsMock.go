package mock

import (
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

type kleverAccountsMock struct {
	accounts map[string]*data.Account
}

func newKleverAccountsMock() *kleverAccountsMock {
	return &kleverAccountsMock{
		accounts: make(map[string]*data.Account),
	}
}

func (mock *kleverAccountsMock) getOrCreate(address core.AddressHandler) *data.Account {
	addrAsString := string(address.AddressBytes())
	acc, found := mock.accounts[addrAsString]
	if !found {
		acc = &data.Account{}
		mock.accounts[addrAsString] = acc
	}

	return acc
}

func (mock *kleverAccountsMock) updateNonce(address core.AddressHandler, nonce uint64) {
	acc := mock.getOrCreate(address)
	acc.Nonce = nonce
}
