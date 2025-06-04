package mock

import (
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
)

type kleverBlockchainAccountsMock struct {
	accounts map[string]*models.Account
}

func newKleverBlockchainAccountsMock() *kleverBlockchainAccountsMock {
	return &kleverBlockchainAccountsMock{
		accounts: make(map[string]*models.Account),
	}
}

func (mock *kleverBlockchainAccountsMock) getOrCreate(address address.Address) *models.Account {
	addrAsString := string(address.Bytes())
	acc, found := mock.accounts[addrAsString]
	if !found {
		acc = &models.Account{}
		mock.accounts[addrAsString] = acc
	}

	return acc
}

func (mock *kleverBlockchainAccountsMock) updateNonce(address address.Address, nonce uint64) {
	acc := mock.getOrCreate(address)
	acc.Nonce = nonce
}
