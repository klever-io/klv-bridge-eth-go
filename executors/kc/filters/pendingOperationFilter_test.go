package filters

import (
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/config"
	"github.com/klever-io/klv-bridge-eth-go/parsers"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

const ethTestAddress1 = "0x880ec53af800b5cd051531672ef4fc4de233bd5d"
const ethTestAddress2 = "0x880ebbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
const klvTestAddress1 = "klv1qqqqqqqqqqqqqpgqh46r9zh78lry2py8tq723fpjdr4pp0zgsg8syf6mq0"
const klvTestAddress2 = "klv1qqqqqqqqqqqqqpgqxjgmvqe9kvvr4xvvxflue3a7cjjeyvx9sg8snh0ljc"

var testLog = logger.GetOrCreate("filters")
var ethTestAddress1Bytes, _ = hex.DecodeString(ethTestAddress1[2:])

func createTestConfig() config.PendingOperationsFilterConfig {
	return config.PendingOperationsFilterConfig{
		DeniedEthAddresses:  nil,
		AllowedEthAddresses: []string{"*"},

		DeniedKlvAddresses:  nil,
		AllowedKlvAddresses: []string{"*"},

		DeniedTokens:  nil,
		AllowedTokens: []string{"*"},
	}
}

func TestNewPendingOperationFilter(t *testing.T) {
	t.Parallel()

	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		filter, err := NewPendingOperationFilter(createTestConfig(), nil)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errNilLogger)
	})
	t.Run("empty config should error", func(t *testing.T) {
		t.Parallel()

		filter, err := NewPendingOperationFilter(config.PendingOperationsFilterConfig{}, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errNoItemsAllowed)
	})
	t.Run("denied eth list contains wildcard should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.DeniedEthAddresses = []string{"	*  "}

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errUnsupportedMarker)
		assert.Contains(t, err.Error(), "on item at index 0 in list DeniedEthAddresses")
	})
	t.Run("denied kda list contains wildcard should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.DeniedKlvAddresses = []string{"	*  "}

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errUnsupportedMarker)
		assert.Contains(t, err.Error(), "on item at index 0 in list DeniedKlvAddresses")
	})
	t.Run("denied tokens list contains wildcard should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.DeniedTokens = []string{"	*  "}

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errUnsupportedMarker)
		assert.Contains(t, err.Error(), "on item at index 0 in list DeniedTokens")
	})
	t.Run("allowed eth list contains empty string should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.AllowedEthAddresses = append(cfg.AllowedEthAddresses, "	 ")

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errUnsupportedMarker)
		assert.Contains(t, err.Error(), "on item at index 1 in list AllowedEthAddresses")
	})
	t.Run("allowed kda list contains empty string should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.AllowedKlvAddresses = append(cfg.AllowedKlvAddresses, "	 ")

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errUnsupportedMarker)
		assert.Contains(t, err.Error(), "on item at index 1 in list AllowedKlvAddresses")
	})
	t.Run("allowed tokens list contains empty string should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.AllowedTokens = append(cfg.AllowedTokens, "	 ")

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errUnsupportedMarker)
		assert.Contains(t, err.Error(), "on item at index 1 in list AllowedTokens")
	})
	t.Run("invalid address in AllowedEthAddresses should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.AllowedEthAddresses = append(cfg.AllowedEthAddresses, "invalid address")

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errMissingEthPrefix)
		assert.Contains(t, err.Error(), "on item at index 1 in list AllowedEthAddresses")
	})
	t.Run("invalid address in DeniedEthAddresses should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.DeniedEthAddresses = append(cfg.DeniedEthAddresses, "invalid address")

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errMissingEthPrefix)
		assert.Contains(t, err.Error(), "on item at index 0 in list DeniedEthAddresses")
	})
	t.Run("invalid address in AllowedKlvAddresses should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.AllowedKlvAddresses = append(cfg.AllowedKlvAddresses, "invalid address")

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "on item at index 1 in list AllowedKlvAddresses")
	})
	t.Run("invalid address in DeniedKlvAddresses should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.DeniedKlvAddresses = append(cfg.DeniedKlvAddresses, "invalid address")

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "on item at index 0 in list DeniedKlvAddresses")
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.AllowedEthAddresses = append(cfg.AllowedKlvAddresses, ethTestAddress1)
		cfg.DeniedEthAddresses = append(cfg.DeniedEthAddresses, ethTestAddress1)
		cfg.AllowedKlvAddresses = append(cfg.AllowedKlvAddresses, klvTestAddress1)
		cfg.DeniedKlvAddresses = append(cfg.DeniedKlvAddresses, klvTestAddress1)
		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.NotNil(t, filter)
		assert.Nil(t, err)
	})
}

func TestPendingOperationFilter_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *pendingOperationFilter
	assert.True(t, instance.IsInterfaceNil())

	instance = &pendingOperationFilter{}
	assert.False(t, instance.IsInterfaceNil())
}

func TestPendingOperationFilter_ShouldExecute(t *testing.T) {
	t.Parallel()

	t.Run("nil callData.To should return false", func(t *testing.T) {
		t.Parallel()

		callData := parsers.ProxySCCompleteCallData{
			To: nil,
		}

		cfg := createTestConfig()
		filter, _ := NewPendingOperationFilter(cfg, testLog)

		assert.False(t, filter.ShouldExecute(callData))
	})
	t.Run("callData.To is not a valid KLV address should return false", func(t *testing.T) {
		t.Parallel()

		addr, _ := address.NewAddressFromBytes([]byte{0x1, 0x2})
		callData := parsers.ProxySCCompleteCallData{
			To: addr,
		}

		cfg := createTestConfig()
		filter, _ := NewPendingOperationFilter(cfg, testLog)

		assert.False(t, filter.ShouldExecute(callData))
	})
	t.Run("eth address", func(t *testing.T) {
		t.Parallel()

		callData := parsers.ProxySCCompleteCallData{
			From: common.BytesToAddress(ethTestAddress1Bytes),
		}
		callData.To, _ = address.NewAddress(klvTestAddress1)
		t.Run("is denied should return false", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.DeniedEthAddresses = []string{ethTestAddress1}
			cfg.AllowedEthAddresses = []string{ethTestAddress1}

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))

			cfg.AllowedEthAddresses = []string{"*"}
			filter, _ = NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))
		})
		t.Run("is not denied but allowed should return true", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.AllowedEthAddresses = []string{ethTestAddress1}

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.True(t, filter.ShouldExecute(callData))

			cfg.AllowedEthAddresses = []string{"*"}
			filter, _ = NewPendingOperationFilter(cfg, testLog)
			assert.True(t, filter.ShouldExecute(callData))
		})
		t.Run("is not denied but not allowed should return false", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.AllowedEthAddresses = []string{ethTestAddress2}
			cfg.AllowedTokens = nil
			cfg.AllowedKlvAddresses = nil

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))
		})
	})
	t.Run("kda address", func(t *testing.T) {
		t.Parallel()

		callData := parsers.ProxySCCompleteCallData{
			From: common.BytesToAddress(ethTestAddress1Bytes),
		}
		callData.To, _ = address.NewAddress(klvTestAddress1)
		t.Run("is denied should return false", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.DeniedKlvAddresses = []string{klvTestAddress1}
			cfg.AllowedKlvAddresses = []string{klvTestAddress1}

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))

			cfg.AllowedKlvAddresses = []string{"*"}
			filter, _ = NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))
		})
		t.Run("is not denied but allowed should return true", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.AllowedKlvAddresses = []string{klvTestAddress1}

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.True(t, filter.ShouldExecute(callData))

			cfg.AllowedKlvAddresses = []string{"*"}
			filter, _ = NewPendingOperationFilter(cfg, testLog)
			assert.True(t, filter.ShouldExecute(callData))
		})
		t.Run("is not denied but not allowed should return false", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.AllowedKlvAddresses = []string{klvTestAddress2}
			cfg.AllowedTokens = nil
			cfg.AllowedEthAddresses = nil

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))
		})
	})
	t.Run("tokens", func(t *testing.T) {
		t.Parallel()

		token1 := "tkn1"
		token2 := "tkn2"
		callData := parsers.ProxySCCompleteCallData{
			From:  common.BytesToAddress(ethTestAddress1Bytes),
			Token: token1,
		}
		callData.To, _ = address.NewAddress(klvTestAddress1)

		t.Run("is denied should return false", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.DeniedTokens = []string{token1}
			cfg.AllowedTokens = []string{token1}

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))

			cfg.AllowedTokens = []string{"*"}
			filter, _ = NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))
		})
		t.Run("is not denied but allowed should return true", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.AllowedTokens = []string{token1}

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.True(t, filter.ShouldExecute(callData))

			cfg.AllowedTokens = []string{"*"}
			filter, _ = NewPendingOperationFilter(cfg, testLog)
			assert.True(t, filter.ShouldExecute(callData))
		})
		t.Run("is not denied but not allowed should return false", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.AllowedTokens = []string{token2}
			cfg.AllowedKlvAddresses = nil
			cfg.AllowedEthAddresses = nil

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))
		})
	})
}
