package balanceValidator

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klv-bridge-eth-go/clients"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// ArgsBalanceValidator represents the DTO struct used in the NewBalanceValidator constructor function
type ArgsBalanceValidator struct {
	Log               logger.Logger
	KleverchainClient KleverchainClient
	EthereumClient    EthereumClient
}

type balanceValidator struct {
	log               logger.Logger
	kleverchainClient KleverchainClient
	ethereumClient    EthereumClient
}

// NewBalanceValidator creates a new instance of type balanceValidator
func NewBalanceValidator(args ArgsBalanceValidator) (*balanceValidator, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &balanceValidator{
		log:               args.Log,
		kleverchainClient: args.KleverchainClient,
		ethereumClient:    args.EthereumClient,
	}, nil
}

func checkArgs(args ArgsBalanceValidator) error {
	if check.IfNil(args.Log) {
		return ErrNilLogger
	}
	if check.IfNil(args.KleverchainClient) {
		return ErrNilKleverchainClient
	}
	if check.IfNil(args.EthereumClient) {
		return ErrNilEthereumClient
	}

	return nil
}

// CheckToken returns error if the bridge can not happen to the provided token due to faulty balance values in the contracts
func (validator *balanceValidator) CheckToken(ctx context.Context, ethToken common.Address, kdaToken []byte, amount *big.Int, direction batchProcessor.Direction) error {
	err := validator.checkRequiredBalance(ctx, ethToken, kdaToken, amount, direction)
	if err != nil {
		return err
	}

	isMintBurnOnEthereum, err := validator.isMintBurnOnEthereum(ctx, ethToken)
	if err != nil {
		return err
	}

	isMintBurnOnKleverchain, err := validator.isMintBurnOnKleverchain(ctx, kdaToken)
	if err != nil {
		return err
	}

	isNativeOnEthereum, err := validator.isNativeOnEthereum(ctx, ethToken)
	if err != nil {
		return err
	}

	isNativeOnKleverchain, err := validator.isNativeOnKleverchain(ctx, kdaToken)
	if err != nil {
		return err
	}

	if !isNativeOnEthereum && !isMintBurnOnEthereum {
		return fmt.Errorf("%w isNativeOnEthereum = %v, isMintBurnOnEthereum = %v", ErrInvalidSetup, isNativeOnEthereum, isMintBurnOnEthereum)
	}

	if !isNativeOnKleverchain && !isMintBurnOnKleverchain {
		return fmt.Errorf("%w isNativeOnKleverchain = %v, isMintBurnOnKleverchain = %v", ErrInvalidSetup, isNativeOnKleverchain, isMintBurnOnKleverchain)
	}

	if isNativeOnEthereum == isNativeOnKleverchain {
		return fmt.Errorf("%w isNativeOnEthereum = %v, isNativeOnKleverchain = %v", ErrInvalidSetup, isNativeOnEthereum, isNativeOnKleverchain)
	}

	ethAmount, err := validator.computeEthAmount(ctx, ethToken, isMintBurnOnEthereum, isNativeOnEthereum)
	if err != nil {
		return err
	}
	kdaAmount, err := validator.computeKdaAmount(ctx, kdaToken, isMintBurnOnKleverchain, isNativeOnKleverchain)
	if err != nil {
		return err
	}

	validator.log.Debug("balanceValidator.CheckToken",
		"ERC20 token", ethToken.String(),
		"ERC20 balance", ethAmount.String(),
		"KDA token", kdaToken,
		"KDA balance", kdaAmount.String(),
		"amount", amount.String(),
	)

	if ethAmount.Cmp(kdaAmount) != 0 {
		return fmt.Errorf("%w, balance for ERC20 token %s is %s and the balance for KDA token %s is %s, direction %s",
			ErrBalanceMismatch, ethToken.String(), ethAmount.String(), kdaToken, kdaAmount.String(), direction)
	}
	return nil
}

func (validator *balanceValidator) checkRequiredBalance(ctx context.Context, ethToken common.Address, kdaToken []byte, amount *big.Int, direction batchProcessor.Direction) error {
	switch direction {
	case batchProcessor.FromKleverchain:
		return validator.ethereumClient.CheckRequiredBalance(ctx, ethToken, amount)
	case batchProcessor.ToKleverchain:
		return validator.kleverchainClient.CheckRequiredBalance(ctx, kdaToken, amount)
	default:
		return fmt.Errorf("%w, direction: %s", ErrInvalidDirection, direction)
	}
}

func (validator *balanceValidator) isMintBurnOnEthereum(ctx context.Context, erc20Address common.Address) (bool, error) {
	isMintBurn, err := validator.ethereumClient.MintBurnTokens(ctx, erc20Address)
	if err != nil {
		return false, err
	}

	return isMintBurn, nil
}

func (validator *balanceValidator) isNativeOnEthereum(ctx context.Context, erc20Address common.Address) (bool, error) {
	isNative, err := validator.ethereumClient.NativeTokens(ctx, erc20Address)
	if err != nil {
		return false, err
	}
	return isNative, nil
}

func (validator *balanceValidator) isMintBurnOnKleverchain(ctx context.Context, token []byte) (bool, error) {
	isMintBurn, err := validator.kleverchainClient.IsMintBurnToken(ctx, token)
	if err != nil {
		return false, err
	}
	return isMintBurn, nil
}

func (validator *balanceValidator) isNativeOnKleverchain(ctx context.Context, token []byte) (bool, error) {
	isNative, err := validator.kleverchainClient.IsNativeToken(ctx, token)
	if err != nil {
		return false, err
	}
	return isNative, nil
}

func (validator *balanceValidator) computeEthAmount(
	ctx context.Context,
	token common.Address,
	isMintBurn bool,
	isNative bool,
) (*big.Int, error) {
	ethAmountInPendingBatches, err := validator.getTotalTransferAmountInPendingEthBatches(ctx, token)
	if err != nil {
		return nil, err
	}

	if !isMintBurn {
		// we need to subtract all locked balances on the Ethereum side (all pending, un-executed batches) so the balances
		// with the minted Kleverchain tokens will match
		total, errTotal := validator.ethereumClient.TotalBalances(ctx, token)
		if errTotal != nil {
			return nil, errTotal
		}

		return total.Sub(total, ethAmountInPendingBatches), nil
	}

	burnBalances, err := validator.ethereumClient.BurnBalances(ctx, token)
	if err != nil {
		return nil, err
	}
	mintBalances, err := validator.ethereumClient.MintBalances(ctx, token)
	if err != nil {
		return nil, err
	}

	// we need to cancel out what was burned in advance when the deposit was registered in the contract
	burnBalances.Sub(burnBalances, ethAmountInPendingBatches)

	var ethAmount *big.Int
	if isNative {
		ethAmount = big.NewInt(0).Sub(burnBalances, mintBalances)
	} else {
		ethAmount = big.NewInt(0).Sub(mintBalances, burnBalances)
	}

	if ethAmount.Cmp(big.NewInt(0)) < 0 {
		return big.NewInt(0), fmt.Errorf("%w, ethAmount: %s", ErrNegativeAmount, ethAmount.String())
	}
	return ethAmount, nil
}

func (validator *balanceValidator) computeKdaAmount(
	ctx context.Context,
	token []byte,
	isMintBurn bool,
	isNative bool,
) (*big.Int, error) {
	kdaAmountInPendingBatches, err := validator.getTotalTransferAmountInPendingKlvBatches(ctx, token)
	if err != nil {
		return nil, err
	}

	if !isMintBurn {
		// we need to subtract all locked balances on the Kleverchain side (all pending, un-executed batches) so the balances
		// with the minted Ethereum tokens will match
		total, errTotal := validator.kleverchainClient.TotalBalances(ctx, token)
		if errTotal != nil {
			return nil, errTotal
		}

		return total.Sub(total, kdaAmountInPendingBatches), nil
	}

	burnBalances, err := validator.kleverchainClient.BurnBalances(ctx, token)
	if err != nil {
		return nil, err
	}
	mintBalances, err := validator.kleverchainClient.MintBalances(ctx, token)
	if err != nil {
		return nil, err
	}
	var kdaAmount *big.Int

	// we need to cancel out what was burned in advance when the deposit was registered in the contract
	burnBalances.Sub(burnBalances, kdaAmountInPendingBatches)

	if isNative {
		kdaAmount = big.NewInt(0).Sub(burnBalances, mintBalances)
	} else {
		kdaAmount = big.NewInt(0).Sub(mintBalances, burnBalances)
	}

	if kdaAmount.Cmp(big.NewInt(0)) < 0 {
		return big.NewInt(0), fmt.Errorf("%w, kdaAmount: %s", ErrNegativeAmount, kdaAmount.String())
	}
	return kdaAmount, nil
}

func getTotalAmountFromBatch(batch *bridgeCore.TransferBatch, token []byte) *big.Int {
	amount := big.NewInt(0)
	for _, deposit := range batch.Deposits {
		if bytes.Equal(deposit.SourceTokenBytes, token) {
			amount.Add(amount, deposit.Amount)
		}
	}

	return amount
}

func (validator *balanceValidator) getTotalTransferAmountInPendingKlvBatches(ctx context.Context, kdaToken []byte) (*big.Int, error) {
	batchID, err := validator.kleverchainClient.GetLastKlvBatchID(ctx)
	if err != nil {
		return nil, err
	}

	var batch *bridgeCore.TransferBatch
	amount := big.NewInt(0)
	for {
		batch, err = validator.kleverchainClient.GetBatch(ctx, batchID)
		if errors.Is(err, clients.ErrNoBatchAvailable) {
			return amount, nil
		}
		if err != nil {
			return nil, err
		}

		wasExecuted, errWasExecuted := validator.ethereumClient.WasExecuted(ctx, batch.ID)
		if errWasExecuted != nil {
			return nil, errWasExecuted
		}
		if wasExecuted {
			return amount, nil
		}

		amountFromBatch := getTotalAmountFromBatch(batch, kdaToken)
		amount.Add(amount, amountFromBatch)
		batchID-- // go to the previous batch
	}
}

func (validator *balanceValidator) getTotalTransferAmountInPendingEthBatches(ctx context.Context, ethToken common.Address) (*big.Int, error) {
	batchID, err := validator.kleverchainClient.GetLastExecutedEthBatchID(ctx)
	if err != nil {
		return nil, err
	}

	var batch *bridgeCore.TransferBatch
	amount := big.NewInt(0)
	for {
		batch, _, err = validator.ethereumClient.GetBatch(ctx, batchID+1) // we take all batches, regardless if they are final or not
		if err != nil {
			return nil, err
		}

		isBatchInvalid := batch.ID != batchID+1 || len(batch.Deposits) == 0
		if isBatchInvalid {
			return amount, nil
		}

		amountFromBatch := getTotalAmountFromBatch(batch, ethToken.Bytes())
		amount.Add(amount, amountFromBatch)
		batchID++
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (validator *balanceValidator) IsInterfaceNil() bool {
	return validator == nil
}
