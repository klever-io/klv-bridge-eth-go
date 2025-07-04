package module

import (
	"testing"

	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	"github.com/klever-io/klv-bridge-eth-go/config"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon"
	"github.com/stretchr/testify/assert"
)

func createTestConfigs() config.ScCallsModuleConfig {
	return config.ScCallsModuleConfig{
		ScProxyBech32Address:            "klv1qqqqqqqqqqqqqpgqu2jcktadaq8mmytwglc704yfv7rezv5usg8sgzuah3",
		ExtraGasToExecute:               6000000,
		MaxGasLimitToUse:                249999999,
		GasLimitForOutOfGasTransactions: 30000000,
		NetworkAddress:                  "http://127.0.0.1:8079",
		ProxyMaxNoncesDelta:             5,
		ProxyFinalityCheck:              false,
		ProxyCacherExpirationSeconds:    60,
		ProxyRestAPIEntityType:          string(models.ObserverNode),
		IntervalToResendTxsInSeconds:    1,
		PrivateKeyFile:                  "testdata/grace.pem",
		PollingIntervalInMillis:         10000,
		Filter: config.PendingOperationsFilterConfig{
			DeniedEthAddresses:  nil,
			AllowedEthAddresses: []string{"*"},
			DeniedKlvAddresses:  nil,
			AllowedKlvAddresses: []string{"*"},
			DeniedTokens:        nil,
			AllowedTokens:       []string{"*"},
		},
	}
}

func TestNewScCallsModule(t *testing.T) {
	t.Parallel()

	t.Run("invalid filter config should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.Filter.DeniedTokens = []string{"*"}

		module, err := NewScCallsModule(cfg, &testsCommon.LoggerStub{}, nil)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "unsupported marker * on item at index 0 in list DeniedTokens")
		assert.Nil(t, module)
	})
	t.Run("invalid proxy cacher interval expiration should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.ProxyCacherExpirationSeconds = 0

		module, err := NewScCallsModule(cfg, &testsCommon.LoggerStub{}, nil)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid caching duration, provided: 0s, minimum: 1s")
		assert.Nil(t, module)
	})
	t.Run("invalid resend interval should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.IntervalToResendTxsInSeconds = 0

		module, err := NewScCallsModule(cfg, &testsCommon.LoggerStub{}, nil)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid value for intervalToResend in NewNonceTransactionHandlerV2")
		assert.Nil(t, module)
	})
	t.Run("invalid private key file should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.PrivateKeyFile = ""

		module, err := NewScCallsModule(cfg, &testsCommon.LoggerStub{}, nil)
		assert.NotNil(t, err)
		assert.Nil(t, module)
	})
	t.Run("invalid polling interval should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.PollingIntervalInMillis = 0

		module, err := NewScCallsModule(cfg, &testsCommon.LoggerStub{}, nil)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid value for PollingInterval")
		assert.Nil(t, module)
	})
	t.Run("should work with nil close app chan", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		module, err := NewScCallsModule(cfg, &testsCommon.LoggerStub{}, nil)
		assert.Nil(t, err)
		assert.NotNil(t, module)

		assert.Zero(t, module.GetNumSentTransaction())

		err = module.Close()
		assert.Nil(t, err)
	})
	t.Run("should work with nil close app chan", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.TransactionChecks.CheckTransactionResults = true
		cfg.TransactionChecks.TimeInSecondsBetweenChecks = 1
		cfg.TransactionChecks.ExecutionTimeoutInSeconds = 1
		cfg.TransactionChecks.CloseAppOnError = true
		module, err := NewScCallsModule(cfg, &testsCommon.LoggerStub{}, make(chan struct{}, 1))
		assert.Nil(t, err)
		assert.NotNil(t, module)

		assert.Zero(t, module.GetNumSentTransaction())

		err = module.Close()
		assert.Nil(t, err)
	})
}
