package kleverchain

import (
	"context"

	"github.com/klever-io/klever-go/data/transaction"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/parsers"
)

// NonceTransactionsHandler represents the interface able to handle the current nonce and the transactions resend mechanism
type NonceTransactionsHandler interface {
	ApplyNonceAndGasPrice(ctx context.Context, address address.Address, tx *transaction.Transaction) error
	SendTransaction(ctx context.Context, tx *transaction.Transaction) (string, error)
	Close() error
	IsInterfaceNil() bool
}

// ScCallsExecuteFilter defines the operations supported by a filter that allows selective executions of batches
type ScCallsExecuteFilter interface {
	ShouldExecute(callData parsers.ProxySCCompleteCallData) bool
	IsInterfaceNil() bool
}

// Codec defines the operations implemented by a Kleverchain codec
type Codec interface {
	DecodeProxySCCompleteCallData(buff []byte) (parsers.ProxySCCompleteCallData, error)
	ExtractGasLimitFromRawCallData(buff []byte) (uint64, error)
	IsInterfaceNil() bool
}
