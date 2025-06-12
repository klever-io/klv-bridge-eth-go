package mappers

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/klever-io/klv-bridge-eth-go/clients"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

type kcToErc20 struct {
	dg DataGetter
}

// NewKcToErc20Mapper returns a new instance of kcToErc20
func NewKcToErc20Mapper(dg DataGetter) (*kcToErc20, error) {
	if check.IfNil(dg) {
		return nil, clients.ErrNilDataGetter
	}

	return &kcToErc20{
		dg: dg,
	}, nil
}

// ConvertToken will return klv token id given a specific erc20 address
func (mapper *kcToErc20) ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error) {

	response, err := mapper.dg.GetERC20AddressForTokenId(ctx, sourceBytes)
	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("%w for provided %s", errUnknownToken, hex.EncodeToString(sourceBytes))
	}

	return response[0], nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mapper *kcToErc20) IsInterfaceNil() bool {
	return mapper == nil
}
