package relayers

import (
	"context"
	"fmt"
	"math/big"
	"path"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/klever-io/klv-bridge-eth-go/clients/chain"
	"github.com/klever-io/klv-bridge-eth-go/config"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon"
	bridgeTests "github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	chainConfig "github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("integrationTests/relayers")

func createMockErc20ContractsHolder(tokens []common.Address, safeContractEthAddress common.Address, availableBalances []*big.Int) *bridgeTests.ERC20ContractsHolderStub {
	return &bridgeTests.ERC20ContractsHolderStub{
		BalanceOfCalled: func(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error) {
			for i, tk := range tokens {
				if tk != erc20Address {
					continue
				}

				if address == safeContractEthAddress {
					return availableBalances[i], nil
				}

				return big.NewInt(0), nil
			}

			return nil, fmt.Errorf("unregistered token %s", erc20Address.Hex())
		},
	}
}

func availableTokensMapToSlices(erc20Map map[common.Address]*big.Int) ([]common.Address, []*big.Int) {
	tokens := make([]common.Address, 0, len(erc20Map))
	availableBalances := make([]*big.Int, 0, len(erc20Map))

	for addr, val := range erc20Map {
		tokens = append(tokens, addr)
		availableBalances = append(availableBalances, val)
	}

	return tokens, availableBalances
}

func closeRelayers(relayers []bridgeComponents) {
	for _, r := range relayers {
		_ = r.Close()
	}
}

// CreateBridgeComponentsConfig -
func CreateBridgeComponentsConfig(index int, workingDir string, gasStationURL string) config.Config {
	stateMachineConfig := config.ConfigStateMachine{
		StepDurationInMillis:       1000,
		IntervalForLeaderInSeconds: 60,
	}

	return config.Config{
		Eth: config.EthereumConfig{
			Chain:                        chain.Ethereum,
			NetworkAddress:               "mock",
			MultisigContractAddress:      "3009d97FfeD62E57d444e552A9eDF9Ee6Bc8644c",
			PrivateKeyFile:               fmt.Sprintf("testdata/ethereum%d.sk", index),
			IntervalToResendTxsInSeconds: 10,
			GasLimitBase:                 350000,
			GasLimitForEach:              30000,
			GasStation: config.GasStationConfig{
				Enabled:                    len(gasStationURL) > 0,
				URL:                        gasStationURL,
				PollingIntervalInSeconds:   1,
				GasPriceMultiplier:         1,
				GasPriceSelector:           "SafeGasPrice",
				MaxFetchRetries:            3,
				MaximumAllowedGasPrice:     math.MaxUint64 / 2,
				RequestRetryDelayInSeconds: 1,
				RequestTimeInSeconds:       1,
			},
			MaxRetriesOnQuorumReached:          1,
			IntervalToWaitForTransferInSeconds: 1,
			ClientAvailabilityAllowDelta:       5,
			EventsBlockRangeFrom:               -5,
			EventsBlockRangeTo:                 50,
		},
		Klever: config.KleverConfig{
			NetworkAddress:                  "mock",
			MultisigContractAddress:         "klv1qqqqqqqqqqqqqpgqh46r9zh78lry2py8tq723fpjdr4pp0zgsg8syf6mq0",
			SafeContractAddress:             "klv1qqqqqqqqqqqqqpgqxjgmvqe9kvvr4xvvxflue3a7cjjeyvx9sg8snh0ljc",
			PrivateKeyFile:                  path.Join(workingDir, fmt.Sprintf("klever%d.pem", index)),
			IntervalToResendTxsInSeconds:    10,
			GasMap:                          testsCommon.CreateTestKleverGasMap(),
			MaxRetriesOnQuorumReached:       1,
			MaxRetriesOnWasTransferProposed: 3,
			ClientAvailabilityAllowDelta:    5,
			Proxy: config.ProxyConfig{
				CacherExpirationSeconds: 600,
				RestAPIEntityType:       "observer",
				MaxNoncesDelta:          10,
				FinalityCheck:           true,
			},
		},
		P2P: config.ConfigP2P{},
		StateMachine: map[string]config.ConfigStateMachine{
			"EthereumToKleverBlockchain": stateMachineConfig,
			"KleverBlockchainToEthereum": stateMachineConfig,
		},
		Relayer: config.ConfigRelayer{
			Marshalizer: chainConfig.MarshalizerConfig{
				Type:           "json",
				SizeCheckDelta: 10,
			},
			RoleProvider: config.RoleProviderConfig{
				PollingIntervalInMillis: 1000,
			},
		},
	}
}
