package mappers

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/klever-io/klv-bridge-eth-go/clients"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

type kleverchainToErc20 struct {
	dg DataGetter
}

// NewKleverchainToErc20Mapper returns a new instance of kleverchainToErc20
func NewKleverchainToErc20Mapper(dg DataGetter) (*kleverchainToErc20, error) {
	if check.IfNil(dg) {
		return nil, clients.ErrNilDataGetter
	}

	return &kleverchainToErc20{
		dg: dg,
	}, nil
}

// ConvertToken will return klv token id given a specific erc20 address
func (mapper *kleverchainToErc20) ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error) {

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
func (mapper *kleverchainToErc20) IsInterfaceNil() bool {
	return mapper == nil
}
