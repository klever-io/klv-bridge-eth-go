package module

import (
	"time"

	"github.com/klever-io/klever-go/tools"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/interactors/nonceHandlerV2"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/mock"
	"github.com/klever-io/klv-bridge-eth-go/config"
	"github.com/klever-io/klv-bridge-eth-go/executors/multiversx"
	"github.com/klever-io/klv-bridge-eth-go/executors/multiversx/filters"
	"github.com/klever-io/klv-bridge-eth-go/parsers"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/core/polling"
)

var suite = ed25519.NewEd25519()
var keyGen = signing.NewKeyGenerator(suite)
var singleSigner = &singlesig.Ed25519Signer{}

type scCallsModule struct {
	nonceTxsHandler  nonceTransactionsHandler
	pollingHandler   pollingHandler
	executorInstance executor
}

// NewScCallsModule creates a starts a new scCallsModule instance
func NewScCallsModule(cfg config.ScCallsModuleConfig, log logger.Logger, chCloseApp chan struct{}) (*scCallsModule, error) {
	filter, err := filters.NewPendingOperationFilter(cfg.Filter, log)
	if err != nil {
		return nil, err
	}

	// TODO: change to real klever proxy when available
	proxy := mock.CreateMockProxyKLV()

	module := &scCallsModule{}

	argNonceHandler := nonceHandlerV2.ArgsNonceTransactionsHandlerV2{
		Proxy:            proxy,
		IntervalToResend: time.Second * time.Duration(cfg.IntervalToResendTxsInSeconds),
	}
	module.nonceTxsHandler, err = nonceHandlerV2.NewNonceTransactionHandlerV2(argNonceHandler)
	if err != nil {
		return nil, err
	}

	// Public key not used in this case
	kleverChainPrivateKeyBytes, _, err := tools.LoadSkPkFromPemFile(cfg.PrivateKeyFile, 0, "")
	if err != nil {
		return nil, err
	}

	privateKey, err := keyGen.PrivateKeyFromByteArray(kleverChainPrivateKeyBytes)
	if err != nil {
		return nil, err
	}

	argsExecutor := multiversx.ArgsScCallExecutor{
		ScProxyBech32Address:            cfg.ScProxyBech32Address,
		Proxy:                           proxy,
		Codec:                           &parsers.MultiversxCodec{},
		Filter:                          filter,
		Log:                             log,
		ExtraGasToExecute:               cfg.ExtraGasToExecute,
		MaxGasLimitToUse:                cfg.MaxGasLimitToUse,
		GasLimitForOutOfGasTransactions: cfg.GasLimitForOutOfGasTransactions,
		NonceTxHandler:                  module.nonceTxsHandler,
		PrivateKey:                      privateKey,
		SingleSigner:                    singleSigner,
		CloseAppChan:                    chCloseApp,
		TransactionChecks:               cfg.TransactionChecks,
	}
	module.executorInstance, err = multiversx.NewScCallExecutor(argsExecutor)
	if err != nil {
		return nil, err
	}

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              log,
		Name:             "MultiversX SC calls",
		PollingInterval:  time.Duration(cfg.PollingIntervalInMillis) * time.Millisecond,
		PollingWhenError: time.Duration(cfg.PollingIntervalInMillis) * time.Millisecond,
		Executor:         module.executorInstance,
	}

	module.pollingHandler, err = polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return nil, err
	}

	err = module.pollingHandler.StartProcessingLoop()
	if err != nil {
		return nil, err
	}

	return module, nil
}

// GetNumSentTransaction returns the total sent transactions
func (module *scCallsModule) GetNumSentTransaction() uint32 {
	return module.executorInstance.GetNumSentTransaction()
}

// Close closes any components started
func (module *scCallsModule) Close() error {
	errPollingHandler := module.pollingHandler.Close()
	errNonceTxsHandler := module.nonceTxsHandler.Close()

	if errPollingHandler != nil {
		return errPollingHandler
	}
	return errNonceTxsHandler
}
