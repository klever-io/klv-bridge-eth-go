package chain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ethToKleverBlockchainName(t *testing.T) {
	assert.Equal(t, "EthereumToKleverBlockchain", Ethereum.EvmCompatibleChainToKleverBlockchainName())
	assert.Equal(t, "BscToKleverBlockchain", Bsc.EvmCompatibleChainToKleverBlockchainName())
}

func Test_kleverBlockchainToEthName(t *testing.T) {
	assert.Equal(t, "KleverBlockchainToEthereum", Ethereum.KleverBlockchainToEvmCompatibleChainName())
	assert.Equal(t, "KleverBlockchainToBsc", Bsc.KleverBlockchainToEvmCompatibleChainName())
}

func Test_baseLogId(t *testing.T) {
	assert.Equal(t, "EthereumKleverBlockchain-Base", Ethereum.BaseLogId())
	assert.Equal(t, "BscKleverBlockchain-Base", Bsc.BaseLogId())
}

func Test_kleverBlockchainClientLogId(t *testing.T) {
	assert.Equal(t, "EthereumKleverBlockchain-KleverBlockchainClient", Ethereum.KleverBlockchainClientLogId())
	assert.Equal(t, "BscKleverBlockchain-KleverBlockchainClient", Bsc.KleverBlockchainClientLogId())
}

func Test_kleverBlockchainDataGetterLogId(t *testing.T) {
	assert.Equal(t, "EthereumKleverBlockchain-KleverBlockchainDataGetter", Ethereum.KleverBlockchainDataGetterLogId())
	assert.Equal(t, "BscKleverBlockchain-KleverBlockchainDataGetter", Bsc.KleverBlockchainDataGetterLogId())
}

func Test_ethClientLogId(t *testing.T) {
	assert.Equal(t, "EthereumKleverBlockchain-EthereumClient", Ethereum.EvmCompatibleChainClientLogId())
	assert.Equal(t, "BscKleverBlockchain-BscClient", Bsc.EvmCompatibleChainClientLogId())
}

func Test_kleverBlockchainRoleProviderLogId(t *testing.T) {
	assert.Equal(t, "EthereumKleverBlockchain-KleverBlockchainRoleProvider", Ethereum.KleverBlockchainRoleProviderLogId())
	assert.Equal(t, "BscKleverBlockchain-KleverBlockchainRoleProvider", Bsc.KleverBlockchainRoleProviderLogId())
}

func Test_ethRoleProviderLogId(t *testing.T) {
	assert.Equal(t, "EthereumKleverBlockchain-EthereumRoleProvider", Ethereum.EvmCompatibleChainRoleProviderLogId())
	assert.Equal(t, "BscKleverBlockchain-BscRoleProvider", Bsc.EvmCompatibleChainRoleProviderLogId())
}

func Test_broadcasterLogId(t *testing.T) {
	assert.Equal(t, "EthereumKleverBlockchain-Broadcaster", Ethereum.BroadcasterLogId())
	assert.Equal(t, "BscKleverBlockchain-Broadcaster", Bsc.BroadcasterLogId())
}

func TestToLower(t *testing.T) {
	assert.Equal(t, "klv", KleverBlockchain.ToLower())
	assert.Equal(t, "ethereum", Ethereum.ToLower())
	assert.Equal(t, "bsc", Bsc.ToLower())
}
