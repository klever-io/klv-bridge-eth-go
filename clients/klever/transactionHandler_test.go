package klever

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/klever-io/klever-go/data/transaction"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/builders"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	bridgeTests "github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	cryptoMock "github.com/klever-io/klv-bridge-eth-go/testsCommon/crypto"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon/interactors"
	roleproviders "github.com/klever-io/klv-bridge-eth-go/testsCommon/roleProviders"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testSigner          = &singlesig.Ed25519Signer{}
	skBytes             = bytes.Repeat([]byte{1}, 32)
	testMultisigAddress = "klv1qqqqqqqqqqqqqpgqh46r9zh78lry2py8tq723fpjdr4pp0zgsg8syf6mq0"
	relayerAddress      = "klv132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqhje7v0"
)

func createTransactionHandlerWithMockComponents() *transactionHandler {
	sk, _ := testKeyGen.PrivateKeyFromByteArray(skBytes)
	pk := sk.GeneratePublic()
	pkBytes, _ := pk.ToByteArray()
	relayerAddress, _ := address.NewAddressFromBytes(pkBytes)

	return &transactionHandler{
		proxy:                   &interactors.ProxyStub{},
		relayerAddress:          relayerAddress,
		multisigAddressAsBech32: testMultisigAddress,
		nonceTxHandler:          &bridgeTests.NonceTransactionsHandlerStub{},
		relayerPrivateKey:       sk,
		singleSigner:            testSigner,
		roleProvider:            &roleproviders.KleverRoleProviderStub{},
	}
}

func TestTransactionHandler_SendTransactionReturnHash(t *testing.T) {
	t.Parallel()

	builder := builders.NewTxDataBuilder().Function("function").ArgBytes([]byte("buff")).ArgInt64(22)
	gasLimit := uint64(2000000)

	t.Run("get network configs errors", func(t *testing.T) {
		expectedErr := errors.New("expected error in get network configs")
		txHandlerInstance := createTransactionHandlerWithMockComponents()
		txHandlerInstance.proxy = &interactors.ProxyStub{
			GetNetworkConfigCalled: func(ctx context.Context) (*models.NetworkConfig, error) {
				return nil, expectedErr
			},
		}

		hash, err := txHandlerInstance.SendTransactionReturnHash(context.Background(), builder, gasLimit)
		assert.Empty(t, hash)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("get nonce errors", func(t *testing.T) {
		expectedErr := errors.New("expected error in get nonce")
		txHandlerInstance := createTransactionHandlerWithMockComponents()
		txHandlerInstance.nonceTxHandler = &bridgeTests.NonceTransactionsHandlerStub{
			ApplyNonceAndGasPriceCalled: func(ctx context.Context, address address.Address, tx *transaction.Transaction) error {
				return expectedErr
			},
		}

		hash, err := txHandlerInstance.SendTransactionReturnHash(context.Background(), builder, gasLimit)
		assert.Empty(t, hash)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("builder errors", func(t *testing.T) {
		txHandlerInstance := createTransactionHandlerWithMockComponents()
		erroredBuilder := builders.NewTxDataBuilder().ArgAddress(nil)

		hash, err := txHandlerInstance.SendTransactionReturnHash(context.Background(), erroredBuilder, gasLimit)
		assert.Empty(t, hash)
		assert.NotNil(t, err)
		assert.Equal(t, "nil address handler in builder.checkAddress", err.Error())
	})
	t.Run("signer errors", func(t *testing.T) {
		txHandlerInstance := createTransactionHandlerWithMockComponents()
		expectedErr := errors.New("expected error in single signer")
		txHandlerInstance.singleSigner = &cryptoMock.SingleSignerStub{
			SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}

		hash, err := txHandlerInstance.SendTransactionReturnHash(context.Background(), builder, gasLimit)
		assert.Empty(t, hash)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("relayer not whitelisted", func(t *testing.T) {
		wasWhiteListedCalled := false
		wasSendTransactionCalled := false
		txHandlerInstance := createTransactionHandlerWithMockComponents()
		txHandlerInstance.roleProvider = &roleproviders.KleverRoleProviderStub{
			IsWhitelistedCalled: func(address address.Address) bool {
				wasWhiteListedCalled = true
				return false
			},
		}
		txHandlerInstance.nonceTxHandler = &bridgeTests.NonceTransactionsHandlerStub{
			SendTransactionCalled: func(ctx context.Context, tx *transaction.Transaction) (string, error) {
				wasSendTransactionCalled = true
				return "", nil
			},
		}

		hash, err := txHandlerInstance.SendTransactionReturnHash(context.Background(), builder, gasLimit)
		assert.Empty(t, hash)
		assert.Equal(t, errRelayerNotWhitelisted, err)
		assert.True(t, wasWhiteListedCalled)
		assert.False(t, wasSendTransactionCalled)
	})
	t.Run("should work", func(t *testing.T) {
		nonce := uint64(55273)
		txHandlerInstance := createTransactionHandlerWithMockComponents()
		txHash := "tx hash"
		sendWasCalled := false

		chainID := "chain ID"

		txHandlerInstance.proxy = &interactors.ProxyStub{
			GetNetworkConfigCalled: func(ctx context.Context) (*models.NetworkConfig, error) {
				return &models.NetworkConfig{
					ChainID: chainID,
				}, nil
			},
		}

		txHandlerInstance.nonceTxHandler = &bridgeTests.NonceTransactionsHandlerStub{
			ApplyNonceAndGasPriceCalled: func(ctx context.Context, address address.Address, tx *transaction.Transaction) error {
				if getBech32Address(address) == relayerAddress {
					tx.GetRawData().Nonce = nonce

					return nil
				}

				return errors.New("unexpected address to fetch the nonce")
			},
			SendTransactionCalled: func(ctx context.Context, tx *transaction.Transaction) (string, error) {
				sendWasCalled = true
				sender, err := address.NewAddressFromBytes(tx.GetSender())
				require.Nil(t, err)
				assert.Equal(t, relayerAddress, sender.Bech32())

				require.Len(t, tx.GetContracts(), 1)
				sc, err := tx.GetContracts()[0].GetSmartContract()
				require.Nil(t, err)

				scAddr, err := address.NewAddressFromBytes(sc.Address)
				require.Nil(t, err)
				assert.Equal(t, testMultisigAddress, scAddr)

				assert.Equal(t, nonce, tx.GetNonce())
				require.Len(t, tx.GetData(), 1)
				assert.Equal(t, "function@62756666@16", string(tx.GetData()[0]))

				require.Len(t, tx.GetSignature(), 1)
				assert.Equal(t, "4d1578a5ea204fa65b209b62a508add5a003de6c8cae2908fceadb810e137ebc74fcdce534cccd05502df697d41276faf3e7decf4896dd378d88b223eef53107", hex.EncodeToString(tx.Signature[0]))
				assert.Equal(t, chainID, string(tx.GetRawData().GetChainID()))

				return txHash, nil
			},
		}

		hash, err := txHandlerInstance.SendTransactionReturnHash(context.Background(), builder, gasLimit)

		assert.Nil(t, err)
		assert.Equal(t, txHash, hash)
		assert.True(t, sendWasCalled)
	})
}
