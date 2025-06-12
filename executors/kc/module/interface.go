package module

import (
	"context"

	"github.com/klever-io/klever-go/data/transaction"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
)

type nonceTransactionsHandler interface {
	ApplyNonceAndGasPrice(ctx context.Context, address address.Address, tx *transaction.Transaction) error
	SendTransaction(ctx context.Context, tx *transaction.Transaction) (string, error)
	Close() error
	IsInterfaceNil() bool
}

type pollingHandler interface {
	StartProcessingLoop() error
	Close() error
	IsInterfaceNil() bool
}

type executor interface {
	Execute(ctx context.Context) error
	GetNumSentTransaction() uint32
	IsInterfaceNil() bool
}
