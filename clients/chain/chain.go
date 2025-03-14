package chain

import (
	"fmt"
	"strings"
)

const (
	evmCompatibleChainToKleverBlockchainNameTemplate = "%sToKleverBlockchain"
	kleverBlockchainToEvmCompatibleChainNameTemplate = "KleverBlockchainTo%s"
	baseLogIdTemplate                                = "%sKleverBlockchain-Base"
	kleverBlockchainClientLogIdTemplate              = "%sKleverBlockchain-KleverBlockchainClient"
	kleverBlockchainDataGetterLogIdTemplate          = "%sKleverBlockchain-KleverBlockchainDataGetter"
	evmCompatibleChainClientLogIdTemplate            = "%sKleverBlockchain-%sClient"
	kleverBlockchainRoleProviderLogIdTemplate        = "%sKleverBlockchain-KleverBlockchainRoleProvider"
	evmCompatibleChainRoleProviderLogIdTemplate      = "%sKleverBlockchain-%sRoleProvider"
	broadcasterLogIdTemplate                         = "%sKleverBlockchain-Broadcaster"
)

// Chain defines all the chain supported
type Chain string

const (
	// KleverBlockchain is the string representation of the KleverBlockchain chain
	KleverBlockchain Chain = "klv"

	// Ethereum is the string representation of the Ethereum chain
	Ethereum Chain = "Ethereum"

	// Bsc is the string representation of the Binance smart chain
	Bsc Chain = "Bsc"

	// Polygon is the string representation of the Polygon chain
	Polygon Chain = "Polygon"
)

// ToLower returns the lowercase string of chain
func (c Chain) ToLower() string {
	return strings.ToLower(string(c))
}

// EvmCompatibleChainToKleverBlockchainName returns the string using chain value and evmCompatibleChainToKleverBlockchainNameTemplate
func (c Chain) EvmCompatibleChainToKleverBlockchainName() string {
	return fmt.Sprintf(evmCompatibleChainToKleverBlockchainNameTemplate, c)
}

// KleverBlockchainToEvmCompatibleChainName returns the string using chain value and kleverBlockchainToEvmCompatibleChainNameTemplate
func (c Chain) KleverBlockchainToEvmCompatibleChainName() string {
	return fmt.Sprintf(kleverBlockchainToEvmCompatibleChainNameTemplate, c)
}

// BaseLogId returns the string using chain value and baseLogIdTemplate
func (c Chain) BaseLogId() string {
	return fmt.Sprintf(baseLogIdTemplate, c)
}

// KleverBlockchainClientLogId returns the string using chain value and kleverBlockchainClientLogIdTemplate
func (c Chain) KleverBlockchainClientLogId() string {
	return fmt.Sprintf(kleverBlockchainClientLogIdTemplate, c)
}

// KleverBlockchainDataGetterLogId returns the string using chain value and kleverBlockchainDataGetterLogIdTemplate
func (c Chain) KleverBlockchainDataGetterLogId() string {
	return fmt.Sprintf(kleverBlockchainDataGetterLogIdTemplate, c)
}

// EvmCompatibleChainClientLogId returns the string using chain value and evmCompatibleChainClientLogIdTemplate
func (c Chain) EvmCompatibleChainClientLogId() string {
	return fmt.Sprintf(evmCompatibleChainClientLogIdTemplate, c, c)
}

// KleverBlockchainRoleProviderLogId returns the string using chain value and kleverBlockchainRoleProviderLogIdTemplate
func (c Chain) KleverBlockchainRoleProviderLogId() string {
	return fmt.Sprintf(kleverBlockchainRoleProviderLogIdTemplate, c)
}

// EvmCompatibleChainRoleProviderLogId returns the string using chain value and evmCompatibleChainRoleProviderLogIdTemplate
func (c Chain) EvmCompatibleChainRoleProviderLogId() string {
	return fmt.Sprintf(evmCompatibleChainRoleProviderLogIdTemplate, c, c)
}

// BroadcasterLogId returns the string using chain value and broadcasterLogIdTemplate
func (c Chain) BroadcasterLogId() string {
	return fmt.Sprintf(broadcasterLogIdTemplate, c)
}
