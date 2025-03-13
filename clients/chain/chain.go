package chain

import (
	"fmt"
	"strings"
)

const (
	evmCompatibleChainToKleverBlockchainNameTemplate = "%sToKleverBlockchain"
	multiversXToEvmCompatibleChainNameTemplate       = "KleverBlockchainTo%s"
	baseLogIdTemplate                                = "%sKleverBlockchain-Base"
	multiversXClientLogIdTemplate                    = "%sKleverBlockchain-KleverBlockchainClient"
	multiversXDataGetterLogIdTemplate                = "%sKleverBlockchain-KleverBlockchainDataGetter"
	evmCompatibleChainClientLogIdTemplate            = "%sKleverBlockchain-%sClient"
	multiversXRoleProviderLogIdTemplate              = "%sKleverBlockchain-KleverBlockchainRoleProvider"
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

// KleverBlockchainToEvmCompatibleChainName returns the string using chain value and multiversXToEvmCompatibleChainNameTemplate
func (c Chain) KleverBlockchainToEvmCompatibleChainName() string {
	return fmt.Sprintf(multiversXToEvmCompatibleChainNameTemplate, c)
}

// BaseLogId returns the string using chain value and baseLogIdTemplate
func (c Chain) BaseLogId() string {
	return fmt.Sprintf(baseLogIdTemplate, c)
}

// KleverBlockchainClientLogId returns the string using chain value and multiversXClientLogIdTemplate
func (c Chain) KleverBlockchainClientLogId() string {
	return fmt.Sprintf(multiversXClientLogIdTemplate, c)
}

// KleverBlockchainDataGetterLogId returns the string using chain value and multiversXDataGetterLogIdTemplate
func (c Chain) KleverBlockchainDataGetterLogId() string {
	return fmt.Sprintf(multiversXDataGetterLogIdTemplate, c)
}

// EvmCompatibleChainClientLogId returns the string using chain value and evmCompatibleChainClientLogIdTemplate
func (c Chain) EvmCompatibleChainClientLogId() string {
	return fmt.Sprintf(evmCompatibleChainClientLogIdTemplate, c, c)
}

// KleverBlockchainRoleProviderLogId returns the string using chain value and multiversXRoleProviderLogIdTemplate
func (c Chain) KleverBlockchainRoleProviderLogId() string {
	return fmt.Sprintf(multiversXRoleProviderLogIdTemplate, c)
}

// EvmCompatibleChainRoleProviderLogId returns the string using chain value and evmCompatibleChainRoleProviderLogIdTemplate
func (c Chain) EvmCompatibleChainRoleProviderLogId() string {
	return fmt.Sprintf(evmCompatibleChainRoleProviderLogIdTemplate, c, c)
}

// BroadcasterLogId returns the string using chain value and broadcasterLogIdTemplate
func (c Chain) BroadcasterLogId() string {
	return fmt.Sprintf(broadcasterLogIdTemplate, c)
}
