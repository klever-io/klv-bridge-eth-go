package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	ethereumClient "github.com/klever-io/klv-bridge-eth-go/clients/ethereum"
	"github.com/klever-io/klv-bridge-eth-go/clients/gasManagement"
	"github.com/klever-io/klv-bridge-eth-go/clients/gasManagement/factory"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	"github.com/klever-io/klv-bridge-eth-go/cmd/migration/disabled"
	"github.com/klever-io/klv-bridge-eth-go/config"
	"github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/executors/ethereum"
	"github.com/klever-io/klv-bridge-eth-go/executors/ethereum/bridgeV2Wrappers"
	"github.com/klever-io/klv-bridge-eth-go/executors/ethereum/bridgeV2Wrappers/contract"
	chainCore "github.com/multiversx/mx-chain-core-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/urfave/cli"
)

const (
	filePathPlaceholder  = "[path]"
	queryMode            = "query"
	signMode             = "sign"
	executeMode          = "execute"
	configPath           = "config"
	timestampPlaceholder = "[timestamp]"
	publicKeyPlaceholder = "[public-key]"
)

var log = logger.GetOrCreate("main")

type internalComponents struct {
	creator              BatchCreator
	batch                *ethereum.BatchInfo
	cryptoHandler        ethereumClient.CryptoHandler
	ethClient            *ethclient.Client
	ethereumChainWrapper ethereum.EthereumChainWrapper
}

func main() {
	app := cli.NewApp()
	app.Name = "Funds migration CLI tool"
	app.Usage = "This is the entry point for the migration CLI tool"
	app.Flags = getFlags()
	app.Authors = []cli.Author{
		{
			Name:  "The Klever Blockchain Team",
			Email: "contact@klever.io",
		},
	}

	app.Action = func(c *cli.Context) error {
		return execute(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	log.Info("process finished successfully")
}

func execute(ctx *cli.Context) error {
	flagsConfig := getFlagsConfig(ctx)

	err := logger.SetLogLevel(flagsConfig.LogLevel)
	if err != nil {
		return err
	}

	cfg, err := loadConfig(flagsConfig.ConfigurationFile)
	if err != nil {
		return err
	}

	log.Info("starting migration help tool", "pid", os.Getpid())

	operationMode := strings.ToLower(ctx.GlobalString(mode.Name))
	switch operationMode {
	case queryMode:
		return executeQuery(cfg)
	case signMode:
		_, err = generateAndSign(ctx, cfg)
		return err
	case executeMode:
		return executeTransfer(ctx, cfg)
	}

	return fmt.Errorf("unknown execution mode: %s", operationMode)
}

func executeQuery(cfg config.MigrationToolConfig) error {
	components, err := createInternalComponentsWithBatchCreator(cfg)
	if err != nil {
		return err
	}

	dummyEthAddress := common.Address{}
	info, err := components.creator.CreateBatchInfo(context.Background(), dummyEthAddress, nil)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Token balances for ERC20 safe address %s\n%s",
		cfg.Eth.SafeContractAddress,
		ethereum.TokensBalancesDisplayString(info),
	))

	return nil
}

func createInternalComponentsWithBatchCreator(cfg config.MigrationToolConfig) (*internalComponents, error) {
	argsProxy := proxy.ArgsProxy{
		ProxyURL:            cfg.Klever.NetworkAddress,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       cfg.Klever.Proxy.FinalityCheck,
		AllowedDeltaToFinal: cfg.Klever.Proxy.MaxNoncesDelta,
		CacheExpirationTime: time.Second * time.Duration(cfg.Klever.Proxy.CacherExpirationSeconds),
		EntityType:          models.RestAPIEntityType(cfg.Klever.Proxy.RestAPIEntityType),
	}
	proxy, err := proxy.NewProxy(argsProxy)
	if err != nil {
		return nil, err
	}

	dummyAddress, err := address.NewAddressFromBytes(bytes.Repeat([]byte{0x1}, 32))
	if err != nil {
		return nil, err
	}

	multisigAddress, err := address.NewAddress(cfg.Klever.MultisigContractAddress)
	if err != nil {
		return nil, err
	}

	safeAddress, err := address.NewAddress(cfg.Klever.SafeContractAddress)
	if err != nil {
		return nil, err
	}

	argsKLVClientDataGetter := klever.ArgsKLVClientDataGetter{
		MultisigContractAddress: multisigAddress,
		SafeContractAddress:     safeAddress,
		RelayerAddress:          dummyAddress,
		Proxy:                   proxy,
		Log:                     log,
	}
	KLVDataGetter, err := klever.NewKLVClientDataGetter(argsKLVClientDataGetter)
	if err != nil {
		return nil, err
	}

	ethClient, err := ethclient.Dial(cfg.Eth.NetworkAddress)
	if err != nil {
		return nil, err
	}

	argsContractsHolder := ethereumClient.ArgsErc20SafeContractsHolder{
		EthClient:              ethClient,
		EthClientStatusHandler: &disabled.StatusHandler{},
	}
	erc20ContractsHolder, err := ethereumClient.NewErc20SafeContractsHolder(argsContractsHolder)
	if err != nil {
		return nil, err
	}

	safeEthAddress := common.HexToAddress(cfg.Eth.SafeContractAddress)

	bridgeEthAddress := common.HexToAddress(cfg.Eth.MultisigContractAddress)
	multiSigInstance, err := contract.NewBridge(bridgeEthAddress, ethClient)
	if err != nil {
		return nil, err
	}

	argsClientWrapper := bridgeV2Wrappers.ArgsEthereumChainWrapper{
		StatusHandler:    &disabled.StatusHandler{},
		MultiSigContract: multiSigInstance,
		BlockchainClient: ethClient,
	}
	ethereumChainWrapper, err := bridgeV2Wrappers.NewEthereumChainWrapper(argsClientWrapper)
	if err != nil {
		return nil, err
	}

	argsCreator := ethereum.ArgsMigrationBatchCreator{
		KlvDataGetter:        KLVDataGetter,
		Erc20ContractsHolder: erc20ContractsHolder,
		SafeContractAddress:  safeEthAddress,
		EthereumChainWrapper: ethereumChainWrapper,
		Logger:               log,
	}

	creator, err := ethereum.NewMigrationBatchCreator(argsCreator)
	if err != nil {
		return nil, err
	}

	return &internalComponents{
		creator:              creator,
		ethClient:            ethClient,
		ethereumChainWrapper: ethereumChainWrapper,
	}, nil
}

func generateAndSign(ctx *cli.Context, cfg config.MigrationToolConfig) (*internalComponents, error) {
	components, err := createInternalComponentsWithBatchCreator(cfg)
	if err != nil {
		return nil, err
	}

	newSafeAddressString := ctx.GlobalString(newSafeAddress.Name)
	if len(newSafeAddressString) == 0 {
		return nil, fmt.Errorf("invalid new safe address for Ethereum")
	}
	newSafeAddressValue := common.HexToAddress(ctx.GlobalString(newSafeAddress.Name))

	partialMigration, err := ethereum.ConvertPartialMigrationStringToMap(ctx.GlobalString(partialMigration.Name))
	if err != nil {
		return nil, err
	}

	components.batch, err = components.creator.CreateBatchInfo(context.Background(), newSafeAddressValue, partialMigration)
	if err != nil {
		return nil, err
	}

	val, err := json.MarshalIndent(components.batch, "", "  ")
	if err != nil {
		return nil, err
	}

	components.cryptoHandler, err = ethereumClient.NewCryptoHandler(cfg.Eth.PrivateKeyFile)
	if err != nil {
		return nil, err
	}

	log.Info("signing batch", "message hash", components.batch.MessageHash.String(),
		"public key", components.cryptoHandler.GetAddress().String())

	signature, err := components.cryptoHandler.Sign(components.batch.MessageHash)
	if err != nil {
		return nil, err
	}

	log.Info("Migration .json file contents: \n" + string(val))

	jsonFilename := ctx.GlobalString(migrationJsonFile.Name)
	jsonFilename = applyTimestamp(jsonFilename)
	err = os.WriteFile(jsonFilename, val, os.ModePerm)
	if err != nil {
		return nil, err
	}

	sigInfo := &ethereum.SignatureInfo{
		Address:     components.cryptoHandler.GetAddress().String(),
		MessageHash: components.batch.MessageHash.String(),
		Signature:   hex.EncodeToString(signature),
	}

	sigFilename := ctx.GlobalString(signatureJsonFile.Name)
	sigFilename = applyTimestamp(sigFilename)
	sigFilename = applyPublicKey(sigFilename, sigInfo.Address)
	val, err = json.MarshalIndent(sigInfo, "", "  ")
	if err != nil {
		return nil, err
	}

	log.Info("Signature .json file contents: \n" + string(val))

	err = os.WriteFile(sigFilename, val, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return components, nil
}

func executeTransfer(ctx *cli.Context, cfg config.MigrationToolConfig) error {
	components, err := generateAndSign(ctx, cfg)
	if err != nil {
		return err
	}

	gasStationConfig := cfg.Eth.GasStation
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

	args := ethereum.ArgsMigrationBatchExecutor{
		EthereumChainWrapper:    components.ethereumChainWrapper,
		CryptoHandler:           components.cryptoHandler,
		Batch:                   *components.batch,
		Signatures:              ethereum.LoadAllSignatures(log, configPath),
		Logger:                  log,
		GasHandler:              gs,
		TransferGasLimitBase:    cfg.Eth.GasLimitBase,
		TransferGasLimitForEach: cfg.Eth.GasLimitForEach,
	}

	executor, err := ethereum.NewMigrationBatchExecutor(args)
	if err != nil {
		return err
	}

	return executor.ExecuteTransfer(context.Background())
}

func loadConfig(filepath string) (config.MigrationToolConfig, error) {
	cfg := config.MigrationToolConfig{}
	err := chainCore.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.MigrationToolConfig{}, err
	}

	return cfg, nil
}

func applyTimestamp(input string) string {
	actualTimestamp := time.Now().Format("2006-01-02T15-04-05")
	actualTimestamp = strings.Replace(actualTimestamp, "T", "-", 1)

	return strings.Replace(input, timestampPlaceholder, actualTimestamp, 1)
}

func applyPublicKey(input string, publickey string) string {
	return strings.Replace(input, publicKeyPlaceholder, publickey, 1)
}
