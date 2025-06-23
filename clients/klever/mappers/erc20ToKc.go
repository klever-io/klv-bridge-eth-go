package mappers

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/klever-io/klv-bridge-eth-go/clients"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

type erc20ToKC struct {
	dg DataGetter
}

// NewErc20ToKCMapper returns a new instance of erc20ToKC
func NewErc20ToKCMapper(dg DataGetter) (*erc20ToKC, error) {
	if check.IfNil(dg) {
		return nil, clients.ErrNilDataGetter
	}

	return &erc20ToKC{
		dg: dg,
	}, nil
}

// ConvertToken will return klv token id given a specific erc20 address
func (mapper *erc20ToKC) ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error) {

	response, err := mapper.dg.GetTokenIdForErc20Address(ctx, sourceBytes)
	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("%w for provided %s", errUnknownToken, hex.EncodeToString(sourceBytes))
	}

	return response[0], nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mapper *erc20ToKC) IsInterfaceNil() bool {
	return mapper == nil
}
