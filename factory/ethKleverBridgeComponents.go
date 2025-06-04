package factory

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klever-go/tools"
	ethklever "github.com/klever-io/klv-bridge-eth-go/bridges/ethMultiversX"
	"github.com/klever-io/klv-bridge-eth-go/bridges/ethMultiversX/disabled"
	ethtoklever "github.com/klever-io/klv-bridge-eth-go/bridges/ethMultiversX/steps/ethToMultiversX"
	multiversxtoeth "github.com/klever-io/klv-bridge-eth-go/bridges/ethMultiversX/steps/multiversxToEth"
	"github.com/klever-io/klv-bridge-eth-go/bridges/ethMultiversX/topology"
	"github.com/klever-io/klv-bridge-eth-go/clients"
	balanceValidatorManagement "github.com/klever-io/klv-bridge-eth-go/clients/balanceValidator"
	"github.com/klever-io/klv-bridge-eth-go/clients/chain"
	"github.com/klever-io/klv-bridge-eth-go/clients/ethereum"
	"github.com/klever-io/klv-bridge-eth-go/clients/gasManagement"
	"github.com/klever-io/klv-bridge-eth-go/clients/gasManagement/factory"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/mappers"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy"
	roleproviders "github.com/klever-io/klv-bridge-eth-go/clients/roleProviders"
	"github.com/klever-io/klv-bridge-eth-go/config"
	"github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/core/converters"
	"github.com/klever-io/klv-bridge-eth-go/core/timer"
	"github.com/klever-io/klv-bridge-eth-go/p2p"
	"github.com/klever-io/klv-bridge-eth-go/stateMachine"
	"github.com/klever-io/klv-bridge-eth-go/status"
	chainCore "github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
	chainConfig "github.com/multiversx/mx-chain-go/config"
	antifloodFactory "github.com/multiversx/mx-chain-go/process/throttle/antiflood/factory"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/core/polling"
)

const (
	minTimeForBootstrap     = time.Millisecond * 100
	minTimeBeforeRepeatJoin = time.Second * 30
	pollingDurationOnError  = time.Second * 5
)

var suite = ed25519.NewEd25519()
var keyGen = signing.NewKeyGenerator(suite)
var singleSigner = &singlesig.Ed25519Signer{}

// ArgsEthereumToKleverBridge is the arguments DTO used for creating an Ethereum to Klever bridge
type ArgsEthereumToKleverBridge struct {
	Configs                   config.Configs
	Messenger                 p2p.NetMessenger
	StatusStorer              core.Storer
	Proxy                     proxy.Proxy
	KleverClientStatusHandler core.StatusHandler
	Erc20ContractsHolder      ethereum.Erc20ContractsHolder
	ClientWrapper             ethereum.ClientWrapper
	TimeForBootstrap          time.Duration
	TimeBeforeRepeatJoin      time.Duration
	MetricsHolder             core.MetricsHolder
	AppStatusHandler          chainCore.AppStatusHandler
}

type ethKleverBridgeComponents struct {
	baseLogger                    logger.Logger
	messenger                     p2p.NetMessenger
	statusStorer                  core.Storer
	multiversXClient              ethklever.MultiversXClient
	ethClient                     ethklever.EthereumClient
	evmCompatibleChain            chain.Chain
	kleverMultisigContractAddress address.Address
	kleverSafeContractAddress     address.Address
	kleverRelayerPrivateKey       crypto.PrivateKey
	kleverRelayerAddress          address.Address
	ethereumRelayerAddress        common.Address
	mxDataGetter                  dataGetter
	proxy                         proxy.Proxy
	kleverRoleProvider            KleverRoleProvider
	ethereumRoleProvider          EthereumRoleProvider
	broadcaster                   Broadcaster
	timer                         core.Timer
	timeForBootstrap              time.Duration
	metricsHolder                 core.MetricsHolder
	addressConverter              core.AddressConverter

	ethtoKleverMachineStates    core.MachineStates
	ethtoKleverStepDuration     time.Duration
	ethtoKleverStatusHandler    core.StatusHandler
	ethtoKleverStateMachine     StateMachine
	ethtoKleverSignaturesHolder ethklever.SignaturesHolder

	multiversXToEthMachineStates core.MachineStates
	multiversXToEthStepDuration  time.Duration
	multiversXToEthStatusHandler core.StatusHandler
	multiversXToEthStateMachine  StateMachine

	mutClosableHandlers sync.RWMutex
	closableHandlers    []io.Closer

	pollingHandlers []PollingHandler

	timeBeforeRepeatJoin time.Duration
	cancelFunc           func()
	appStatusHandler     chainCore.AppStatusHandler
}

// NewethKleverBridgeComponents creates a new eth-multiversx bridge components holder
func NewEthKleverBridgeComponents(args ArgsEthereumToKleverBridge) (*ethKleverBridgeComponents, error) {
	err := checkArgsEthereumToKleverBridge(args)
	if err != nil {
		return nil, err
	}
	evmCompatibleChain := args.Configs.GeneralConfig.Eth.Chain
	ethtokleverName := evmCompatibleChain.EvmCompatibleChainToKleverBlockchainName()
	baseLogId := evmCompatibleChain.BaseLogId()
	components := &ethKleverBridgeComponents{
		baseLogger:           core.NewLoggerWithIdentifier(logger.GetOrCreate(ethtokleverName), baseLogId),
		evmCompatibleChain:   evmCompatibleChain,
		messenger:            args.Messenger,
		statusStorer:         args.StatusStorer,
		closableHandlers:     make([]io.Closer, 0),
		proxy:                args.Proxy,
		timer:                timer.NewNTPTimer(),
		timeForBootstrap:     args.TimeForBootstrap,
		timeBeforeRepeatJoin: args.TimeBeforeRepeatJoin,
		metricsHolder:        args.MetricsHolder,
		appStatusHandler:     args.AppStatusHandler,
	}

	addressConverter, err := converters.NewAddressConverter()
	if err != nil {
		return nil, clients.ErrNilAddressConverter
	}
	components.addressConverter = addressConverter

	components.addClosableComponent(components.timer)

	err = components.createKleverKeysAndAddresses(args.Configs.GeneralConfig.Klever)
	if err != nil {
		return nil, err
	}

	err = components.createDataGetter()
	if err != nil {
		return nil, err
	}

	err = components.createKleverRoleProvider(args)
	if err != nil {
		return nil, err
	}

	err = components.createMultiversXClient(args)
	if err != nil {
		return nil, err
	}

	err = components.createEthereumRoleProvider(args)
	if err != nil {
		return nil, err
	}

	err = components.createEthereumClient(args)
	if err != nil {
		return nil, err
	}

	err = components.createEthereumToKleverBlockchainBridge(args)
	if err != nil {
		return nil, err
	}

	err = components.createEthereumToKleverBlockchainStateMachine()
	if err != nil {
		return nil, err
	}

	err = components.createMultiversXToEthereumBridge(args)
	if err != nil {
		return nil, err
	}

	err = components.createMultiversXToEthereumStateMachine()
	if err != nil {
		return nil, err
	}

	return components, nil
}

func (components *ethKleverBridgeComponents) addClosableComponent(closable io.Closer) {
	components.mutClosableHandlers.Lock()
	components.closableHandlers = append(components.closableHandlers, closable)
	components.mutClosableHandlers.Unlock()
}

func checkArgsEthereumToKleverBridge(args ArgsEthereumToKleverBridge) error {
	if check.IfNil(args.Proxy) {
		return errNilProxy
	}
	if check.IfNil(args.Messenger) {
		return errNilMessenger
	}
	if check.IfNil(args.ClientWrapper) {
		return errNilEthClient
	}
	if check.IfNil(args.StatusStorer) {
		return errNilStatusStorer
	}
	if check.IfNil(args.Erc20ContractsHolder) {
		return errNilErc20ContractsHolder
	}
	if args.TimeForBootstrap < minTimeForBootstrap {
		return fmt.Errorf("%w for TimeForBootstrap, received: %v, minimum: %v", errInvalidValue, args.TimeForBootstrap, minTimeForBootstrap)
	}
	if args.TimeBeforeRepeatJoin < minTimeBeforeRepeatJoin {
		return fmt.Errorf("%w for TimeBeforeRepeatJoin, received: %v, minimum: %v", errInvalidValue, args.TimeBeforeRepeatJoin, minTimeBeforeRepeatJoin)
	}
	if check.IfNil(args.MetricsHolder) {
		return errNilMetricsHolder
	}
	if check.IfNil(args.AppStatusHandler) {
		return errNilStatusHandler
	}

	return nil
}

func (components *ethKleverBridgeComponents) createKleverKeysAndAddresses(chainConfigs config.KleverConfig) error {
	encodedSk, pbkString, err := tools.LoadSkPkFromPemFile(chainConfigs.PrivateKeyFile, 0, "")
	if err != nil {
		return err
	}

	kleverPrivateKeyBytes, err := hex.DecodeString(string(encodedSk))
	if err != nil {
		return fmt.Errorf("%w for encoded secret key", err)
	}

	components.kleverRelayerPrivateKey, err = keyGen.PrivateKeyFromByteArray(kleverPrivateKeyBytes)
	if err != nil {
		return err
	}

	components.kleverRelayerAddress, err = address.NewAddress(pbkString)
	if err != nil {
		return err
	}

	// TODO: change decoder to use klever from string to hex
	components.kleverMultisigContractAddress, err = address.NewAddress(chainConfigs.MultisigContractAddress)
	if err != nil {
		return fmt.Errorf("%w for chainConfigs.MultisigContractAddress", err)
	}

	// TODO: change decoder to use klever from string to hex
	components.kleverSafeContractAddress, err = address.NewAddress(chainConfigs.SafeContractAddress)
	if err != nil {
		return fmt.Errorf("%w for chainConfigs.SafeContractAddress", err)
	}

	return nil
}

func (components *ethKleverBridgeComponents) createDataGetter() error {
	multiversXDataGetterLogId := components.evmCompatibleChain.KleverBlockchainDataGetterLogId()
	argsKLVClientDataGetter := klever.ArgsKLVClientDataGetter{
		MultisigContractAddress: components.kleverMultisigContractAddress,
		SafeContractAddress:     components.kleverSafeContractAddress,
		RelayerAddress:          components.kleverRelayerAddress,
		Proxy:                   components.proxy,
		Log:                     core.NewLoggerWithIdentifier(logger.GetOrCreate(multiversXDataGetterLogId), multiversXDataGetterLogId),
	}

	var err error
	components.mxDataGetter, err = klever.NewKLVClientDataGetter(argsKLVClientDataGetter)

	return err
}

func (components *ethKleverBridgeComponents) createMultiversXClient(args ArgsEthereumToKleverBridge) error {
	chainConfigs := args.Configs.GeneralConfig.Klever
	tokensMapper, err := mappers.NewMultiversXToErc20Mapper(components.mxDataGetter)
	if err != nil {
		return err
	}
	multiversXClientLogId := components.evmCompatibleChain.KleverBlockchainClientLogId()

	clientArgs := klever.ClientArgs{
		GasMapConfig:                 chainConfigs.GasMap,
		Proxy:                        args.Proxy,
		Log:                          core.NewLoggerWithIdentifier(logger.GetOrCreate(multiversXClientLogId), multiversXClientLogId),
		RelayerPrivateKey:            components.kleverRelayerPrivateKey,
		MultisigContractAddress:      components.kleverMultisigContractAddress,
		SafeContractAddress:          components.kleverSafeContractAddress,
		IntervalToResendTxsInSeconds: chainConfigs.IntervalToResendTxsInSeconds,
		TokensMapper:                 tokensMapper,
		RoleProvider:                 components.kleverRoleProvider,
		StatusHandler:                args.KleverClientStatusHandler,
		ClientAvailabilityAllowDelta: chainConfigs.ClientAvailabilityAllowDelta,
	}

	components.multiversXClient, err = klever.NewClient(clientArgs)
	components.addClosableComponent(components.multiversXClient)

	return err
}

func (components *ethKleverBridgeComponents) createEthereumClient(args ArgsEthereumToKleverBridge) error {
	ethereumConfigs := args.Configs.GeneralConfig.Eth

	gasStationConfig := ethereumConfigs.GasStation
	argsGasStation := gasManagement.ArgsGasStation{
		RequestURL:             gasStationConfig.URL,
		RequestPollingInterval: time.Duration(gasStationConfig.PollingIntervalInSeconds) * time.Second,
		RequestRetryDelay:      time.Duration(gasStationConfig.RequestRetryDelayInSeconds) * time.Second,
		MaximumFetchRetries:    gasStationConfig.MaxFetchRetries,
		RequestTime:            time.Duration(gasStationConfig.RequestTimeInSeconds) * time.Second,
		MaximumGasPrice:        gasStationConfig.MaximumAllowedGasPrice,
		GasPriceSelector:       core.EthGasPriceSelector(gasStationConfig.GasPriceSelector),
		GasPriceMultiplier:     gasStationConfig.GasPriceMultiplier,
	}

	gs, err := factory.CreateGasStation(argsGasStation, gasStationConfig.Enabled)
	if err != nil {
		return err
	}

	components.addClosableComponent(gs)

	antifloodComponents, err := components.createAntifloodComponents(args.Configs.GeneralConfig.P2P.AntifloodConfig)
	if err != nil {
		return err
	}

	peerDenialEvaluator, err := p2p.NewPeerDenialEvaluator(antifloodComponents.BlacklistHandler, antifloodComponents.PubKeysCacher)
	if err != nil {
		return err
	}
	err = args.Messenger.SetPeerDenialEvaluator(peerDenialEvaluator)
	if err != nil {
		return err
	}

	broadcasterLogId := components.evmCompatibleChain.BroadcasterLogId()
	ethtokleverName := components.evmCompatibleChain.EvmCompatibleChainToKleverBlockchainName()
	argsBroadcaster := p2p.ArgsBroadcaster{
		Messenger:              args.Messenger,
		Log:                    core.NewLoggerWithIdentifier(logger.GetOrCreate(broadcasterLogId), broadcasterLogId),
		MultiversXRoleProvider: components.kleverRoleProvider,
		SignatureProcessor:     components.ethereumRoleProvider,
		KeyGen:                 keyGen,
		SingleSigner:           singleSigner,
		PrivateKey:             components.kleverRelayerPrivateKey,
		Name:                   ethtokleverName,
		AntifloodComponents:    antifloodComponents,
	}

	components.broadcaster, err = p2p.NewBroadcaster(argsBroadcaster)
	if err != nil {
		return err
	}

	cryptoHandler, err := ethereum.NewCryptoHandler(ethereumConfigs.PrivateKeyFile)
	if err != nil {
		return err
	}

	components.ethereumRelayerAddress = cryptoHandler.GetAddress()

	tokensMapper, err := mappers.NewErc20ToMultiversXMapper(components.mxDataGetter)
	if err != nil {
		return err
	}

	signaturesHolder := ethklever.NewSignatureHolder()
	components.ethtoKleverSignaturesHolder = signaturesHolder
	err = components.broadcaster.AddBroadcastClient(signaturesHolder)
	if err != nil {
		return err
	}

	safeContractAddress := common.HexToAddress(ethereumConfigs.SafeContractAddress)

	ethClientLogId := components.evmCompatibleChain.EvmCompatibleChainClientLogId()
	argsEthClient := ethereum.ArgsEthereumClient{
		ClientWrapper:                args.ClientWrapper,
		Erc20ContractsHandler:        args.Erc20ContractsHolder,
		Log:                          core.NewLoggerWithIdentifier(logger.GetOrCreate(ethClientLogId), ethClientLogId),
		AddressConverter:             components.addressConverter,
		Broadcaster:                  components.broadcaster,
		CryptoHandler:                cryptoHandler,
		TokensMapper:                 tokensMapper,
		SignatureHolder:              signaturesHolder,
		SafeContractAddress:          safeContractAddress,
		GasHandler:                   gs,
		TransferGasLimitBase:         ethereumConfigs.GasLimitBase,
		TransferGasLimitForEach:      ethereumConfigs.GasLimitForEach,
		ClientAvailabilityAllowDelta: ethereumConfigs.ClientAvailabilityAllowDelta,
		EventsBlockRangeFrom:         ethereumConfigs.EventsBlockRangeFrom,
		EventsBlockRangeTo:           ethereumConfigs.EventsBlockRangeTo,
	}

	components.ethClient, err = ethereum.NewEthereumClient(argsEthClient)

	return err
}

func (components *ethKleverBridgeComponents) createKleverRoleProvider(args ArgsEthereumToKleverBridge) error {
	configs := args.Configs.GeneralConfig
	multiversXRoleProviderLogId := components.evmCompatibleChain.KleverBlockchainRoleProviderLogId()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(multiversXRoleProviderLogId), multiversXRoleProviderLogId)

	argsRoleProvider := roleproviders.ArgsKleverRoleProvider{
		DataGetter: components.mxDataGetter,
		Log:        log,
	}

	var err error
	components.kleverRoleProvider, err = roleproviders.NewKleverRoleProvider(argsRoleProvider)
	if err != nil {
		return err
	}

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              log,
		Name:             "KleverBlockchain role provider",
		PollingInterval:  time.Duration(configs.Relayer.RoleProvider.PollingIntervalInMillis) * time.Millisecond,
		PollingWhenError: pollingDurationOnError,
		Executor:         components.kleverRoleProvider,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	components.addClosableComponent(pollingHandler)
	components.pollingHandlers = append(components.pollingHandlers, pollingHandler)

	return nil
}

func (components *ethKleverBridgeComponents) createEthereumRoleProvider(args ArgsEthereumToKleverBridge) error {
	configs := args.Configs.GeneralConfig
	ethRoleProviderLogId := components.evmCompatibleChain.EvmCompatibleChainRoleProviderLogId()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(ethRoleProviderLogId), ethRoleProviderLogId)
	argsRoleProvider := roleproviders.ArgsEthereumRoleProvider{
		EthereumChainInteractor: args.ClientWrapper,
		Log:                     log,
	}

	var err error
	components.ethereumRoleProvider, err = roleproviders.NewEthereumRoleProvider(argsRoleProvider)
	if err != nil {
		return err
	}

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              log,
		Name:             string(components.evmCompatibleChain) + " role provider",
		PollingInterval:  time.Duration(configs.Relayer.RoleProvider.PollingIntervalInMillis) * time.Millisecond,
		PollingWhenError: pollingDurationOnError,
		Executor:         components.ethereumRoleProvider,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	components.addClosableComponent(pollingHandler)
	components.pollingHandlers = append(components.pollingHandlers, pollingHandler)

	return nil
}

func (components *ethKleverBridgeComponents) createEthereumToKleverBlockchainBridge(args ArgsEthereumToKleverBridge) error {
	ethtokleverName := components.evmCompatibleChain.EvmCompatibleChainToKleverBlockchainName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(ethtokleverName), ethtokleverName)

	configs, found := args.Configs.GeneralConfig.StateMachine[ethtokleverName]
	if !found {
		return fmt.Errorf("%w for %q", errMissingConfig, ethtokleverName)
	}

	components.ethtoKleverStepDuration = time.Duration(configs.StepDurationInMillis) * time.Millisecond

	argsTopologyHandler := topology.ArgsTopologyHandler{
		PublicKeysProvider: components.kleverRoleProvider,
		Timer:              components.timer,
		IntervalForLeader:  time.Second * time.Duration(configs.IntervalForLeaderInSeconds),
		AddressBytes:       components.kleverRelayerAddress.Bytes(),
		Log:                log,
		AddressConverter:   components.addressConverter,
	}

	topologyHandler, err := topology.NewTopologyHandler(argsTopologyHandler)
	if err != nil {
		return err
	}

	components.ethtoKleverStatusHandler, err = status.NewStatusHandler(ethtokleverName, components.statusStorer)
	if err != nil {
		return err
	}

	err = components.metricsHolder.AddStatusHandler(components.ethtoKleverStatusHandler)
	if err != nil {
		return err
	}

	timeForTransferExecution := time.Second * time.Duration(args.Configs.GeneralConfig.Eth.IntervalToWaitForTransferInSeconds)

	balanceValidator, err := components.createBalanceValidator()
	if err != nil {
		return err
	}

	argsBridgeExecutor := ethklever.ArgsBridgeExecutor{
		Log:                          log,
		TopologyProvider:             topologyHandler,
		MultiversXClient:             components.multiversXClient,
		EthereumClient:               components.ethClient,
		StatusHandler:                components.ethtoKleverStatusHandler,
		TimeForWaitOnEthereum:        timeForTransferExecution,
		SignaturesHolder:             disabled.NewDisabledSignaturesHolder(),
		BalanceValidator:             balanceValidator,
		MaxQuorumRetriesOnEthereum:   args.Configs.GeneralConfig.Eth.MaxRetriesOnQuorumReached,
		MaxQuorumRetriesOnMultiversX: args.Configs.GeneralConfig.Klever.MaxRetriesOnQuorumReached,
		MaxRestriesOnWasProposed:     args.Configs.GeneralConfig.Klever.MaxRetriesOnWasTransferProposed,
	}

	bridge, err := ethklever.NewBridgeExecutor(argsBridgeExecutor)
	if err != nil {
		return err
	}

	components.ethtoKleverMachineStates, err = ethtoklever.CreateSteps(bridge)
	if err != nil {
		return err
	}

	return nil
}

func (components *ethKleverBridgeComponents) createMultiversXToEthereumBridge(args ArgsEthereumToKleverBridge) error {
	multiversXToEthName := components.evmCompatibleChain.KleverBlockchainToEvmCompatibleChainName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(multiversXToEthName), multiversXToEthName)

	configs, found := args.Configs.GeneralConfig.StateMachine[multiversXToEthName]
	if !found {
		return fmt.Errorf("%w for %q", errMissingConfig, multiversXToEthName)
	}

	components.multiversXToEthStepDuration = time.Duration(configs.StepDurationInMillis) * time.Millisecond
	argsTopologyHandler := topology.ArgsTopologyHandler{
		PublicKeysProvider: components.kleverRoleProvider,
		Timer:              components.timer,
		IntervalForLeader:  time.Second * time.Duration(configs.IntervalForLeaderInSeconds),
		AddressBytes:       components.kleverRelayerAddress.Bytes(),
		Log:                log,
		AddressConverter:   components.addressConverter,
	}

	topologyHandler, err := topology.NewTopologyHandler(argsTopologyHandler)
	if err != nil {
		return err
	}

	components.multiversXToEthStatusHandler, err = status.NewStatusHandler(multiversXToEthName, components.statusStorer)
	if err != nil {
		return err
	}

	err = components.metricsHolder.AddStatusHandler(components.multiversXToEthStatusHandler)
	if err != nil {
		return err
	}

	timeForWaitOnEthereum := time.Second * time.Duration(args.Configs.GeneralConfig.Eth.IntervalToWaitForTransferInSeconds)

	balanceValidator, err := components.createBalanceValidator()
	if err != nil {
		return err
	}

	argsBridgeExecutor := ethklever.ArgsBridgeExecutor{
		Log:                          log,
		TopologyProvider:             topologyHandler,
		MultiversXClient:             components.multiversXClient,
		EthereumClient:               components.ethClient,
		StatusHandler:                components.multiversXToEthStatusHandler,
		TimeForWaitOnEthereum:        timeForWaitOnEthereum,
		SignaturesHolder:             components.ethtoKleverSignaturesHolder,
		BalanceValidator:             balanceValidator,
		MaxQuorumRetriesOnEthereum:   args.Configs.GeneralConfig.Eth.MaxRetriesOnQuorumReached,
		MaxQuorumRetriesOnMultiversX: args.Configs.GeneralConfig.Klever.MaxRetriesOnQuorumReached,
		MaxRestriesOnWasProposed:     args.Configs.GeneralConfig.Klever.MaxRetriesOnWasTransferProposed,
	}

	bridge, err := ethklever.NewBridgeExecutor(argsBridgeExecutor)
	if err != nil {
		return err
	}

	components.multiversXToEthMachineStates, err = multiversxtoeth.CreateSteps(bridge)
	if err != nil {
		return err
	}

	return nil
}

func (components *ethKleverBridgeComponents) startPollingHandlers() error {
	for _, pollingHandler := range components.pollingHandlers {
		err := pollingHandler.StartProcessingLoop()
		if err != nil {
			return err
		}
	}

	return nil
}

// Start will start the bridge
func (components *ethKleverBridgeComponents) Start() error {
	err := components.messenger.Bootstrap()
	if err != nil {
		return err
	}

	components.baseLogger.Info("waiting for p2p bootstrap", "time", components.timeForBootstrap)
	time.Sleep(components.timeForBootstrap)

	err = components.broadcaster.RegisterOnTopics()
	if err != nil {
		return err
	}

	components.broadcaster.BroadcastJoinTopic()

	err = components.startPollingHandlers()
	if err != nil {
		return err
	}

	var ctx context.Context
	ctx, components.cancelFunc = context.WithCancel(context.Background())
	go components.startBroadcastJoinRetriesLoop(ctx)

	return nil
}

func (components *ethKleverBridgeComponents) createBalanceValidator() (ethklever.BalanceValidator, error) {
	argsBalanceValidator := balanceValidatorManagement.ArgsBalanceValidator{
		Log:              components.baseLogger,
		MultiversXClient: components.multiversXClient,
		EthereumClient:   components.ethClient,
	}

	return balanceValidatorManagement.NewBalanceValidator(argsBalanceValidator)
}

func (components *ethKleverBridgeComponents) createEthereumToKleverBlockchainStateMachine() error {
	ethtokleverName := components.evmCompatibleChain.EvmCompatibleChainToKleverBlockchainName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(ethtokleverName), ethtokleverName)

	argsStateMachine := stateMachine.ArgsStateMachine{
		StateMachineName:     ethtokleverName,
		Steps:                components.ethtoKleverMachineStates,
		StartStateIdentifier: ethtoklever.GettingPendingBatchFromEthereum,
		Log:                  log,
		StatusHandler:        components.ethtoKleverStatusHandler,
	}

	var err error
	components.ethtoKleverStateMachine, err = stateMachine.NewStateMachine(argsStateMachine)
	if err != nil {
		return err
	}

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              log,
		Name:             ethtokleverName + " State machine",
		PollingInterval:  components.ethtoKleverStepDuration,
		PollingWhenError: pollingDurationOnError,
		Executor:         components.ethtoKleverStateMachine,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	components.addClosableComponent(pollingHandler)
	components.pollingHandlers = append(components.pollingHandlers, pollingHandler)

	return nil
}

func (components *ethKleverBridgeComponents) createMultiversXToEthereumStateMachine() error {
	multiversXToEthName := components.evmCompatibleChain.KleverBlockchainToEvmCompatibleChainName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(multiversXToEthName), multiversXToEthName)

	argsStateMachine := stateMachine.ArgsStateMachine{
		StateMachineName:     multiversXToEthName,
		Steps:                components.multiversXToEthMachineStates,
		StartStateIdentifier: multiversxtoeth.GettingPendingBatchFromMultiversX,
		Log:                  log,
		StatusHandler:        components.multiversXToEthStatusHandler,
	}

	var err error
	components.multiversXToEthStateMachine, err = stateMachine.NewStateMachine(argsStateMachine)
	if err != nil {
		return err
	}

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              log,
		Name:             multiversXToEthName + " State machine",
		PollingInterval:  components.multiversXToEthStepDuration,
		PollingWhenError: pollingDurationOnError,
		Executor:         components.multiversXToEthStateMachine,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	components.addClosableComponent(pollingHandler)
	components.pollingHandlers = append(components.pollingHandlers, pollingHandler)

	return nil
}

func (components *ethKleverBridgeComponents) createAntifloodComponents(antifloodConfig chainConfig.AntifloodConfig) (*antifloodFactory.AntiFloodComponents, error) {
	var err error
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancelFunc()
		}
	}()

	cfg := chainConfig.Config{
		Antiflood: antifloodConfig,
	}
	antiFloodComponents, err := antifloodFactory.NewP2PAntiFloodComponents(ctx, cfg, components.appStatusHandler, components.messenger.ID())
	if err != nil {
		return nil, err
	}
	return antiFloodComponents, nil
}

func (components *ethKleverBridgeComponents) startBroadcastJoinRetriesLoop(ctx context.Context) {
	broadcastTimer := time.NewTimer(components.timeBeforeRepeatJoin)
	defer broadcastTimer.Stop()

	for {
		broadcastTimer.Reset(components.timeBeforeRepeatJoin)

		select {
		case <-broadcastTimer.C:
			components.baseLogger.Info("broadcast again join topic")
			components.broadcaster.BroadcastJoinTopic()
		case <-ctx.Done():
			components.baseLogger.Info("closing broadcast join topic loop")
			return

		}
	}
}

// Close will close any sub-components started
func (components *ethKleverBridgeComponents) Close() error {
	components.mutClosableHandlers.RLock()
	defer components.mutClosableHandlers.RUnlock()

	if components.cancelFunc != nil {
		components.cancelFunc()
	}

	var lastError error
	for _, closable := range components.closableHandlers {
		if closable == nil {
			components.baseLogger.Warn("programming error, nil closable component")
			continue
		}

		err := closable.Close()
		if err != nil {
			lastError = err

			components.baseLogger.Error("error closing component", "error", err)
		}
	}

	return lastError
}

// KleverRelayerAddress returns the Klever's address associated to this relayer
func (components *ethKleverBridgeComponents) KleverRelayerAddress() address.Address {
	return components.kleverRelayerAddress
}

// EthereumRelayerAddress returns the Ethereum's address associated to this relayer
func (components *ethKleverBridgeComponents) EthereumRelayerAddress() common.Address {
	return components.ethereumRelayerAddress
}
