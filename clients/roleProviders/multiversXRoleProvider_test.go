package roleproviders

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/klever-io/klever-go-sdk/core/address"
	"github.com/klever-io/klv-bridge-eth-go/clients"
	bridgeTests "github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

func createMockArgs() ArgsMultiversXRoleProvider {
	return ArgsMultiversXRoleProvider{
		Log:        logger.GetOrCreate("test"),
		DataGetter: &bridgeTests.DataGetterStub{},
	}
}

func TestNewMultiversXRoleProvider(t *testing.T) {
	t.Parallel()

	t.Run("nil data getter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.DataGetter = nil

		erp, err := NewMultiversXRoleProvider(args)
		assert.True(t, check.IfNil(erp))
		assert.Equal(t, clients.ErrNilDataGetter, err)
	})
	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Log = nil

		erp, err := NewMultiversXRoleProvider(args)
		assert.True(t, check.IfNil(erp))
		assert.Equal(t, clients.ErrNilLogger, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		erp, err := NewMultiversXRoleProvider(args)
		assert.False(t, check.IfNil(erp))
		assert.Nil(t, err)
	})
}

func TestMultiversXRoleProvider_ExecuteErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	args := createMockArgs()
	args.DataGetter = &bridgeTests.DataGetterStub{
		GetAllStakedRelayersCalled: func(ctx context.Context) ([][]byte, error) {
			return nil, expectedErr
		},
	}

	erp, _ := NewMultiversXRoleProvider(args)
	err := erp.Execute(context.TODO())
	assert.Equal(t, expectedErr, err)
}

func TestMultiversXRoleProvider_ExecuteShouldWork(t *testing.T) {
	t.Parallel()

	whitelistedAddresses := [][]byte{
		bytes.Repeat([]byte("1"), 32),
		bytes.Repeat([]byte("3"), 32),
		bytes.Repeat([]byte("2"), 32),
	}
	expectedSortedPublicKeys := [][]byte{
		bytes.Repeat([]byte("1"), 32),
		bytes.Repeat([]byte("2"), 32),
		bytes.Repeat([]byte("3"), 32),
	}

	t.Run("nil whitelisted", testMultiversXExecuteShouldWork(nil, make([][]byte, 0)))
	t.Run("empty whitelisted", testMultiversXExecuteShouldWork(make([][]byte, 0), make([][]byte, 0)))
	t.Run("with whitelisted", testMultiversXExecuteShouldWork(whitelistedAddresses, expectedSortedPublicKeys))
}

func testMultiversXExecuteShouldWork(whitelistedAddresses [][]byte, expectedSortedPublicKeys [][]byte) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.DataGetter = &bridgeTests.DataGetterStub{
			GetAllStakedRelayersCalled: func(ctx context.Context) ([][]byte, error) {
				return whitelistedAddresses, nil
			},
		}

		erp, _ := NewMultiversXRoleProvider(args)
		err := erp.Execute(context.TODO())
		assert.Nil(t, err)

		for _, addr := range whitelistedAddresses {
			addressHandler, err := address.NewAddressFromBytes(addr)
			assert.Nil(t, err)
			assert.True(t, erp.IsWhitelisted(addressHandler))
		}

		// TODO: check if error should be verified
		randomAddress, _ := address.NewAddressFromBytes([]byte("random address"))

		assert.False(t, erp.IsWhitelisted(randomAddress))
		assert.False(t, erp.IsWhitelisted(nil))
		erp.mut.RLock()
		assert.Equal(t, len(whitelistedAddresses), len(erp.whitelistedAddresses))
		erp.mut.RUnlock()
		sortedPublicKeys := erp.SortedPublicKeys()
		assert.Equal(t, expectedSortedPublicKeys, sortedPublicKeys)
	}
}

func TestMultiversXRoleProvider_MisconfiguredAddressesShouldError(t *testing.T) {
	t.Parallel()

	misconfiguredAddresses := [][]byte{
		bytes.Repeat([]byte("1"), 32),
		bytes.Repeat([]byte("2"), 32),
		[]byte("bad address"),
	}

	args := createMockArgs()
	args.DataGetter = &bridgeTests.DataGetterStub{
		GetAllStakedRelayersCalled: func(ctx context.Context) ([][]byte, error) {
			return misconfiguredAddresses, nil
		},
	}

	erp, _ := NewMultiversXRoleProvider(args)
	err := erp.Execute(context.TODO())
	assert.True(t, errors.Is(err, ErrInvalidAddressBytes))
	assert.True(t, strings.Contains(err.Error(), hex.EncodeToString(misconfiguredAddresses[2])))
	assert.Zero(t, len(erp.whitelistedAddresses))
}
