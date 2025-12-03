package balanceValidator

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klv-bridge-eth-go/clients"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/core/batchProcessor"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

var (
	ethToken = common.BytesToAddress([]byte("eth token"))
	kdaToken = []byte("kda token")
	amount   = big.NewInt(100)
	amount2  = big.NewInt(200)
)

func createMockArgsBalanceValidator() ArgsBalanceValidator {
	return ArgsBalanceValidator{
		Log:            &testscommon.LoggerStub{},
		KCClient:       &bridge.KCClientStub{},
		EthereumClient: &bridge.EthereumClientStub{},
	}
}

type testConfiguration struct {
	isNativeOnEth      bool
	isMintBurnOnEth    bool
	totalBalancesOnEth *big.Int
	burnBalancesOnEth  *big.Int
	mintBalancesOnEth  *big.Int

	isNativeOnKlv      bool
	isMintBurnOnKlv    bool
	totalBalancesOnKlv *big.Int
	burnBalancesOnKlv  *big.Int
	mintBalancesOnKlv  *big.Int

	errorsOnCalls map[string]error

	ethToken  common.Address
	kdaToken  []byte
	amount    *big.Int
	direction batchProcessor.Direction

	lastExecutedEthBatch       uint64
	pendingKlvBatchId          uint64
	amountsOnKlvPendingBatches map[uint64][]*big.Int
	amountsOnEthPendingBatches map[uint64][]*big.Int

	// conversionMultiplier is used to simulate decimal conversion.
	// If nil, conversion returns the same value (same decimals).
	// If set, the converted value = amount * conversionMultiplier / conversionDivisor
	conversionMultiplier *big.Int
	conversionDivisor    *big.Int

	// nilConvertedAmountOnKlvBatch simulates a batch where ConvertedAmount is nil
	nilConvertedAmountOnKlvBatch bool
}

func (cfg *testConfiguration) deepClone() testConfiguration {
	result := testConfiguration{
		isNativeOnEth:                cfg.isNativeOnEth,
		isMintBurnOnEth:              cfg.isMintBurnOnEth,
		isNativeOnKlv:                cfg.isNativeOnKlv,
		isMintBurnOnKlv:              cfg.isMintBurnOnKlv,
		errorsOnCalls:                make(map[string]error),
		ethToken:                     common.HexToAddress(cfg.ethToken.Hex()),
		kdaToken:                     make([]byte, len(cfg.kdaToken)),
		direction:                    cfg.direction,
		lastExecutedEthBatch:         cfg.lastExecutedEthBatch,
		pendingKlvBatchId:            cfg.pendingKlvBatchId,
		amountsOnKlvPendingBatches:   make(map[uint64][]*big.Int),
		amountsOnEthPendingBatches:   make(map[uint64][]*big.Int),
		nilConvertedAmountOnKlvBatch: cfg.nilConvertedAmountOnKlvBatch,
	}
	if cfg.totalBalancesOnEth != nil {
		result.totalBalancesOnEth = big.NewInt(0).Set(cfg.totalBalancesOnEth)
	}
	if cfg.burnBalancesOnEth != nil {
		result.burnBalancesOnEth = big.NewInt(0).Set(cfg.burnBalancesOnEth)
	}
	if cfg.mintBalancesOnEth != nil {
		result.mintBalancesOnEth = big.NewInt(0).Set(cfg.mintBalancesOnEth)
	}
	if cfg.totalBalancesOnKlv != nil {
		result.totalBalancesOnKlv = big.NewInt(0).Set(cfg.totalBalancesOnKlv)
	}
	if cfg.burnBalancesOnKlv != nil {
		result.burnBalancesOnKlv = big.NewInt(0).Set(cfg.burnBalancesOnKlv)
	}
	if cfg.mintBalancesOnKlv != nil {
		result.mintBalancesOnKlv = big.NewInt(0).Set(cfg.mintBalancesOnKlv)
	}
	if cfg.amount != nil {
		result.amount = big.NewInt(0).Set(cfg.amount)
	}
	if cfg.conversionMultiplier != nil {
		result.conversionMultiplier = big.NewInt(0).Set(cfg.conversionMultiplier)
	}
	if cfg.conversionDivisor != nil {
		result.conversionDivisor = big.NewInt(0).Set(cfg.conversionDivisor)
	}

	for key, err := range cfg.errorsOnCalls {
		result.errorsOnCalls[key] = err
	}
	copy(result.kdaToken, cfg.kdaToken)
	for nonce, values := range cfg.amountsOnKlvPendingBatches {
		result.amountsOnKlvPendingBatches[nonce] = make([]*big.Int, 0, len(values))
		for _, value := range values {
			result.amountsOnKlvPendingBatches[nonce] = append(result.amountsOnKlvPendingBatches[nonce], big.NewInt(0).Set(value))
		}
	}
	for nonce, values := range cfg.amountsOnEthPendingBatches {
		result.amountsOnEthPendingBatches[nonce] = make([]*big.Int, 0, len(values))
		for _, value := range values {
			result.amountsOnEthPendingBatches[nonce] = append(result.amountsOnEthPendingBatches[nonce], big.NewInt(0).Set(value))
		}
	}

	return result
}

type testResult struct {
	checkRequiredBalanceOnEthCalled bool
	checkRequiredBalanceOnKlvCalled bool
	error                           error
}

func TestNewBalanceValidator(t *testing.T) {
	t.Parallel()

	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBalanceValidator()
		args.Log = nil
		instance, err := NewBalanceValidator(args)
		assert.Nil(t, instance)
		assert.Equal(t, ErrNilLogger, err)
	})
	t.Run("nil Klever Blockchain client should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBalanceValidator()
		args.KCClient = nil
		instance, err := NewBalanceValidator(args)
		assert.Nil(t, instance)
		assert.Equal(t, ErrNilKCClient, err)
	})
	t.Run("nil Ethereum client should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBalanceValidator()
		args.EthereumClient = nil
		instance, err := NewBalanceValidator(args)
		assert.Nil(t, instance)
		assert.Equal(t, ErrNilEthereumClient, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBalanceValidator()
		instance, err := NewBalanceValidator(args)
		assert.NotNil(t, instance)
		assert.Nil(t, err)
	})
}

func TestBalanceValidator_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *balanceValidator
	assert.True(t, instance.IsInterfaceNil())

	instance = &balanceValidator{}
	assert.False(t, instance.IsInterfaceNil())
}

func TestBridgeExecutor_CheckToken(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	t.Run("unknown direction should error", func(t *testing.T) {
		t.Parallel()

		cfg := testConfiguration{
			direction: "",
		}
		result := validatorTester(cfg)
		assert.ErrorIs(t, result.error, ErrInvalidDirection)
	})
	t.Run("query operations error", func(t *testing.T) {
		t.Parallel()

		t.Run("on isMintBurnOnEthereum", func(t *testing.T) {
			cfg := testConfiguration{
				direction: batchProcessor.FromKC,
				errorsOnCalls: map[string]error{
					"MintBurnTokensEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on isMintBurnOnKC", func(t *testing.T) {
			cfg := testConfiguration{
				direction: batchProcessor.ToKC,
				errorsOnCalls: map[string]error{
					"IsMintBurnTokenKlv": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on isNativeOnEthereum", func(t *testing.T) {
			cfg := testConfiguration{
				direction: batchProcessor.ToKC,
				errorsOnCalls: map[string]error{
					"NativeTokensEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on isNativeOnKC", func(t *testing.T) {
			cfg := testConfiguration{
				direction: batchProcessor.FromKC,
				errorsOnCalls: map[string]error{
					"IsNativeTokenKlv": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeEthAmount, TotalBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.FromKC,
				isMintBurnOnKlv: true,
				isNativeOnEth:   true,
				errorsOnCalls: map[string]error{
					"TotalBalancesEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeEthAmount, BurnBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.FromKC,
				isNativeOnKlv:   true,
				isMintBurnOnEth: true,
				errorsOnCalls: map[string]error{
					"BurnBalancesEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeEthAmount, MintBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.FromKC,
				isNativeOnKlv:   true,
				isMintBurnOnEth: true,
				errorsOnCalls: map[string]error{
					"MintBalancesEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeEthAmount, GetLastExecutedEthBatchID", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.FromKC,
				isNativeOnKlv:   true,
				isMintBurnOnEth: true,
				errorsOnCalls: map[string]error{
					"GetLastExecutedEthBatchIDKlv": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeEthAmount, GetBatch", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.FromKC,
				isNativeOnKlv:   true,
				isMintBurnOnEth: true,
				errorsOnCalls: map[string]error{
					"GetBatchEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeKlvAmount, TotalBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToKC,
				isNativeOnKlv:   true,
				isMintBurnOnEth: true,
				errorsOnCalls: map[string]error{
					"TotalBalancesKlv": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeKlvAmount, BurnBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToKC,
				isMintBurnOnKlv: true,
				isNativeOnEth:   true,
				errorsOnCalls: map[string]error{
					"BurnBalancesKlv": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeKlvAmount, MintBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToKC,
				isMintBurnOnKlv: true,
				isNativeOnEth:   true,
				errorsOnCalls: map[string]error{
					"MintBalancesKlv": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeKlvAmount, GetLastKCBatchID", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToKC,
				isMintBurnOnKlv: true,
				isNativeOnEth:   true,
				errorsOnCalls: map[string]error{
					"GetLastKCBatchID": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeKlvAmount, GetBatch", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToKC,
				isMintBurnOnKlv: true,
				isNativeOnEth:   true,
				errorsOnCalls: map[string]error{
					"GetBatchKlv": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeKlvAmount, WasExecuted", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToKC,
				isMintBurnOnKlv: true,
				isNativeOnEth:   true,
				errorsOnCalls: map[string]error{
					"WasExecutedEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeKlvAmount, nil ConvertedAmount in batch should error", func(t *testing.T) {
			t.Parallel()

			cfg := testConfiguration{
				direction:         batchProcessor.ToKC,
				isMintBurnOnKlv:   true,
				isNativeOnEth:     true,
				burnBalancesOnKlv: big.NewInt(100),
				mintBalancesOnKlv: big.NewInt(0),
				burnBalancesOnEth: big.NewInt(0),
				mintBalancesOnEth: big.NewInt(100),
				pendingKlvBatchId: 1,
				amountsOnKlvPendingBatches: map[uint64][]*big.Int{
					1: {big.NewInt(50)}, // Pending batch with amount but no ConvertedAmount
				},
				nilConvertedAmountOnKlvBatch: true,
				kdaToken:                     kdaToken,
				ethToken:                     ethToken,
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, clients.ErrMissingConvertedAmount)
			assert.Contains(t, result.error.Error(), "deposit nonce 0")
		})
	})
	t.Run("invalid setup", func(t *testing.T) {
		t.Parallel()

		t.Run("on Ethereum is not native nor is mint/burn, should error", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToKC,
				isMintBurnOnKlv: true,
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrInvalidSetup)
			assert.Contains(t, result.error.Error(), "isNativeOnEthereum = false, isMintBurnOnEthereum = false")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on Klever Blockchain is not native nor is mint/burn, should error", func(t *testing.T) {
			cfg := testConfiguration{
				direction:     batchProcessor.ToKC,
				isNativeOnEth: true,
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrInvalidSetup)
			assert.Contains(t, result.error.Error(), "isNativeOnKC = false, isMintBurnOnKC = false")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("native on both chains, should error", func(t *testing.T) {
			cfg := testConfiguration{
				direction:     batchProcessor.ToKC,
				isNativeOnEth: true,
				isNativeOnKlv: true,
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrInvalidSetup)
			assert.Contains(t, result.error.Error(), "isNativeOnEthereum = true, isNativeOnKC = true")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
	})
	t.Run("bad values on mint & burn", func(t *testing.T) {
		t.Parallel()

		t.Run("on Ethereum, native", func(t *testing.T) {
			t.Parallel()

			cfg := testConfiguration{
				direction:         batchProcessor.ToKC,
				isMintBurnOnEth:   true,
				isNativeOnEth:     true,
				isMintBurnOnKlv:   true,
				burnBalancesOnEth: big.NewInt(37),
				mintBalancesOnEth: big.NewInt(38),
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrNegativeAmount)
			assert.Contains(t, result.error.Error(), "ethAmount: -1")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on Ethereum, non-native", func(t *testing.T) {
			t.Parallel()

			cfg := testConfiguration{
				direction:         batchProcessor.ToKC,
				isMintBurnOnEth:   true,
				isNativeOnKlv:     true,
				burnBalancesOnEth: big.NewInt(38),
				mintBalancesOnEth: big.NewInt(37),
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrNegativeAmount)
			assert.Contains(t, result.error.Error(), "ethAmount: -1")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on KC, native", func(t *testing.T) {
			t.Parallel()

			cfg := testConfiguration{
				direction:         batchProcessor.ToKC,
				isMintBurnOnEth:   true,
				isMintBurnOnKlv:   true,
				isNativeOnKlv:     true,
				burnBalancesOnKlv: big.NewInt(37),
				mintBalancesOnKlv: big.NewInt(38),
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrNegativeAmount)
			assert.Contains(t, result.error.Error(), "kdaAmount: -1")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on KC, non-native", func(t *testing.T) {
			t.Parallel()

			cfg := testConfiguration{
				direction:         batchProcessor.ToKC,
				isNativeOnEth:     true,
				isMintBurnOnKlv:   true,
				burnBalancesOnKlv: big.NewInt(38),
				mintBalancesOnKlv: big.NewInt(37),
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrNegativeAmount)
			assert.Contains(t, result.error.Error(), "kdaAmount: -1")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
	})
	t.Run("scenarios", func(t *testing.T) {
		t.Parallel()

		t.Run("Eth -> Klv", func(t *testing.T) {
			t.Parallel()

			t.Run("native on Klv, mint-burn on Eth, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.ToKC,
					isMintBurnOnEth:    true,
					isNativeOnKlv:      true,
					burnBalancesOnEth:  big.NewInt(1100),  // initial burn (1000) + burn from this transfer (100)
					mintBalancesOnEth:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnKlv: big.NewInt(10000),
					amount:             amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Klv, mint-burn on Eth, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.ToKC,
					isMintBurnOnEth:    true,
					isNativeOnKlv:      true,
					burnBalancesOnEth:  big.NewInt(1220),  // initial burn (1000) + burn from this transfer (100) + burn from next batches (120)
					mintBalancesOnEth:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnKlv: big.NewInt(10000),
					amount:             amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Klv but with mint-burn, mint-burn on Eth, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.ToKC,
					isMintBurnOnEth:   true,
					isNativeOnKlv:     true,
					isMintBurnOnKlv:   true,
					burnBalancesOnEth: big.NewInt(1100),  // initial burn (1000) + burn from this transfer (100)
					mintBalancesOnEth: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnKlv: big.NewInt(12000),
					mintBalancesOnKlv: big.NewInt(2000), // burn - mint on Klv === mint - burn on Eth
					amount:            amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Klv but with mint-burn, mint-burn on Eth, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.ToKC,
					isMintBurnOnEth:   true,
					isNativeOnKlv:     true,
					isMintBurnOnKlv:   true,
					burnBalancesOnEth: big.NewInt(1220),  // initial burn (1000) + burn from this transfer (100) + next batches (120)
					mintBalancesOnEth: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnKlv: big.NewInt(12000),
					mintBalancesOnKlv: big.NewInt(2000), // burn - mint on Klv === mint - burn on Eth
					amount:            amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Eth, mint-burn on Klv, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.ToKC,
					isMintBurnOnKlv:    true,
					isNativeOnEth:      true,
					burnBalancesOnKlv:  big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnKlv:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnEth: big.NewInt(10100), // initial (10000) + locked from this transfer (100)
					amount:             amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnKlv.Add(cfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Eth, mint-burn on Klv, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.ToKC,
					isMintBurnOnKlv:    true,
					isNativeOnEth:      true,
					burnBalancesOnKlv:  big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnKlv:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnEth: big.NewInt(10220), // initial (10000) + locked from this transfer (100) + next batches (120)
					amount:             amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnKlv.Add(cfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Eth but with mint-burn, mint-burn on Klv, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.ToKC,
					isMintBurnOnKlv:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnKlv: big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnKlv: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnEth: big.NewInt(12100),
					mintBalancesOnEth: big.NewInt(2000), // burn - mint - transfer on Eth === mint - burn on Klv
					amount:            amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnKlv.Add(cfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Eth but with mint-burn, mint-burn on Klv, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.ToKC,
					isMintBurnOnKlv:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnKlv: big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnKlv: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnEth: big.NewInt(12220),
					mintBalancesOnEth: big.NewInt(2000), // burn - mint - transfer on Eth - next transfers === mint - burn on Klv
					amount:            amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnKlv.Add(cfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
		})

		t.Run("Klv -> Eth", func(t *testing.T) {
			t.Parallel()

			t.Run("native on Klv, mint-burn on Eth, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.FromKC,
					isMintBurnOnEth:    true,
					isNativeOnKlv:      true,
					burnBalancesOnEth:  big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnEth:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnKlv: big.NewInt(10100), // initial (10000) + transfer from this batch (100)
					amount:             amount,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Klv, mint-burn on Eth, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.FromKC,
					isMintBurnOnEth:    true,
					isNativeOnKlv:      true,
					burnBalancesOnEth:  big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnEth:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnKlv: big.NewInt(10220), // initial (10000) + transfer from this batch (100) + next batches (120)
					amount:             amount,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Klv but with mint-burn, mint-burn on Eth, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.FromKC,
					isMintBurnOnEth:   true,
					isNativeOnKlv:     true,
					isMintBurnOnKlv:   true,
					burnBalancesOnEth: big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnEth: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnKlv: big.NewInt(12100),
					mintBalancesOnKlv: big.NewInt(2000), // burn - mint - transfer on Klv === mint - burn on Eth
					amount:            amount,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Klv but with mint-burn, mint-burn on Eth, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.FromKC,
					isMintBurnOnEth:   true,
					isNativeOnKlv:     true,
					isMintBurnOnKlv:   true,
					burnBalancesOnEth: big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnEth: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnKlv: big.NewInt(12220),
					mintBalancesOnKlv: big.NewInt(2000), // burn - mint - transfer - next batches on Klv === mint - burn on Eth
					amount:            amount,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Eth, mint-burn on Klv, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.FromKC,
					isMintBurnOnKlv:    true,
					isNativeOnEth:      true,
					burnBalancesOnKlv:  big.NewInt(1100),  // initial burn (1000) + transfer from this batch (100)
					mintBalancesOnKlv:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnEth: big.NewInt(10000), // initial (10000)
					amount:             amount,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnKlv.Add(cfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Eth, mint-burn on Klv, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.FromKC,
					isMintBurnOnKlv:    true,
					isNativeOnEth:      true,
					burnBalancesOnKlv:  big.NewInt(1220),  // initial burn (1000) + transfer from this batch (100) + next batches (120)
					mintBalancesOnKlv:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnEth: big.NewInt(10000), // initial (10000)
					amount:             amount,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnKlv.Add(cfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Eth but with mint-burn, mint-burn on Klv, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.FromKC,
					isMintBurnOnKlv:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnKlv: big.NewInt(1100),  // initial burn (1000) + transfer from this batch (100)
					mintBalancesOnKlv: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnEth: big.NewInt(12000),
					mintBalancesOnEth: big.NewInt(2000), // burn - mint on Eth === mint - burn - transfer on Klv
					amount:            amount,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnKlv.Add(cfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Eth but with mint-burn, mint-burn on Klv, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.FromKC,
					isMintBurnOnKlv:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnKlv: big.NewInt(1220),  // initial burn (1000) + transfer from this batch (100) + transfer from next batches
					mintBalancesOnKlv: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnEth: big.NewInt(12000),
					mintBalancesOnEth: big.NewInt(2000), // burn - mint on Eth === mint - burn - transfer - next batches on Klv
					amount:            amount,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnKlv.Add(cfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
		})

		t.Run("Klv <-> Eth", func(t *testing.T) {
			t.Parallel()

			t.Run("from Eth: native on Klv, mint-burn on Eth, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				existingNativeBalanceKlv := int64(10000)
				existingBurnEth := int64(150000)
				existingMintEth := int64(160000)

				cfg := testConfiguration{
					direction:          batchProcessor.ToKC,
					isMintBurnOnEth:    true,
					isNativeOnKlv:      true,
					burnBalancesOnEth:  big.NewInt(existingBurnEth + 100 + 30 + 40 + 50),
					mintBalancesOnEth:  big.NewInt(existingMintEth),
					totalBalancesOnKlv: big.NewInt(existingNativeBalanceKlv + 60 + 80 + 100 + 200),
					amount:             amount,
					pendingKlvBatchId:  1,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					copiedCfg := cfg.deepClone()
					copiedCfg.burnBalancesOnEth.Add(copiedCfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(copiedCfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from Eth: native on Klv but with mint-burn, mint-burn on Eth, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				existingBurnKlv := int64(370000) // burn > mint because the token is native
				existingMintKlv := int64(360000)
				existingBurnEth := int64(150000)
				existingMintEth := int64(160000)

				cfg := testConfiguration{
					direction:         batchProcessor.ToKC,
					isMintBurnOnEth:   true,
					isNativeOnKlv:     true,
					isMintBurnOnKlv:   true,
					burnBalancesOnEth: big.NewInt(existingBurnEth + 100 + 30 + 40 + 50),
					mintBalancesOnEth: big.NewInt(existingMintEth),
					burnBalancesOnKlv: big.NewInt(existingBurnKlv + 60 + 80 + 100 + 200),
					mintBalancesOnKlv: big.NewInt(existingMintKlv),
					amount:            amount,
					pendingKlvBatchId: 1,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					copiedCfg := cfg.deepClone()
					copiedCfg.burnBalancesOnEth.Add(copiedCfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(copiedCfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from Eth: native on Eth, mint-burn on Klv, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				existingBurnKlv := int64(360000)
				existingMintKlv := int64(370000)
				existingNativeBalanceEth := int64(10000)

				cfg := testConfiguration{
					direction:          batchProcessor.ToKC,
					isMintBurnOnKlv:    true,
					isNativeOnEth:      true,
					burnBalancesOnKlv:  big.NewInt(existingBurnKlv + 200 + 60 + 80 + 100),
					mintBalancesOnKlv:  big.NewInt(existingMintKlv),
					totalBalancesOnEth: big.NewInt(existingNativeBalanceEth + 100 + 30 + 40 + 50),
					amount:             amount,
					pendingKlvBatchId:  1,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					copiedCfg := cfg.deepClone()
					copiedCfg.burnBalancesOnKlv.Add(copiedCfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(copiedCfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from Eth: native on Eth but with mint-burn, mint-burn on Klv, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				existingBurnKlv := int64(360000)
				existingMintKlv := int64(370000)
				existingBurnEth := int64(160000) // burn > mint because the token is native
				existingMintEth := int64(150000)

				cfg := testConfiguration{
					direction:         batchProcessor.ToKC,
					isMintBurnOnKlv:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnKlv: big.NewInt(existingBurnKlv + 200 + 60 + 80 + 100),
					mintBalancesOnKlv: big.NewInt(existingMintKlv),
					burnBalancesOnEth: big.NewInt(existingBurnEth + 100 + 30 + 40 + 50),
					mintBalancesOnEth: big.NewInt(existingMintEth),
					amount:            amount,
					pendingKlvBatchId: 1,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					copiedCfg := cfg.deepClone()
					copiedCfg.burnBalancesOnKlv.Add(copiedCfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(copiedCfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from Klv: native on Klv, mint-burn on Eth, ok values, with next pending batches on both chains", func(t *testing.T) {
				t.Parallel()

				existingNativeBalanceKlv := int64(10000)
				existingBurnEth := int64(150000)
				existingMintEth := int64(160000)

				cfg := testConfiguration{
					direction:          batchProcessor.FromKC,
					isMintBurnOnEth:    true,
					isNativeOnKlv:      true,
					burnBalancesOnEth:  big.NewInt(existingBurnEth + 200 + 60 + 80 + 100),
					mintBalancesOnEth:  big.NewInt(existingMintEth),
					totalBalancesOnKlv: big.NewInt(existingNativeBalanceKlv + 30 + 40 + 50 + 100),
					amount:             amount,
					pendingKlvBatchId:  1,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					copiedCfg := cfg.deepClone()
					copiedCfg.burnBalancesOnEth.Add(copiedCfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(copiedCfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from Klv: native on Klv but with mint-burn, mint-burn on Eth, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				existingBurnKlv := int64(370000) // burn > mint because the token is native
				existingMintKlv := int64(360000)
				existingBurnEth := int64(150000)
				existingMintEth := int64(160000)

				cfg := testConfiguration{
					direction:         batchProcessor.FromKC,
					isMintBurnOnEth:   true,
					isNativeOnKlv:     true,
					isMintBurnOnKlv:   true,
					burnBalancesOnEth: big.NewInt(existingBurnEth + 200 + 60 + 80 + 100),
					mintBalancesOnEth: big.NewInt(existingMintEth),
					burnBalancesOnKlv: big.NewInt(existingBurnKlv + 30 + 40 + 50 + 100),
					mintBalancesOnKlv: big.NewInt(existingMintKlv),
					amount:            amount,
					pendingKlvBatchId: 1,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					copiedCfg := cfg.deepClone()
					copiedCfg.burnBalancesOnEth.Add(copiedCfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(copiedCfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from Klv: native on Eth, mint-burn on Klv, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				existingBurnKlv := int64(360000)
				existingMintKlv := int64(370000)
				existingNativeBalanceEth := int64(10000)

				cfg := testConfiguration{
					direction:          batchProcessor.FromKC,
					isMintBurnOnKlv:    true,
					isNativeOnEth:      true,
					burnBalancesOnKlv:  big.NewInt(existingBurnKlv + 100 + 30 + 40 + 50),
					mintBalancesOnKlv:  big.NewInt(existingMintKlv),
					totalBalancesOnEth: big.NewInt(existingNativeBalanceEth + 200 + 60 + 80 + 100),
					amount:             amount,
					pendingKlvBatchId:  1,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					copiedCfg := cfg.deepClone()
					copiedCfg.burnBalancesOnKlv.Add(copiedCfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(copiedCfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from Klv: native on Eth but with mint-burn, mint-burn on Klv, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				existingBurnKlv := int64(360000)
				existingMintKlv := int64(370000)
				existingBurnEth := int64(160000) // burn > mint because the token is native
				existingMintEth := int64(150000)

				cfg := testConfiguration{
					direction:         batchProcessor.FromKC,
					isMintBurnOnKlv:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnKlv: big.NewInt(existingBurnKlv + 100 + 30 + 40 + 50),
					mintBalancesOnKlv: big.NewInt(existingMintKlv),
					burnBalancesOnEth: big.NewInt(existingBurnEth + 200 + 60 + 80 + 100),
					mintBalancesOnEth: big.NewInt(existingMintEth),
					amount:            amount,
					pendingKlvBatchId: 1,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					copiedCfg := cfg.deepClone()
					copiedCfg.burnBalancesOnKlv.Add(copiedCfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(copiedCfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
		})
	})
	t.Run("decimal conversion", func(t *testing.T) {
		t.Parallel()

		t.Run("same decimals (6 ETH, 6 KDA) - no conversion needed", func(t *testing.T) {
			t.Parallel()

			// Both chains use 6 decimals (realistic since KDA max is 8)
			// In this case, no conversion is needed
			ethBalance := big.NewInt(1e6) // 1 token in ETH decimals
			kdaBalance := big.NewInt(1e6) // 1 token in KDA decimals

			cfg := testConfiguration{
				direction:          batchProcessor.ToKC,
				isMintBurnOnEth:    true,
				isNativeOnKlv:      true,
				burnBalancesOnEth:  big.NewInt(0).Add(big.NewInt(1e6), ethBalance), // initial + current
				mintBalancesOnEth:  big.NewInt(0).Add(big.NewInt(1e6), big.NewInt(1e6)),
				totalBalancesOnKlv: kdaBalance,
				amount:             big.NewInt(1e6),
				amountsOnEthPendingBatches: map[uint64][]*big.Int{
					1: {big.NewInt(1e6)},
				},
				kdaToken: kdaToken,
				ethToken: ethToken,
				// No conversion = multiplier and divisor are 1
				conversionMultiplier: big.NewInt(1),
				conversionDivisor:    big.NewInt(1),
			}

			result := validatorTester(cfg)
			assert.Nil(t, result.error)
		})

		t.Run("different decimals (18 ETH -> 6 KDA) - conversion by smart contract", func(t *testing.T) {
			t.Parallel()

			// Token has 18 decimals on ETH and 6 decimals on KDA
			// ETH amount: 1000000000000000000 (1 token with 18 decimals)
			// KDA amount: 1000000 (1 token with 6 decimals)
			// Conversion: ETH / 10^12 = KDA
			ethBalance := big.NewInt(0).Mul(big.NewInt(1), big.NewInt(1e18)) // 1 token in ETH (18 decimals)
			kdaBalance := big.NewInt(1e6)                                    // 1 token in KDA (6 decimals)

			cfg := testConfiguration{
				direction:          batchProcessor.ToKC,
				isMintBurnOnEth:    true,
				isNativeOnKlv:      true,
				burnBalancesOnEth:  big.NewInt(0).Add(big.NewInt(1e18), ethBalance),
				mintBalancesOnEth:  big.NewInt(0).Add(big.NewInt(1e18), big.NewInt(1e18)),
				totalBalancesOnKlv: kdaBalance,
				amount:             big.NewInt(1e18),
				amountsOnEthPendingBatches: map[uint64][]*big.Int{
					1: {big.NewInt(1e18)},
				},
				kdaToken: kdaToken,
				ethToken: ethToken,
				// Conversion: divide by 10^12 (from 18 decimals to 6 decimals)
				conversionMultiplier: big.NewInt(1),
				conversionDivisor:    big.NewInt(1e12),
			}

			result := validatorTester(cfg)
			assert.Nil(t, result.error)
		})

		t.Run("different decimals (18 ETH -> 8 KDA) - conversion by smart contract", func(t *testing.T) {
			t.Parallel()

			// Token has 18 decimals on ETH and 8 decimals on KDA (max KDA decimals)
			// ETH amount: 1e18 (1 token with 18 decimals)
			// KDA amount: 1e8 (1 token with 8 decimals)
			// Conversion: ETH / 10^10 = KDA
			ethBalance := big.NewInt(1e18) // 1 token in ETH (18 decimals)
			kdaBalance := big.NewInt(1e8)  // 1 token in KDA (8 decimals)

			cfg := testConfiguration{
				direction:          batchProcessor.ToKC,
				isMintBurnOnEth:    true,
				isNativeOnKlv:      true,
				burnBalancesOnEth:  big.NewInt(0).Add(big.NewInt(1e18), ethBalance),
				mintBalancesOnEth:  big.NewInt(0).Add(big.NewInt(1e18), big.NewInt(1e18)),
				totalBalancesOnKlv: kdaBalance,
				amount:             big.NewInt(1e18),
				amountsOnEthPendingBatches: map[uint64][]*big.Int{
					1: {big.NewInt(1e18)},
				},
				kdaToken: kdaToken,
				ethToken: ethToken,
				// Conversion: divide by 10^10 (from 18 decimals to 8 decimals)
				conversionMultiplier: big.NewInt(1),
				conversionDivisor:    big.NewInt(1e10),
			}

			result := validatorTester(cfg)
			assert.Nil(t, result.error)
		})

		t.Run("different decimals with mismatch should error", func(t *testing.T) {
			t.Parallel()

			// 18 ETH decimals -> 6 KDA decimals with wrong KDA balance
			ethBalance := big.NewInt(1e18)
			kdaBalance := big.NewInt(1e6)

			cfg := testConfiguration{
				direction:          batchProcessor.ToKC,
				isMintBurnOnEth:    true,
				isNativeOnKlv:      true,
				burnBalancesOnEth:  big.NewInt(0).Add(big.NewInt(1e18), ethBalance),
				mintBalancesOnEth:  big.NewInt(0).Add(big.NewInt(1e18), big.NewInt(1e18)),
				totalBalancesOnKlv: big.NewInt(0).Add(kdaBalance, big.NewInt(1)), // Off by 1
				amount:             big.NewInt(1e18),
				amountsOnEthPendingBatches: map[uint64][]*big.Int{
					1: {big.NewInt(1e18)},
				},
				kdaToken:             kdaToken,
				ethToken:             ethToken,
				conversionMultiplier: big.NewInt(1),
				conversionDivisor:    big.NewInt(1e12),
			}

			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrBalanceMismatch)
		})

		t.Run("conversion error should propagate", func(t *testing.T) {
			t.Parallel()

			expectedError := errors.New("conversion error")
			cfg := testConfiguration{
				direction:          batchProcessor.ToKC,
				isMintBurnOnEth:    true,
				isNativeOnKlv:      true,
				burnBalancesOnEth:  big.NewInt(1e18),
				mintBalancesOnEth:  big.NewInt(1e18),
				totalBalancesOnKlv: big.NewInt(1e8), // 8 KDA decimals
				amount:             big.NewInt(1e18),
				kdaToken:           kdaToken,
				ethToken:           ethToken,
				errorsOnCalls: map[string]error{
					"ConvertEthToKdaAmount": expectedError,
				},
			}

			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
		})

		t.Run("bad conversion from smart contract should cause mismatch", func(t *testing.T) {
			t.Parallel()

			// Simulate a bug in the smart contract where conversion returns wrong value
			// Token has 18 decimals on ETH and 6 decimals on KDA
			// Correct conversion: ETH / 10^12 = KDA
			// Bad conversion: returns wrong multiplier/divisor
			ethBalance := big.NewInt(1e18) // 1 token in ETH (18 decimals)
			kdaBalance := big.NewInt(1e6)  // 1 token in KDA (6 decimals)

			cfg := testConfiguration{
				direction:          batchProcessor.ToKC,
				isMintBurnOnEth:    true,
				isNativeOnKlv:      true,
				burnBalancesOnEth:  big.NewInt(0).Add(big.NewInt(1e18), ethBalance),
				mintBalancesOnEth:  big.NewInt(0).Add(big.NewInt(1e18), big.NewInt(1e18)),
				totalBalancesOnKlv: kdaBalance,
				amount:             big.NewInt(1e18),
				amountsOnEthPendingBatches: map[uint64][]*big.Int{
					1: {big.NewInt(1e18)},
				},
				kdaToken: kdaToken,
				ethToken: ethToken,
				// BAD conversion: uses wrong divisor (10^10 instead of 10^12)
				// This simulates a misconfigured smart contract
				// Expected: 1e18 / 1e12 = 1e6, but returns: 1e18 / 1e10 = 1e8
				conversionMultiplier: big.NewInt(1),
				conversionDivisor:    big.NewInt(1e10), // Wrong! Should be 1e12
			}

			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrBalanceMismatch)
		})

		// FromKC direction tests
		t.Run("FromKC: same decimals (6 ETH, 6 KDA) - no conversion needed", func(t *testing.T) {
			t.Parallel()

			// Both chains use 6 decimals (realistic since KDA max is 8)
			// For FromKC with native on KDA: totalBalancesOnKlv is used
			// For FromKC with mint-burn on ETH: mint - burn on ETH should equal totalBalancesOnKlv (after pending adjustments)
			existingNativeBalanceKlv := big.NewInt(10e6) // 10 tokens in KDA (6 decimals)
			existingBurnEth := big.NewInt(1e6)           // 1 token burned on ETH
			existingMintEth := big.NewInt(11e6)          // 11 tokens minted on ETH

			cfg := testConfiguration{
				direction:          batchProcessor.FromKC,
				isMintBurnOnEth:    true,
				isNativeOnKlv:      true,
				burnBalancesOnEth:  big.NewInt(0).Add(existingBurnEth, big.NewInt(1e6)), // + pending ETH transfer
				mintBalancesOnEth:  existingMintEth,
				totalBalancesOnKlv: big.NewInt(0).Add(existingNativeBalanceKlv, big.NewInt(1e6)), // + pending KLV transfer
				amount:             big.NewInt(1e6),
				pendingKlvBatchId:  1,
				amountsOnKlvPendingBatches: map[uint64][]*big.Int{
					1: {big.NewInt(1e6)},
				},
				amountsOnEthPendingBatches: map[uint64][]*big.Int{
					1: {big.NewInt(1e6)},
				},
				kdaToken: kdaToken,
				ethToken: ethToken,
				// No conversion = same decimals
			}

			result := validatorTester(cfg)
			assert.Nil(t, result.error)
		})

		t.Run("FromKC: different decimals (18 ETH, 6 KDA) - conversion by smart contract", func(t *testing.T) {
			t.Parallel()

			// Token has 18 decimals on ETH and 6 decimals on KDA
			// Native balance on KDA: 10e6 (10 tokens in 6 decimals)
			// Pending transfers from KDA: 1e6 (1 token in 6 decimals)
			// ETH values need to be in 18 decimals
			// After conversion (ETH / 10^12 = KDA), ETH mint - burn should equal KDA totalBalances
			existingNativeBalanceKlv := big.NewInt(10e6)                           // 10 tokens in KDA (6 decimals)
			existingBurnEth := big.NewInt(1e18)                                    // 1 token burned on ETH (18 decimals)
			existingMintEth := big.NewInt(0).Mul(big.NewInt(11), big.NewInt(1e18)) // 11 tokens minted on ETH (18 decimals)
			pendingEthTransfer := big.NewInt(1e18)                                 // 1 token pending (ETH decimals)

			cfg := testConfiguration{
				direction:          batchProcessor.FromKC,
				isMintBurnOnEth:    true,
				isNativeOnKlv:      true,
				burnBalancesOnEth:  big.NewInt(0).Add(existingBurnEth, pendingEthTransfer), // + 1 token pending (ETH decimals)
				mintBalancesOnEth:  existingMintEth,
				totalBalancesOnKlv: big.NewInt(0).Add(existingNativeBalanceKlv, big.NewInt(1e6)), // + 1 token pending (KDA decimals)
				amount:             big.NewInt(1e6),                                              // amount in KDA decimals
				pendingKlvBatchId:  1,
				amountsOnKlvPendingBatches: map[uint64][]*big.Int{
					1: {big.NewInt(1e6)}, // ConvertedAmount in KDA decimals
				},
				amountsOnEthPendingBatches: map[uint64][]*big.Int{
					1: {big.NewInt(1e18)}, // Amount in ETH decimals
				},
				kdaToken: kdaToken,
				ethToken: ethToken,
				// Conversion: divide by 10^12 (from 18 ETH decimals to 6 KDA decimals)
				conversionMultiplier: big.NewInt(1),
				conversionDivisor:    big.NewInt(1e12),
			}

			result := validatorTester(cfg)
			assert.Nil(t, result.error)
		})

		t.Run("FromKC: different decimals (18 ETH, 8 KDA) - conversion by smart contract", func(t *testing.T) {
			t.Parallel()

			// Token has 18 decimals on ETH and 8 decimals on KDA (max KDA decimals)
			// Conversion: ETH / 10^10 = KDA
			existingNativeBalanceKlv := big.NewInt(10e8)                           // 10 tokens in KDA (8 decimals)
			existingBurnEth := big.NewInt(1e18)                                    // 1 token burned on ETH (18 decimals)
			existingMintEth := big.NewInt(0).Mul(big.NewInt(11), big.NewInt(1e18)) // 11 tokens minted on ETH (18 decimals)

			cfg := testConfiguration{
				direction:          batchProcessor.FromKC,
				isMintBurnOnEth:    true,
				isNativeOnKlv:      true,
				burnBalancesOnEth:  big.NewInt(0).Add(existingBurnEth, big.NewInt(1e18)), // + 1 token pending (ETH decimals)
				mintBalancesOnEth:  existingMintEth,
				totalBalancesOnKlv: big.NewInt(0).Add(existingNativeBalanceKlv, big.NewInt(1e8)), // + 1 token pending (KDA decimals)
				amount:             big.NewInt(1e8),
				pendingKlvBatchId:  1,
				amountsOnKlvPendingBatches: map[uint64][]*big.Int{
					1: {big.NewInt(1e8)}, // ConvertedAmount in KDA decimals
				},
				amountsOnEthPendingBatches: map[uint64][]*big.Int{
					1: {big.NewInt(1e18)}, // Amount in ETH decimals
				},
				kdaToken: kdaToken,
				ethToken: ethToken,
				// Conversion: divide by 10^10 (from 18 ETH decimals to 8 KDA decimals)
				conversionMultiplier: big.NewInt(1),
				conversionDivisor:    big.NewInt(1e10),
			}

			result := validatorTester(cfg)
			assert.Nil(t, result.error)
		})

		t.Run("FromKC: different decimals (18 ETH -> 8 KDA) with mismatch should error", func(t *testing.T) {
			t.Parallel()

			existingNativeBalanceKlv := big.NewInt(10e8) // 10 tokens in KDA decimals (8)
			// existingBurnEth: 1 token in ETH decimals (18)
			existingBurnEth := big.NewInt(1e18)
			// existingMintEth: 11 tokens in ETH decimals (18)
			existingMintEth := big.NewInt(0).Mul(big.NewInt(11), big.NewInt(1e18))

			cfg := testConfiguration{
				direction:          batchProcessor.FromKC,
				isMintBurnOnEth:    true,
				isNativeOnKlv:      true,
				burnBalancesOnEth:  big.NewInt(0).Add(existingBurnEth, big.NewInt(1e18)), // + 1 token pending
				mintBalancesOnEth:  existingMintEth,
				totalBalancesOnKlv: big.NewInt(0).Add(existingNativeBalanceKlv, big.NewInt(1e8+1)), // Off by 1 in KDA decimals
				amount:             big.NewInt(1e8),                                                // 1 token in KDA decimals
				pendingKlvBatchId:  1,
				amountsOnKlvPendingBatches: map[uint64][]*big.Int{
					1: {big.NewInt(1e8)}, // ConvertedAmount in KDA decimals
				},
				amountsOnEthPendingBatches: map[uint64][]*big.Int{
					1: {big.NewInt(1e18)}, // Amount in ETH decimals
				},
				kdaToken: kdaToken,
				ethToken: ethToken,
				// Conversion: divide by 10^10 (from 18 ETH decimals to 8 KDA decimals)
				conversionMultiplier: big.NewInt(1),
				conversionDivisor:    big.NewInt(1e10),
			}

			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrBalanceMismatch)
		})

		t.Run("FromKC: conversion error should propagate", func(t *testing.T) {
			t.Parallel()

			expectedError := errors.New("conversion error")
			cfg := testConfiguration{
				direction:          batchProcessor.FromKC,
				isMintBurnOnEth:    true,
				isNativeOnKlv:      true,
				burnBalancesOnEth:  big.NewInt(1e8),
				mintBalancesOnEth:  big.NewInt(10e8),
				totalBalancesOnKlv: big.NewInt(10e6),
				amount:             big.NewInt(1e6),
				kdaToken:           kdaToken,
				ethToken:           ethToken,
				errorsOnCalls: map[string]error{
					"ConvertEthToKdaAmount": expectedError,
				},
			}

			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
		})

		t.Run("FromKC: bad conversion from smart contract should cause mismatch", func(t *testing.T) {
			t.Parallel()

			// Simulate a bug in the smart contract where conversion returns wrong value
			// Token has 18 decimals on ETH and 6 decimals on KDA
			// Correct conversion: ETH / 10^12 = KDA
			// Bad conversion: returns wrong multiplier/divisor
			existingNativeBalanceKlv := big.NewInt(10e6)                           // 10 tokens in KDA (6 decimals)
			existingBurnEth := big.NewInt(1e18)                                    // 1 token burned on ETH (18 decimals)
			existingMintEth := big.NewInt(0).Mul(big.NewInt(11), big.NewInt(1e18)) // 11 tokens minted on ETH (18 decimals)
			pendingEthTransfer := big.NewInt(1e18)                                 // 1 token pending (ETH decimals)

			cfg := testConfiguration{
				direction:          batchProcessor.FromKC,
				isMintBurnOnEth:    true,
				isNativeOnKlv:      true,
				burnBalancesOnEth:  big.NewInt(0).Add(existingBurnEth, pendingEthTransfer), // + 1 token pending (ETH decimals)
				mintBalancesOnEth:  existingMintEth,
				totalBalancesOnKlv: big.NewInt(0).Add(existingNativeBalanceKlv, big.NewInt(1e6)), // + 1 token pending (KDA decimals)
				amount:             big.NewInt(1e6),                                              // amount in KDA decimals
				pendingKlvBatchId:  1,
				amountsOnKlvPendingBatches: map[uint64][]*big.Int{
					1: {big.NewInt(1e6)}, // ConvertedAmount in KDA decimals
				},
				amountsOnEthPendingBatches: map[uint64][]*big.Int{
					1: {big.NewInt(1e18)}, // Amount in ETH decimals
				},
				kdaToken: kdaToken,
				ethToken: ethToken,
				// BAD conversion: uses wrong divisor (10^10 instead of 10^12)
				// This simulates a misconfigured smart contract
				// Expected: 1e18 / 1e12 = 1e6, but returns: 1e18 / 1e10 = 1e8
				conversionMultiplier: big.NewInt(1),
				conversionDivisor:    big.NewInt(1e10), // Wrong! Should be 1e12
			}

			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrBalanceMismatch)
		})
	})
}

func validatorTester(cfg testConfiguration) testResult {
	args := createMockArgsBalanceValidator()

	result := testResult{}

	lastKlvBatchID := uint64(0)
	for key := range cfg.amountsOnKlvPendingBatches {
		if key > lastKlvBatchID {
			lastKlvBatchID = key
		}
	}

	args.KCClient = &bridge.KCClientStub{
		CheckRequiredBalanceCalled: func(ctx context.Context, token []byte, value *big.Int) error {
			result.checkRequiredBalanceOnKlvCalled = true
			return nil
		},
		IsMintBurnTokenCalled: func(ctx context.Context, token []byte) (bool, error) {
			err := cfg.errorsOnCalls["IsMintBurnTokenKlv"]
			if err != nil {
				return false, err
			}

			return cfg.isMintBurnOnKlv, nil
		},
		IsNativeTokenCalled: func(ctx context.Context, token []byte) (bool, error) {
			err := cfg.errorsOnCalls["IsNativeTokenKlv"]
			if err != nil {
				return false, err
			}

			return cfg.isNativeOnKlv, nil
		},
		TotalBalancesCalled: func(ctx context.Context, token []byte) (*big.Int, error) {
			err := cfg.errorsOnCalls["TotalBalancesKlv"]
			if err != nil {
				return nil, err
			}

			return returnBigIntOrZeroIfNil(cfg.totalBalancesOnKlv), nil
		},
		MintBalancesCalled: func(ctx context.Context, token []byte) (*big.Int, error) {
			err := cfg.errorsOnCalls["MintBalancesKlv"]
			if err != nil {
				return nil, err
			}

			return returnBigIntOrZeroIfNil(cfg.mintBalancesOnKlv), nil
		},
		BurnBalancesCalled: func(ctx context.Context, token []byte) (*big.Int, error) {
			err := cfg.errorsOnCalls["BurnBalancesKlv"]
			if err != nil {
				return nil, err
			}

			return returnBigIntOrZeroIfNil(cfg.burnBalancesOnKlv), nil
		},
		GetPendingBatchCalled: func(ctx context.Context) (*bridgeCore.TransferBatch, error) {
			err := cfg.errorsOnCalls["GetPendingBatchKlv"]
			if err != nil {
				return nil, err
			}

			batch := &bridgeCore.TransferBatch{
				ID: cfg.pendingKlvBatchId,
			}
			applyDummyFromKlvDepositsToBatch(cfg, batch)

			return batch, nil
		},
		GetBatchCalled: func(ctx context.Context, batchID uint64) (*bridgeCore.TransferBatch, error) {
			err := cfg.errorsOnCalls["GetBatchKlv"]
			if err != nil {
				return nil, err
			}

			if batchID > getMaxKlvPendingBatchID(cfg) {
				return nil, clients.ErrNoBatchAvailable
			}
			batch := &bridgeCore.TransferBatch{
				ID: batchID,
			}
			applyDummyFromKlvDepositsToBatch(cfg, batch)

			return batch, nil
		},
		GetLastExecutedEthBatchIDCalled: func(ctx context.Context) (uint64, error) {
			err := cfg.errorsOnCalls["GetLastExecutedEthBatchIDKlv"]
			if err != nil {
				return 0, err
			}

			return cfg.lastExecutedEthBatch, nil
		},
		GetLastKCBatchIDCalled: func(ctx context.Context) (uint64, error) {
			err := cfg.errorsOnCalls["GetLastKCBatchID"]
			if err != nil {
				return 0, err
			}

			return lastKlvBatchID, nil
		},
		ConvertEthToKdaAmountCalled: func(ctx context.Context, token []byte, amount *big.Int) (*big.Int, error) {
			err := cfg.errorsOnCalls["ConvertEthToKdaAmount"]
			if err != nil {
				return nil, err
			}

			// If no conversion ratio is set, return the same value (same decimals)
			if cfg.conversionMultiplier == nil || cfg.conversionDivisor == nil {
				return big.NewInt(0).Set(amount), nil
			}

			// Apply conversion: result = amount * multiplier / divisor
			result := big.NewInt(0).Mul(amount, cfg.conversionMultiplier)
			result.Div(result, cfg.conversionDivisor)
			return result, nil
		},
	}
	args.EthereumClient = &bridge.EthereumClientStub{
		CheckRequiredBalanceCalled: func(ctx context.Context, erc20Address common.Address, value *big.Int) error {
			result.checkRequiredBalanceOnEthCalled = true
			return nil
		},
		MintBurnTokensCalled: func(ctx context.Context, account common.Address) (bool, error) {
			err := cfg.errorsOnCalls["MintBurnTokensEth"]
			if err != nil {
				return false, err
			}

			return cfg.isMintBurnOnEth, nil
		},
		NativeTokensCalled: func(ctx context.Context, account common.Address) (bool, error) {
			err := cfg.errorsOnCalls["NativeTokensEth"]
			if err != nil {
				return false, err
			}

			return cfg.isNativeOnEth, nil
		},
		TotalBalancesCalled: func(ctx context.Context, account common.Address) (*big.Int, error) {
			err := cfg.errorsOnCalls["TotalBalancesEth"]
			if err != nil {
				return nil, err
			}

			return returnBigIntOrZeroIfNil(cfg.totalBalancesOnEth), nil
		},
		MintBalancesCalled: func(ctx context.Context, account common.Address) (*big.Int, error) {
			err := cfg.errorsOnCalls["MintBalancesEth"]
			if err != nil {
				return nil, err
			}

			return returnBigIntOrZeroIfNil(cfg.mintBalancesOnEth), nil
		},
		BurnBalancesCalled: func(ctx context.Context, account common.Address) (*big.Int, error) {
			err := cfg.errorsOnCalls["BurnBalancesEth"]
			if err != nil {
				return nil, err
			}

			return returnBigIntOrZeroIfNil(cfg.burnBalancesOnEth), nil
		},
		GetBatchCalled: func(ctx context.Context, nonce uint64) (*bridgeCore.TransferBatch, bool, error) {
			err := cfg.errorsOnCalls["GetBatchEth"]
			if err != nil {
				return nil, false, err
			}

			batch := &bridgeCore.TransferBatch{
				ID: nonce,
			}
			applyDummyFromEthDepositsToBatch(cfg, batch)

			return batch, false, nil
		},
		WasExecutedCalled: func(ctx context.Context, batchID uint64) (bool, error) {
			err := cfg.errorsOnCalls["WasExecutedEth"]
			if err != nil {
				return false, err
			}

			_, found := cfg.amountsOnKlvPendingBatches[batchID]
			return !found, nil
		},
	}

	validator, err := NewBalanceValidator(args)
	if err != nil {
		result.error = err
		return result
	}

	result.error = validator.CheckToken(context.Background(), cfg.ethToken, cfg.kdaToken, cfg.amount, cfg.direction)

	return result
}

func applyDummyFromKlvDepositsToBatch(cfg testConfiguration, batch *bridgeCore.TransferBatch) {
	if cfg.amountsOnKlvPendingBatches != nil {
		values, found := cfg.amountsOnKlvPendingBatches[batch.ID]
		if found {
			depositCounter := uint64(0)

			for _, deposit := range values {
				deposit := &bridgeCore.DepositTransfer{
					Nonce:            depositCounter,
					Amount:           big.NewInt(0).Set(deposit),
					SourceTokenBytes: kdaToken,
				}
				// Only set ConvertedAmount if not simulating nil case
				if !cfg.nilConvertedAmountOnKlvBatch {
					deposit.ConvertedAmount = big.NewInt(0).Set(deposit.Amount)
				}
				batch.Deposits = append(batch.Deposits, deposit)
			}
		}
	}
}

func applyDummyFromEthDepositsToBatch(cfg testConfiguration, batch *bridgeCore.TransferBatch) {
	if cfg.amountsOnEthPendingBatches != nil {
		values, found := cfg.amountsOnEthPendingBatches[batch.ID]
		if found {
			depositCounter := uint64(0)

			for _, deposit := range values {
				batch.Deposits = append(batch.Deposits, &bridgeCore.DepositTransfer{
					Nonce:            depositCounter,
					Amount:           big.NewInt(0).Set(deposit),
					ConvertedAmount:  big.NewInt(0).Set(deposit),
					SourceTokenBytes: ethToken.Bytes(),
				})
			}
		}
	}
}

func getMaxKlvPendingBatchID(cfg testConfiguration) uint64 {
	if cfg.amountsOnKlvPendingBatches == nil {
		return 0
	}

	maxBatchIDFound := uint64(0)
	for batchID := range cfg.amountsOnKlvPendingBatches {
		if batchID > maxBatchIDFound {
			maxBatchIDFound = batchID
		}
	}

	return maxBatchIDFound
}

func returnBigIntOrZeroIfNil(value *big.Int) *big.Int {
	if value == nil {
		return big.NewInt(0)
	}

	return big.NewInt(0).Set(value)
}
