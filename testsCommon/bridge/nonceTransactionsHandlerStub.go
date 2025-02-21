package bridge

import (
	"context"

	"github.com/klever-io/klever-go/data/transaction"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
)

// NonceTransactionsHandlerStub -
type NonceTransactionsHandlerStub struct {
	ApplyNonceAndGasPriceCalled func(ctx context.Context, address address.Address, tx *transaction.Transaction) error
	SendTransactionCalled       func(ctx context.Context, tx *transaction.Transaction) (string, error)
	CloseCalled                 func() error
}

// ApplyNonceAndGasPrice -
func (stub *NonceTransactionsHandlerStub) ApplyNonceAndGasPrice(ctx context.Context, address address.Address, tx *transaction.Transaction) error {
	if stub.ApplyNonceAndGasPriceCalled != nil {
		return stub.ApplyNonceAndGasPriceCalled(ctx, address, tx)
	}

	return nil
}

// SendTransaction -
func (stub *NonceTransactionsHandlerStub) SendTransaction(ctx context.Context, tx *transaction.Transaction) (string, error) {
	if stub.SendTransactionCalled != nil {
		return stub.SendTransactionCalled(ctx, tx)
	}

	return "", nil
}

// Close -
func (stub *NonceTransactionsHandlerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}
