package mock

import (
	"github.com/klever-io/klever-go-sdk/core/address"
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

func (mock *kleverAccountsMock) getOrCreate(address address.Address) *data.Account {
	addrAsString := string(address.Bytes())
	acc, found := mock.accounts[addrAsString]
	if !found {
		acc = &data.Account{}
		mock.accounts[addrAsString] = acc
	}

	return acc
}

func (mock *kleverAccountsMock) updateNonce(address address.Address, nonce uint64) {
	acc := mock.getOrCreate(address)
	acc.Nonce = nonce
}
