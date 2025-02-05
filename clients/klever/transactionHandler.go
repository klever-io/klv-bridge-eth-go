package klever

import (
	"context"
	"encoding/json"

	"github.com/klever-io/klever-go/data/transaction"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/builders"
	crypto "github.com/multiversx/mx-chain-crypto-go"
)

type transactionHandler struct {
	proxy                   Proxy
	relayerAddress          address.Address
	multisigAddressAsBech32 string
	nonceTxHandler          NonceTransactionsHandler
	relayerPrivateKey       crypto.PrivateKey
	singleSigner            crypto.SingleSigner
	roleProvider            roleProvider
}

// SendTransactionReturnHash will try to assemble a transaction, sign it, send it and, if everything is OK, returns the transaction's hash
func (txHandler *transactionHandler) SendTransactionReturnHash(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (string, error) {
	if !txHandler.roleProvider.IsWhitelisted(txHandler.relayerAddress) {
		return "", errRelayerNotWhitelisted
	}
	tx, err := txHandler.signTransaction(ctx, builder, gasLimit)
	if err != nil {
		return "", err
	}

	// send transaction using nonceTxHandler, which just handles the nonce logic, the send proccess and broadcast is done by proxy interface
	// that he receives in the contructor
	return txHandler.nonceTxHandler.SendTransaction(context.Background(), tx)
}

func (txHandler *transactionHandler) signTransaction(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (*transaction.Transaction, error) {
	// networkConfig, err := txHandler.proxy.GetNetworkConfig(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	dataBytes, err := builder.ToDataBytes()
	if err != nil {
		return nil, err
	}

	senderByteAddress := txHandler.relayerAddress.Bytes()

	// building transaction to be signed, and send using proxy interface, but noncehandler as intermediare to help with nonce logic

	tx := transaction.NewBaseTransaction(senderByteAddress, 0, [][]byte{dataBytes}, 0, 0)
	// tx := &transaction.FrontendTransaction{
	// 	ChainID:  networkConfig.ChainID,
	// 	Version:  networkConfig.MinTransactionVersion,
	// 	GasLimit: gasLimit,
	// 	Data:     dataBytes,
	// 	Sender:   bech32Address,
	// 	Receiver: txHandler.multisigAddressAsBech32,
	// 	Value:    "0",
	// }

	// uses addressNonceHandler to fetch gas price using proxy endpoint GetNetworkConfig, in case of klever should
	// use node simulate transaction probably
	err = txHandler.nonceTxHandler.ApplyNonceAndGasPrice(context.Background(), txHandler.relayerAddress, tx)
	if err != nil {
		return nil, err
	}

	err = txHandler.signTransactionWithPrivateKey(tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// signTransactionWithPrivateKey signs a transaction with the client's private key
func (txHandler *transactionHandler) signTransactionWithPrivateKey(tx *transaction.Transaction) error {
	bytes, err := json.Marshal(&tx)
	if err != nil {
		return err
	}

	signature, err := txHandler.singleSigner.Sign(txHandler.relayerPrivateKey, bytes)
	if err != nil {
		return err
	}

	tx.AddSignature(signature)

	return nil
}

// Close will close any sub-components it uses
func (txHandler *transactionHandler) Close() error {
	return txHandler.nonceTxHandler.Close()
}
