package mappers

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/klever-io/klv-bridge-eth-go/clients"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

type erc20ToKc struct {
	dg DataGetter
}

// NewErc20ToKcMapper returns a new instance of erc20ToKc
func NewErc20ToKcMapper(dg DataGetter) (*erc20ToKc, error) {
	if check.IfNil(dg) {
		return nil, clients.ErrNilDataGetter
	}

	return &erc20ToKc{
		dg: dg,
	}, nil
}

// ConvertToken will return klv token id given a specific erc20 address
func (mapper *erc20ToKc) ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error) {

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
func (mapper *erc20ToKc) IsInterfaceNil() bool {
	return mapper == nil
}
