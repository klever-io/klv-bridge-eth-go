package roleproviders

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/klever-io/klever-go/tools/check"
	"github.com/klever-io/klv-bridge-eth-go/clients"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	bridgeTests "github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

func createMockArgs() ArgsKleverRoleProvider {
	return ArgsKleverRoleProvider{
		Log:        logger.GetOrCreate("test"),
		DataGetter: &bridgeTests.DataGetterStub{},
	}
}

func TestNewKleverRoleProvider(t *testing.T) {
	t.Parallel()

	t.Run("nil data getter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.DataGetter = nil

		krp, err := NewKleverRoleProvider(args)
		assert.True(t, check.IfNil(krp))
		assert.Equal(t, clients.ErrNilDataGetter, err)
	})
	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Log = nil

		krp, err := NewKleverRoleProvider(args)
		assert.True(t, check.IfNil(krp))
		assert.Equal(t, clients.ErrNilLogger, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		krp, err := NewKleverRoleProvider(args)
		assert.False(t, check.IfNil(krp))
		assert.Nil(t, err)
	})
}

func TestKleverRoleProvider_ExecuteErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	args := createMockArgs()
	args.DataGetter = &bridgeTests.DataGetterStub{
		GetAllStakedRelayersCalled: func(ctx context.Context) ([][]byte, error) {
			return nil, expectedErr
		},
	}

	krp, _ := NewKleverRoleProvider(args)
	err := krp.Execute(context.TODO())
	assert.Equal(t, expectedErr, err)
}

func TestKleverRoleProvider_ExecuteShouldWork(t *testing.T) {
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

	t.Run("nil whitelisted", testKleverExecuteShouldWork(nil, make([][]byte, 0)))
	t.Run("empty whitelisted", testKleverExecuteShouldWork(make([][]byte, 0), make([][]byte, 0)))
	t.Run("with whitelisted", testKleverExecuteShouldWork(whitelistedAddresses, expectedSortedPublicKeys))
}

func testKleverExecuteShouldWork(whitelistedAddresses [][]byte, expectedSortedPublicKeys [][]byte) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.DataGetter = &bridgeTests.DataGetterStub{
			GetAllStakedRelayersCalled: func(ctx context.Context) ([][]byte, error) {
				return whitelistedAddresses, nil
			},
		}

		krp, _ := NewKleverRoleProvider(args)
		err := krp.Execute(context.TODO())
		assert.Nil(t, err)

		for _, addr := range whitelistedAddresses {
			addressHandler, err := address.NewAddressFromBytes(addr)
			assert.Nil(t, err)
			assert.True(t, krp.IsWhitelisted(addressHandler))
		}

		// TODO: check if error should be verified
		randomAddress, _ := address.NewAddressFromBytes([]byte("random address"))

		assert.False(t, krp.IsWhitelisted(randomAddress))
		assert.False(t, krp.IsWhitelisted(nil))
		krp.mut.RLock()
		assert.Equal(t, len(whitelistedAddresses), len(krp.whitelistedAddresses))
		krp.mut.RUnlock()
		sortedPublicKeys := krp.SortedPublicKeys()
		assert.Equal(t, expectedSortedPublicKeys, sortedPublicKeys)
	}
}

func TestKleverRoleProvider_MisconfiguredAddressesShouldError(t *testing.T) {
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

	krp, _ := NewKleverRoleProvider(args)
	err := krp.Execute(context.TODO())
	assert.True(t, errors.Is(err, ErrInvalidAddressBytes))
	assert.True(t, strings.Contains(err.Error(), hex.EncodeToString(misconfiguredAddresses[2])))
	assert.Zero(t, len(krp.whitelistedAddresses))
}
