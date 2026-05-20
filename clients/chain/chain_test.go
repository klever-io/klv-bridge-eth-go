package chain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ethToKleverBlockchainName(t *testing.T) {
	assert.Equal(t, "EthereumToKleverBlockchain", Ethereum.EvmCompatibleChainToKleverBlockchainName())
	assert.Equal(t, "BscToKleverBlockchain", Bsc.EvmCompatibleChainToKleverBlockchainName())
	assert.Equal(t, "TronToKleverBlockchain", Tron.EvmCompatibleChainToKleverBlockchainName())
}

func Test_kleverBlockchainToEthName(t *testing.T) {
	assert.Equal(t, "KleverBlockchainToEthereum", Ethereum.KleverBlockchainToEvmCompatibleChainName())
	assert.Equal(t, "KleverBlockchainToBsc", Bsc.KleverBlockchainToEvmCompatibleChainName())
	assert.Equal(t, "KleverBlockchainToTron", Tron.KleverBlockchainToEvmCompatibleChainName())
}

func Test_baseLogId(t *testing.T) {
	assert.Equal(t, "EthereumKleverBlockchain-Base", Ethereum.BaseLogId())
	assert.Equal(t, "BscKleverBlockchain-Base", Bsc.BaseLogId())
	assert.Equal(t, "TronKleverBlockchain-Base", Tron.BaseLogId())
}

func Test_kleverBlockchainClientLogId(t *testing.T) {
	assert.Equal(t, "EthereumKleverBlockchain-KleverBlockchainClient", Ethereum.KleverBlockchainClientLogId())
	assert.Equal(t, "BscKleverBlockchain-KleverBlockchainClient", Bsc.KleverBlockchainClientLogId())
	assert.Equal(t, "TronKleverBlockchain-KleverBlockchainClient", Tron.KleverBlockchainClientLogId())
}

func Test_kleverBlockchainDataGetterLogId(t *testing.T) {
	assert.Equal(t, "EthereumKleverBlockchain-KleverBlockchainDataGetter", Ethereum.KleverBlockchainDataGetterLogId())
	assert.Equal(t, "BscKleverBlockchain-KleverBlockchainDataGetter", Bsc.KleverBlockchainDataGetterLogId())
	assert.Equal(t, "TronKleverBlockchain-KleverBlockchainDataGetter", Tron.KleverBlockchainDataGetterLogId())
}

func Test_ethClientLogId(t *testing.T) {
	assert.Equal(t, "EthereumKleverBlockchain-EthereumClient", Ethereum.EvmCompatibleChainClientLogId())
	assert.Equal(t, "BscKleverBlockchain-BscClient", Bsc.EvmCompatibleChainClientLogId())
	assert.Equal(t, "TronKleverBlockchain-TronClient", Tron.EvmCompatibleChainClientLogId())
}

func Test_kleverBlockchainRoleProviderLogId(t *testing.T) {
	assert.Equal(t, "EthereumKleverBlockchain-KleverBlockchainRoleProvider", Ethereum.KleverBlockchainRoleProviderLogId())
	assert.Equal(t, "BscKleverBlockchain-KleverBlockchainRoleProvider", Bsc.KleverBlockchainRoleProviderLogId())
	assert.Equal(t, "TronKleverBlockchain-KleverBlockchainRoleProvider", Tron.KleverBlockchainRoleProviderLogId())
}

func Test_ethRoleProviderLogId(t *testing.T) {
	assert.Equal(t, "EthereumKleverBlockchain-EthereumRoleProvider", Ethereum.EvmCompatibleChainRoleProviderLogId())
	assert.Equal(t, "BscKleverBlockchain-BscRoleProvider", Bsc.EvmCompatibleChainRoleProviderLogId())
	assert.Equal(t, "TronKleverBlockchain-TronRoleProvider", Tron.EvmCompatibleChainRoleProviderLogId())
}

func Test_broadcasterLogId(t *testing.T) {
	assert.Equal(t, "EthereumKleverBlockchain-Broadcaster", Ethereum.BroadcasterLogId())
	assert.Equal(t, "BscKleverBlockchain-Broadcaster", Bsc.BroadcasterLogId())
	assert.Equal(t, "TronKleverBlockchain-Broadcaster", Tron.BroadcasterLogId())
}

func TestToLower(t *testing.T) {
	assert.Equal(t, "klv", KleverBlockchain.ToLower())
	assert.Equal(t, "ethereum", Ethereum.ToLower())
	assert.Equal(t, "bsc", Bsc.ToLower())
	assert.Equal(t, "tron", Tron.ToLower())
}
