package nonceHandlerV2

import (
	"github.com/klever-io/klever-go/data/transaction"
	"github.com/klever-io/klever-go/tools/check"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/interactors"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy"
)

// NewAddressNonceHandlerWithPrivateAccess -
func NewAddressNonceHandlerWithPrivateAccess(proxy proxy.Proxy, address address.Address) (*addressNonceHandler, error) {
	if check.IfNil(proxy) {
		return nil, interactors.ErrNilProxy
	}
	if check.IfNil(address) {
		return nil, interactors.ErrNilAddress
	}
	return &addressNonceHandler{
		address:      address,
		proxy:        proxy,
		transactions: make(map[uint64]*transaction.Transaction),
	}, nil
}
