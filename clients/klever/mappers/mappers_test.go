package mappers

import (
	"context"
	"errors"
	"testing"

	"github.com/klever-io/klv-bridge-eth-go/clients"
	bridgeTests "github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewMapper(t *testing.T) {
	t.Parallel()
	{
		t.Run("Erc20ToKC: nil dataGetter", func(t *testing.T) {
			mapper, err := NewErc20ToKCMapper(nil)
			assert.Equal(t, clients.ErrNilDataGetter, err)
			assert.True(t, check.IfNil(mapper))
		})
		t.Run("Erc20ToKC: should work", func(t *testing.T) {
			mapper, err := NewErc20ToKCMapper(&bridgeTests.DataGetterStub{})
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))
		})
	}
	{
		t.Run("KCToErc20: nil dataGetter", func(t *testing.T) {
			mapper, err := NewKCToErc20Mapper(nil)
			assert.Equal(t, clients.ErrNilDataGetter, err)
			assert.True(t, check.IfNil(mapper))
		})
		t.Run("KCToErc20: should work", func(t *testing.T) {
			mapper, err := NewKCToErc20Mapper(&bridgeTests.DataGetterStub{})
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))
		})
	}
}

func TestConvertToken(t *testing.T) {
	t.Parallel()

	{
		t.Run("KCToErc20: dataGetter returns error", func(t *testing.T) {
			expectedError := errors.New("expected error")
			dg := &bridgeTests.DataGetterStub{
				GetERC20AddressForTokenIdCalled: func(ctx context.Context, tokenId []byte) ([][]byte, error) {
					return nil, expectedError
				}}
			mapper, err := NewKCToErc20Mapper(dg)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))

			_, err = mapper.ConvertToken(context.Background(), []byte("klvAddress"))
			assert.Equal(t, expectedError, err)
		})
		t.Run("KCToErc20: should work", func(t *testing.T) {
			expectedErc20Address := []byte("erc20Address")
			dg := &bridgeTests.DataGetterStub{
				GetERC20AddressForTokenIdCalled: func(ctx context.Context, tokenId []byte) ([][]byte, error) {
					return [][]byte{expectedErc20Address}, nil
				}}
			mapper, err := NewKCToErc20Mapper(dg)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))
			erc20AddressReturned, err := mapper.ConvertToken(context.Background(), []byte("klvAddress"))
			assert.Nil(t, err)
			assert.Equal(t, expectedErc20Address, erc20AddressReturned)
		})
	}
	{
		t.Run("Erc20ToKC: dataGetter returns error", func(t *testing.T) {
			expectedError := errors.New("expected error")
			dg := &bridgeTests.DataGetterStub{
				GetTokenIdForErc20AddressCalled: func(ctx context.Context, erc20Address []byte) ([][]byte, error) {
					return nil, expectedError
				}}
			mapper, err := NewErc20ToKCMapper(dg)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))

			_, err = mapper.ConvertToken(context.Background(), []byte("erc20Address"))
			assert.Equal(t, expectedError, err)
		})
		t.Run("Erc20ToKC: should work", func(t *testing.T) {
			expectedKlvAddress := []byte("klvAddress")
			dg := &bridgeTests.DataGetterStub{
				GetTokenIdForErc20AddressCalled: func(ctx context.Context, erc20Address []byte) ([][]byte, error) {
					return [][]byte{expectedKlvAddress}, nil
				}}
			mapper, err := NewErc20ToKCMapper(dg)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))
			klvAddressReturned, err := mapper.ConvertToken(context.Background(), []byte("erc20Address"))
			assert.Nil(t, err)
			assert.Equal(t, expectedKlvAddress, klvAddressReturned)
		})
	}
}
