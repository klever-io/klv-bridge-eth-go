package framework

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/klever-io/klv-bridge-eth-go/clients/ethereum"
	"github.com/klever-io/klv-bridge-eth-go/config"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/factory"
	"github.com/klever-io/klv-bridge-eth-go/integrationTests"
	testsRelayers "github.com/klever-io/klv-bridge-eth-go/integrationTests/relayers"
	"github.com/klever-io/klv-bridge-eth-go/status"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/require"
)

const (
	relayerETHKeyPathFormat = "../testdata/ethereum%d.sk"
)

// BridgeComponents holds and manages the relayers components
type BridgeComponents struct {
	testing.TB
	RelayerInstances   []Relayer
	gasStationInstance *gasStation
}

// NewBridgeComponents will create the bridge components (relayers)
func NewBridgeComponents(
	tb testing.TB,
	workingDir string,
	chainSimulator ChainSimulatorWrapper,
	ethereumChain ethereum.ClientWrapper,
	erc20ContractsHolder ethereum.Erc20ContractsHolder,
	ethBackend *simulated.Backend,
	numRelayers int,
	ethSafeContractAddress string,
	kdaSafeAddress *KlvAddress,
	kdaMultisigAddress *KlvAddress,
) *BridgeComponents {
	bridge := &BridgeComponents{
		TB:                 tb,
		RelayerInstances:   make([]Relayer, 0, numRelayers),
		gasStationInstance: NewGasStation(ethBackend),
	}

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	gasStationURL := bridge.gasStationInstance.URL()
	log.Info("started gas station server", "URL", gasStationURL)

	wg := sync.WaitGroup{}
	wg.Add(numRelayers)

	for i := 0; i < numRelayers; i++ {
		generalConfigs := testsRelayers.CreateBridgeComponentsConfig(i, workingDir, gasStationURL)
		generalConfigs.Eth.PrivateKeyFile = fmt.Sprintf(relayerETHKeyPathFormat, i)
		argsBridgeComponents := factory.ArgsEthereumToKleverBridge{
			Configs: config.Configs{
				GeneralConfig:   generalConfigs,
				ApiRoutesConfig: config.ApiRoutesConfig{},
				FlagsConfig: config.ContextFlagsConfig{
					RestApiInterface: bridgeCore.WebServerOffString,
				},
			},
			Proxy:                     chainSimulator.Proxy(),
			ClientWrapper:             ethereumChain,
			Messenger:                 messengers[i],
			StatusStorer:              testsCommon.NewStorerMock(),
			TimeForBootstrap:          time.Second * 5,
			TimeBeforeRepeatJoin:      time.Second * 30,
			MetricsHolder:             status.NewMetricsHolder(),
			AppStatusHandler:          &statusHandler.AppStatusHandlerStub{},
			KleverClientStatusHandler: &testsCommon.StatusHandlerStub{},
		}
		argsBridgeComponents.Configs.GeneralConfig.Eth.SafeContractAddress = ethSafeContractAddress
		argsBridgeComponents.Erc20ContractsHolder = erc20ContractsHolder
		argsBridgeComponents.Configs.GeneralConfig.Klever.NetworkAddress = chainSimulator.GetNetworkAddress()
		argsBridgeComponents.Configs.GeneralConfig.Klever.SafeContractAddress = kdaSafeAddress.Bech32()
		argsBridgeComponents.Configs.GeneralConfig.Klever.MultisigContractAddress = kdaMultisigAddress.Bech32()
		argsBridgeComponents.Configs.GeneralConfig.Klever.GasMap = config.KleverGasMapConfig{
			Sign:                   8000000,
			ProposeTransferBase:    11000000,
			ProposeTransferForEach: 5500000,
			ProposeStatusBase:      10000000,
			ProposeStatusForEach:   7000000,
			PerformActionBase:      40000000,
			PerformActionForEach:   5500000,
			ScCallPerByte:          100000,
			ScCallPerformForEach:   10000000,
		}
		relayer, err := factory.NewEthKleverBridgeComponents(argsBridgeComponents)
		require.Nil(bridge, err)

		go func() {
			err = relayer.Start()
			log.LogIfError(err)
			require.Nil(bridge, err)
			wg.Done()
		}()

		bridge.RelayerInstances = append(bridge.RelayerInstances, relayer)
	}

	// ensure all relayers are successfully started before returning the bridge components instance
	wg.Wait()

	return bridge
}

// CloseRelayers will call close on all created relayers
func (bridge *BridgeComponents) CloseRelayers() {
	bridge.gasStationInstance.Close()

	for _, r := range bridge.RelayerInstances {
		_ = r.Close()
	}
}
